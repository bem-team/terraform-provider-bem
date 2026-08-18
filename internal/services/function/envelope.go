package function

import (
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
)

// FunctionFunctionEnvelope mirrors WorkflowWorkflowEnvelope: the API wraps a
// function payload as {"function": {...}}.
//
// Deliberately declared here rather than in model.go, which carries the
// Stainless generated-file header. Stainless generates the workflow envelope
// but not this one, so keeping it in a hand-written file means a regen cannot
// silently drop it.
type FunctionFunctionEnvelope struct {
	Function FunctionModel `json:"function"`
}

// functionMirror isolates the read-only `function` attribute so it can be
// decoded without touching anything else on the model. See hydrateFromResponse.
type functionMirror struct {
	Function customfield.NestedObject[FunctionFunctionModel] `json:"function,computed"`
}

// hydrateFromResponse decodes a function API response into data.
//
// FunctionModel has two competing claims on the same payload:
//
//   - the resource's own attributes (function_name, type, config,
//     output_schema, tags, ...), whose keys live *inside* the {"function": ...}
//     wrapper; and
//   - the read-only `function` mirror attribute, tagged json:"function", which
//     therefore matches the wrapper itself.
//
// A single decode can only satisfy one of them, and the wrapper contains no
// nested "function" key of its own, so they are mutually exclusive:
//
//	envelope only -> attributes hydrate, the `function` mirror goes null
//	direct only   -> the mirror hydrates, attributes keep their plan values
//
// Create and Update originally decoded straight into the model, so `config`
// was never hydrated: the response's config lives at function.config, not at
// the root. The server fills in defaults for the computed_optional leaves
// (steps[].top_k / search_mode / score_threshold and
// endpoints[].match_top_k / max_candidates / max_pages), but state kept null
// for all six. The framework marks a Computed attribute unknown whenever its
// configuration value is null (fwserver.MarkComputedNilsAsUnknown), so every
// subsequent plan proposed an in-place update, applied it, got null back, and
// proposed it again - one new function version per apply, indefinitely, with
// the workflow's node diff cascading off it.
//
// Dropping the mirror instead is not an option: FunctionModel has no top-level
// version_num or function_id, so `function` is the only place to read them
// from, and configs interpolate function.function_id / function.version_num
// into workflow nodes.
//
// Hence both passes. The envelope runs first for the real attributes, then the
// mirror is decoded on its own.
//
// The second pass is deliberately a one-field struct rather than a second full
// decode of the model, for the reason spelled out in workflow/envelope.go: the
// struct decoder nulls any field whose key is absent from the JSON, and at the
// root every attribute's key is absent - they all live one level down, inside
// the wrapper. A full second pass therefore undoes the envelope pass for every
// field it is allowed to touch.
//
// Under UnmarshalComputed it was allowed to touch almost nothing (non-computed
// fields decode as OnlyNested, so top-level values survive), which is why a full
// second pass appeared to work. UnmarshalForImport is not computed-only, so it
// wiped all 20-odd no_refresh attributes straight back to null and import landed
// state holding only the mirror. Keeping the blast radius to one field makes the
// pass correct regardless of which decoder entry point the caller passes.
func hydrateFromResponse(bytes []byte, data *FunctionModel, unmarshal func([]byte, any) error) error {
	env := FunctionFunctionEnvelope{Function: *data}
	if err := unmarshal(bytes, &env); err != nil {
		return err
	}
	*data = env.Function

	mirror := functionMirror{Function: data.Function}
	if err := unmarshal(bytes, &mirror); err != nil {
		return err
	}
	data.Function = mirror.Function

	return nil
}
