package function_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/bem-team/bem-go-sdk"
	"github.com/bem-team/bem-go-sdk/option"
	"github.com/bem-team/terraform-provider-bem/internal"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/bem-team/terraform-provider-bem/internal/services/function"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// prodAPIHost is the SDK's default host. Reaching it from an acceptance test
// means BEM_BASE_URL was never set, not that someone chose production.
const prodAPIHost = "api.bem.ai"

func testAccPreCheck(t *testing.T) {
	if os.Getenv("BEM_API_KEY") == "" {
		t.Fatal("BEM_API_KEY must be set for acceptance tests (point it at a staging environment, not production)")
	}

	// These tests create and destroy real functions and workflows. With
	// BEM_BASE_URL unset the SDK falls back to production, so "forgot one
	// export" plus a production key is all it takes to run the whole suite
	// against live customer data - the failure mode is silent unless the key
	// happens to be environment-scoped. Require the host to be chosen
	// explicitly and refuse production outright; a comment asking people to
	// point at staging isn't a control.
	baseURL := os.Getenv("BEM_BASE_URL")
	if baseURL == "" {
		t.Fatalf("BEM_BASE_URL must be set explicitly for acceptance tests - leaving it unset defaults to production (https://%s). Use a staging host, e.g. https://api.stg.us1.bem.ai", prodAPIHost)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("BEM_BASE_URL %q is not a valid URL: %v", baseURL, err)
	}
	if strings.EqualFold(parsed.Hostname(), prodAPIHost) {
		t.Fatalf("refusing to run acceptance tests against production (%s) - these create and destroy real functions and workflows. Use a staging host, e.g. https://api.stg.us1.bem.ai", prodAPIHost)
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
		Tags:         customfield.NewListMust[types.String](ctx, []attr.Value{types.StringValue("bem-acc-test")}),
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
			bem.FunctionGetParams{},
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
//     ImportState used to leave every no_refresh-tagged writable attribute
//     (tags, type, semantic_page_split_config, ...) null immediately after
//     import. BEM-1396 fixed that - UnmarshalForImport lifts the skip, and
//     the values are read out of the response's {"function": {...}} wrapper
//     via the envelope. ImportStateCheck now asserts it, which is what this
//     step was missing: it previously asserted nothing at all about the
//     state the import produced.
//
//     Note the earlier attempt recorded here: decoding no_refresh fields
//     without going through the envelope nulls function_name too, which
//     breaks the immediately-following refresh. That failure mode is real,
//     and it is why the fix is envelope-plus-mirror rather than a
//     root-level decode.
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
//
// checkImportHydratesWritableAttributes asserts that import populates the
// no_refresh-tagged writable attributes.
//
// This is the check the suite was missing. ImportState used to leave every one of
// them null - the values only exist inside the response's {"function": {...}}
// wrapper, and the no_refresh skip that keeps Read from re-writing state has
// nothing to protect on an import. The consequences are all downstream of state:
// the first plan after `terraform import` proposes every writable field changing
// from null, and `plan -generate-config-out` cannot emit config for a Required
// no_refresh attribute at all. Nothing here failed, because the import step
// asserted nothing about the state it produced.
//
// It stayed hidden through two attempts at the fix. The first nulled
// function_name as well and was caught only because the *following* step's
// refresh then failed; the second hydrated the wrapper correctly and then had a
// full second decode pass wipe it again at the root, which no test saw and only a
// live import surfaced.
func checkImportHydratesWritableAttributes(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("expected 1 state after import, got %d", len(states))
	}
	attrs := states[0].Attributes

	// function_name is checked separately: it varies per run.
	want := map[string]string{
		"type":         "split",
		"display_name": "TF Acc Test Function",
		"tags.#":       "1",
		"tags.0":       "bem-acc-test",
		"semantic_page_split_config.item_classes.#": "2",
	}

	for key, expected := range want {
		got, ok := attrs[key]
		if !ok || got == "" {
			return fmt.Errorf("after import, %s is absent from state. Every writable attribute "+
				"is no_refresh, so a null here is a field the first post-import plan proposes "+
				"changing from null - see hydrateFromResponse", key)
		}
		if got != expected {
			return fmt.Errorf("after import, %s = %q, want %q", key, got, expected)
		}
	}

	if attrs["function_name"] == "" {
		return fmt.Errorf("after import, function_name is empty; it is Required, so a null " +
			"also breaks plan -generate-config-out entirely")
	}
	// The read-only mirror is the half that survived when the rest did not, so
	// assert it too rather than assuming the attributes above cover the decode.
	if attrs["function.version_num"] == "" {
		return fmt.Errorf("after import, the `function` mirror is not populated; " +
			"configurations read function.function_id / function.version_num from it")
	}
	return nil
}

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
				ImportStateCheck:   checkImportHydratesWritableAttributes,
			},
			{
				// The check a practitioner actually experiences, and the one
				// this suite was missing: after importing a resource that
				// already matches the configuration, is the next plan empty?
				//
				// ImportStateCheck above proves the attributes hydrated. It says
				// nothing about whether the resulting state agrees with the
				// config, which is a separate question and the one that bites:
				// import writes server values that Read never wrote, so any
				// attribute where the server's value differs from the config's -
				// including a server default of "" against an omitted attribute -
				// leaves the plan dirty forever, cutting a new function version
				// on every apply.
				Config:   testAccFunctionConfig(functionName),
				PlanOnly: true,
			},
			{
				// Re-apply the same config against the imported state.
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

// testAccCreateMinimalExtractOutOfBand creates an extract function with the
// smallest body the API accepts, so that every attribute the server defaults is
// server-assigned rather than echoed back from the request.
func testAccCreateMinimalExtractOutOfBand(t *testing.T, name string) {
	t.Helper()
	client := testAccBemClient()
	ctx := context.Background()

	body := fmt.Sprintf(`{
	  "functionName": %q,
	  "type": "extract",
	  "outputSchemaName": "TFAccMinimalSchema",
	  "outputSchema": {"type": "object", "properties": {"a": {"type": "string"}}}
	}`, name)

	res := new(http.Response)
	_, err := client.Functions.New(
		ctx,
		bem.FunctionNewParams{},
		option.WithRequestBody("application/json", []byte(body)),
		option.WithResponseBodyInto(&res),
	)
	if err != nil {
		t.Fatalf("failed to create minimal out-of-band function %q: %v", name, err)
	}

	t.Cleanup(func() {
		_ = client.Functions.Delete(context.Background(), name)
	})
}

func testAccMinimalExtractConfig(name string) string {
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_function" "minimal" {
  function_name      = %[1]q
  type               = "extract"
  output_schema_name = "TFAccMinimalSchema"
  output_schema      = jsonencode({
    type       = "object"
    properties = { a = { type = "string" } }
  })
}
`, name)
}

// TestAccFunctionResource_ImportMinimalPlansClean is the guard for the class of
// defect BEM-1396 found on import: attributes the API assigns a default for.
//
// Before `UnmarshalForImport`, no import ever put a server value into state for a
// no_refresh attribute, so this could not arise. Now it does, and any attribute
// where the server supplies a value the configuration omits leaves the plan dirty
// **forever** - each apply cutting a new function version, and the churn
// cascading into every workflow that interpolates the function's id or version.
//
// The configuration here is deliberately minimal: no description, no tags, and
// none of the extract booleans. A probe against the API confirmed the server
// assigns all of them when unset:
//
//	description              -> ""
//	tags                     -> []
//	tabular_chunking_enabled -> false
//	enable_bounding_boxes    -> false
//	pre_count                -> false
//	config                   -> {"steps": null}
//
// This cannot be repaired in ModifyPlan. Terraform core requires the planned
// value of a non-Computed attribute to equal the configuration value, so writing
// prior state over a plain-Optional attribute is rejected outright:
//
//	Error: Provider produced invalid plan
//	.description: planned value cty.StringVal("") for a non-computed attribute
//
// (verified, not assumed). The fix therefore has to be in the schema: an
// attribute the server assigns is Computed by definition, and Optional+Computed
// is what lets prior state be preserved when the configuration omits it.
func TestAccFunctionResource_ImportMinimalPlansClean(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	functionName := acctest.RandomWithPrefix("tf-acc-bem-minimal")
	testAccCreateMinimalExtractOutOfBand(t, functionName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				ResourceName:       "bem_function.minimal",
				Config:             testAccMinimalExtractConfig(functionName),
				ImportState:        true,
				ImportStateId:      functionName,
				ImportStatePersist: true,
			},
			{
				// Apply once, then require stability. Every non-PlanOnly step
				// already asserts its own post-apply plan is empty, so a
				// server-defaulted attribute that churns perpetually fails here.
				Config: testAccMinimalExtractConfig(functionName),
			},
			{
				Config:   testAccMinimalExtractConfig(functionName),
				PlanOnly: true,
			},
		},
	})
}

func testAccTagsConfig(name, tags string) string {
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_function" "tagged" {
  function_name      = %[1]q
  type               = "extract"
  output_schema_name = "TFAccTagsSchema"
  output_schema      = jsonencode({ type = "object", properties = { a = { type = "string" } } })
  tags               = %[2]s
}
`, name, tags)
}

// TestAccFunctionResource_EmptyTagsListClears guards the escape hatch for this
// release's one breaking change.
//
// `tags` is Optional+Computed, because the API returns [] for it and an imported
// function whose configuration omits it could otherwise never plan clean. The
// cost of that is standard Optional+Computed semantics: **removing** tags from a
// configuration no longer clears them, it preserves prior state. Verified
// against the API - the tags survive the removal.
//
// So `tags = []` is the documented way to clear them, and this asserts it works.
// If it ever stops working there is no way to clear tags at all, which would turn
// a documented behaviour change into a genuine regression.
func TestAccFunctionResource_EmptyTagsListClears(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	functionName := acctest.RandomWithPrefix("tf-acc-bem-tags")

	checkServerTags := func(want int) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			client := testAccBemClient()
			var raw map[string]any
			if _, err := client.Functions.Get(context.Background(), functionName,
				bem.FunctionGetParams{}, option.WithResponseBodyInto(&raw)); err != nil {
				return fmt.Errorf("reading %q back from the API: %w", functionName, err)
			}
			fn, _ := raw["function"].(map[string]any)
			tags, _ := fn["tags"].([]any)
			if len(tags) != want {
				return fmt.Errorf("the API holds %d tags (%v), want %d - state is not the "+
					"authority here, the server is", len(tags), tags, want)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagsConfig(functionName, `["acc-a", "acc-b"]`),
				Check:  checkServerTags(2),
			},
			{
				// An explicit empty list must still reach the API and clear them.
				Config: testAccTagsConfig(functionName, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.tagged", "tags.#", "0"),
					checkServerTags(0),
				),
			},
		},
	})
}
