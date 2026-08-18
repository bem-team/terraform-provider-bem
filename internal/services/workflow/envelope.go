package workflow

import (
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
)

// workflowMirror isolates the read-only `workflow` attribute so it can be
// decoded without touching anything else on the model. See
// hydrateWorkflowResponse.
type workflowMirror struct {
	Workflow customfield.NestedObject[WorkflowWorkflowModel] `json:"workflow,computed"`
}

// hydrateWorkflowResponse decodes a workflow API response into data, populating
// both the resource's own attributes and the read-only `workflow` mirror.
//
// WorkflowModel has two competing claims on the same payload: its own attributes
// live *inside* the {"workflow": {...}} wrapper, while the mirror attribute is
// tagged json:"workflow" and so matches the wrapper itself. The wrapper contains
// no nested "workflow" key (asserted by
// TestRealWorkflowResponse_HasNoNestedWorkflowKey), so one decode cannot satisfy
// both.
//
// Using the envelope alone fixed the top-level attributes but left the mirror
// permanently null - and once collapseNoOpPlan suppresses the spurious updates
// that used to overwrite it, it stays null, so any configuration reading
// bem_workflow.x.workflow.* gets null.
//
// The second decode is deliberately NOT a full pass over the model. The struct
// decoder nulls any field whose key is absent from the JSON, and at the root
// every attribute's key is absent - they all live one level down, inside the
// wrapper. WorkflowModel has many top-level `computed` attributes, which
// UnmarshalComputed treats as "always update", so a second full decode nulls
// version_num, created_at, updated_at and the rest, undoing the envelope pass.
// Decoding into a one-field struct keeps the blast radius to the mirror.
//
// function/envelope.go's hydrateFromResponse originally did use a full second
// pass, on the reasoning that FunctionModel has no such fields - only its mirror
// is plain `computed`. That held only for computed-only decoders. UnmarshalForImport
// is not one, so it nulled every no_refresh attribute there and import landed
// state holding just the mirror. Both helpers now decode the mirror alone, which
// is correct whichever entry point the caller passes.
func hydrateWorkflowResponse(bytes []byte, data *WorkflowModel, unmarshal func([]byte, any) error) error {
	env := WorkflowWorkflowEnvelope{Workflow: *data}
	if err := unmarshal(bytes, &env); err != nil {
		return err
	}
	*data = env.Workflow

	mirror := workflowMirror{Workflow: data.Workflow}
	if err := unmarshal(bytes, &mirror); err != nil {
		return err
	}
	data.Workflow = mirror.Workflow

	return nil
}
