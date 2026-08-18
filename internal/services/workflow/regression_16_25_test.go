package workflow

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// `connectors` has to satisfy two shapes that pull in opposite directions, and
// getting one of them wrong is what produced a regression on two separate builds.
//
// The API returns [] for connectors, and no_refresh means only an import ever
// writes it. So:
//
//	imported workflow  -> prior state = []      (hydrated by ImportState)
//	created workflow   -> prior state = null    (never hydrated at all)
//
// The configuration omits the attribute in both cases.
//
//	plain Optional          -> imported plans dirty ([] -> null), created is clean
//	Optional+Computed alone -> created plans dirty (unknown vs null), imported is clean
//
// Build 16:25 shipped the second and broke every Terraform-created workflow;
// reverting shipped the first and broke TestAccWorkflowResource_ImportThenUpdate.
// Neither is a fix - the attribute needs Computed *and* null-prior-state adoption
// for collections, which is what preserveServerDefaults now does.
//
// Both directions are asserted here, because fixing either one alone has now
// visibly broken the other.
func TestConnectors_BothPriorStateShapesSettle(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)

	attribute, ok := rschema.Attributes["connectors"]
	if !ok {
		t.Fatal("connectors is missing from the schema")
	}
	if !attribute.IsComputed() || !attribute.IsOptional() {
		t.Fatalf("connectors must be Optional+Computed (optional=%v computed=%v). Without "+
			"Computed, an imported workflow holds the API's [] in state against a config that "+
			"omits the attribute, and plans `[] -> null` forever.",
			attribute.IsOptional(), attribute.IsComputed())
	}

	objectType, _ := rschema.Type().TerraformType(ctx).(tftypes.Object)
	connectorsType := objectType.AttributeTypes["connectors"]

	build := func(connectors tftypes.Value) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range objectType.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["name"] = tftypes.NewValue(tftypes.String, "example-workflow")
		attrs["connectors"] = connectors
		return tftypes.NewValue(objectType, attrs)
	}

	empty := tftypes.NewValue(connectorsType, []tftypes.Value{})

	cases := []struct {
		name  string
		state tftypes.Value
		want  tftypes.Value
		why   string
	}{
		{
			name:  "created workflow: connectors never hydrated, so state is null",
			state: tftypes.NewValue(connectorsType, nil),
			// NOT the prior null. The API returns [] and so does Update, and the
			// planned value must equal what Update returns or apply fails with
			// "was null, but now cty.ListValEmpty(...)" - which it did, on 16:54,
			// *after* applying the change, leaving a correct resource and a red run.
			want: empty,
			why:  "every workflow Terraform created rather than imported",
		},
		{
			name:  "imported workflow: ImportState hydrated the API's empty list",
			state: empty,
			want:  empty,
			why:   "every workflow adopted via terraform import",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The configuration omits connectors; the framework has marked the
			// planned value unknown because the attribute is Computed.
			config := tfsdk.Config{Raw: build(tftypes.NewValue(connectorsType, nil)), Schema: rschema}
			plan := tfsdk.Plan{Raw: build(tftypes.NewValue(connectorsType, tftypes.UnknownValue)), Schema: rschema}
			state := tfsdk.State{Raw: build(tc.state), Schema: rschema}

			preserved, err := preserveServerDefaults(ctx, rschema, config, plan, state)
			if err != nil {
				t.Fatalf("preserveServerDefaults: %v", err)
			}
			got, err := valueAtPath(preserved, tftypes.NewAttributePath().WithAttributeName("connectors"))
			if err != nil {
				t.Fatalf("walking to connectors: %v", err)
			}

			if !got.IsKnown() {
				t.Fatalf("connectors is still unknown, so the guard will diff it against prior "+
					"state and decline: the plan never settles for %s", tc.why)
			}
			if !got.Equal(tc.want) {
				t.Errorf("connectors = %s, want %s.\nThe planned value must equal what Update "+
					"returns - the API sends an empty list, so promising null fails the "+
					"post-apply consistency check for %s.", got.String(), tc.want.String(), tc.why)
			}
		})
	}
}

// The invariant that the three connectors attempts kept violating, stated
// directly: **the planned value must equal what Update will return.**
//
// Terraform rejects any post-apply value that differs from a planned value that
// was not unknown. The API returns [] for connectors, so Update returns [], so the
// plan must promise []. Promising the prior null instead failed apply with
//
//	.connectors: was null, but now cty.ListValEmpty(...)
//
// *after* the change had been applied - correct infrastructure, red pipeline, and
// a re-run that shows "No changes" and explains nothing.
//
// No plan-based test could catch that: the plan was clean. This one asserts the
// property that makes apply safe, so it is checkable without an apply.
func TestPreserveServerDefaults_PlannedCollectionMatchesWhatUpdateReturns(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)
	objectType, _ := rschema.Type().TerraformType(ctx).(tftypes.Object)

	build := func(connectors tftypes.Value) tftypes.Value {
		attrs := map[string]tftypes.Value{}
		for name, ty := range objectType.AttributeTypes {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
		attrs["name"] = tftypes.NewValue(tftypes.String, "example-workflow")
		attrs["connectors"] = connectors
		return tftypes.NewValue(objectType, attrs)
	}
	connectorsType := objectType.AttributeTypes["connectors"]

	// A Terraform-created workflow: connectors was never hydrated, so null.
	config := tfsdk.Config{Raw: build(tftypes.NewValue(connectorsType, nil)), Schema: rschema}
	plan := tfsdk.Plan{Raw: build(tftypes.NewValue(connectorsType, tftypes.UnknownValue)), Schema: rschema}
	state := tfsdk.State{Raw: build(tftypes.NewValue(connectorsType, nil)), Schema: rschema}

	preserved, err := preserveServerDefaults(ctx, rschema, config, plan, state)
	if err != nil {
		t.Fatalf("preserveServerDefaults: %v", err)
	}
	planned, err := valueAtPath(preserved, tftypes.NewAttributePath().WithAttributeName("connectors"))
	if err != nil {
		t.Fatalf("walking to connectors: %v", err)
	}

	// What Update returns, decoded from the API's response.
	returned := tftypes.NewValue(connectorsType, []tftypes.Value{})

	if planned.IsNull() {
		t.Fatalf("the plan promises null while Update returns %s. Apply will fail the "+
			"consistency check after having already applied the change.", returned.String())
	}
	if !planned.Equal(returned) {
		t.Errorf("planned %s, Update returns %s - these must be equal",
			planned.String(), returned.String())
	}

	// And the other half: with nothing to apply, the guard must still see no
	// meaningful change, or the plan renders dirty forever.
	diffs, err := planned.Diff(tftypes.NewValue(connectorsType, nil))
	if err != nil {
		t.Fatalf("diffing: %v", err)
	}
	if len(meaningfulDiffs(diffs)) != 0 {
		t.Error("an empty collection and a null one count as a meaningful difference, so the " +
			"guard will decline and a genuinely empty plan will render dirty")
	}
}
