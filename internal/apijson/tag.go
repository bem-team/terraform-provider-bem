package apijson

import (
	"reflect"
	"strings"
)

const jsonStructTag = "json"
const formatStructTag = "format"

type parsedStructTag struct {
	name              string
	extras            bool
	metadata          bool
	inline            bool
	required          bool
	optional          bool
	computed          bool
	computed_optional bool
	noRefresh         bool
	// Don't skip this value, even if it's computed (no-op for computed optional fields)
	// If encodeStateForUnknown is set on a computed field, this flag should also be set;
	// otherwise this flag will have no effect
	// NOTE: won't work if update behavior is 'patch'
	forceEncode bool
	// If the value in the plan is unknown,
	// encode the value from the state instead
	// This is similar to the UseStateForUnknown plan modifier,
	// but it only impacts serialization of request bodies, not planning.
	// NOTE #1: only use this for computed/computed_optional values that may be changed by the server;
	// otherwise just use the UseStateForUnknown plan modifier
	// NOTE #2: won't work if update behavior is 'patch'
	encodeStateValueWhenPlanUnknown bool
	// Fields sharing the same non-empty atomicGroup value must be sent together
	// on a patch whenever any one of them changes, even if the others are
	// individually unchanged from state. Set via `atomic_group=<name>` on the
	// json tag. Use this when the API rejects a partial update to a set of
	// sibling fields (e.g. a server-side "these must all be provided together"
	// constraint) that JSON Merge Patch's per-field diffing can't express on
	// its own.
	atomicGroup string
	// fullReplace marks a field the API treats as an indivisible block rather
	// than something it will merge: whenever the field has a value, a patch must
	// carry the whole thing, even if nothing inside it changed. Set via
	// `full_replace` on the json tag.
	//
	// This is atomicGroup's counterpart one level up. atomicGroup keeps a set of
	// *sibling* fields together; fullReplace keeps a *single* field's own
	// subtree intact, for the case where the API validates the field against
	// itself on every request and so rejects (or silently discards) a body that
	// omits it. A field with no siblings to trigger it can't be expressed as a
	// group, which is exactly the gap this fills.
	//
	// A field whose value is null or unset is still omitted - this forces
	// completeness, not presence.
	//
	// NOTE: don't combine with atomicGroup on the same field. fullReplace
	// encodes unconditionally, so the field would always look "changed" and
	// would permanently trigger its group.
	fullReplace bool
}

func parseJSONStructTag(field reflect.StructField) (tag parsedStructTag, ok bool) {
	raw, ok := field.Tag.Lookup(jsonStructTag)
	if !ok {
		return
	}
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return tag, false
	}
	tag.name = parts[0]
	for _, part := range parts[1:] {
		switch part {
		case "extras":
			tag.extras = true
		case "metadata":
			tag.metadata = true
		case "inline":
			tag.inline = true
		case "required":
			tag.required = true
		case "optional":
			tag.optional = true
		case "computed":
			tag.computed = true
		case "computed_optional":
			tag.computed_optional = true
		case "no_refresh":
			tag.noRefresh = true
		case "encode_state_for_unknown":
			tag.encodeStateValueWhenPlanUnknown = true
		case "force_encode":
			tag.forceEncode = true
		case "full_replace":
			tag.fullReplace = true
		default:
			if group, ok := strings.CutPrefix(part, "atomic_group="); ok {
				tag.atomicGroup = group
			}
		}
	}
	return
}

func parseFormatStructTag(field reflect.StructField) (format string, ok bool) {
	format, ok = field.Tag.Lookup(formatStructTag)
	return
}
