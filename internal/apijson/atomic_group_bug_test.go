package apijson

import (
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Mirrors the shape that broke BEM-1392: three sibling fields the API
// requires together on any update (mainNodeName/nodes/edges on
// bem_workflow), but where an ordinary apply only ever changes one of them
// (nodes, via a referenced function's version bump). JSON Merge Patch's
// per-field diffing correctly omits the two unchanged siblings on its own -
// atomic_group is what makes MarshalForPatch send them anyway.
type AtomicGroupContainer struct {
	MainNodeName types.String        `tfsdk:"main_node_name" json:"mainNodeName,required,atomic_group=dag"`
	Nodes        *[]*AtomicGroupNode `tfsdk:"nodes" json:"nodes,required,atomic_group=dag"`
	Edges        *[]*AtomicGroupEdge `tfsdk:"edges" json:"edges,optional,atomic_group=dag"`
	DisplayName  types.String        `tfsdk:"display_name" json:"displayName,optional"`
}

type AtomicGroupNode struct {
	Name       types.String `tfsdk:"name" json:"name,optional"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,optional"`
}

type AtomicGroupEdge struct {
	SourceNodeName types.String `tfsdk:"source_node_name" json:"sourceNodeName,required"`
}

func atomicGroupFixture() (plan, state AtomicGroupContainer) {
	state = AtomicGroupContainer{
		MainNodeName: types.StringValue("splitter"),
		Nodes: &[]*AtomicGroupNode{
			{Name: types.StringValue("splitter"), VersionNum: types.Int64Value(2)},
		},
		Edges:       &[]*AtomicGroupEdge{},
		DisplayName: types.StringValue("Customer Test Workflow"),
	}
	plan = AtomicGroupContainer{
		MainNodeName: types.StringValue("splitter"), // unchanged
		Nodes: &[]*AtomicGroupNode{
			{Name: types.StringValue("splitter"), VersionNum: types.Int64Value(3)}, // changed
		},
		Edges:       &[]*AtomicGroupEdge{},                       // unchanged
		DisplayName: types.StringValue("Customer Test Workflow"), // unchanged, no atomic_group - must stay omitted
	}
	return plan, state
}

// Without atomic_group, only "nodes" would appear in the patch body -
// mainNodeName/edges are individually unchanged, so per-field diffing omits
// them, and the bem API rejects the resulting PATCH. This asserts the fixed
// behavior: all three dag-tagged fields are present once any one changes,
// while the untagged, genuinely-unchanged displayName field still stays
// omitted (atomic_group must not defeat patch diffing for unrelated fields).
func TestAtomicGroup_SiblingChange_ForcesGroupTogether(t *testing.T) {
	plan, state := atomicGroupFixture()

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	// Key order is deterministic but not alphabetical across both passes:
	// pass 1 writes only genuinely-changed fields in sorted order ("nodes" is
	// the only one here), pass 2 then appends the rest of the triggered group
	// ("edges", "mainNodeName"), also in sorted order. displayName never
	// appears - it's untagged and genuinely unchanged.
	want := `{"nodes":[{"name":"splitter","versionNum":3}],"edges":[],"mainNodeName":"splitter"}`
	t.Logf("got:  %s", string(got))
	t.Logf("want: %s", want)

	if string(got) != want {
		t.Errorf("atomic_group did not force the unchanged siblings into the patch body.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// The API's other documented path: "Omit all three to update only metadata
// fields." An untagged field changing must NOT drag the group along - if it
// did, every metadata-only edit would start sending a full DAG replacement,
// turning a cheap PATCH into an unnecessary new workflow version.
func TestAtomicGroup_UntaggedFieldChange_DoesNotTriggerGroup(t *testing.T) {
	plan, state := atomicGroupFixture()
	*plan.Nodes = *state.Nodes // DAG fully unchanged
	plan.DisplayName = types.StringValue("Renamed")

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{"displayName":"Renamed"}`
	if string(got) != want {
		t.Errorf("an untagged field's change must not force the atomic group into the body.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// A cleared list member is the case that survived the first version of this
// fix: JSON Merge Patch encodes "plan cleared what state had" as an explicit
// null, and null deserializes server-side to the same nil pointer an absent
// key does - so the group looks partial to the API's present-key count and
// still 400s, despite all three keys literally appearing in the body.
// Confirmed against platform/protocol/workflow.go's WorkflowUpdateRequestV3,
// where Edges is *[]WorkflowEdgeRequest and the validator counts non-nil
// pointers, not present keys.
func TestAtomicGroup_ClearedListMember_SendsEmptyArrayNotNull(t *testing.T) {
	plan, state := atomicGroupFixture()
	state.Edges = &[]*AtomicGroupEdge{{SourceNodeName: types.StringValue("splitter")}}
	plan.Edges = nil // user removed `edges` from config entirely

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{"edges":[],"nodes":[{"name":"splitter","versionNum":3}],"mainNodeName":"splitter"}`
	if string(got) != want {
		t.Errorf("a cleared list member of an atomic group must serialize as [] rather than null.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// atomic_group is a patch-encoding concept - it exists only because per-field
// diffing can omit an unchanged sibling. Non-patch encoding never omits a set
// field in the first place, so the group logic has nothing to correct there
// and must stay out of the way: a create body should look exactly as it did
// before atomic_group existed, with a never-configured optional list still
// absent rather than force-sent as [].
func TestAtomicGroup_NonPatchMarshal_LeavesUnsetMemberAbsent(t *testing.T) {
	plan, _ := atomicGroupFixture()
	plan.Edges = nil

	got, err := MarshalRoot(plan)
	if err != nil {
		t.Fatalf("MarshalRoot failed: %v", err)
	}

	want := `{"displayName":"Customer Test Workflow","mainNodeName":"splitter","nodes":[{"name":"splitter","versionNum":3}]}`
	if string(got) != want {
		t.Errorf("non-patch encoding must not force atomic-group members.\ngot:  %s\nwant: %s", string(got), want)
	}
}

type unencodableGroupContainer struct {
	Nodes types.String `tfsdk:"nodes" json:"nodes,required,atomic_group=dag"`
	// A plain []T rather than *[]T: nil here is not a nil *pointer*, so the
	// empty-array rescue doesn't apply and there is genuinely nothing to send.
	Edges []types.String `tfsdk:"edges" json:"edges,optional,atomic_group=dag"`
}

// If a group member can't be encoded, emitting the rest of the group anyway
// would produce exactly the partial-group body the API rejects - turning a
// provider bug into an unexplained 400 with nothing pointing back here. Fail
// loudly and name the field instead.
func TestAtomicGroup_UnencodableMember_ReturnsErrorNotPartialGroup(t *testing.T) {
	state := unencodableGroupContainer{Nodes: types.StringValue("v2")}
	plan := unencodableGroupContainer{Nodes: types.StringValue("v3")}

	got, err := MarshalForPatch(plan, state)
	if err == nil {
		t.Fatalf("expected an error rather than a partial group body, got: %s", string(got))
	}
	for _, want := range []string{"edges", "dag"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should name %q so the failure is traceable; got: %v", want, err)
		}
	}
}

type twoGroupContainer struct {
	DagMain  types.String        `tfsdk:"dag_main" json:"dagMain,required,atomic_group=dag"`
	DagNodes *[]*AtomicGroupNode `tfsdk:"dag_nodes" json:"dagNodes,required,atomic_group=dag"`
	NetHost  types.String        `tfsdk:"net_host" json:"netHost,required,atomic_group=net"`
	NetPort  types.Int64         `tfsdk:"net_port" json:"netPort,required,atomic_group=net"`
}

// Groups are keyed by name, so a change inside one group must not force
// members of an unrelated group. Guards against the obvious regression of
// collapsing groupChanged into a single "any group fired" boolean.
func TestAtomicGroup_IndependentGroups_DoNotCrossTrigger(t *testing.T) {
	state := twoGroupContainer{
		DagMain:  types.StringValue("splitter"),
		DagNodes: &[]*AtomicGroupNode{{Name: types.StringValue("splitter"), VersionNum: types.Int64Value(2)}},
		NetHost:  types.StringValue("api.bem.ai"),
		NetPort:  types.Int64Value(443),
	}
	plan := state
	plan.DagNodes = &[]*AtomicGroupNode{{Name: types.StringValue("splitter"), VersionNum: types.Int64Value(3)}}

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{"dagNodes":[{"name":"splitter","versionNum":3}],"dagMain":"splitter"}`
	if string(got) != want {
		t.Errorf("the 'net' group must stay omitted when only the 'dag' group changed.\ngot:  %s\nwant: %s", string(got), want)
	}
}

// atomic_group adds a second encoderFunc (noPatchFn) per tagged field, built
// at cache-population time and stored in the same package-level `encoders`
// map BEM-1367 found to be both cache-poisoning-prone and a genuine data
// race. This races many goroutines through a cold cache on the atomic_group
// path specifically; the pre-existing concurrency test doesn't cover it
// because no type it touches carries the tag.
func TestAtomicGroup_ConcurrentRace(t *testing.T) {
	const n = 200
	var wg sync.WaitGroup
	results := make([]string, n)
	errs := make([]error, n)

	const want = `{"nodes":[{"name":"splitter","versionNum":3}],"edges":[],"mainNodeName":"splitter"}`
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			plan, state := atomicGroupFixture()
			got, err := MarshalForPatch(plan, state)
			results[idx] = string(got)
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	failures := 0
	for i, r := range results {
		if errs[i] != nil {
			t.Errorf("goroutine %d: error: %v", i, errs[i])
			continue
		}
		if r != want {
			failures++
			t.Logf("goroutine %d CORRUPTED: %s", i, r)
		}
	}
	if failures > 0 {
		t.Errorf("%d/%d concurrent atomic_group calls corrupted", failures, n)
	}
}

// No member of the group changed - none of them, nor any other field,
// should appear. Confirms atomic_group only fires when triggered, it
// doesn't turn the whole group into an unconditional full-send.
func TestAtomicGroup_NoChange_StaysEmpty(t *testing.T) {
	plan, state := atomicGroupFixture()
	*plan.Nodes = *state.Nodes // make nodes match state too - nothing changed anywhere

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}

	want := `{}`
	if string(got) != want {
		t.Errorf("expected no fields when nothing in the group changed.\ngot:  %s\nwant: %s", string(got), want)
	}
}
