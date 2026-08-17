package workflow_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bem-team/bem-go-sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccFunctionConfig declares a single split function, matching the
// shape of the real bem-customer Atmos component (components/terraform/infra/bem-customer/v0.0.1/main.tf)
// - the exact resource shape a customer hand-writes, not the internal
// for_each abstraction bem-workflows uses.
func testAccFunctionConfig(functionName, displayName string) string {
	return fmt.Sprintf(`
resource "bem_function" "splitter" {
  function_name = %[1]q
  type          = "split"
  display_name  = %[2]q
  split_type    = "semantic_page"
  tags          = ["bem-acc-test"]

  semantic_page_split_config = {
    item_classes = [
      { name = "invoice", description = "Invoice pages" }
    ]
  }
}
`, functionName, displayName)
}

// testAccWorkflowNodeBlock is the node block every workflow below shares:
// pinned to the shared function's current version_num, exactly like opto's
// real trigger (extract_schedule_of_investments / Simons-Example-for-Terraform)
// - not the weaker id-only reference bem-customer's main.tf originally had,
// which turned out not to reproduce BEM-1392 at all (Terraform resolves the
// unknown function_id during apply and finds it unchanged, so the workflow
// update never fires).
const testAccWorkflowNodeBlock = `
  nodes = [
    {
      name = "splitter"
      function = {
        id          = bem_function.splitter.function.function_id
        version_num = bem_function.splitter.function.version_num
      }
    }
  ]
`

// testAccFunctionAndWorkflowConfig: one function, one workflow referencing
// it with version_num pinned. The BEM-1392 regression shape.
func testAccFunctionAndWorkflowConfig(functionName, functionDisplayName, workflowName string) string {
	return fmt.Sprintf(`
provider "bem" {}

%s

resource "bem_workflow" "test" {
  name           = %[2]q
  display_name   = "TF Acc Test Workflow"
  main_node_name = "splitter"
  edges          = []
%[3]s
}
`, testAccFunctionConfig(functionName, functionDisplayName), workflowName, testAccWorkflowNodeBlock)
}

// testAccSharedFunctionTwoWorkflowsConfig: one function, two independent
// workflows both pinning the same function's version_num - the "can a
// function be part of multiple workflows, and does bumping it update all of
// them in one apply" scenario.
func testAccSharedFunctionTwoWorkflowsConfig(functionName, functionDisplayName, workflowNameA, workflowNameB string) string {
	return fmt.Sprintf(`
provider "bem" {}

%s

resource "bem_workflow" "a" {
  name           = %[2]q
  display_name   = "TF Acc Test Workflow A"
  main_node_name = "splitter"
  edges          = []
%[4]s
}

resource "bem_workflow" "b" {
  name           = %[3]q
  display_name   = "TF Acc Test Workflow B"
  main_node_name = "splitter"
  edges          = []
%[4]s
}
`, testAccFunctionConfig(functionName, functionDisplayName), workflowNameA, workflowNameB, testAccWorkflowNodeBlock)
}

// testAccCheckWorkflowNodeVersionMatchesFunction queries the live API
// directly (not Terraform state) and confirms the workflow's single node is
// pinned to the function's current version_num. This is exactly the
// invariant BEM-1392 broke: the encoder's patch-diffing correctly omitted
// mainNodeName/edges since they were individually unchanged, and the bem
// API rejected the resulting partial DAG PATCH outright - so pre-fix, this
// check's own precondition (an apply that actually succeeded) never even
// gets reached on the second step of either test below.
func testAccCheckWorkflowNodeVersionMatchesFunction(workflowName, functionName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccBemClient()
		ctx := context.Background()

		fnRes, err := client.Functions.Get(ctx, functionName, bem.FunctionGetParams{})
		if err != nil {
			return fmt.Errorf("failed to fetch function %q from API: %w", functionName, err)
		}

		wfRes, err := client.Workflows.Get(ctx, workflowName)
		if err != nil {
			return fmt.Errorf("failed to fetch workflow %q from API: %w", workflowName, err)
		}

		if len(wfRes.Workflow.Nodes) == 0 {
			return fmt.Errorf("workflow %q has no nodes", workflowName)
		}

		got := wfRes.Workflow.Nodes[0].Function.VersionNum
		want := fnRes.Function.VersionNum
		if got != want {
			return fmt.Errorf(
				"workflow %q node 0's function.versionNum = %d, want %d (function %q's current version) - "+
					"this is exactly BEM-1392's atomic-DAG-fields bug if it regresses: mainNodeName/edges "+
					"omitted from the PATCH body because they were individually unchanged, so the bem API "+
					"rejected the update and the node never advanced past its old pinned version",
				workflowName, got, want, functionName,
			)
		}
		return nil
	}
}

// TestAccWorkflowResource_FunctionVersionBumpUpdatesDAG is the direct
// regression test for BEM-1392, matching the scenario confirmed live
// against stg-ue1-sr-0001's bem-customer component:
//
//  1. Create a function and a workflow whose sole node pins that function's
//     version_num.
//  2. Change only the function's display_name (forces a new function
//     version server-side - Update's own docs: "displayName ... updates
//     also create a new version"). Nothing in the workflow's own config
//     changes; only the interpolated version_num differs. Pre-fix, this is
//     exactly the apply that 400s ("mainNodeName, nodes, and edges must all
//     be provided together when updating the workflow DAG") because the
//     patch encoder only sent the changed `nodes` field.
func TestAccWorkflowResource_FunctionVersionBumpUpdatesDAG(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	functionName := acctest.RandomWithPrefix("tf-acc-bem-1392-fn")
	workflowName := acctest.RandomWithPrefix("tf-acc-bem-1392-wf")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionAndWorkflowConfig(functionName, "BEM-1392 Acc Test Function", workflowName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_workflow.test", "name", workflowName),
					resource.TestCheckResourceAttr("bem_workflow.test", "nodes.#", "1"),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowName, functionName),
				),
			},
			{
				// Only the function's display_name changes - the workflow's
				// own config is byte-identical to the step above. The only
				// reason nodes differs at all is the interpolated
				// version_num tracking the function's new version.
				Config: testAccFunctionAndWorkflowConfig(functionName, "BEM-1392 Acc Test Function (bumped)", workflowName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.splitter", "display_name", "BEM-1392 Acc Test Function (bumped)"),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowName, functionName),
				),
			},
		},
	})
}

// TestAccWorkflowResource_SharedFunctionAcrossTwoWorkflows confirms a
// function can be referenced by more than one workflow (platform confirmed:
// protocol/function.go's UsedInWorkflows is a list, and nothing in the
// DB/manager layer constrains a function version to a single workflow's
// node) and that bumping it updates every referencing workflow correctly in
// the same apply - each bem_workflow.Update() call is an independent HTTP
// request with its own plan/state pair, so the atomic-group fix runs once
// per call with no cross-resource coordination needed. This is the "one
// function, many customer workflows" case opto's real usage represents.
func TestAccWorkflowResource_SharedFunctionAcrossTwoWorkflows(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	functionName := acctest.RandomWithPrefix("tf-acc-bem-1392-shared-fn")
	workflowNameA := acctest.RandomWithPrefix("tf-acc-bem-1392-shared-wf-a")
	workflowNameB := acctest.RandomWithPrefix("tf-acc-bem-1392-shared-wf-b")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSharedFunctionTwoWorkflowsConfig(functionName, "BEM-1392 Shared Function", workflowNameA, workflowNameB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_workflow.a", "name", workflowNameA),
					resource.TestCheckResourceAttr("bem_workflow.b", "name", workflowNameB),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowNameA, functionName),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowNameB, functionName),
				),
			},
			{
				// One function update, both workflows must pick it up in
				// this single apply - two independent Update() calls, each
				// hitting the atomic-group fix on its own.
				Config: testAccSharedFunctionTwoWorkflowsConfig(functionName, "BEM-1392 Shared Function (bumped)", workflowNameA, workflowNameB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.splitter", "display_name", "BEM-1392 Shared Function (bumped)"),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowNameA, functionName),
					testAccCheckWorkflowNodeVersionMatchesFunction(workflowNameB, functionName),
				),
			},
		},
	})
}
