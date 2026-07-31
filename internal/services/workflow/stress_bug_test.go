package workflow

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Tests whether the corruption is a genuine concurrency/race issue in the
// shared apijson encoder cache: many goroutines call MarshalJSONForUpdate
// simultaneously in a FRESH process (empty global cache), racing to build
// the same cached type encoders. If this fails while a single-threaded loop
// of the same call never does, that pins the remaining bug as a race in the
// lazy cache-build path, not a deterministic logic bug.
func TestMarshalJSONForUpdate_ConcurrentRace(t *testing.T) {
	const n = 200
	var wg sync.WaitGroup
	results := make([]string, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			state := WorkflowModel{Name: types.StringValue("bem-1367-repro-workflow")}
			plan := WorkflowModel{
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
			got, err := plan.MarshalJSONForUpdate(state)
			results[idx] = string(got)
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	want := `{"displayName":"BEM-1367 Repro Workflow","edges":[],"mainNodeName":"splitter","nodes":[{"function":{"id":"f_abc123"},"name":"splitter"}],"tags":["bem-1367-repro"]}`
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
		t.Errorf("%d/%d concurrent calls corrupted", failures, n)
	}
}
