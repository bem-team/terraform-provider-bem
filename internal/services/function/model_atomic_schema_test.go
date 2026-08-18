package function

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The bem API treats outputSchemaName and outputSchema as an atomic pair on
// PATCH /v3/functions/{name}, for every function type whose update struct
// embeds OutputSchemaConfigCreate - analyze, extract, transform and join:
//
//	func (efu ExtractFunctionUpdate) Validate() error {
//	    if !efu.OutputSchemaConfigCreate.IsEmpty() {      // name alone => not empty
//	        return efu.OutputSchemaConfigCreate.Validate() // schema absent => error
//	    }
//	}
//
// So a body carrying outputSchemaName without outputSchema is rejected with
// 400 "output schema is required". JSON Merge Patch's per-field diffing
// produces exactly that body when a user changes only output_schema_name in
// HCL, which is the same failure shape BEM-1392 fixed for the workflow DAG.
//
// The reverse (schema without name) is not an error - createOutputSchema
// back-fills the existing name - but the group is declared on both fields, so
// both directions send both keys.

const testOutputSchema = `{"type":"object","properties":{"invoiceNumber":{"type":"string"}},"required":["invoiceNumber"]}`

// atomicSchemaFixture returns a converged extract function: both output-schema
// fields populated in state, nothing pending. Callers mutate the returned plan
// to isolate a single attribute change.
func atomicSchemaFixture() (plan, state FunctionModel) {
	state = FunctionModel{
		FunctionName:     types.StringValue("tf-extract"),
		Type:             types.StringValue("extract"),
		DisplayName:      types.StringValue("TF Extract"),
		OutputSchemaName: types.StringValue("TFInvoiceSchema"),
		OutputSchema:     jsontypes.NewNormalizedValue(testOutputSchema),
	}
	return state, state
}

// bodyKeys decodes a PATCH body into its top-level keys. Asserting on keys
// rather than a byte-for-byte string keeps these tests focused on the atomic
// grouping and stops unrelated schema-formatting changes from failing them.
func bodyKeys(t *testing.T, body []byte) map[string]json.RawMessage {
	t.Helper()

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("PATCH body is not a JSON object: %v (body: %s)", err, body)
	}
	return decoded
}

// Tripwire for the two ways this fix can silently disappear, mirroring
// TestWorkflowModel_DAGFieldsCarryAtomicGroupTag in the workflow package:
// a typo'd option (apijson's parser ignores anything it doesn't recognise, so
// the fix just turns off), or a Stainless regen dropping the hand-added tag
// from this generated file. Without this, the behavioural tests below would
// still pass against a correctly-working plain patch encoder.
func TestMarshalJSONForUpdate_NameOnlyChange_SplicesSchemaIn(t *testing.T) {
	plan, state := atomicSchemaFixture()
	plan.OutputSchemaName = types.StringValue("TFInvoiceSchemaV2")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)
	if _, ok := decoded["outputSchemaName"]; !ok {
		t.Fatal("outputSchemaName missing - it is the field that changed")
	}
	schema, ok := decoded["outputSchema"]
	if !ok {
		t.Fatalf("outputSchema missing. The API rejects a name without a schema "+
			"(\"output schema is required\"), so ensureOutputSchemaAccompaniesName must splice "+
			"it in. If this failed right after a Stainless regen, that helper or its call in "+
			"MarshalJSONForUpdate was dropped.\nbody: %s", got)
	}
	if len(schema) == 0 || string(schema) == "null" {
		t.Errorf("outputSchema was sent as %s, which the API treats as absent", schema)
	}
}

// The pair must NOT be an atomic_group. That was tried and reverted: the group is
// symmetric, but the API's requirement is one-directional - a name without a
// schema is rejected, while a schema without a name is fine because the server
// back-fills the name. Grouping them made an ordinary update of a function that
// sets only one of the two fail to serialize at all:
//
//	apijson: cannot encode field "outputSchemaName" of atomic group
//	"output_schema": ... but it has no value to send
//
// output_schema_name is optional, so that is a legitimate configuration, and the
// resource became permanently un-updatable. full_replace on the schema alone
// achieves the same result with no failure mode.
func TestFunctionModel_OutputSchemaPairIsNotAnAtomicGroup(t *testing.T) {
	modelType := reflect.TypeOf(FunctionModel{})

	for _, fieldName := range []string{"OutputSchemaName", "OutputSchema"} {
		field, ok := modelType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("FunctionModel has no field %q", fieldName)
		}
		for _, part := range strings.Split(field.Tag.Get("json"), ",") {
			if strings.HasPrefix(part, "atomic_group=") {
				t.Errorf("FunctionModel.%s carries %q. See the comment above this test: the "+
					"requirement is one-directional and grouping them breaks updates of "+
					"functions that set only one.", fieldName, part)
			}
		}
	}
}

// Regression for exactly that breakage: a function that sets output_schema but
// never output_schema_name must still be updatable.
func TestMarshalJSONForUpdate_SchemaWithoutName_Serializes(t *testing.T) {
	state := FunctionModel{
		FunctionName: types.StringValue("tf-extract"),
		Type:         types.StringValue("extract"),
		OutputSchema: jsontypes.NewNormalizedValue(testOutputSchema),
	}
	plan := state
	plan.OutputSchema = jsontypes.NewNormalizedValue(
		`{"type":"object","properties":{"invoiceNumber":{"type":"string"},"total":{"type":"number"}}}`)

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("update failed to serialize for a function with no output_schema_name: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)
	if _, ok := decoded["outputSchema"]; !ok {
		t.Error("outputSchema missing - it is the field that changed")
	}
	if raw, ok := decoded["outputSchemaName"]; ok {
		t.Errorf("outputSchemaName = %s was sent although it was never configured", raw)
	}
}

// And the symmetric case: a function that sets only output_schema_name must be
// updatable too. The server back-fills the schema, so a name-only body is not
// itself the problem - a hard serialize error was.
func TestMarshalJSONForUpdate_NameWithoutSchema_Serializes(t *testing.T) {
	state := FunctionModel{
		FunctionName:     types.StringValue("tf-extract"),
		Type:             types.StringValue("extract"),
		OutputSchemaName: types.StringValue("Old"),
	}
	plan := state
	plan.OutputSchemaName = types.StringValue("New")

	if _, err := plan.MarshalJSONForUpdate(state); err != nil {
		t.Fatalf("update failed to serialize for a function with no output_schema: %v", err)
	}
}

// The headline bug. Pre-fix this produced
// {"outputSchemaName":"TFInvoiceSchemaV2","functionName":"tf-extract"} -
// captured by running this exact model through MarshalJSONForUpdate against
// released 0.13.0 - which the API rejects with 400 "output schema is
// required".
func TestMarshalJSONForUpdate_OutputSchemaNameOnlyChange_Superseded(t *testing.T) {
	plan, state := atomicSchemaFixture()
	plan.OutputSchemaName = types.StringValue("TFInvoiceSchemaV2")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)

	name, ok := decoded["outputSchemaName"]
	if !ok {
		t.Fatal("outputSchemaName missing from the body - it is the field that changed")
	}
	if string(name) != `"TFInvoiceSchemaV2"` {
		t.Errorf("outputSchemaName = %s, want %q", name, "TFInvoiceSchemaV2")
	}

	schema, ok := decoded["outputSchema"]
	if !ok {
		t.Fatalf(
			"outputSchema missing from the body despite outputSchemaName changing.\n"+
				"The API rejects this with 400 \"output schema is required\".\nbody: %s", got,
		)
	}
	if len(schema) == 0 || string(schema) == "null" {
		t.Errorf("outputSchema was sent as %s, which the API treats as absent", schema)
	}
}

// The other direction. Not an API error on its own (createOutputSchema
// back-fills the existing name when it's absent), but the group is declared on
// both fields, so the name must ride along - and asserting it here is what
// proves the grouping is symmetric rather than a one-way special case.
func TestMarshalJSONForUpdate_OutputSchemaOnlyChange_SendsSchemaAlone(t *testing.T) {
	plan, state := atomicSchemaFixture()
	plan.OutputSchema = jsontypes.NewNormalizedValue(
		`{"type":"object","properties":{"invoiceNumber":{"type":"string"},"totalAmount":{"type":"number"}},"required":["invoiceNumber"]}`,
	)

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)

	if _, ok := decoded["outputSchema"]; !ok {
		t.Error("outputSchema missing from the body - it is the field that changed")
	}
	// The name must NOT ride along: the server back-fills it from the existing
	// schema, so forcing it is unnecessary, and doing so symmetrically is what
	// broke functions that never set it.
	if raw, ok := decoded["outputSchemaName"]; ok {
		t.Errorf("outputSchemaName = %s was sent although only the schema changed", raw)
	}
}

// The regression guard, and the case most likely to break if the grouping
// over-fires. An edit that touches neither output-schema field must send
// neither - BEM-1392's post-release verification called this out as the one
// scenario that could actually regress, since every other test proves the
// group fires when it should rather than staying quiet when it shouldn't.
//
// It also matters practically: a display_name edit is the most common function
// update there is, and forcing an unnecessary output-schema write on every one
// of them would create a new output schema row server-side each time.
func TestMarshalJSONForUpdate_DisplayNameOnlyChange_OmitsOutputSchemaFields(t *testing.T) {
	plan, state := atomicSchemaFixture()
	plan.DisplayName = types.StringValue("TF Extract (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)

	if _, ok := decoded["displayName"]; !ok {
		t.Error("displayName missing from the body - it is the field that changed")
	}
	for _, key := range []string{"outputSchemaName", "outputSchema"} {
		if raw, ok := decoded[key]; ok {
			t.Errorf("%s = %s was sent on a display-name-only edit; the atomic group is over-firing", key, raw)
		}
	}
}

// A function that has never configured either output-schema field must stay
// clear of the group entirely - this is the shape a `split` or `classify`
// function has, and forcing an empty outputSchema onto one of those would turn
// a working update into a 400 rather than fixing one.
func TestMarshalJSONForUpdate_NeverConfiguredOutputSchema_OmitsBothFields(t *testing.T) {
	state := FunctionModel{
		FunctionName: types.StringValue("tf-split"),
		Type:         types.StringValue("split"),
		DisplayName:  types.StringValue("TF Split"),
	}
	plan := state
	plan.DisplayName = types.StringValue("TF Split (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := bodyKeys(t, got)
	for _, key := range []string{"outputSchemaName", "outputSchema"} {
		if raw, ok := decoded[key]; ok {
			t.Errorf("%s = %s was sent for a function that never configured an output schema", key, raw)
		}
	}
}

// Create must not be affected: MarshalJSON has no patch state to diff against,
// so the atomic group has nothing to force and both fields should appear
// simply because they are set. Guards against the group leaking into the
// create path.
func TestMarshalJSON_Create_SendsConfiguredOutputSchemaFields(t *testing.T) {
	_, model := atomicSchemaFixture()

	got, err := model.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	t.Logf("POST body: %s", got)

	decoded := bodyKeys(t, got)
	for _, key := range []string{"outputSchemaName", "outputSchema"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("%s missing from the create body despite being configured", key)
		}
	}
}
