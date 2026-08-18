package workflow

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// noConfiguredAttributeChanged reports whether every attribute a practitioner
// can actually set is identical between plan and state.
//
// Purely-Computed attributes are skipped: they are server-owned, and they are
// exactly the ones Terraform core nulls in the proposed new state (see
// collapseNoOpPlan). Comparing them would always report a change and defeat the
// purpose.
//
// An unknown planned value counts as a change. That matters: it is what keeps a
// genuine update working. When a referenced function's id or version_num is
// unknown because that function is itself changing, the attribute interpolating
// it differs from state, so the plan is left alone.
func noConfiguredAttributeChanged(
	ctx context.Context,
	rschema schema.Schema,
	plan tfsdk.Plan,
	state tfsdk.State,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Blocks are not compared below, so collapsing a plan while any exist would
	// silently discard a block-only edit. Neither schema uses blocks today; if
	// codegen ever emits one, decline rather than guess.
	if len(rschema.Blocks) > 0 {
		tflog.Debug(ctx, "bem-noop-guard: declining, schema has blocks which are not compared", map[string]any{
			"resource": "bem_workflow",
		})
		return false, diags
	}

	for name, attribute := range rschema.Attributes {
		// Purely computed: no configuration input, nothing to compare.
		if attribute.IsComputed() && !attribute.IsOptional() && !attribute.IsRequired() {
			continue
		}

		tfPath := tftypes.NewAttributePath().WithAttributeName(name)

		// Compare the RAW tftypes values, not the framework values.
		//
		// Framework equality routes through the attribute's custom type, and
		// customfield.NestedObject[T].Equal delegates to
		// basetypes.ObjectValue.Equal, which compares attribute *types* as well
		// as values. A live plan showed `config` comparing unequal while both
		// sides rendered byte-identically:
		//
		//	attribute=config equal=false plan_null=false state_null=false
		//	DECLINING: plan="{...}" state="{...}"      <- identical strings
		//
		// Raw comparison is also the more honest question to ask here: the guard
		// wants to know whether the practitioner's input changed, and identical
		// raw values mean it did not, whatever a custom type's Equal decides.
		planValue, planErr := valueAtPath(plan.Raw, tfPath)
		stateValue, stateErr := valueAtPath(state.Raw, tfPath)
		if planErr != nil || stateErr != nil {
			// Cannot compare; be conservative and treat it as changed rather
			// than risk collapsing a real update into a no-op.
			tflog.Debug(ctx, "bem-noop-guard: declining, attribute not readable", map[string]any{
				"resource":  "bem_workflow",
				"attribute": name,
				"plan_err":  fmt.Sprintf("%v", planErr),
				"state_err": fmt.Sprintf("%v", stateErr),
			})
			return false, diags
		}

		// Diff rather than Equal, so numerically-equal Numbers that differ
		// structurally do not count as a change. See numbersEqual.
		rawDiffs, diffErr := planValue.Diff(stateValue)
		if diffErr != nil {
			tflog.Debug(ctx, "bem-noop-guard: declining, cannot diff attribute", map[string]any{
				"resource":  "bem_workflow",
				"attribute": name,
				"error":     diffErr.Error(),
			})
			return false, diags
		}
		realDiffs := meaningfulDiffs(rawDiffs)
		equal := len(realDiffs) == 0

		if !equal {
			// The single most useful line in this whole investigation: it names
			// the attribute that keeps the plan dirty, and prints both sides.
			tflog.Debug(ctx, "bem-noop-guard: DECLINING - attribute differs", map[string]any{
				"resource":  "bem_workflow",
				"attribute": name,
			})
			logValueDiff(ctx, "bem_workflow", name, realDiffs)
			return false, diags
		}
	}

	tflog.Debug(ctx, "bem-noop-guard: collapsing plan to prior state", map[string]any{
		"resource": "bem_workflow",
	})
	return true, diags
}

// collapseNoOpPlan turns a plan that changes nothing a practitioner configured
// back into a genuine no-op, by writing prior state over the plan.
//
// Why this is needed. Terraform core builds the proposed new state before the
// provider sees it, and for a Computed-only attribute with a nested object type
// it does not carry the prior value forward - it proposes null. The framework
// then marks that null as unknown
// (fwserver.MarkComputedNilsAsUnknown, gated on the proposed state already
// differing from prior), and the result is a plan that wants to update the
// resource even though nothing in the configuration moved.
//
// On bem_workflow the exposed attributes are `workflow` and `audit`. Measured from a real plan's JSON, its `after_unknown` contained
// exactly one entry - `.function`, the whole object - and no configured
// attribute differed at all. Every apply then created a new function version,
// and the churn cascaded to every workflow whose nodes interpolate
// function.function_id / function.version_num, so each apply bumped the
// workflow too. BEM-1396 finding 2.
//
// Why not a plan modifier. objectplanmodifier.UseStateForUnknown() fires on
// every plan where the attribute is unknown, including a legitimate update. A
// real display_name change would then plan version_num = N while Update()
// returns N+1, and the framework rejects the apply with "Provider produced
// inconsistent result after apply" - breaking the version-bump propagation
// BEM-1392 fixed. The guard here is what makes this safe: it only collapses the
// plan when nothing configured changed, so a real update is untouched.
func collapseNoOpPlan(
	ctx context.Context,
	rschema schema.Schema,
	plan tfsdk.Plan,
	state tfsdk.State,
	respPlan *tfsdk.Plan,
) diag.Diagnostics {
	unchanged, diags := noConfiguredAttributeChanged(ctx, rschema, plan, state)
	if diags.HasError() || !unchanged {
		return diags
	}

	respPlan.Raw = state.Raw
	return diags
}

// preserveServerDefaults writes prior state back over Optional+Computed
// attributes that the configuration does not set.
//
// This is the behaviour practitioners expect from Optional+Computed and do not
// get for leaves nested inside a collection. Terraform core builds the proposed
// new state by taking the configuration value where one exists; for a leaf the
// HCL omits, that value is null, while prior state holds whatever the server
// filled in. The attribute therefore differs on every plan, forever, and no
// amount of decoding on the provider side changes it - the difference is
// manufactured before the provider is consulted.
//
// Measured on `bem_function`'s enrich `config` block. The server defaults six
// leaves the fixture deliberately leaves unset:
//
//	config.steps[].top_k            -> 1
//	config.steps[].search_mode      -> "semantic"
//	config.steps[].score_threshold  -> 0.6
//	config.endpoints[].match_top_k  -> 1
//	config.endpoints[].max_candidates -> 50
//	config.endpoints[].max_pages    -> 10
//
// Pinning all six in HCL makes the plan clean, which is what identified them,
// but that is not an acceptable end state: they are optional attributes with
// documented server defaults, so omitting them is the normal case and every
// enrich function would hit this. BEM-1396 finding 2.
//
// Deliberately narrow:
//
//   - Optional+Computed only. A required attribute is the practitioner's to
//     set, and a purely-computed one is handled by collapseNoOpPlan.
//   - The configuration value must be null. Once the practitioner sets the
//     attribute, their value wins and this never fires.
//   - Prior state must hold a known, non-null value. Nothing is invented.
//
// Known limitation: paths into a list are index-based, so if a practitioner
// reorders or inserts list elements, an unset leaf can inherit the neighbouring
// element's prior value rather than its own. Both are server defaults for the
// same field, so the practical impact is nil, but it is why this does not try to
// correlate elements by identity.
func preserveServerDefaults(
	ctx context.Context,
	rschema schema.Schema,
	config tfsdk.Config,
	plan tfsdk.Plan,
	state tfsdk.State,
) (tftypes.Value, error) {
	if plan.Raw.IsNull() || state.Raw.IsNull() {
		return plan.Raw, nil
	}

	return tftypes.Transform(plan.Raw, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if len(p.Steps()) < 1 {
			return v, nil
		}
		// Act on null *and* unknown. The framework runs
		// MarkComputedNilsAsUnknown (server_planresourcechange.go:252) before it
		// calls resource-level ModifyPlan (line 347), so by the time this pass
		// runs, an Optional+Computed leaf the configuration omitted is already
		// unknown rather than null - that is what the plan renders as
		// "(known after apply)". Gating on IsNull alone returned early on every
		// live plan and preserved nothing.
		if !v.IsNull() && v.IsKnown() {
			return v, nil
		}

		attribute, err := rschema.AttributeAtTerraformPath(ctx, p)
		if err != nil {
			// Blocks, non-schema paths and values inside atomic attributes have
			// no attribute of their own. Leave them alone.
			return v, nil
		}
		if !attribute.IsComputed() || !attribute.IsOptional() {
			return v, nil
		}

		configValue, err := valueAtPath(config.Raw, p)
		if err != nil || !configValue.IsNull() {
			return v, nil
		}

		stateValue, err := valueAtPath(state.Raw, p)
		// Prior state's null IS the value to preserve, not the absence of one. An
		// Optional+Computed attribute that is null in state and omitted from
		// config - an enrich function's `endpoints` when every step is
		// collection-sourced - must adopt that null, or the guard diffs
		// unknown-vs-null and the plan never settles. This matches the
		// framework's own UseStateForUnknown, which documents that null is a
		// known value and is copied to the planned value.
		//
		// Adopted for primitives only. For objects, lists and maps the bail below
		// is conservative rather than understood: adopting a null there once
		// produced an apply-time "was null, but now cty.ObjectVal{...}" failure.
		// Two plausible causes have since been tested and RULED OUT - the decoder,
		// and the req.Plan.Get / resp.State.Set round trip, both of which preserve
		// a null. See the fuller note in function/noop_plan.go, which these two
		// passes deliberately keep in step.
		if err != nil || !stateValue.IsKnown() {
			return v, nil
		}
		// Prior state's null is itself the value to preserve, for a primitive: the
		// framework marks an Optional+Computed attribute unknown when the
		// configuration omits it, and unknown-vs-null keeps the plan dirty forever.
		// Primitives only - adopting a null for an object, list or map trips the
		// Update-path defect described below. Kept identical to the bem_function
		// copy on purpose; these two passes have diverged before.
		// A null prior state needs different treatment per kind, and getting this
		// wrong has now produced three separate failures on `connectors` alone.
		//
		// The invariant that matters is not "match prior state" - it is **the planned
		// value must equal what Update will return**, because Terraform rejects any
		// post-apply value that differs from a planned value that was not unknown.
		//
		//   - primitives: the API omits an unset scalar, so the decode leaves it null
		//     and Update returns null. Adopting the prior null satisfies the invariant.
		//   - collections: the API returns an **empty container**, not null. Adopting
		//     the prior null made the plan promise null while Update returned [], and
		//     apply failed with "was null, but now cty.ListValEmpty(...)" - after the
		//     change had already been applied, so the resource was correct and the
		//     pipeline was red. Plan the empty collection instead: it equals what
		//     Update returns, and it also equals prior state for a resource that WAS
		//     hydrated, so both shapes agree.
		//   - objects: left unknown. There is no evidence for what an object's "empty"
		//     is here, and inventing one is what produced the original
		//     "was null, but now cty.ObjectVal{steps: null, endpoints: null}".
		//
		// Lifting the object branch is a per-attribute question, and the two attributes
		// that want it have been probed against the API with the attribute omitted:
		//
		//   parse_config  -> {extractEntities: true, linkAcrossDocuments: true}, on a
		//                    BARE GET. Real defaults the model can hold, so adopting
		//                    the server's value is coherent and this looks viable. One
		//                    trap: linkAcrossDocuments defaults to TRUE, so adopt what
		//                    the API returns, never a synthesised zero value.
		//
		//   extra_config  -> 21 populated fields on an extract function, none of them
		//                    the single field FunctionExtraConfigModel represents. The
		//                    decoder matches nothing and yields {enable_bounding_boxes:
		//                    null}: non-null object, null leaf, which is exactly the
		//                    state that regressed ten fixtures.
		//
		// So extra_config is NOT a null-handling problem and must not be fixed with a
		// rule here - it is a model/API surface mismatch (1 field of 21 on extract, 0
		// of 0 on parse), and on extract its leaf shadows the top-level
		// enable_bounding_boxes attribute, whose value the API surfaces at the top
		// level instead. Making it Optional+Computed would have the provider adopt as
		// prior state an object it structurally cannot represent. The prior question is
		// what the attribute is for: read-only mirror, narrowed, or removed.
		//
		// This is why `connectors` needed three attempts: Optional alone broke the
		// imported case, Computed alone broke the created case, and Computed plus
		// null-adoption broke apply. Only the value the API actually returns satisfies
		// plan and apply at once.
		if stateValue.IsNull() {
			empty, isContainer := emptyContainer(stateValue.Type())
			switch {
			case isContainer:
				stateValue = empty
			case !isPrimitiveType(stateValue.Type()):
				// Objects: leave the plan unknown, per the analysis above. Unknown
				// accepts whatever Update returns, which is the only safe promise
				// when we do not know what the API sends for an unset value.
				return v, nil
			}
		}
		if !stateValue.Type().Equal(v.Type()) {
			return v, nil
		}
		// Nothing to do when they already agree. This is not just an
		// optimisation: returning a different-but-equal value makes
		// tftypes.Transform treat the parent as modified and rebuild it, which
		// turns a null parent object into a known object full of nulls. That
		// surfaced as "Provider produced inconsistent result after apply:
		// .config: was null, but now cty.ObjectVal{...}" on a function type that
		// configures no config block at all.
		if v.Equal(stateValue) {
			return v, nil
		}
		if hasNullAncestor(plan.Raw, p) {
			return v, nil
		}

		// Prior state's null IS the value to preserve, not the absence of one.
		// An earlier version bailed on a null state under the heading "nothing is
		// invented", which conflated the two and left the original bug unfixed
		// for a shape that hits it: an enrich function with collection-only steps
		// declares no `endpoints`, so the attribute is null in state and unknown
		// in the plan, the guard diffs unknown-vs-null, declines, and the plan
		// stays dirty forever. This matches the framework's own
		// UseStateForUnknown, whose documentation is explicit that null is a
		// known value and is copied to the planned value.
		return stateValue, nil
	})
}

func valueAtPath(root tftypes.Value, p *tftypes.AttributePath) (tftypes.Value, error) {
	raw, _, err := tftypes.WalkAttributePath(root, p)
	if err != nil {
		return tftypes.Value{}, err
	}
	value, ok := raw.(tftypes.Value)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("unexpected type %T at %s", raw, p)
	}
	return value, nil
}

// logValueDiff names the exact leaf that differs, using tftypes' own diff rather
// than leaving anyone to eyeball two long renderings.
//
// This exists because a live plan showed `config` comparing unequal while plan
// and state rendered byte-identically at both the framework and raw levels - so
// the difference is in something the renderer does not show. Value.Diff walks
// both sides and reports the differing paths, which is the only way to see it.
func logValueDiff(ctx context.Context, resource, attribute string, diffs []tftypes.ValueDiff) {
	for i, d := range diffs {
		entry := map[string]any{
			"resource":  resource,
			"attribute": attribute,
			"index":     i,
			"path":      d.Path.String(),
		}
		if d.Value1 != nil {
			entry["plan"] = d.Value1.String()
			entry["plan_type"] = d.Value1.Type().String()
			entry["plan_exact"] = numberDetail(d.Value1)
		}
		if d.Value2 != nil {
			entry["state"] = d.Value2.String()
			entry["state_type"] = d.Value2.Type().String()
			entry["state_exact"] = numberDetail(d.Value2)
		}
		tflog.Debug(ctx, "bem-noop-guard: DIFF AT PATH", entry)
	}
}

// numbersEqual reports whether two tftypes values are Numbers holding the same
// numeric value.
//
// tftypes.Value.Equal is structural. Number holds a *big.Float, and two
// big.Floats that are numerically identical are not necessarily structurally
// identical, so a value that has round-tripped through the protocol or a
// framework type can compare unequal to one that has not. The guard only wants
// to know whether the practitioner's input changed, so numeric comparison is the
// correct question for Numbers.
//
// Note what this deliberately does NOT do: it does not tolerate a genuine
// numeric difference. 0.6 and 0.5999999999999999778 are not equal here, and a
// diff between them still declines the collapse.
func numbersEqual(a, b *tftypes.Value) bool {
	if a == nil || b == nil {
		return false
	}
	if !a.Type().Is(tftypes.Number) || !b.Type().Is(tftypes.Number) {
		return false
	}
	if a.IsNull() != b.IsNull() || a.IsKnown() != b.IsKnown() {
		return false
	}
	if a.IsNull() || !a.IsKnown() {
		return true
	}

	af, bf := new(big.Float), new(big.Float)
	if err := a.As(&af); err != nil {
		return false
	}
	if err := b.As(&bf); err != nil {
		return false
	}
	return af.Cmp(bf) == 0
}

// meaningfulDiffs returns the diffs that represent a real change, discarding
// Number entries that are numerically equal (see numbersEqual).
func meaningfulDiffs(diffs []tftypes.ValueDiff) []tftypes.ValueDiff {
	out := make([]tftypes.ValueDiff, 0, len(diffs))
	for _, d := range diffs {
		if numbersEqual(d.Value1, d.Value2) {
			continue
		}
		if emptyEqualsNullCollection(d.Value1, d.Value2) {
			continue
		}
		out = append(out, d)
	}
	return out
}

// emptyEqualsNullCollection reports whether the two values are a null collection
// and an empty one of the same type - which this API does not distinguish.
//
// Needed because preserveServerDefaults plans the empty container for a
// collection whose prior state is null: it has to, because that is what Update
// returns and a planned value must match. Without this rule the guard then sees
// plan [] against state null, declines, and a genuinely empty plan renders dirty -
// which is the failure this whole sequence started from.
//
// Safe in both directions. When there IS something to apply, the guard declines
// for the real reason and the plan keeps the empty container, so apply stays
// consistent. When there is not, the guard collapses to prior state, no Update is
// called, and the null in state is never contradicted.
//
// A practitioner cannot express the difference either: removing a list from
// configuration and setting it to [] both reach the API as "no entries".
func emptyEqualsNullCollection(a, b *tftypes.Value) bool {
	if a == nil || b == nil {
		return false
	}
	if !a.Type().Equal(b.Type()) {
		return false
	}
	if _, isContainer := emptyContainer(a.Type()); !isContainer {
		return false
	}
	if !a.IsKnown() || !b.IsKnown() {
		return false
	}

	empty := func(v *tftypes.Value) (isEmpty bool, isNull bool) {
		if v.IsNull() {
			return false, true
		}
		var elements []tftypes.Value
		if err := v.As(&elements); err == nil {
			return len(elements) == 0, false
		}
		var entries map[string]tftypes.Value
		if err := v.As(&entries); err == nil {
			return len(entries) == 0, false
		}
		return false, false
	}

	aEmpty, aNull := empty(a)
	bEmpty, bNull := empty(b)
	return (aNull && bEmpty) || (bNull && aEmpty)
}

// numberDetail renders a Number at full precision. Value.String() formats via
// big.Float's 10-significant-digit default, so 0.6 and 0.5999999999999999778
// both print as "0.6" - which is exactly what made a real difference look like a
// comparison artefact for two rounds of debugging.
func numberDetail(v *tftypes.Value) string {
	if v == nil {
		return "<nil>"
	}
	if !v.Type().Is(tftypes.Number) || v.IsNull() || !v.IsKnown() {
		return v.String()
	}
	f := new(big.Float)
	if err := v.As(&f); err != nil {
		return v.String() + " (undecodable: " + err.Error() + ")"
	}
	return f.Text('g', 50)
}

// normalizeNumberRepresentation rewrites plan Numbers to prior state's
// representation whenever the two agree at float64 precision.
//
// This is the root cause of BEM-1396 finding 2. The framework quantises a
// Float64 attribute's plan value to float64, while prior state read from the
// state file carries the exact decimal the API sent. Captured from a live plan
// for config.steps[0].score_threshold:
//
//	plan  0.59999999999999997779553950749686919152736663818359   (float64(0.6))
//	state 0.6                                                    (exact decimal)
//
// Those are different numbers, so the planned state Terraform gets back differs
// from prior state, and core proposes an in-place update on every plan - which
// then cascades to every workflow interpolating the function's id/version_num.
//
// float64 is the correct granularity to normalise at: it is the resolution the
// schema can express. score_threshold is the only Float64Attribute in this
// provider and there are no NumberAttributes at all, so nothing here relies on
// precision finer than float64. That is also why enrich was the only resource
// affected - every other numeric attribute is Int64, and integers round-trip
// through float64 exactly.
//
// A genuine numeric change is untouched: the two must be equal as float64 before
// state's representation is adopted, so lowering score_threshold from 0.6 to 0.5
// still plans and applies.
func normalizeNumberRepresentation(ctx context.Context, plan, state tftypes.Value) (tftypes.Value, error) {
	if plan.IsNull() || state.IsNull() {
		return plan, nil
	}

	return tftypes.Transform(plan, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if len(p.Steps()) < 1 {
			return v, nil
		}
		if !v.Type().Is(tftypes.Number) || v.IsNull() || !v.IsKnown() {
			return v, nil
		}

		stateValue, err := valueAtPath(state, p)
		if err != nil || stateValue.IsNull() || !stateValue.IsKnown() || !stateValue.Type().Is(tftypes.Number) {
			return v, nil
		}
		if v.Equal(stateValue) {
			return v, nil
		}

		planFloat, stateFloat := new(big.Float), new(big.Float)
		if err := v.As(&planFloat); err != nil {
			return v, nil
		}
		if err := stateValue.As(&stateFloat); err != nil {
			return v, nil
		}

		planF64, _ := planFloat.Float64()
		stateF64, _ := stateFloat.Float64()
		if planF64 != stateF64 {
			// A real change. Leave it alone.
			return v, nil
		}

		tflog.Debug(ctx, "bem-noop-guard: normalising number representation", map[string]any{
			"path":  p.String(),
			"plan":  planFloat.Text('g', 50),
			"state": stateFloat.Text('g', 50),
		})
		return stateValue, nil
	})
}

// hasNullAncestor reports whether any parent of p is null in root.
//
// Substituting a value underneath a null parent is what materialises that
// parent: tftypes.Transform still visits child paths beneath a null object, and
// any substitution there is a real change, so Transform rebuilds the parent as a
// KNOWN object full of nulls. That fails the plan/apply consistency check with
// ".config: was null, but now cty.ObjectVal{...}".
func hasNullAncestor(root tftypes.Value, p *tftypes.AttributePath) bool {
	steps := p.Steps()
	for i := 1; i < len(steps); i++ {
		ancestor, err := valueAtPath(root, tftypes.NewAttributePathWithSteps(steps[:i]))
		if err != nil {
			// Cannot tell; be conservative and leave the value alone.
			return true
		}
		if ancestor.IsNull() {
			return true
		}
	}
	return false
}

// isPrimitiveType reports whether t is a scalar - a type whose null value cannot
// be "materialised" into anything. See preserveServerDefaults.
func isPrimitiveType(t tftypes.Type) bool {
	return t.Is(tftypes.String) || t.Is(tftypes.Bool) || t.Is(tftypes.Number)
}

// emptyContainer returns the empty value for a collection type, and false for
// anything else.
//
// Used for a null prior state: the API returns an empty container rather than a
// null for the collections it assigns, so the plan has to promise the same thing
// or apply fails the consistency check. See preserveServerDefaults.
func emptyContainer(t tftypes.Type) (tftypes.Value, bool) {
	switch {
	case t.Is(tftypes.List{}), t.Is(tftypes.Set{}):
		// Tuples are deliberately excluded: tftypes.NewValue panics on a tuple type
		// that has element types, which would crash the provider inside ModifyPlan.
		return tftypes.NewValue(t, []tftypes.Value{}), true
	case t.Is(tftypes.Map{}), t.Is(tftypes.Object{}):
		// Objects are deliberately excluded - see preserveServerDefaults. Maps are
		// keyed containers with a genuine empty form, so they behave like lists.
		if t.Is(tftypes.Map{}) {
			return tftypes.NewValue(t, map[string]tftypes.Value{}), true
		}
		return tftypes.Value{}, false
	}
	return tftypes.Value{}, false
}
