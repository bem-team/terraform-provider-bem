package function

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// BEM-1396 finding 2. `config.steps[].top_k` is Optional+Computed and the
// server defaults it to 1. When the HCL omits it, Terraform core's proposed new
// state takes null from the configuration while prior state holds 1, so
// `config` differs on every plan forever - and because `config` is
// practitioner-settable, collapseNoOpPlan correctly declines to collapse the
// plan. Pinning all six leaves in HCL made the plan clean, which is how they
// were identified, but omitting an optional attribute with a documented server
// default is the normal case.
//
// These tests build the three tftypes values the framework hands ModifyPlan and
// drive the transform directly, which is the only way to exercise this without
// a live plan.

// buildFunctionValues returns config/plan/state raw values for an enrich
// function's single step, in the shape the framework actually hands
// ModifyPlan.
//
// The plan value is the subtle part. For an Optional+Computed leaf the
// configuration omits, core proposes null and then
// MarkComputedNilsAsUnknown converts it to *unknown* - before resource-level
// ModifyPlan runs (server_planresourcechange.go:252 vs 347). So the plan holds
// unknown, not null, which is what the plan renders as "(known after apply)".
//
// Building it as null was a real bug in an earlier version of these tests: it
// exercised a path the live plan never takes, so all four passed while the live
// plan stayed dirty.
//
// configTopK nil means the HCL omits it (plan = unknown); non-nil means the HCL
// sets it (plan = that value). stateTopK is what prior state holds.
func buildFunctionValues(t *testing.T, configTopK, stateTopK *float64) (tfsdk.Config, tfsdk.Plan, tfsdk.State) {
	t.Helper()

	ctx := context.TODO()
	rschema := ResourceSchema(ctx)
	objectType, ok := rschema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resource schema type = %T, want tftypes.Object", rschema.Type().TerraformType(ctx))
	}

	stepType, ok := objectType.AttributeTypes["config"].(tftypes.Object).
		AttributeTypes["steps"].(tftypes.List)
	if !ok {
		t.Fatal("could not resolve the config.steps list type from the schema")
	}
	stepObject, ok := stepType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatalf("config.steps element type = %T, want tftypes.Object", stepType.ElementType)
	}
	configObject := objectType.AttributeTypes["config"].(tftypes.Object)

	// A step with everything null except the two fields the fixture sets, plus
	// top_k which varies between config/plan and state.
	step := func(topK *float64, unknownTopK bool) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range stepObject.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["source_field"] = tftypes.NewValue(tftypes.String, "lineItems[*].description")
		attrs["target_field"] = tftypes.NewValue(tftypes.String, "lineItems[*].enrichedProducts")
		switch {
		case unknownTopK:
			attrs["top_k"] = tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)
		case topK != nil:
			attrs["top_k"] = tftypes.NewValue(tftypes.Number, *topK)
		}
		return tftypes.NewValue(stepObject, attrs)
	}

	config := func(topK *float64, unknownTopK bool) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range configObject.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["steps"] = tftypes.NewValue(stepType, []tftypes.Value{step(topK, unknownTopK)})
		return tftypes.NewValue(configObject, attrs)
	}

	resource := func(cfg tftypes.Value) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range objectType.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["function_name"] = tftypes.NewValue(tftypes.String, "tf-enrich-enricher")
		attrs["type"] = tftypes.NewValue(tftypes.String, "enrich")
		attrs["config"] = cfg
		return tftypes.NewValue(objectType, attrs)
	}

	configValue := resource(config(configTopK, false))
	// Omitted in config => unknown in the plan by the time ModifyPlan runs.
	planValue := resource(config(configTopK, configTopK == nil))
	stateValue := resource(config(stateTopK, false))

	return tfsdk.Config{Raw: configValue, Schema: rschema},
		tfsdk.Plan{Raw: planValue, Schema: rschema},
		tfsdk.State{Raw: stateValue, Schema: rschema}
}

func topKFromPlan(t *testing.T, plan tftypes.Value) tftypes.Value {
	t.Helper()

	p := tftypes.NewAttributePath().WithAttributeName("config").
		WithAttributeName("steps").WithElementKeyInt(0).WithAttributeName("top_k")
	got, err := valueAtPath(plan, p)
	if err != nil {
		t.Fatalf("walking to config.steps[0].top_k: %v", err)
	}
	return got
}

// The fix: an omitted Optional+Computed leaf inherits the value prior state
// already holds, so the plan stops differing from state.
func TestPreserveServerDefaults_RestoresOmittedOptionalComputedLeaf(t *testing.T) {
	ctx := context.TODO()
	one := 1.0
	config, plan, state := buildFunctionValues(t, nil, &one)

	// The framework has already turned the omitted leaf into unknown; that is
	// the state this pass has to cope with.
	if got := topKFromPlan(t, plan.Raw); got.IsKnown() {
		t.Fatalf("precondition failed: plan's top_k = %v, expected unknown", got)
	}

	preserved, err := preserveServerDefaults(ctx, ResourceSchema(ctx), config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}

	got := topKFromPlan(t, preserved)
	if !got.IsKnown() {
		t.Fatal("config.steps[0].top_k is still unknown; the plan will keep differing from state forever")
	}
	number := new(big.Float)
	if err := got.As(&number); err != nil {
		t.Fatalf("decoding top_k: %v", err)
	}
	if number.Cmp(big.NewFloat(1)) != 0 {
		t.Errorf("config.steps[0].top_k = %v, want 1 (the value prior state holds)", number)
	}
}

// Prior state's null is itself the value to preserve, for a primitive.
//
// An Optional+Computed scalar that is null in state and omitted from config has
// to adopt that null: the framework marks it unknown, the guard then diffs
// unknown-vs-null, declines, and the plan never settles. This is not
// hypothetical - it is what made an enrich function churn forever once
// pre_count / tabular_chunking_enabled / enable_bounding_boxes became
// Optional+Computed, because the API returns those for an extract function and
// omits them for every other type.
//
// Restricted to primitives on purpose; see the object case below.
func TestPreserveServerDefaults_AdoptsNullStateForPrimitives(t *testing.T) {
	ctx := context.TODO()
	config, plan, state := buildFunctionValues(t, nil, nil)

	preserved, err := preserveServerDefaults(ctx, ResourceSchema(ctx), config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}

	got := topKFromPlan(t, preserved)
	if !got.IsKnown() {
		t.Fatal("config.steps[0].top_k is still unknown against a null prior state; the guard " +
			"will diff unknown-vs-null, decline, and the plan will never settle")
	}
	if !got.IsNull() {
		t.Errorf("config.steps[0].top_k = %v, want null - prior state's null is the value to "+
			"preserve, not a licence to invent one", got)
	}
}

// The practitioner's value always wins. Once the configuration sets the
// attribute, this pass must not touch it - otherwise setting top_k = 5 against a
// state holding 1 would be silently reverted.
func TestPreserveServerDefaults_ConfiguredValueWins(t *testing.T) {
	ctx := context.TODO()
	one := 1.0
	five := 5.0

	// Configuration explicitly sets 5 while prior state still holds the server's 1.
	config, plan, state := buildFunctionValues(t, &five, &one)

	preserved, err := preserveServerDefaults(ctx, ResourceSchema(ctx), config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}

	got := topKFromPlan(t, preserved)
	number := new(big.Float)
	if err := got.As(&number); err != nil {
		t.Fatalf("decoding top_k: %v", err)
	}
	if number.Cmp(big.NewFloat(5)) != 0 {
		t.Errorf("config.steps[0].top_k = %v, want the configured 5", number)
	}
}

// Required attributes are the practitioner's alone, and purely-computed ones are
// collapseNoOpPlan's business. Neither may be rewritten by this pass: doing so
// to a Required attribute would mask a genuine change.
func TestPreserveServerDefaults_IgnoresNonOptionalComputedAttributes(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)

	for name, attribute := range rschema.Attributes {
		if attribute.IsRequired() {
			// `function_name` and `type` are Required; confirm the guard the
			// pass relies on actually distinguishes them.
			if attribute.IsComputed() && attribute.IsOptional() {
				t.Errorf("attribute %q is Required yet reports Optional+Computed; the guard would rewrite it", name)
			}
		}
	}

	one := 1.0
	config, plan, state := buildFunctionValues(t, nil, &one)

	preserved, err := preserveServerDefaults(ctx, rschema, config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}

	// function_name is Required and set identically in plan and state; it must
	// be untouched.
	p := tftypes.NewAttributePath().WithAttributeName("function_name")
	got, err := valueAtPath(preserved, p)
	if err != nil {
		t.Fatalf("walking to function_name: %v", err)
	}
	var name string
	if err := got.As(&name); err != nil {
		t.Fatalf("decoding function_name: %v", err)
	}
	if name != "tf-enrich-enricher" {
		t.Errorf("function_name = %q, want it untouched", name)
	}
}

// The integration these tests kept missing: run both passes in the order
// ModifyPlan runs them, then ask whether the guard fires.
//
// Each pass was verified in isolation and each worked, while the live plan
// stayed dirty three builds running. The gap was always between them - here,
// that preserveServerDefaults rebuilds `config`'s value via tftypes.Transform
// and the resulting framework value compares unequal to state's under
// customfield.NestedObject[T].Equal (which delegates to
// basetypes.ObjectValue.Equal, comparing attribute *types* as well as values)
// even though both render byte-identically. Confirmed from a live plan:
//
//	attribute=config equal=false plan_null=false state_null=false
//	DECLINING - attribute differs: plan="{...}" state="{...}"   <- identical strings
func TestModifyPlanPasses_CollapseAfterPreserve(t *testing.T) {
	ctx := context.TODO()
	one := 1.0

	// Config omits top_k, so the plan carries unknown; prior state holds 1.
	config, plan, state := buildFunctionValues(t, nil, &one)

	preserved, err := preserveServerDefaults(ctx, ResourceSchema(ctx), config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}
	plan.Raw = preserved

	unchanged, diags := noConfiguredAttributeChanged(ctx, ResourceSchema(ctx), plan, state)
	if diags.HasError() {
		t.Fatalf("noConfiguredAttributeChanged: %v", diags)
	}
	if !unchanged {
		t.Fatal("the guard declined after preserveServerDefaults restored the omitted leaf. " +
			"Nothing a practitioner set differs, so the plan should collapse - this is the " +
			"gap that kept bem-pipeline-enrich dirty across three builds.")
	}
}

// The root cause of BEM-1396 finding 2, reproduced.
//
// Terraform's plan value for a Float64 attribute is quantised to float64 by the
// framework, while prior state read from the state file carries the exact
// decimal the API sent. So `score_threshold` arrives as
//
//	plan  0.59999999999999997779553950749686919152736663818359   (float64(0.6))
//	state 0.6                                                    (exact decimal)
//
// captured verbatim from a live plan. Those are genuinely different numbers, so
// Terraform sees `config` as changed and proposes an update on every plan -
// forever. score_threshold is the ONLY Float64Attribute in this provider, and
// there are no arbitrary-precision NumberAttributes at all, which is exactly why
// enrich is the only resource that churns: every other numeric attribute is
// Int64, and integers round-trip through float64 exactly.
func TestNormalizeNumberRepresentation_AdoptsStateForFloat64EqualValues(t *testing.T) {
	ctx := context.TODO()

	planFloat := big.NewFloat(0.6)                                          // float64-derived
	stateFloat, _, err := big.ParseFloat("0.6", 10, 512, big.ToNearestEven) // exact decimal
	if err != nil {
		t.Fatalf("ParseFloat: %v", err)
	}

	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"score_threshold": tftypes.Number}}
	planRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"score_threshold": tftypes.NewValue(tftypes.Number, planFloat),
	})
	stateRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"score_threshold": tftypes.NewValue(tftypes.Number, stateFloat),
	})

	// Precondition: these really are different values, so the test is testing
	// something. If this ever stops holding, the bug is gone upstream.
	if planRaw.Equal(stateRaw) {
		t.Skip("float64(0.6) and decimal 0.6 now compare equal; the upstream behaviour changed")
	}

	normalized, err := normalizeNumberRepresentation(ctx, planRaw, stateRaw)
	if err != nil {
		t.Fatalf("normalizeNumberRepresentation: %v", err)
	}

	if !normalized.Equal(stateRaw) {
		t.Fatalf("plan still differs from state after normalisation.\n got: %s\nwant: %s",
			normalized.String(), stateRaw.String())
	}
}

// A real numeric change must survive untouched - otherwise this pass would
// silently revert a practitioner lowering score_threshold from 0.6 to 0.5.
func TestNormalizeNumberRepresentation_PreservesGenuineNumericChange(t *testing.T) {
	ctx := context.TODO()

	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"score_threshold": tftypes.Number}}
	planRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"score_threshold": tftypes.NewValue(tftypes.Number, big.NewFloat(0.5)),
	})
	stateRaw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"score_threshold": tftypes.NewValue(tftypes.Number, big.NewFloat(0.6)),
	})

	normalized, err := normalizeNumberRepresentation(ctx, planRaw, stateRaw)
	if err != nil {
		t.Fatalf("normalizeNumberRepresentation: %v", err)
	}

	got, err := valueAtPath(normalized, tftypes.NewAttributePath().WithAttributeName("score_threshold"))
	if err != nil {
		t.Fatalf("walking to score_threshold: %v", err)
	}
	f := new(big.Float)
	if err := got.As(&f); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if f.Cmp(big.NewFloat(0.5)) != 0 {
		t.Errorf("score_threshold = %v, want the configured 0.5 preserved", f)
	}
}

// The reason adopting a null prior state is safe only alongside
// hasNullAncestor. tftypes.Transform visits child paths beneath a null object,
// and substituting there is a real change, so Transform rebuilds the parent as a
// KNOWN object of nulls. That failed the plan/apply consistency check with
// ".config: was null, but now cty.ObjectVal{endpoints: null, steps: null}",
// caught by TestAccFunctionResource_ImportThenUpdate against a function type
// that configures no config block at all.
func TestPreserveServerDefaults_LeavesNullParentNull(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)

	objectType, ok := rschema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resource schema type = %T", rschema.Type().TerraformType(ctx))
	}

	// A split-shaped resource: config null everywhere, as it is for every
	// function type that has no config block.
	build := func(configUnknown bool) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range objectType.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["function_name"] = tftypes.NewValue(tftypes.String, "tf-split")
		attrs["type"] = tftypes.NewValue(tftypes.String, "split")
		if configUnknown {
			attrs["config"] = tftypes.NewValue(objectType.AttributeTypes["config"], tftypes.UnknownValue)
		}
		return tftypes.NewValue(objectType, attrs)
	}

	config := tfsdk.Config{Raw: build(false), Schema: rschema}
	plan := tfsdk.Plan{Raw: build(true), Schema: rschema}
	state := tfsdk.State{Raw: build(false), Schema: rschema}

	preserved, err := preserveServerDefaults(ctx, rschema, config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}

	got, err := valueAtPath(preserved, tftypes.NewAttributePath().WithAttributeName("config"))
	if err != nil {
		t.Fatalf("walking to config: %v", err)
	}
	// Either null or still unknown is fine - both are un-materialised. What must
	// never happen is a KNOWN object whose members are null, which is what fails
	// the plan/apply consistency check on every function type with no config
	// block.
	if got.IsKnown() && !got.IsNull() {
		t.Errorf("config = %s, want it left null or unknown rather than materialised into a "+
			"known object of nulls.", got.String())
	}
}
