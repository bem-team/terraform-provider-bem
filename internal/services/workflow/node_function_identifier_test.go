package workflow

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API accepts exactly one identifier per node function:
//
//	400 node 0 function is invalid: function identifier must have either an ID or
//	    a name, but not both
//
// Confirmed against the API, not inferred. Both keys reach the body without anyone
// asking for it: a name-based configuration leaves `id` null, `id` is
// computed_optional so the response decode fills it in, prior state carries it into
// the plan, and any later DAG edit re-sends the whole node through
// atomic_group=dag.
//
// So this is the encoder's job, and it has to preserve whichever half the
// practitioner actually chose.
func TestMarshalJSONForUpdate_NodeFunctionSendsOneIdentifier(t *testing.T) {
	build := func(id, name string, version int64) WorkflowModel {
		fn := &WorkflowNodesFunctionModel{VersionNum: types.Int64Value(version)}
		if id != "" {
			fn.ID = types.StringValue(id)
		}
		if name != "" {
			fn.Name = types.StringValue(name)
		}
		return WorkflowModel{
			Name:         types.StringValue("example-workflow"),
			MainNodeName: types.StringValue("node_one"),
			Nodes:        &[]*WorkflowNodesModel{{Name: types.StringValue("node_one"), Function: fn}},
			Edges:        &[]*WorkflowEdgesModel{},
		}
	}

	for _, tc := range []struct {
		name       string
		id, fnName string
		wantKey    string
		notWant    string
	}{
		{
			// The failing shape: configuration referenced the function by name, and
			// the decode filled in the id behind it.
			name: "name configured, id hydrated from the response",
			id:   "f_exampleFunctionID", fnName: "example-extractor",
			wantKey: `"name":"example-extractor"`, notWant: `"id"`,
		},
		{
			// An id-based configuration must keep sending the id. `name` is plain
			// optional and is not filled in from a response, so nothing to strip.
			name: "id configured, no name",
			id:   "f_exampleFunctionID", fnName: "",
			wantKey: `"id":"f_exampleFunctionID"`, notWant: `"name":"example-extractor"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := build(tc.id, tc.fnName, 3).MarshalJSONForUpdate(build(tc.id, tc.fnName, 2))
			if err != nil {
				t.Fatalf("MarshalJSONForUpdate: %v", err)
			}
			got := string(body)
			if !strings.Contains(got, `"nodes"`) {
				t.Fatalf("the DAG is absent from the body, so this asserts nothing: %s", got)
			}
			if !strings.Contains(got, tc.wantKey) {
				t.Errorf("body is missing %s\ngot: %s", tc.wantKey, got)
			}
			// Only meaningful for the first case; harmless for the second.
			if tc.notWant == `"id"` && strings.Contains(got, `"id"`) {
				t.Errorf("body carries both id and name, which the API rejects with 400\ngot: %s", got)
			}
		})
	}
}
