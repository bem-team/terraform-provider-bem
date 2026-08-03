package function_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/bem-team/bem-go-sdk"
	"github.com/bem-team/bem-go-sdk/option"
	"github.com/bem-team/terraform-provider-bem/internal"
	"github.com/bem-team/terraform-provider-bem/internal/services/function"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bem": providerserver.NewProtocol6WithError(internal.NewProvider("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("BEM_API_KEY") == "" {
		t.Fatal("BEM_API_KEY must be set for acceptance tests (point it at a staging environment, not production)")
	}
}

// testAccBemClient builds a client the same way BemProvider.Configure does,
// so it talks to the same environment the test itself runs against.
func testAccBemClient() *bem.Client {
	opts := []option.RequestOption{}
	if v := os.Getenv("BEM_BASE_URL"); v != "" {
		opts = append(opts, option.WithBaseURL(v))
	}
	if v := os.Getenv("BEM_API_KEY"); v != "" {
		opts = append(opts, option.WithAPIKey(v))
	}
	client := bem.NewClient(opts...)
	return &client
}

// testAccCreateFunctionOutOfBand creates a function directly via the SDK
// client, bypassing Terraform entirely - matching "someone built this
// manually in the UI" with no Terraform state anywhere for it yet. Mirrors
// FunctionResource.Create's own request-building (FunctionModel.MarshalJSON)
// rather than hand-crafting raw JSON.
//
// Registers a best-effort cleanup so the function isn't orphaned on the
// backend if the test fails before Terraform's own state ever adopts (and
// later destroys) it.
func testAccCreateFunctionOutOfBand(t *testing.T, name string) {
	t.Helper()
	client := testAccBemClient()
	ctx := context.Background()

	data := &function.FunctionModel{
		FunctionName: types.StringValue(name),
		DisplayName:  types.StringValue("TF Acc Test Function"),
		Type:         types.StringValue("split"),
		SplitType:    types.StringValue("semantic_page"),
		Tags:         &[]types.String{types.StringValue("bem-acc-test")},
		SemanticPageSplitConfig: &function.FunctionSemanticPageSplitConfigModel{
			ItemClasses: &[]*function.FunctionSemanticPageSplitConfigItemClassesModel{
				{Name: types.StringValue("class_one"), Description: types.StringValue("First class")},
				{Name: types.StringValue("class_two"), Description: types.StringValue("Second class")},
			},
		},
	}

	dataBytes, err := data.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to build out-of-band function payload: %v", err)
	}

	res := new(http.Response)
	_, err = client.Functions.New(
		ctx,
		bem.FunctionNewParams{},
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
	)
	if err != nil {
		t.Fatalf("failed to create out-of-band function %q: %v", name, err)
	}

	t.Cleanup(func() {
		// Best-effort: if Terraform's own destroy already removed this (the
		// normal path once the test successfully imports it), this 404s and
		// is ignored.
		_ = client.Functions.Delete(context.Background(), name)
	})
}

func testAccFunctionConfig(name string) string {
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_function" "test" {
  function_name = %[1]q
  display_name  = "TF Acc Test Function"
  type          = "split"
  split_type    = "semantic_page"
  tags          = ["bem-acc-test"]
  semantic_page_split_config = {
    item_classes = [
      { name = "class_one", description = "First class" },
      { name = "class_two", description = "Second class" },
    ]
  }
}
`, name)
}

func testAccCheckFunctionDestroy(s *terraform.State) error {
	client := testAccBemClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bem_function" {
			continue
		}

		res := new(http.Response)
		_, err := client.Functions.Get(
			ctx,
			rs.Primary.Attributes["function_name"],
			option.WithResponseBodyInto(&res),
		)
		if res != nil && res.StatusCode == http.StatusNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error checking bem_function %q was destroyed: %w", rs.Primary.Attributes["function_name"], err)
		}
		return fmt.Errorf("bem_function %q still exists", rs.Primary.Attributes["function_name"])
	}

	return nil
}

// TestAccFunctionResource_ImportThenUpdate exercises the actual bug fixed by
// BEM-1367 end to end against a real backend:
//
//  1. Import a manually-created function (created out of band, above, with
//     no Terraform state for it yet). This is literally the first TestStep -
//     ImportStatePersist carries the imported state forward as the test
//     case's real state (rather than the isolated, throwaway copy used when
//     ImportStatePersist is unset), and since nothing has been applied via
//     Terraform yet, there's no pre-existing state for the import to
//     conflict with.
//
//     ImportState still leaves every no_refresh-tagged writable attribute
//     (tags, type, semantic_page_split_config, ...) null immediately after
//     import - that's a known, separate, NOT-fixed-by-BEM-1367 gap (their
//     real values only exist in the response's nested, computed `function`
//     union, which has no flat top-level representation for the decoder to
//     hydrate these from; an earlier attempt to "fix" this by decoding
//     no_refresh fields anyway made things worse - see git history - since
//     it nulled out FunctionName too, breaking the immediately-following
//     ReadResource refresh). This is intentional and expected here, not a
//     test bug.
//
//  2. Re-apply the same config against that under-hydrated, persisted,
//     imported state. This is the core BEM-1367 regression: before the fix,
//     this apply failed outright with "missing required path_function_name
//     parameter", because Update() keyed off a field that's never populated
//     unless the user explicitly sets it. With the fix, Update() correctly
//     uses state.FunctionName (always populated) for the API path, so this
//     apply now succeeds - and since Update() decodes the response via
//     UnmarshalComputed (which preserves the plan's already-set flat fields
//     rather than trying to re-derive them), the Check functions below
//     confirm state fully converges to match config after this one apply,
//     exactly like the original manual verification during BEM-1367 found.
//
//  3. Rename the function (change function_name only) and re-apply. This
//     exercises resource.go's Update(), which must use the OLD name
//     (state.FunctionName) for the API path while the request body carries
//     the NEW name (from plan/data) - the exact fix BEM-1367 made. Before
//     that fix nothing regression-tested this path.
func TestAccFunctionResource_ImportThenUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	functionName := acctest.RandomWithPrefix("tf-acc-bem-function")
	renamedFunctionName := functionName + "-renamed"

	testAccCreateFunctionOutOfBand(t, functionName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				// No prior TestStep exists to infer the import ID from
				// (this is the first step), so it must be set explicitly.
				ResourceName:       "bem_function.test",
				Config:             testAccFunctionConfig(functionName),
				ImportState:        true,
				ImportStateId:      functionName,
				ImportStatePersist: true,
			},
			{
				// Re-apply the same config against the under-hydrated
				// imported state - see point 2 above. Expect a real diff
				// here (not an empty plan) - that's the known, separate
				// Finding B gap, not this test failing.
				Config: testAccFunctionConfig(functionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.test", "function_name", functionName),
					resource.TestCheckResourceAttr("bem_function.test", "display_name", "TF Acc Test Function"),
					resource.TestCheckResourceAttr("bem_function.test", "type", "split"),
					resource.TestCheckResourceAttr("bem_function.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("bem_function.test", "tags.0", "bem-acc-test"),
					resource.TestCheckResourceAttr("bem_function.test", "semantic_page_split_config.item_classes.#", "2"),
				),
			},
			{
				// Rename: exercises Update()'s state.FunctionName (old name,
				// for the URL path) vs data.FunctionName (new name, for the
				// request body) split.
				Config: testAccFunctionConfig(renamedFunctionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.test", "function_name", renamedFunctionName),
					resource.TestCheckResourceAttr("bem_function.test", "id", renamedFunctionName),
				),
			},
		},
	})
}
