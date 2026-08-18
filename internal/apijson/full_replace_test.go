package apijson

import (
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// full_replace covers the case atomic_group structurally cannot: a field the
// API validates against itself on every request, so a patch that omits it is
// rejected (or worse, silently overwrites the stored value with an empty one).
//
// The live instance is bem_function's `config` on an enrich function. The API
// validates `config` on every update whether or not the request carries it, and
// cannot distinguish "omitted" from "sent empty" - so a patch that leaves it out
// fails with "enrich function must have at least one step", even when the only
// thing being changed is display_name. There is no
// sibling field whose change could trigger a group, which is why this needs a
// per-field mechanism rather than a group.
type FullReplaceContainer struct {
	Config      *FullReplaceConfig `tfsdk:"config" json:"config,optional,full_replace"`
	DisplayName types.String       `tfsdk:"display_name" json:"displayName,optional"`
	Tags        *[]types.String    `tfsdk:"tags" json:"tags,optional"`
}

type FullReplaceConfig struct {
	Steps     *[]*FullReplaceStep     `tfsdk:"steps" json:"steps,required"`
	Endpoints *[]*FullReplaceEndpoint `tfsdk:"endpoints" json:"endpoints,optional"`
}

type FullReplaceStep struct {
	SourceField  types.String `tfsdk:"source_field" json:"sourceField,required"`
	EndpointName types.String `tfsdk:"endpoint_name" json:"endpointName,optional"`
}

type FullReplaceEndpoint struct {
	Name types.String `tfsdk:"name" json:"name,required"`
	URL  types.String `tfsdk:"url" json:"url,required"`
}

func fullReplaceFixture() (plan, state FullReplaceContainer) {
	config := func() *FullReplaceConfig {
		return &FullReplaceConfig{
			Steps: &[]*FullReplaceStep{
				{
					SourceField:  types.StringValue("lineItems[*].description"),
					EndpointName: types.StringValue("catalog"),
				},
			},
			Endpoints: &[]*FullReplaceEndpoint{
				{
					Name: types.StringValue("catalog"),
					URL:  types.StringValue("https://sink.example.com/enrich"),
				},
			},
		}
	}

	state = FullReplaceContainer{
		Config:      config(),
		DisplayName: types.StringValue("TF Enrich"),
		Tags:        &[]types.String{types.StringValue("terraform-managed")},
	}
	plan = FullReplaceContainer{
		Config:      config(), // byte-identical to state - patch diffing would omit it
		DisplayName: types.StringValue("TF Enrich"),
		Tags:        &[]types.String{types.StringValue("terraform-managed")},
	}
	return plan, state
}

// The headline case. A metadata-only edit leaves config untouched, so ordinary
// per-field diffing drops it entirely - and the server then rejects the
// request. full_replace re-encodes the whole block regardless.
func TestFullReplace_UnrelatedFieldChange_StillSendsWholeBlock(t *testing.T) {
	plan, state := fullReplaceFixture()
	plan.DisplayName = types.StringValue("TF Enrich (renamed)")

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}
	body := string(got)
	t.Logf("body: %s", body)

	for _, want := range []string{`"displayName":"TF Enrich (renamed)"`, `"steps"`, `"endpoints"`} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %s\nbody: %s", want, body)
		}
	}
	// The whole point: not merely present, but complete.
	if !strings.Contains(body, `"name":"catalog"`) || !strings.Contains(body, `"endpointName":"catalog"`) {
		t.Errorf("config block was sent but not in full\nbody: %s", body)
	}
	// full_replace must not defeat patch diffing for anything else.
	if strings.Contains(body, `"tags"`) {
		t.Errorf("unchanged, untagged tags field was sent\nbody: %s", body)
	}
}

// A change *inside* the block must send the block whole too, not just the
// changed leaf. This is the shape that produced the observed
// "step 0 references unknown endpoint" 400: steps re-encoded, endpoints
// dropped, so the server validated steps against an empty endpoint set.
func TestFullReplace_NestedChange_SendsSiblingsInsideBlock(t *testing.T) {
	plan, state := fullReplaceFixture()
	plan.Config = &FullReplaceConfig{
		Steps: &[]*FullReplaceStep{
			{
				SourceField:  types.StringValue("lineItems[*].sku"), // changed
				EndpointName: types.StringValue("catalog"),
			},
		},
		Endpoints: &[]*FullReplaceEndpoint{
			{
				Name: types.StringValue("catalog"), // unchanged
				URL:  types.StringValue("https://sink.example.com/enrich"),
			},
		},
	}

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}
	body := string(got)
	t.Logf("body: %s", body)

	if !strings.Contains(body, `"sourceField":"lineItems[*].sku"`) {
		t.Errorf("changed step field missing\nbody: %s", body)
	}
	if !strings.Contains(body, `"endpoints"`) || !strings.Contains(body, `"url":"https://sink.example.com/enrich"`) {
		t.Errorf("unchanged endpoints were dropped from the block - this is the 400\nbody: %s", body)
	}
}

// full_replace forces completeness, not presence. A field nobody configured
// must stay absent, or every non-enrich function type would start sending an
// empty config block on ordinary updates.
func TestFullReplace_NilField_StaysOmitted(t *testing.T) {
	state := FullReplaceContainer{
		DisplayName: types.StringValue("TF Split"),
	}
	plan := state
	plan.DisplayName = types.StringValue("TF Split (renamed)")

	got, err := MarshalForPatch(plan, state)
	if err != nil {
		t.Fatalf("MarshalForPatch failed: %v", err)
	}
	body := string(got)
	t.Logf("body: %s", body)

	if strings.Contains(body, `"config"`) {
		t.Errorf("a nil full_replace field was sent; it must stay omitted\nbody: %s", body)
	}
	if !strings.Contains(body, `"displayName":"TF Split (renamed)"`) {
		t.Errorf("the field that actually changed is missing\nbody: %s", body)
	}
}

// Non-patch marshalling must be untouched - full_replace only has meaning
// against a state to diff from.
func TestFullReplace_NonPatchMarshalUnaffected(t *testing.T) {
	_, model := fullReplaceFixture()

	got, err := Marshal(model)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	body := string(got)
	t.Logf("body: %s", body)

	for _, want := range []string{`"config"`, `"displayName"`, `"tags"`} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %s\nbody: %s", want, body)
		}
	}
}

// The encoder cache is a package-level global whose closures get invoked
// concurrently once cached, and full_replace adds a second encoder (noPatchFn)
// per tagged field. BEM-1367 shipped a data race in exactly this cache, so a
// new field-level encoder path gets the same concurrency guard the atomic_group
// work did. Meaningful only under -race.
func TestFullReplace_ConcurrentMarshalIsRaceFree(t *testing.T) {
	plan, state := fullReplaceFixture()
	plan.DisplayName = types.StringValue("TF Enrich (renamed)")

	const goroutines = 100
	var wg sync.WaitGroup
	bodies := make([]string, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got, err := MarshalForPatch(plan, state)
			bodies[i], errs[i] = string(got), err
		}(i)
	}
	wg.Wait()

	for i := range goroutines {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: %v", i, errs[i])
		}
		if !strings.Contains(bodies[i], `"endpoints"`) || !strings.Contains(bodies[i], `"steps"`) {
			t.Fatalf("goroutine %d produced an incomplete block: %s", i, bodies[i])
		}
		if bodies[i] != bodies[0] {
			t.Fatalf("goroutine %d diverged:\n got: %s\nwant: %s", i, bodies[i], bodies[0])
		}
	}
}
