package workflow

import (
	"testing"

	"github.com/bem-team/terraform-provider-bem/internal/apijson"
)

// BEM-1396 finding 2: bem_workflow re-planned an in-place update forever, and
// the churn was state-only - the server data was correct throughout.
//
// Cause: Update() decoded the response straight into WorkflowModel while
// Create(), Read() and ImportState() all decode through
// WorkflowWorkflowEnvelope. The API wraps its response as
// {"workflow": {...}} and WorkflowModel.Workflow is itself tagged
// json:"workflow", so the two paths populate disjoint halves of the model:
//
//	with envelope    -> top-level computed attributes populated, nested null
//	without envelope -> nested "workflow" object populated, top-level null
//
// Create wrote the top-level attributes; Update nulled them and wrote the
// nested object; the next plan saw nulls and proposed another update. The plan
// notation was the tell - `~ created_at = "..."` before the update (state had a
// value) became `+ created_at` after it (state now holds null).
//
// The payload below is the real shape, captured from
// GET /v3/workflows/tf-enrich-wf on staging. The important property is what it
// does NOT contain: any nested "workflow" key inside the wrapper.

const realWorkflowResponse = `{
  "workflow": {
    "id": "wf_3Hbc",
    "name": "tf-enrich-wf",
    "displayName": "TF Enrich Workflow",
    "mainNodeName": "extract",
    "versionNum": 2,
    "createdAt": "2026-08-07T22:03:14Z",
    "updatedAt": "2026-08-11T17:11:56Z",
    "emailAddress": "eml_example@example.com",
    "restricted": false,
    "tags": ["terraform-managed"],
    "nodes": [{"name": "extract", "function": {"id": "f_3Hbc", "name": "tf-enrich-extractor", "versionNum": 4}}],
    "edges": [],
    "audit": {"createdBy": "usr_3Hbc"}
  }
}`

// Guards the fix. Every top-level computed attribute must come back populated
// after an update, because that is what Create leaves in state and what the
// next plan compares against.
func TestUnmarshalComputed_ViaEnvelope_PopulatesTopLevelComputedAttributes(t *testing.T) {
	env := WorkflowWorkflowEnvelope{}
	if err := apijson.UnmarshalComputed([]byte(realWorkflowResponse), &env); err != nil {
		t.Fatalf("UnmarshalComputed: %v", err)
	}
	model := env.Workflow

	if model.VersionNum.IsNull() {
		t.Error("version_num is null; the next plan would propose an update to fill it")
	} else if got := model.VersionNum.ValueInt64(); got != 2 {
		t.Errorf("version_num = %d, want 2", got)
	}

	for name, isNull := range map[string]bool{
		"created_at":    model.CreatedAt.IsNull(),
		"updated_at":    model.UpdatedAt.IsNull(),
		"email_address": model.EmailAddress.IsNull(),
		"restricted":    model.Restricted.IsNull(),
		"audit":         model.Audit.IsNull(),
	} {
		if isNull {
			t.Errorf("%s is null after decoding through the envelope; "+
				"Create populates it, so Update leaving it null is what caused the perpetual diff", name)
		}
	}
}

// The bug itself, pinned so it cannot come back by someone "simplifying" the
// envelope away. Decoding the same payload straight into the model puts the
// response under the nested attribute and leaves the top level empty.
func TestUnmarshalComputed_WithoutEnvelope_LeavesTopLevelNull(t *testing.T) {
	var model WorkflowModel
	if err := apijson.UnmarshalComputed([]byte(realWorkflowResponse), &model); err != nil {
		t.Fatalf("UnmarshalComputed: %v", err)
	}

	if !model.VersionNum.IsNull() {
		t.Fatalf("version_num = %v without the envelope. If this now populates, the response shape "+
			"or the model changed and Update's envelope wrap should be re-examined - see the fix in "+
			"resource.go's Update.", model.VersionNum)
	}
	if !model.CreatedAt.IsNull() {
		t.Errorf("created_at = %v without the envelope, expected null", model.CreatedAt)
	}
}

// The response wrapper has no nested "workflow" key, which is why the two
// decode paths are mutually exclusive rather than additive. If the API ever
// starts returning one, both halves could be populated at once and the envelope
// asymmetry above would stop being a problem - so assert the assumption the fix
// rests on rather than leaving it implicit.
func TestRealWorkflowResponse_HasNoNestedWorkflowKey(t *testing.T) {
	env := WorkflowWorkflowEnvelope{}
	if err := apijson.UnmarshalComputed([]byte(realWorkflowResponse), &env); err != nil {
		t.Fatalf("UnmarshalComputed: %v", err)
	}
	if !env.Workflow.Workflow.IsNull() {
		t.Errorf("the nested workflow attribute is populated (%v), meaning the wrapper now contains "+
			"a nested \"workflow\" key. Re-check whether Update still needs the envelope.",
			env.Workflow.Workflow)
	}
}

// Both halves must survive an Update decode. The envelope-only version fixed the
// top-level attributes but left the `workflow` mirror permanently null, which any
// configuration reading bem_workflow.x.workflow.* would see.
func TestHydrateWorkflowResponse_PopulatesTopLevelAndMirror(t *testing.T) {
	data := &WorkflowModel{}
	if err := hydrateWorkflowResponse([]byte(realWorkflowResponse), data, apijson.UnmarshalComputed); err != nil {
		t.Fatalf("hydrateWorkflowResponse: %v", err)
	}

	if data.VersionNum.IsNull() {
		t.Error("top-level version_num is null; Create populates it, so Update must too")
	}
	if data.CreatedAt.IsNull() {
		t.Error("top-level created_at is null")
	}
	if data.Workflow.IsNull() {
		t.Fatal("the `workflow` mirror is null - a single envelope decode nulls it, which is " +
			"why Update needs the second pass")
	}
}

// The import path, which no test covered until the live import proved it broken.
//
// UnmarshalForImport is not computed-only, so unlike UnmarshalComputed it is
// allowed to write every field - including the no_refresh ones the envelope pass
// just populated. Since the struct decoder nulls fields whose key is absent, and
// no attribute's key exists at the root, any pass over the whole model at root
// level wipes them. Both halves must survive.
func TestHydrateWorkflowResponse_ForImport_PopulatesNoRefreshAndMirror(t *testing.T) {
	data := &WorkflowModel{}
	if err := hydrateWorkflowResponse([]byte(realWorkflowResponse), data, apijson.UnmarshalForImport); err != nil {
		t.Fatalf("hydrateWorkflowResponse: %v", err)
	}

	// display_name and main_node_name are no_refresh, so they are exactly the
	// fields import exists to populate. main_node_name is also Required, so a
	// null there breaks plan -generate-config-out outright.
	if got := data.DisplayName.ValueString(); got != "TF Enrich Workflow" {
		t.Errorf("display_name = %q after an import decode, want it populated", got)
	}
	if got := data.MainNodeName.ValueString(); got != "extract" {
		t.Errorf("main_node_name = %q after an import decode, want %q", got, "extract")
	}
	if data.VersionNum.IsNull() {
		t.Error("version_num is null after an import decode")
	}
	if data.Workflow.IsNull() {
		t.Error("the `workflow` mirror is null after an import decode; configurations read " +
			"workflow.version_num from it")
	}
}
