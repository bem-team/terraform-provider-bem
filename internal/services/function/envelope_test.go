package function

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bem-team/terraform-provider-bem/internal/apijson"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
)

// BEM-1396 finding 2, second half. bem_function re-planned an in-place update
// forever - one new function version per apply - and the workflow's node diff
// cascaded off it.
//
// The payload below is the real staging response for tf-enrich-enricher,
// trimmed to the fields that matter. Two properties of it drive everything:
// the resource's attributes live *inside* the {"function": ...} wrapper, and
// the server fills in defaults for the computed_optional leaves the config
// deliberately leaves unset (steps[].topK / searchMode / scoreThreshold,
// endpoints[].matchTopK / maxCandidates / maxPages).
//
// Decoding straight into FunctionModel - what Create and Update used to do -
// looks for `config` at the root, where it does not exist, so all six stayed
// null in state. fwserver.MarkComputedNilsAsUnknown marks a Computed attribute
// unknown whenever its configuration value is null, so each plan saw null,
// proposed an update, applied it, got null back, and proposed it again.
const realEnrichResponse = `{
  "function": {
    "functionID": "f_exampleFunctionID",
    "versionNum": 3,
    "functionName": "tf-enrich-enricher",
    "type": "enrich",
    "displayName": "TF Enrich",
    "tags": ["terraform-managed", "tf-enrich"],
    "description": "",
    "config": {
      "steps": [
        {
          "source": "endpoint",
          "sourceField": "lineItems[*].description",
          "targetField": "lineItems[*].enrichedProducts",
          "topK": 1,
          "searchMode": "semantic",
          "scoreThreshold": 0.6,
          "endpointName": "catalog"
        }
      ],
      "endpoints": [
        {
          "name": "catalog",
          "url": "https://sink.example.com/enrich",
          "method": "POST",
          "matchTopK": 1,
          "maxCandidates": 50,
          "maxPages": 10
        }
      ]
    }
  }
}`

// planModel is the enrich fixture as HCL configures it: the six
// server-defaulted leaves are all unset.
func planModel(ctx context.Context) FunctionModel {
	return FunctionModel{
		FunctionName: types.StringValue("tf-enrich-enricher"),
		Type:         types.StringValue("enrich"),
		DisplayName:  types.StringValue("TF Enrich"),
		Config: customfield.NewObjectMust(ctx, &FunctionConfigModel{
			Steps: &[]*FunctionConfigStepsModel{{
				SourceField:  types.StringValue("lineItems[*].description"),
				TargetField:  types.StringValue("lineItems[*].enrichedProducts"),
				EndpointName: types.StringValue("catalog"),
				Source:       types.StringValue("endpoint"),
			}},
			Endpoints: customfield.NewObjectListMust(ctx, []FunctionConfigEndpointsModel{{
				Name:   types.StringValue("catalog"),
				Method: types.StringValue("POST"),
				URL:    types.StringValue("https://sink.example.com/enrich"),
			}}),
		}),
	}
}

func firstStep(t *testing.T, m FunctionModel) *FunctionConfigStepsModel {
	t.Helper()

	cfg, diags := m.Config.Value(context.TODO())
	if diags.HasError() {
		t.Fatalf("reading config: %v", diags)
	}
	if cfg == nil || cfg.Steps == nil || len(*cfg.Steps) == 0 {
		t.Fatal("config.steps is empty")
	}
	return (*cfg.Steps)[0]
}

// The fix. Both halves of the model must come out populated.
func TestHydrateFromResponse_PopulatesConfigAndMirror(t *testing.T) {
	ctx := context.TODO()
	data := planModel(ctx)

	if err := hydrateFromResponse([]byte(realEnrichResponse), &data, apijson.UnmarshalComputed); err != nil {
		t.Fatalf("hydrateFromResponse: %v", err)
	}

	step := firstStep(t, data)
	for name, v := range map[string]bool{
		"top_k":           step.TopK.IsNull(),
		"search_mode":     step.SearchMode.IsNull(),
		"score_threshold": step.ScoreThreshold.IsNull(),
	} {
		if v {
			t.Errorf("config.steps[0].%s is null. The server returned a value, so a null here is "+
				"marked unknown on every plan and produces a perpetual in-place update.", name)
		}
	}
	if got := step.TopK.ValueInt64(); got != 1 {
		t.Errorf("config.steps[0].top_k = %d, want 1", got)
	}
	if got := step.SearchMode.ValueString(); got != "semantic" {
		t.Errorf("config.steps[0].search_mode = %q, want %q", got, "semantic")
	}

	// The endpoints list is a customfield list rather than a plain slice, and it
	// has the same server-defaulted leaves - it broke identically, so assert it
	// separately rather than assuming the steps check covers it.
	cfg, diags := data.Config.Value(ctx)
	if diags.HasError() {
		t.Fatalf("reading config: %v", diags)
	}
	endpointsAny, diags := cfg.Endpoints.AsStructSlice(ctx)
	if diags.HasError() {
		t.Fatalf("reading endpoints: %v", diags)
	}
	endpoints, ok := endpointsAny.([]FunctionConfigEndpointsModel)
	if !ok {
		t.Fatalf("AsStructSlice returned %T", endpointsAny)
	}
	if len(endpoints) == 0 {
		t.Fatal("config.endpoints is empty")
	}
	if endpoints[0].MatchTopK.IsNull() || endpoints[0].MaxCandidates.IsNull() || endpoints[0].MaxPages.IsNull() {
		t.Errorf("config.endpoints[0] server defaults are null: matchTopK=%v maxCandidates=%v maxPages=%v",
			endpoints[0].MatchTopK, endpoints[0].MaxCandidates, endpoints[0].MaxPages)
	}

	// And the mirror must survive. It is the only place version_num and
	// function_id are readable - FunctionModel has no top-level equivalents -
	// and configs interpolate them into workflow nodes, so nulling it would
	// break every apply that builds a workflow from a function.
	if data.Function.IsNull() {
		t.Fatal("the `function` mirror is null; configs read function.function_id / function.version_num from it")
	}
	fn, diags := data.Function.Value(ctx)
	if diags.HasError() {
		t.Fatalf("reading function mirror: %v", diags)
	}
	if fn.FunctionID.IsNull() || fn.VersionNum.IsNull() {
		t.Errorf("function.function_id = %v, function.version_num = %v; both must be readable",
			fn.FunctionID, fn.VersionNum)
	}

	// Configured attributes must not be clobbered by either pass.
	if got := data.FunctionName.ValueString(); got != "tf-enrich-enricher" {
		t.Errorf("function_name = %q after hydration", got)
	}
	if got := data.Type.ValueString(); got != "enrich" {
		t.Errorf("type = %q after hydration", got)
	}
}

// The import path, which no test covered until the live import proved it broken.
//
// ImportState passes UnmarshalForImport so that no_refresh attributes hydrate -
// on import there is no prior state for the refresh skip to protect. But that
// decoder is not computed-only, so unlike UnmarshalComputed it is allowed to
// write every field, including the no_refresh ones the envelope pass has just
// populated. The struct decoder nulls any field whose key is absent from the
// JSON, and at the root no attribute's key exists - they are all inside the
// wrapper. So the second pass must touch the mirror and nothing else.
//
// Before the fix this left state holding only the mirror plus the import ID, and
// the first plan after `terraform import` proposed every writable field changing
// from null - the exact symptom the import work set out to remove.
func TestHydrateFromResponse_ForImport_PopulatesNoRefreshAndMirror(t *testing.T) {
	ctx := context.TODO()
	data := new(FunctionModel)

	if err := hydrateFromResponse([]byte(realEnrichResponse), data, apijson.UnmarshalForImport); err != nil {
		t.Fatalf("hydrateFromResponse: %v", err)
	}

	for name, got := range map[string]string{
		"function_name": data.FunctionName.ValueString(),
		"type":          data.Type.ValueString(),
		"display_name":  data.DisplayName.ValueString(),
	} {
		if got == "" {
			t.Errorf("%s is empty after an import decode. Every writable attribute is "+
				"no_refresh, so a null here is a field the first post-import plan proposes "+
				"changing from null.", name)
		}
	}
	if got := data.Type.ValueString(); got != "enrich" {
		t.Errorf("type = %q, want %q", got, "enrich")
	}
	if data.Tags.IsNull() || data.Tags.IsUnknown() {
		t.Error("tags is empty after an import decode")
	}
	if data.Config.IsNull() {
		t.Error("config is null after an import decode")
	}

	// And the mirror, which is the half that survived when the rest did not.
	if data.Function.IsNull() {
		t.Fatal("the `function` mirror is null after an import decode")
	}
	fn, diags := data.Function.Value(ctx)
	if diags.HasError() {
		t.Fatalf("reading function mirror: %v", diags)
	}
	if fn.FunctionID.IsNull() || fn.VersionNum.IsNull() {
		t.Errorf("function.function_id = %v, function.version_num = %v; both must be readable",
			fn.FunctionID, fn.VersionNum)
	}
}

// Pins the mechanism the fix turns on, so a full second pass cannot be
// reintroduced: at the root, a whole-model decode sees none of the attribute
// keys and therefore nulls them.
func TestUnmarshalForImport_DirectIntoModel_NullsTopLevelFields(t *testing.T) {
	data := FunctionModel{
		FunctionName: types.StringValue("tf-enrich-enricher"),
		Type:         types.StringValue("enrich"),
	}

	if err := apijson.UnmarshalForImport([]byte(realEnrichResponse), &data); err != nil {
		t.Fatalf("UnmarshalForImport: %v", err)
	}

	if !data.Type.IsNull() {
		t.Errorf("type = %v after a direct import decode, expected it nulled. If a root-level "+
			"decode now preserves top-level attributes, re-read hydrateFromResponse - the "+
			"mirror-only second pass may no longer be necessary.", data.Type)
	}
	if data.Function.IsNull() {
		t.Error("the `function` mirror should be the one thing a direct decode does populate")
	}
}

// Pins the bug, so the two-pass cannot be "simplified" back into one. A single
// direct decode cannot see config at all, because the response has no
// root-level config key.
func TestUnmarshalComputed_DirectIntoModel_LeavesConfigUnhydrated(t *testing.T) {
	ctx := context.TODO()
	data := planModel(ctx)

	if err := apijson.UnmarshalComputed([]byte(realEnrichResponse), &data); err != nil {
		t.Fatalf("UnmarshalComputed: %v", err)
	}

	step := firstStep(t, data)
	if !step.TopK.IsNull() {
		t.Fatalf("config.steps[0].top_k = %v from a direct decode. If this now hydrates, the "+
			"response shape or the model changed and the two-pass in envelope.go should be "+
			"re-examined.", step.TopK)
	}
}

// And the counterpart: an envelope-only decode hydrates config but destroys the
// mirror. Documents why the fix needs both passes rather than just swapping in
// the envelope, as was done for bem_workflow.
func TestUnmarshalComputed_EnvelopeOnly_NullsTheFunctionMirror(t *testing.T) {
	ctx := context.TODO()
	env := FunctionFunctionEnvelope{Function: planModel(ctx)}

	if err := apijson.UnmarshalComputed([]byte(realEnrichResponse), &env); err != nil {
		t.Fatalf("UnmarshalComputed: %v", err)
	}

	if step := firstStep(t, env.Function); step.TopK.IsNull() {
		t.Error("envelope decode should hydrate config.steps[0].top_k")
	}
	if !env.Function.Function.IsNull() {
		t.Fatal("the `function` mirror is populated by an envelope-only decode. If the wrapper now " +
			"contains a nested \"function\" key, the second pass may be unnecessary.")
	}
}
