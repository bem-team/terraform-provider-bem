package workflow

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// bem1367Fixture returns the plan/state pair and expected JSON body shared by
// TestMarshalJSONForUpdate_PostImportNilState and
// TestMarshalJSONForUpdate_ConcurrentRace (stress_bug_test.go): after
// `terraform import`, state has every no_refresh field null (ImportState
// never hydrates them), while the plan/config carries the real desired
// values. This is exactly what MarshalJSONForUpdate sees on the first
// post-import apply.
func bem1367Fixture() (plan, state WorkflowModel, want string) {
	state = WorkflowModel{
		// ImportState explicitly sets Name; everything else no_refresh stays null.
		Name: types.StringValue("bem-1367-repro-workflow"),
	}

	plan = WorkflowModel{
		Name:         types.StringValue("bem-1367-repro-workflow"),
		MainNodeName: types.StringValue("splitter"),
		DisplayName:  types.StringValue("BEM-1367 Repro Workflow"),
		Tags:         &[]types.String{types.StringValue("bem-1367-repro")},
		Nodes: &[]*WorkflowNodesModel{
			{
				Name:     types.StringValue("splitter"),
				Function: &WorkflowNodesFunctionModel{ID: types.StringValue("f_abc123")},
			},
		},
		Edges: &[]*WorkflowEdgesModel{},
	}

	want = `{"displayName":"BEM-1367 Repro Workflow","edges":[],"mainNodeName":"splitter","nodes":[{"function":{"id":"f_abc123"},"name":"splitter"}],"tags":["bem-1367-repro"]}`
	return plan, state, want
}

// Uses the real WorkflowModel/WorkflowNodesModel/WorkflowNodesFunctionModel
// types and the real MarshalJSONForUpdate method - no hand-rolled mirror
// types.
func TestMarshalJSONForUpdate_PostImportNilState(t *testing.T) {
	plan, state, want := bem1367Fixture()

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	t.Logf("got: %s", string(got))

	if string(got) != want {
		t.Errorf("MarshalJSONForUpdate corrupted the request body.\ngot:  %s\nwant: %s", string(got), want)
	}
}
