package workflow_test

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
	"github.com/bem-team/terraform-provider-bem/internal/services/workflow"
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

// testAccCreateFunctionOutOfBand creates a bare function directly via the
// SDK for the workflow's node to reference - the workflow resource under
// test doesn't manage this function itself, matching how a real customer's
// workflow can reference a function that isn't (or isn't yet) under
// Terraform management at all.
func testAccCreateFunctionOutOfBand(t *testing.T, name string) {
	t.Helper()
	client := testAccBemClient()
	ctx := context.Background()

	res := new(http.Response)
	_, err := client.Functions.New(
		ctx,
		bem.FunctionNewParams{},
		option.WithRequestBody("application/json", []byte(fmt.Sprintf(
			`{"functionName":%q,"displayName":"TF Acc Test Function (workflow dep)","type":"split","splitType":"semantic_page","semanticPageSplitConfig":{"itemClasses":[{"name":"class_one"}]}}`,
			name,
		))),
		option.WithResponseBodyInto(&res),
	)
	if err != nil {
		t.Fatalf("failed to create out-of-band function %q: %v", name, err)
	}

	t.Cleanup(func() {
		_ = client.Functions.Delete(context.Background(), name)
	})
}

// testAccCreateWorkflowOutOfBand creates a function and a workflow directly
// via the SDK client, bypassing Terraform entirely - matching "someone built
// this manually in the UI" with no Terraform state anywhere for it yet.
// Mirrors WorkflowResource.Create's own request-building
// (WorkflowModel.MarshalJSON) rather than hand-crafting raw JSON.
func testAccCreateWorkflowOutOfBand(t *testing.T, workflowName, functionName string) {
	t.Helper()
	client := testAccBemClient()
	ctx := context.Background()

	data := &workflow.WorkflowModel{
		Name:         types.StringValue(workflowName),
		DisplayName:  types.StringValue("TF Acc Test Workflow"),
		MainNodeName: types.StringValue("main"),
		Tags:         &[]types.String{types.StringValue("bem-acc-test")},
		Nodes: &[]*workflow.WorkflowNodesModel{
			{
				Name: types.StringValue("main"),
				Function: &workflow.WorkflowNodesFunctionModel{
					Name: types.StringValue(functionName),
				},
			},
		},
	}

	dataBytes, err := data.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to build out-of-band workflow payload: %v", err)
	}

	res := new(http.Response)
	_, err = client.Workflows.New(
		ctx,
		bem.WorkflowNewParams{},
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
	)
	if err != nil {
		t.Fatalf("failed to create out-of-band workflow %q: %v", workflowName, err)
	}

	t.Cleanup(func() {
		// Best-effort: if Terraform's own destroy already removed this (the
		// normal path once the test successfully imports it), this 404s and
		// is ignored.
		_ = client.Workflows.Delete(context.Background(), workflowName)
	})
}

func testAccWorkflowConfig(workflowName, functionName string, tags ...string) string {
	quoted := ""
	for i, tag := range tags {
		if i > 0 {
			quoted += ", "
		}
		quoted += fmt.Sprintf("%q", tag)
	}
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_workflow" "test" {
  name           = %[1]q
  display_name   = "TF Acc Test Workflow"
  main_node_name = "main"
  tags           = [%[3]s]
  edges          = []

  nodes = [
    {
      name = "main"
      function = {
        name = %[2]q
      }
    }
  ]
}
`, workflowName, functionName, quoted)
}

func testAccCheckWorkflowDestroy(s *terraform.State) error {
	client := testAccBemClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bem_workflow" {
			continue
		}

		res := new(http.Response)
		_, err := client.Workflows.Get(
			ctx,
			rs.Primary.Attributes["name"],
			option.WithResponseBodyInto(&res),
		)
		if res != nil && res.StatusCode == http.StatusNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error checking bem_workflow %q was destroyed: %w", rs.Primary.Attributes["name"], err)
		}
		return fmt.Errorf("bem_workflow %q still exists", rs.Primary.Attributes["name"])
	}

	return nil
}

// testAccCheckWorkflowNodeFunctionIntact queries the live API directly
// (not Terraform state) and confirms node 0's function reference is a real,
// non-empty value - the exact field the encoder race (BEM-1367, Finding 3)
// silently corrupted to `{}` on update, with no error, no matter what
// Terraform's own state or "apply succeeded" claimed.
func testAccCheckWorkflowNodeFunctionIntact(workflowName, expectedFunctionName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccBemClient()
		ctx := context.Background()

		res, err := client.Workflows.Get(ctx, workflowName)
		if err != nil {
			return fmt.Errorf("failed to fetch workflow %q from API: %w", workflowName, err)
		}

		if len(res.Workflow.Nodes) == 0 {
			return fmt.Errorf("workflow %q has no nodes - this is exactly the corruption BEM-1367's encoder race caused (nodes silently wiped)", workflowName)
		}

		fn := res.Workflow.Nodes[0].Function
		if fn.ID == "" && fn.Name == "" {
			return fmt.Errorf("workflow %q node 0's function reference is empty ({}) - this is exactly the silent corruption bug BEM-1367 fixed in internal/apijson/encoder.go", workflowName)
		}
		if fn.Name != "" && fn.Name != expectedFunctionName {
			return fmt.Errorf("workflow %q node 0's function name = %q, want %q", workflowName, fn.Name, expectedFunctionName)
		}

		return nil
	}
}

// TestAccWorkflowResource_ImportThenUpdate exercises the encoder-race bug
// fixed by BEM-1367 end to end against a real backend, for bem_workflow
// specifically - the resource type the original corruption hit hardest
// (nodes[0].function silently sent as `{}`, wiping the reference
// server-side with no error at all):
//
//  1. Import a manually-created workflow (created out of band, above, with
//     no Terraform state for it yet), referencing a function that also
//     exists out of band. ImportStatePersist carries the imported state
//     forward as the test case's real state, exactly like the function
//     resource's equivalent test.
//
//     ImportState still leaves every no_refresh-tagged writable attribute
//     (main_node_name, nodes, tags, ...) null immediately after import -
//     that's the known, separate, NOT-fixed-by-BEM-1367 ImportState gap
//     (Finding 1's "defect B" / the doc's "Importing an existing resource
//     leaves most attributes null" section). This test does not use the
//     `ignore_changes = [nodes, edges]` lifecycle workaround the
//     bem-workflows Atmos component uses, so it isn't exercising the
//     separate ignore_changes interaction (Finding 4) - just the core
//     encoder correctness this ticket actually fixes.
//
//  2. Re-apply the same config against that under-hydrated, persisted,
//     imported state. Before the fix, this is exactly the shape that
//     triggered the race: a scalar/list field (tags) transitioning from
//     null to populated in the same MarshalJSONForUpdate call as the
//     `nodes` field. The live-API check below confirms node 0's function
//     reference is still a real, correct value after this apply - not just
//     that "apply succeeded" or that Terraform's own state looks right.
//
//  3. Add a second tag (any small scalar/list change) and re-apply again,
//     re-checking the same live-API invariant. This exercises the update
//     path a second time on an already-converged resource, the more common
//     real-world case ("someone tweaks a workflow's tags") than the
//     immediately-post-import case in step 2.
func TestAccWorkflowResource_ImportThenUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	workflowName := acctest.RandomWithPrefix("tf-acc-bem-workflow")
	functionName := acctest.RandomWithPrefix("tf-acc-bem-workflow-fn")

	testAccCreateFunctionOutOfBand(t, functionName)
	testAccCreateWorkflowOutOfBand(t, workflowName, functionName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				// No prior TestStep exists to infer the import ID from
				// (this is the first step), so it must be set explicitly.
				ResourceName:       "bem_workflow.test",
				Config:             testAccWorkflowConfig(workflowName, functionName, "bem-acc-test"),
				ImportState:        true,
				ImportStateId:      workflowName,
				ImportStatePersist: true,
			},
			{
				// Re-apply the same config against the under-hydrated
				// imported state - see point 2 above. Expect a real diff
				// here (not an empty plan) - that's the known, separate
				// ImportState gap, not this test failing.
				Config: testAccWorkflowConfig(workflowName, functionName, "bem-acc-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_workflow.test", "name", workflowName),
					resource.TestCheckResourceAttr("bem_workflow.test", "display_name", "TF Acc Test Workflow"),
					resource.TestCheckResourceAttr("bem_workflow.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("bem_workflow.test", "nodes.#", "1"),
					testAccCheckWorkflowNodeFunctionIntact(workflowName, functionName),
				),
			},
			{
				// Second update on an already-converged resource - the more
				// common real-world case than immediately-post-import.
				Config: testAccWorkflowConfig(workflowName, functionName, "bem-acc-test", "bem-acc-test-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_workflow.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("bem_workflow.test", "nodes.#", "1"),
					testAccCheckWorkflowNodeFunctionIntact(workflowName, functionName),
				),
			},
		},
	})
}
