package apijson

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
)

// Regression guard for a data race in newTerraformTypeDecoder that shipped in
// 0.13.0 and earlier, found by ./scripts/test-acc on its first full run -
// TestAccWorkflowResource_SharedFunctionAcrossTwoWorkflows updates two
// workflows concurrently, which is what it took to expose it.
//
// Four branches (NestedObjectList, List, NestedObjectMap, Map) each declared
//
//	var dec decoderFunc
//	originalDec := d.typeDecoder(...)
//	return func(...) { ...; dec = <one or the other>; ...; dec(...) }
//
// with `dec` captured by the returned closure. That closure is cached in the
// package-level `decoders` map and invoked concurrently, so every caller wrote
// the same variable and then read back whoever won. The consequence was not
// merely a reported race: a goroutine whose value was unset could select
// alwaysUpdateDecoder and then have a sibling overwrite it with originalDec
// before the call, decoding that goroutine's nested list with the wrong update
// behaviour - i.e. silently wrong state, the read-path twin of BEM-1367.
//
// Why no existing test caught it: CI runs ./scripts/test -race with TF_ACC
// unset, so no acceptance test runs; the unit tests were single-threaded, and
// -race only reports races it actually observes. It needed -race AND
// concurrency AND a customfield nested type in the same run.
//
// Meaningful only under -race.

type raceDecoderModel struct {
	Name       types.String                                    `tfsdk:"name" json:"name,computed_optional"`
	Nodes      customfield.NestedObjectList[raceDecoderNode]   `tfsdk:"nodes" json:"nodes,computed"`
	Tags       customfield.List[basetypes.StringValue]         `tfsdk:"tags" json:"tags,computed"`
	Labels     customfield.Map[basetypes.StringValue]          `tfsdk:"labels" json:"labels,computed"`
	NodesByKey customfield.NestedObjectMap[raceDecoderNode]    `tfsdk:"nodes_by_key" json:"nodesByKey,computed"`
	Nested     customfield.NestedObject[raceDecoderNodeHolder] `tfsdk:"nested" json:"nested,computed"`
}

type raceDecoderNode struct {
	Name       types.String `tfsdk:"name" json:"name,computed_optional"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,computed"`
}

type raceDecoderNodeHolder struct {
	Inner customfield.NestedObjectList[raceDecoderNode] `tfsdk:"inner" json:"inner,computed"`
}

const raceDecoderPayload = `{
  "name": "wf",
  "nodes": [{"name": "splitter", "versionNum": 3}, {"name": "extractor", "versionNum": 2}],
  "tags": ["a", "b"],
  "labels": {"env": "stg"},
  "nodesByKey": {"splitter": {"name": "splitter", "versionNum": 3}},
  "nested": {"inner": [{"name": "deep", "versionNum": 1}]}
}`

// nullModel is the shape that made the race consequential rather than merely
// reported: with the collections null, UnmarshalComputed takes the IfUnset
// branch and selects alwaysUpdateDecoder, so two goroutines disagree about
// which decoder to use for the same cached closure.
func nullModel() raceDecoderModel {
	ctx := context.TODO()
	return raceDecoderModel{
		Nodes:      customfield.NullObjectList[raceDecoderNode](ctx),
		Tags:       customfield.NullList[basetypes.StringValue](ctx),
		Labels:     customfield.NullMap[basetypes.StringValue](ctx),
		NodesByKey: customfield.NullObjectMap[raceDecoderNode](ctx),
		Nested:     customfield.NullObject[raceDecoderNodeHolder](ctx),
	}
}

func TestUnmarshalComputed_ConcurrentNestedCollections_IsRaceFree(t *testing.T) {
	const goroutines = 100

	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	results := make([]raceDecoderModel, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			model := nullModel()
			errs[i] = UnmarshalComputed([]byte(raceDecoderPayload), &model)
			results[i] = model
		}(i)
	}
	wg.Wait()

	for i := range goroutines {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: %v", i, errs[i])
		}
	}

	// Every goroutine decoded identical input from identical state, so every
	// result must be identical. A divergence here means the wrong decoder was
	// selected for some callers - the corruption the race could cause, which a
	// -race report alone would not prove.
	ctx := context.TODO()
	for i := 1; i < goroutines; i++ {
		if !results[i].Nodes.Equal(results[0].Nodes) {
			t.Fatalf("goroutine %d decoded different nodes than goroutine 0", i)
		}
		if !results[i].Tags.Equal(results[0].Tags) {
			t.Fatalf("goroutine %d decoded different tags than goroutine 0", i)
		}
		if !results[i].Labels.Equal(results[0].Labels) {
			t.Fatalf("goroutine %d decoded different labels than goroutine 0", i)
		}
		if !results[i].NodesByKey.Equal(results[0].NodesByKey) {
			t.Fatalf("goroutine %d decoded different nodesByKey than goroutine 0", i)
		}
	}

	// And the decode has to have actually populated something, or the loop above
	// is comparing empty values and proves nothing.
	decoded, diags := results[0].Nodes.AsStructSlice(ctx)
	if diags.HasError() {
		t.Fatalf("reading decoded nodes: %v", diags)
	}
	nodes, ok := decoded.([]raceDecoderNode)
	if !ok {
		t.Fatalf("AsStructSlice returned %T, want []raceDecoderNode", decoded)
	}
	if len(nodes) != 2 {
		t.Fatalf("decoded %d nodes, want 2 - the payload was not applied", len(nodes))
	}
}

// The same closures are shared between Unmarshal and UnmarshalComputed only if
// the cache key distinguishes them; exercising both concurrently checks the key
// as well as the closure bodies.
func TestUnmarshalMixedEntryPoints_Concurrent_IsRaceFree(t *testing.T) {
	const goroutines = 100

	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			model := nullModel()
			if i%2 == 0 {
				errs[i] = UnmarshalComputed([]byte(raceDecoderPayload), &model)
			} else {
				errs[i] = Unmarshal([]byte(raceDecoderPayload), &model)
			}
		}(i)
	}
	wg.Wait()

	for i := range goroutines {
		if errs[i] != nil && !strings.Contains(errs[i].Error(), "cannot marshal into invalid value") {
			t.Fatalf("goroutine %d: %v", i, errs[i])
		}
	}
}
