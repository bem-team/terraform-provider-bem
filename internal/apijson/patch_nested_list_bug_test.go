package apijson

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Mirrors internal/services/workflow/model.go's WorkflowNodesModel /
// WorkflowNodesFunctionModel shape: a *[]*struct{...} field where the struct
// has a plain types.String leaf, and one leaf nested one level deeper inside
// a *struct.
type PatchBugFunctionRef struct {
	ID types.String `tfsdk:"id" json:"id,optional"`
}

type PatchBugNode struct {
	Name     types.String         `tfsdk:"name" json:"name,optional"`
	Function *PatchBugFunctionRef `tfsdk:"function" json:"function,required"`
}

// json tag "aMainNodeName" sorts BEFORE "nodes" - MainNodeName gets encoded
// first. This is the real WorkflowModel field order (mainNodeName < nodes
// alphabetically).
type PatchBugContainerFieldFirst struct {
	MainNodeName types.String     `tfsdk:"main_node_name" json:"aMainNodeName,required,no_refresh"`
	Nodes        *[]*PatchBugNode `tfsdk:"nodes" json:"nodes,optional,no_refresh"`
}

// Same fields, but MainNodeName's json tag sorts AFTER "nodes" - Nodes gets
// encoded first instead.
type PatchBugContainerFieldSecond struct {
	MainNodeName types.String     `tfsdk:"main_node_name" json:"zMainNodeName,required,no_refresh"`
	Nodes        *[]*PatchBugNode `tfsdk:"nodes" json:"nodes,optional,no_refresh"`
}

// Reproduces the exact scenario from BEM-1367 testing: after `terraform
// import`, state has every no_refresh field null (ImportState never
// hydrates them), while the plan/config carries real values. This is what
// MarshalJSONForUpdate/MarshalForPatch sees on the first post-import apply.
//
// FAILS: when a plain types.String field (null state -> value plan) is
// encoded BEFORE a *[]*struct{...} field (nil state -> populated plan) in
// the same MarshalForPatch call, the list's nested string leaves
// (nodes[].name, nodes[].function.id) are silently dropped/nulled instead
// of being sent - even though the top-level "nodes" key IS present in the
// output. This matches the live PATCH body captured during BEM-1367
// testing byte-for-byte: {"nodes":[{"function":{}}], ...} instead of
// {"nodes":[{"function":{"id":"f_abc123"},"name":"splitter"}], ...}.
func TestPatchBug_StringFieldBeforeList_CorruptsNestedList(t *testing.T) {
	state := PatchBugContainerFieldFirst{Nodes: nil}
	plan := PatchBugContainerFieldFirst{
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*PatchBugNode{
			{
				Name:     types.StringValue("splitter"),
				Function: &PatchBugFunctionRef{ID: types.StringValue("f_abc123")},
			},
		},
	}

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{"aMainNodeName":"splitter","nodes":[{"function":{"id":"f_abc123"},"name":"splitter"}]}`
	t.Logf("got:  %s", string(got))
	t.Logf("want: %s", want)

	if string(got) != want {
		t.Errorf("MarshalForPatch corrupted the nested list when a preceding string field was encoded first.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// PASSES in isolation (go test -run TestPatchBug_ListFieldFirst_NoCorruption)
// with the identical data, once the list field is encoded FIRST
// (alphabetically before the plain string field) instead of second - proving
// the corruption is an order-dependent side effect, not a fundamental flaw
// in list-of-struct encoding.
//
// IMPORTANT: this test FAILS when run in the same process AFTER
// TestPatchBug_StringFieldBeforeList_CorruptsNestedList (e.g. `go test ./...`
// running the whole package), even though PatchBugNode/PatchBugFunctionRef's
// OWN field order here is fine. That's because internal/apijson's per-type
// encoder cache (`var encoders sync.Map`, encoder.go) is a PACKAGE-LEVEL
// global that persists for the life of the process. Once
// PatchBugNode/PatchBugFunctionRef's encoder gets built once under the "bad"
// ordering, it's permanently cached in that corrupted form and reused by
// EVERY subsequent MarshalForPatch call for that type - regardless of that
// later call's own field order. In the real provider, this means whichever
// resource happens to be processed first in a given `terraform apply` can
// poison node/tag/classification encoding for every other resource sharing
// that nested struct shape for the rest of that apply.
func TestPatchBug_ListFieldFirst_NoCorruption(t *testing.T) {
	state := PatchBugContainerFieldSecond{Nodes: nil}
	plan := PatchBugContainerFieldSecond{
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*PatchBugNode{
			{
				Name:     types.StringValue("splitter"),
				Function: &PatchBugFunctionRef{ID: types.StringValue("f_abc123")},
			},
		},
	}

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{"nodes":[{"function":{"id":"f_abc123"},"name":"splitter"}],"zMainNodeName":"splitter"}`
	t.Logf("got:  %s", string(got))
	t.Logf("want: %s", want)

	if string(got) != want {
		t.Errorf("expected no corruption when list field is encoded first.\ngot:  %s\nwant: %s", string(got), want)
	}
}
