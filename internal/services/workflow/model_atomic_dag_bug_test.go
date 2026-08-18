package workflow

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Tripwire for the two ways the BEM-1392 fix can silently disappear without
// any other test noticing:
//
//   - a typo in the tag option. apijson's tag parser ignores every option it
//     doesn't recognise, so `atomic_groups=dag` or `atomic_group =dag` parses
//     clean and just quietly turns the fix off.
//   - a Stainless regen. model.go carries the "File generated from our OpenAPI
//     spec" header and these tags are hand-added, so a spec sync drops them.
//
// Either way the encoder falls back to plain patch diffing and BEM-1392 comes
// straight back, with nothing failing until a customer's apply 400s. Asserting
// on the struct tags directly is what makes that loud instead of silent - the
// behavioural tests above can't catch it, because with the tags gone they'd
// still be testing a correctly-working plain patch encoder.
func TestWorkflowModel_DAGFieldsCarryAtomicGroupTag(t *testing.T) {
	const wantOption = "atomic_group=dag"
	modelType := reflect.TypeOf(WorkflowModel{})

	for _, fieldName := range []string{"MainNodeName", "Nodes", "Edges"} {
		field, ok := modelType.FieldByName(fieldName)
		if !ok {
			t.Errorf("WorkflowModel has no field %q - if it was renamed, the atomic_group tag must move with it", fieldName)
			continue
		}

		// Match a whole comma-separated option, exactly as apijson's
		// parseJSONStructTag splits it - a substring check would accept
		// malformed neighbours like "xatomic_group=dagy".
		found := false
		for _, part := range strings.Split(field.Tag.Get("json"), ",") {
			if part == wantOption {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(
				"WorkflowModel.%s is missing the %q json tag option (got tag: %q).\n"+
					"The bem API requires mainNodeName, nodes and edges to be sent together on any DAG "+
					"update; without this tag the patch encoder omits the unchanged ones and the PATCH "+
					"400s (BEM-1392). If this failed right after a Stainless regen, re-apply the tag.",
				fieldName, wantOption, field.Tag.Get("json"),
			)
		}
	}
}

// Reproduces BEM-1392: a workflow that's already in sync (state fully
// populated, not the post-import nil-state case BEM-1367's fixture covers)
// where only a referenced function's version_num changes. mainNodeName and
// edges are identical in plan and state - JSON Merge Patch's per-field
// diffing correctly omits them on its own, which is exactly what produced
// the bem API's 400 ("mainNodeName, nodes, and edges must all be provided
// together when updating the workflow DAG") in the practitioner report that
// opened BEM-1392.
func bem1392Fixture() (plan, state WorkflowModel) {
	state = WorkflowModel{
		Name:         types.StringValue("example-workflow"),
		MainNodeName: types.StringValue("node_one"),
		Nodes: &[]*WorkflowNodesModel{
			{
				Name: types.StringValue("node_one"),
				Function: &WorkflowNodesFunctionModel{
					Name:       types.StringValue("example-extractor"),
					VersionNum: types.Int64Value(2),
				},
			},
		},
		Edges: &[]*WorkflowEdgesModel{},
	}

	plan = WorkflowModel{
		Name:         types.StringValue("example-workflow"),
		MainNodeName: types.StringValue("node_one"), // unchanged
		Nodes: &[]*WorkflowNodesModel{
			{
				Name: types.StringValue("node_one"),
				Function: &WorkflowNodesFunctionModel{
					Name:       types.StringValue("example-extractor"),
					VersionNum: types.Int64Value(3), // changed - the only real diff
				},
			},
		},
		Edges: &[]*WorkflowEdgesModel{}, // unchanged
	}
	return plan, state
}

// Pre-fix, this produced exactly the PATCH body captured from that report -
// {"nodes":[{"function":{"name":"example-extractor","versionNum":3},"name":"node_one"}]}
// - missing mainNodeName and edges, which the bem API rejects. Post-fix,
// atomic_group=dag on WorkflowModel's MainNodeName/Nodes/Edges forces all
// three into the body together whenever any one of them changes.
func TestMarshalJSONForUpdate_NodeVersionBumpSendsFullDAG(t *testing.T) {
	plan, state := bem1392Fixture()

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{"nodes":[{"function":{"name":"example-extractor","versionNum":3},"name":"node_one"}],"edges":[],"mainNodeName":"node_one"}`
	t.Logf("got:  %s", string(got))
	t.Logf("want: %s", want)

	if string(got) != want {
		t.Errorf("MarshalJSONForUpdate omitted mainNodeName/edges despite the DAG changing.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// Reproduces the gap the first version of this fix missed, caught live against
// a staging environment: a single-node
// workflow whose configuration never sets `edges` at all, so Edges is a genuine
// nil pointer in both plan and state - not an unchanged-but-present empty
// slice like bem1392Fixture uses. mainNodeName made it into the patch body
// correctly; edges didn't, because a nil-in-both-sides pointer has nothing
// for even the non-patch encoder to encode. The API still needs the key
// present. Live wire body pre-fix-for-this-gap:
// {"nodes":[{"function":{"id":"...","versionNum":3},"name":"splitter"}],"mainNodeName":"splitter"}
// - edges silently missing, same 400.
func TestMarshalJSONForUpdate_NeverConfiguredEdges_SendsEmptyArray(t *testing.T) {
	state := WorkflowModel{
		Name:         types.StringValue("single-node-workflow"),
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*WorkflowNodesModel{
			{
				Name: types.StringValue("splitter"),
				Function: &WorkflowNodesFunctionModel{
					ID: types.StringValue("f_exampleFunctionID"),
				},
			},
		},
		Edges: nil,
	}

	plan := WorkflowModel{
		Name:         types.StringValue("single-node-workflow"),
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*WorkflowNodesModel{
			{
				Name: types.StringValue("splitter"),
				Function: &WorkflowNodesFunctionModel{
					ID:         types.StringValue("f_exampleFunctionID"),
					VersionNum: types.Int64Value(3),
				},
			},
		},
		Edges: nil,
	}

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{"nodes":[{"function":{"id":"f_exampleFunctionID","versionNum":3},"name":"splitter"}],"edges":[],"mainNodeName":"splitter"}`
	t.Logf("got:  %s", string(got))
	t.Logf("want: %s", want)

	if string(got) != want {
		t.Errorf("expected an explicit empty edges array when edges was never configured.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// The other half of the API contract: "Omit all three to update only
// metadata fields." A display_name-only edit on the workflow itself must
// still send just displayName - forcing the DAG in would make every cosmetic
// edit create a new workflow version server-side.
func TestMarshalJSONForUpdate_MetadataOnlyChange_OmitsDAGFields(t *testing.T) {
	plan, state := bem1392Fixture()
	(*plan.Nodes)[0].Function.VersionNum = types.Int64Value(2) // DAG identical
	state.DisplayName = types.StringValue("Example Workflow")
	plan.DisplayName = types.StringValue("Example Workflow (Managed by Terraform)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{"displayName":"Example Workflow (Managed by Terraform)"}`
	if string(got) != want {
		t.Errorf("a metadata-only edit must not send the DAG fields.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// The trigger is symmetric: mainNodeName is the group member most likely to
// change on its own (repointing a workflow's entry node without touching the
// node list), and it must pull nodes and edges along just as nodes pulls it.
func TestMarshalJSONForUpdate_MainNodeNameChangeSendsFullDAG(t *testing.T) {
	plan, state := bem1392Fixture()
	(*plan.Nodes)[0].Function.VersionNum = types.Int64Value(2) // nodes identical
	plan.MainNodeName = types.StringValue("node_replacement")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{"mainNodeName":"node_replacement","edges":[],"nodes":[{"function":{"name":"example-extractor","versionNum":2},"name":"node_one"}]}`
	if string(got) != want {
		t.Errorf("a mainNodeName-only change must send the whole DAG.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// Removing `edges` from config while the DAG also changes: patch semantics
// encode the cleared field as an explicit null, which platform's
// WorkflowUpdateRequestV3 unmarshals into a nil *[]WorkflowEdgeRequest -
// indistinguishable from an absent key to its dagFieldCount check, so the
// request would still 400 with all three keys present in the body. Must
// serialize as [] instead.
func TestMarshalJSONForUpdate_ClearedEdges_SendsEmptyArrayNotNull(t *testing.T) {
	plan, state := bem1392Fixture()
	state.Edges = &[]*WorkflowEdgesModel{
		{SourceNodeName: types.StringValue("node_one")},
	}
	plan.Edges = nil // `edges` removed from HCL

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{"edges":[],"nodes":[{"function":{"name":"example-extractor","versionNum":3},"name":"node_one"}],"mainNodeName":"node_one"}`
	if string(got) != want {
		t.Errorf("cleared edges must serialize as [] so the API sees all three DAG fields.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// Creates go through MarshalJSON (non-patch), where nothing is ever omitted
// for being unchanged, so atomic_group has no work to do. A workflow whose
// config never sets `edges` must produce the same create body it did before
// this fix - no injected "edges":[].
func TestMarshalJSON_Create_LeavesNeverConfiguredEdgesAbsent(t *testing.T) {
	m := WorkflowModel{
		Name:         types.StringValue("single-node-workflow"),
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*WorkflowNodesModel{
			{
				Name: types.StringValue("splitter"),
				Function: &WorkflowNodesFunctionModel{
					ID: types.StringValue("f_exampleFunctionID"),
				},
			},
		},
		Edges: nil,
	}

	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	want := `{"mainNodeName":"splitter","name":"single-node-workflow","nodes":[{"function":{"id":"f_exampleFunctionID"},"name":"splitter"}]}`
	if string(got) != want {
		t.Errorf("the create path must be unaffected by atomic_group.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// Sanity check: if nothing in the DAG changes, none of the three
// atomic_group fields should appear - confirms the fix only fires when
// actually needed, not on every apply regardless of content.
func TestMarshalJSONForUpdate_NoNodeChange_OmitsDAGFields(t *testing.T) {
	plan, state := bem1392Fixture()
	(*plan.Nodes)[0].Function.VersionNum = types.Int64Value(2) // match state - nothing changed

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}

	want := `{}`
	if string(got) != want {
		t.Errorf("expected no DAG fields when nothing changed.\ngot:  %s\nwant: %s", string(got), want)
	}
}
