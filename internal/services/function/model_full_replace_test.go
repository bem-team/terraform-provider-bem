package function

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BEM-1396 finding 4 — the preventative audit. Sweeping every case of the
// function update handler turned up two more instances of the same family
// beyond the enrich `config` one: a set of fields the server rebuilds wholesale
// on every update, so JSON Merge Patch omitting the unchanged members produces
// a request the server silently applies with the missing pieces zeroed.
//
// Both of these fail *quietly*, which is why neither had been noticed - unlike
// the enrich case, which announced itself with a 400:
//
//   - send: validation passes when destinationType is absent, and the API only
//     rebuilds the destination block when destinationType is present.
//     A webhook_url-only edit therefore sends no destinationType,
//     the server discards the change, and returns 200. Because webhookUrl is
//     no_refresh, state keeps the new value and the divergence never surfaces
//     as a diff at all.
//
//   - payload_shaping: the API assigns shapingSchema from the request
//     unconditionally, with no "was it provided" check and no validation, and an
//     absent key is indistinguishable from an empty string. So it wipes the
//     stored schema on any edit that does not carry it - a display_name change, for
//     instance.
//
// full_replace on each affected field makes the block always travel complete.
//
// Audited and found clean, for the record: parse and render both merge
// explicitly, carrying omitted fields forward from the existing config
// (functions.go:2118-2135, 2154-2178); route and classify are pointer-guarded
// per field (updateRouteFunction, functions.go:2436-2461). split was
// considered and deliberately left alone - see
// TestMarshalJSONForUpdate_SplitConfigsAreNotFullReplace below.

// fullReplaceTaggedFields is the audit's conclusion, encoded. Adding a field to
// the model that the server rebuilds wholesale means adding it here too.
var fullReplaceTaggedFields = map[string]string{
	// send - destination block, rebuilt wholesale from the request
	"DestinationType":       "destinationType",
	"WebhookURL":            "webhookUrl",
	"WebhookSigningEnabled": "webhookSigningEnabled",
	"S3Bucket":              "s3Bucket",
	"S3Prefix":              "s3Prefix",
	"GoogleDriveFolderID":   "googleDriveFolderId",
	// payload_shaping - assigned unconditionally, no omitempty
	"ShapingSchema": "shapingSchema",
	// enrich - validated against itself on every request
	"Config": "config",
}

func TestFunctionModel_FullReplaceFieldsCarryTag(t *testing.T) {
	modelType := reflect.TypeOf(FunctionModel{})

	for fieldName, jsonName := range fullReplaceTaggedFields {
		t.Run(fieldName, func(t *testing.T) {
			field, ok := modelType.FieldByName(fieldName)
			if !ok {
				t.Fatalf("FunctionModel has no field %q - if it was renamed, the full_replace tag must move with it", fieldName)
			}

			parts := strings.Split(field.Tag.Get("json"), ",")
			if parts[0] != jsonName {
				t.Errorf("FunctionModel.%s json name = %q, want %q", fieldName, parts[0], jsonName)
			}
			for _, part := range parts[1:] {
				if part == "full_replace" {
					return
				}
			}
			t.Errorf(
				"FunctionModel.%s is missing the \"full_replace\" json tag option (got tag: %q).\n"+
					"The bem API rebuilds this field's block wholesale on update, so a patch that omits "+
					"it either discards the change or zeroes the stored value - usually with a 200 and no "+
					"error. If this failed right after a Stainless regen, re-apply the tag.",
				fieldName, field.Tag.Get("json"),
			)
		})
	}
}

func decodeBody(t *testing.T, body []byte) map[string]json.RawMessage {
	t.Helper()

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("body is not a JSON object: %v (body: %s)", err, body)
	}
	return decoded
}

func sendFixture() FunctionModel {
	return FunctionModel{
		FunctionName:          types.StringValue("tf-send"),
		Type:                  types.StringValue("send"),
		DisplayName:           types.StringValue("TF Send"),
		DestinationType:       types.StringValue("webhook"),
		WebhookURL:            types.StringValue("https://sink.example.com/hook"),
		WebhookSigningEnabled: types.BoolValue(false),
	}
}

// The silent one. Pre-fix a webhook_url-only edit sent
// {"webhookUrl":"...","functionName":"..."} with no destinationType, so the
// server skipped the SendConfig rebuild entirely and returned 200 having
// changed nothing.
func TestMarshalJSONForUpdate_SendWebhookURLOnlyChange_SendsDestinationType(t *testing.T) {
	state := sendFixture()
	plan := state
	plan.WebhookURL = types.StringValue("https://sink.example.com/hook-v2")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := decodeBody(t, got)
	if string(decoded["webhookUrl"]) != `"https://sink.example.com/hook-v2"` {
		t.Errorf("webhookUrl = %s, want the new value", decoded["webhookUrl"])
	}
	if _, ok := decoded["destinationType"]; !ok {
		t.Fatalf("destinationType missing - the server skips the SendConfig rebuild without it and "+
			"discards this change with a 200\nbody: %s", got)
	}
}

// A metadata-only edit must still carry the whole destination block, because
// the server rebuilds SendConfig from the request and would zero anything
// absent.
func TestMarshalJSONForUpdate_SendDisplayNameOnlyChange_SendsWholeDestinationBlock(t *testing.T) {
	state := sendFixture()
	plan := state
	plan.DisplayName = types.StringValue("TF Send (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := decodeBody(t, got)
	for _, key := range []string{"destinationType", "webhookUrl", "webhookSigningEnabled"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("%s missing from the body; the server would zero it on the rebuild\nbody: %s", key, got)
		}
	}
	// An explicit false must survive - it is a configured value, not an absence.
	if string(decoded["webhookSigningEnabled"]) != "false" {
		t.Errorf("webhookSigningEnabled = %s, want false", decoded["webhookSigningEnabled"])
	}
	// Fields belonging to other destination types were never set and must stay out.
	for _, key := range []string{"s3Bucket", "s3Prefix", "googleDriveFolderId"} {
		if raw, ok := decoded[key]; ok {
			t.Errorf("%s = %s was sent for a webhook destination that never configured it", key, raw)
		}
	}
}

// payload_shaping: the shaping schema must ride along on every update, or the
// server's unconditional assignment overwrites it with "".
func TestMarshalJSONForUpdate_PayloadShapingDisplayNameOnlyChange_StillSendsSchema(t *testing.T) {
	state := FunctionModel{
		FunctionName:  types.StringValue("tf-shape"),
		Type:          types.StringValue("payload_shaping"),
		DisplayName:   types.StringValue("TF Shape"),
		ShapingSchema: types.StringValue(`{"invoiceNumber":"{{ .invoiceNumber }}"}`),
	}
	plan := state
	plan.DisplayName = types.StringValue("TF Shape (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := decodeBody(t, got)
	if _, ok := decoded["shapingSchema"]; !ok {
		t.Fatalf("shapingSchema missing - the server assigns it unconditionally, so an absent key "+
			"wipes the stored schema and still returns 200\nbody: %s", got)
	}
}

// Deliberately NOT fixed here, recorded so the next person does not re-derive
// it: an update carrying splitType clears both split config blocks and re-applies
// only what the request body contained, and the validation that would have caught
// a type/config mismatch does not run on the update path.
//
// It looks like the same family as the fixes above, but no provider-side repro
// exists: a split function carries exactly one config block, so changing
// split_type necessarily changes the block too and both travel together. Adding
// full_replace here would only affect the case where a user changes split_type
// and leaves the *old* block in config, and there it makes things worse - the
// stale block would be persisted against the new type instead of the update
// landing visibly empty. The real gap is the missing server-side validation.
//
// This test pins the current, deliberate behaviour so a future full_replace
// addition has to be a conscious decision rather than a reflex.
func TestMarshalJSONForUpdate_SplitConfigsAreNotFullReplace(t *testing.T) {
	modelType := reflect.TypeOf(FunctionModel{})

	for _, fieldName := range []string{"PrintPageSplitConfig", "SemanticPageSplitConfig"} {
		field, ok := modelType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("FunctionModel has no field %q", fieldName)
		}
		for _, part := range strings.Split(field.Tag.Get("json"), ",") {
			if part == "full_replace" {
				t.Errorf("FunctionModel.%s carries full_replace. That was considered and rejected - "+
					"see the comment above this test. If it is being added deliberately, update the "+
					"comment with the repro that justifies it.", fieldName)
			}
		}
	}
}

// The counterpart guard for all of the above: a function type that configures
// none of these fields must send none of them. This is what stops the audit's
// fix from turning working updates for other types into failures.
func TestMarshalJSONForUpdate_ExtractFunction_SendsNoOtherTypesBlocks(t *testing.T) {
	state := FunctionModel{
		FunctionName: types.StringValue("tf-extract"),
		Type:         types.StringValue("extract"),
		DisplayName:  types.StringValue("TF Extract"),
	}
	plan := state
	plan.DisplayName = types.StringValue("TF Extract (renamed)")

	got, err := plan.MarshalJSONForUpdate(state)
	if err != nil {
		t.Fatalf("MarshalJSONForUpdate failed: %v", err)
	}
	t.Logf("PATCH body: %s", got)

	decoded := decodeBody(t, got)
	for _, jsonName := range fullReplaceTaggedFields {
		if raw, ok := decoded[jsonName]; ok {
			t.Errorf("%s = %s was sent for an extract function that never configured it", jsonName, raw)
		}
	}
	if _, ok := decoded["displayName"]; !ok {
		t.Errorf("displayName missing from the body\nbody: %s", got)
	}
}
