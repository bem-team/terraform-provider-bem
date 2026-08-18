package function

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
)

// BEM-1396 finding 1. An `enrich` function's `config` is validated against
// itself on every PATCH, whether or not the request carries it - the API cannot
// distinguish an omitted config from an empty one, so
//
//   - config absent entirely => zero value => "enrich function must have at
//     least one step". This is what a display_name-only edit produced, making
//     an enrich function effectively immutable once created.
//   - config present but missing endpoints => every endpoint-sourced step is
//     validated against an empty endpoint set => "step N references unknown
//     endpoint". This is the error the fixture actually hit.
//
// There is a third, quieter case that decided the shape of the fix: the API
// stores whatever config the request carried with no "was it provided" check,
// and validation of a *collection*-sourced step never references endpoint names
// at all. So a collection-sourced enrich
// function whose steps change would pass validation with endpoints omitted and
// have its stored endpoints silently replaced by nil - data loss, not a 400.
//
// `full_replace` on Config is what covers all three: whenever config has a
// value, the whole block goes out.

func enrichFixture() FunctionModel {
	ctx := context.TODO()
	return FunctionModel{
		FunctionName: types.StringValue("tf-enrich-enricher"),
		Type:         types.StringValue("enrich"),
		DisplayName:  types.StringValue("TF Enrich"),
		Config: customfield.NewObjectMust(ctx, &FunctionConfigModel{
			Steps: &[]*FunctionConfigStepsModel{
				{
					SourceField:  types.StringValue("lineItems[*].description"),
					TargetField:  types.StringValue("lineItems[*].enrichedProducts"),
					EndpointName: types.StringValue("catalog"),
					Source:       types.StringValue("endpoint"),
				},
			},
			Endpoints: customfield.NewObjectListMust(ctx, []FunctionConfigEndpointsModel{
				{
					Name:         types.StringValue("catalog"),
					Method:       types.StringValue("POST"),
					URL:          types.StringValue("https://sink.example.com/enrich"),
					BodyTemplate: types.StringValue(`{"query": "{value}", "limit": 5}`),
					Headers:      jsontypes.NewNormalizedValue(`{"X-Bem-Fixture":"tf-enrich"}`),
				},
			}),
		}),
	}
}

// assertCompleteEnrichConfig checks the body carries a config block with both
// co-validated halves populated - the invariant the server enforces.
func assertCompleteEnrichConfig(t *testing.T, body []byte) {
	t.Helper()

	var decoded struct {
		Config *struct {
			Steps []struct {
				SourceField  string `json:"sourceField"`
				EndpointName string `json:"endpointName"`
			} `json:"steps"`
			Endpoints []struct {
				Name string `json:"name"`
			} `json:"endpoints"`
		} `json:"config"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("body is not valid JSON: %v (body: %s)", err, body)
	}

	if decoded.Config == nil {
		t.Fatalf("config absent from the PATCH body - the API rejects this with "+
			"\"enrich function must have at least one step\"\nbody: %s", body)
	}
	if len(decoded.Config.Steps) == 0 {
		t.Errorf("config.steps is empty - the API requires at least one\nbody: %s", body)
	}
	if len(decoded.Config.Endpoints) == 0 {
		t.Fatalf("config.endpoints absent or empty while steps were sent. Endpoint-sourced steps "+
			"are validated against this set, so the API rejects it with \"step N references unknown "+
			"endpoint\"; a collection-sourced function would instead have its endpoints silently "+
			"wiped.\nbody: %s", body)
	}

	// Names must line up, otherwise the block is present but internally
	// inconsistent - which fails the same server-side check.
	endpointNames := map[string]bool{}
	for _, ep := range decoded.Config.Endpoints {
		endpointNames[ep.Name] = true
	}
	for i, step := range decoded.Config.Steps {
		if step.EndpointName != "" && !endpointNames[step.EndpointName] {
			t.Errorf("step %d references endpoint %q, which is not in the body's endpoint set %v",
				i, step.EndpointName, endpointNames)
		}
	}
}

// Tripwire, matching the ones for atomic_group on the DAG trio and the
// output-schema pair. model.go is Stainless-generated and this tag is
// hand-added, so a regen silently drops it; apijson's tag parser also ignores
// any option it doesn't recognise, so a typo turns the fix off with nothing
// failing. The behavioural tests below cannot catch either, because without the
// tag they would just be exercising a correctly-working plain patch encoder.
func TestFunctionModel_ConfigCarriesFullReplaceTag(t *testing.T) {
	const wantOption = "full_replace"

	field, ok := reflect.TypeOf(FunctionModel{}).FieldByName("Config")
	if !ok {
		t.Fatal("FunctionModel has no Config field - if it was renamed, the full_replace tag must move with it")
	}

	for _, part := range strings.Split(field.Tag.Get("json"), ",") {
		if part == wantOption {
			return
		}
	}
	t.Errorf(
		"FunctionModel.Config is missing the %q json tag option (got tag: %q).\n"+
			"An enrich function's config is validated against itself on every update, so a patch "+
			"that omits it fails with \"enrich function must have at least one step\" and a patch "+
			"that sends it partially fails with \"step N references unknown endpoint\". If this "+
			"failed right after a Stainless regen, re-apply the tag.",
		wantOption, field.Tag.Get("json"),
	)
}

// The case the original brief did not cover, and the most common edit there
// is. Pre-fix this produced {"displayName":"...","functionName":"..."} with no
// config at all, which the API rejects outright - so renaming an enrich
// function was impossible.
func TestMarshalJSONForUpdate_EnrichDisplayNameOnlyChange_StillSendsConfig(t *testing.T) {
	state := enrichFixture()
	plan := state
	plan.DisplayName = types.StringValue("TF Enrich (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	if !strings.Contains(string(got), `"displayName":"TF Enrich (renamed)"`) {
		t.Errorf("the field that changed is missing\nbody: %s", got)
	}
	assertCompleteEnrichConfig(t, got)
}

// Finding 1 as reported: a change inside config.steps must carry
// config.endpoints along with it.
func TestMarshalJSONForUpdate_EnrichStepChange_SendsEndpointsToo(t *testing.T) {
	ctx := context.TODO()
	state := enrichFixture()
	plan := state
	plan.Config = customfield.NewObjectMust(ctx, &FunctionConfigModel{
		Steps: &[]*FunctionConfigStepsModel{
			{
				SourceField:  types.StringValue("lineItems[*].description"),
				TargetField:  types.StringValue("lineItems[*].enrichedThings"), // changed
				EndpointName: types.StringValue("catalog"),
				Source:       types.StringValue("endpoint"),
			},
		},
		Endpoints: customfield.NewObjectListMust(ctx, []FunctionConfigEndpointsModel{
			{
				Name:         types.StringValue("catalog"),
				Method:       types.StringValue("POST"),
				URL:          types.StringValue("https://sink.example.com/enrich"),
				BodyTemplate: types.StringValue(`{"query": "{value}", "limit": 5}`),
				Headers:      jsontypes.NewNormalizedValue(`{"X-Bem-Fixture":"tf-enrich"}`),
			},
		}),
	})

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	if !strings.Contains(string(got), `"targetField":"lineItems[*].enrichedThings"`) {
		t.Errorf("the changed step field is missing\nbody: %s", got)
	}
	assertCompleteEnrichConfig(t, got)
}

// The silent-data-loss path, and the reason this is fixed at the config level
// rather than only inside it. A collection-sourced step never consults the
// endpoint set server-side, so a body with steps and no endpoints passes
// validation and then overwrites the stored config - dropping endpoints with no
// error at all. Terraform would report success.
func TestMarshalJSONForUpdate_EnrichCollectionSourcedStepChange_StillSendsEndpoints(t *testing.T) {
	ctx := context.TODO()

	buildConfig := func(targetField string) customfield.NestedObject[FunctionConfigModel] {
		return customfield.NewObjectMust(ctx, &FunctionConfigModel{
			Steps: &[]*FunctionConfigStepsModel{
				{
					SourceField:    types.StringValue("orderNumber"),
					TargetField:    types.StringValue(targetField),
					CollectionName: types.StringValue("orders"),
					Source:         types.StringValue("collection"),
				},
			},
			Endpoints: customfield.NewObjectListMust(ctx, []FunctionConfigEndpointsModel{
				{
					Name:   types.StringValue("catalog"),
					Method: types.StringValue("POST"),
					URL:    types.StringValue("https://sink.example.com/enrich"),
				},
			}),
		})
	}

	state := enrichFixture()
	state.Config = buildConfig("enrichedOrder")
	plan := state
	plan.Config = buildConfig("enrichedOrderV2")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	assertCompleteEnrichConfig(t, got)
}

// full_replace must force completeness, not presence. Every other function
// type leaves config null, and sending an empty config block for them would
// turn working updates into failures.
func TestMarshalJSONForUpdate_NonEnrichFunction_SendsNoConfigBlock(t *testing.T) {
	state := FunctionModel{
		FunctionName: types.StringValue("tf-split"),
		Type:         types.StringValue("split"),
		DisplayName:  types.StringValue("TF Split"),
		SplitType:    types.StringValue("semantic_page"),
	}
	plan := state
	plan.DisplayName = types.StringValue("TF Split (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("body is not a JSON object: %v", err)
	}
	if raw, ok := decoded["config"]; ok {
		t.Errorf("config = %s was sent for a split function that never configured one", raw)
	}
	if _, ok := decoded["displayName"]; !ok {
		t.Errorf("displayName missing from the body\nbody: %s", got)
	}
}
