package function

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The audit BEM-1396 asked for, as a test rather than a one-off grep.
//
// Background. `no_refresh` means Read never writes the attribute, so before
// UnmarshalForImport nothing ever put a server value into state for it. Import
// now does. Any attribute where the API supplies a value that the configuration
// omits therefore leaves the plan dirty **forever** - each apply cutting a new
// function version, with the churn cascading into every workflow that
// interpolates the function's id or version.
//
// The critical constraint: this cannot be repaired in ModifyPlan for an attribute
// that is Optional *without* Computed. Terraform core requires the planned value
// of a non-computed attribute to equal the configuration value, and rejects the
// plan outright otherwise:
//
//	Error: Provider produced invalid plan
//	.description: planned value cty.StringVal("") for a non-computed attribute
//
// That was verified by implementing it, not reasoned about. So membership of this
// set is decided in the schema, and an attribute the API assigns is Computed by
// definition.
//
// Probed against the API (create with a minimal body, then read back) to find
// which attributes it actually assigns. NOTE the limitation this method has: a
// bare GET omits any field the API gates behind a query parameter, so absence is
// only evidence for ungated fields. extraConfig is gated behind
// includeExtraSettings and was initially misread this way.
//
//	description               ""                        every type
//	tags                      []                        every type
//	pre_count                 false                     extract only
//	tabular_chunking_enabled  false                     extract only
//	enable_bounding_boxes     false                     extract only
//	parse_config              {extract_entities: true,  parse only
//	                           link_across_documents: true}
//	config                    {"steps": null}           every type (already Optional+Computed)
//
// Note that two of those are not zero values, so "preserve when prior state holds
// the type's zero value" would not have covered them.
func TestAudit_PlainOptionalNoRefreshAttributes(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)

	// Confirmed server-assigned, and therefore made Optional+Computed. schema.go
	// carries a codegen header, so a regen drops these - hence the assertion.
	mustBeComputed := []string{
		"description",
		"tags",
		"pre_count",
		"tabular_chunking_enabled",
		"enable_bounding_boxes",
	}
	for _, name := range mustBeComputed {
		attribute, ok := rschema.Attributes[name]
		if !ok {
			t.Errorf("%s is missing from the schema", name)
			continue
		}
		if !attribute.IsComputed() || !attribute.IsOptional() {
			t.Errorf("%s must be Optional+Computed (optional=%v computed=%v). The API assigns it "+
				"when the configuration omits it, so without Computed an imported resource can "+
				"never plan clean and ModifyPlan cannot fix it - core rejects a planned value "+
				"that differs from a non-computed attribute's config value. If a codegen run "+
				"dropped this, re-add it.", name, attribute.IsOptional(), attribute.IsComputed())
		}
	}

	// Everything still Optional-without-Computed + no_refresh. Not known to be
	// server-assigned, with one exception recorded below.
	known := map[string]string{
		"classifications":  "not server-assigned",
		"destination_type": "not server-assigned",
		"display_name":     "not server-assigned",
		// NOT an Optional+Computed candidate, despite looking like one. Probed against
		// the API with the attribute omitted:
		//
		//   extract, ?includeExtraSettings=true -> 21 populated fields (ocrProvider,
		//   llmProvider, every fallback* variant, service versions), NONE of them the
		//   single field FunctionExtraConfigModel represents.
		//
		// The decoder matches nothing and produces {enable_bounding_boxes: null} - a
		// non-null object with a null leaf. That is what regressed ten of eleven
		// consumer fixtures on build 16:25, and it is a model/API surface mismatch
		// (1 field of 21 on extract, 0 of 0 on parse), not a null-handling problem.
		// Making it Optional+Computed would have the provider adopt as prior state an
		// object it structurally cannot represent.
		//
		// Also: on extract this leaf shadows the top-level enable_bounding_boxes
		// attribute, and the API surfaces that value at the top level rather than
		// inside extraConfig - so what the nested one does, if anything, is unclear.
		//
		// The open question is therefore what the attribute is FOR - read-only mirror
		// of what the API sends, narrowed to what the model covers, or removed - not
		// which plan rule to apply. Track it separately from parse_config, which is a
		// mechanical change.
		//
		// Meanwhile ImportState does not request extraConfig (see functionGetParams),
		// which keeps state null and the plan clean, at the cost of the attribute
		// being absent after an import.
		"extra_config":               "server-assigned (gated) - OPEN, awaiting the acceptance run",
		"google_drive_folder_id":     "not server-assigned",
		"join_type":                  "not server-assigned",
		"native_visual_input":        "not server-assigned",
		"output_schema":              "not server-assigned, but see TestAccFunctionResource_ImportMinimalPlansClean - the server's JSON formatting differs from jsonencode's, costing one apply after import",
		"output_schema_name":         "not server-assigned",
		"print_page_split_config":    "not server-assigned",
		"render_config":              "not server-assigned",
		"s3_bucket":                  "not server-assigned",
		"s3_prefix":                  "not server-assigned",
		"semantic_page_split_config": "not server-assigned",
		"shaping_schema":             "not server-assigned",
		"split_type":                 "not server-assigned",
		"webhook_signing_enabled":    "not server-assigned",
		"webhook_url":                "not server-assigned",

		// OPEN, and unlike extra_config this one is a viable Optional+Computed
		// candidate. The API assigns it for a parse function, so an imported parse
		// function whose HCL omits it never plans clean.
		//
		// Probed with the attribute omitted: a BARE GET returns
		// {extractEntities: true, linkAcrossDocuments: true} - real defaults the model
		// can hold, so adopting the server's value is coherent. No gating flag needed.
		//
		// Needs the two mechanical changes (customfield.NestedObject model field,
		// Computed in the schema) plus the object branch of null-prior-state adoption
		// in preserveServerDefaults. The trap: linkAcrossDocuments defaults to TRUE,
		// so adopt what the API returns and never a synthesised zero value.
		"parse_config": "server-assigned - viable Optional+Computed candidate, see comment",
	}

	noRefresh := map[string]bool{}
	modelType := reflect.TypeOf(FunctionModel{})
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		name := field.Tag.Get("tfsdk")
		for _, part := range strings.Split(tag, ",")[1:] {
			if part == "no_refresh" {
				noRefresh[name] = true
			}
		}
	}

	var found []string
	for name, attribute := range rschema.Attributes {
		if !attribute.IsOptional() || attribute.IsComputed() || attribute.IsRequired() {
			continue
		}
		if !noRefresh[name] {
			continue
		}
		found = append(found, name)
	}
	sort.Strings(found)

	for _, name := range found {
		if _, ok := known[name]; !ok {
			t.Errorf("%s is Optional-without-Computed and carries no_refresh, and is not recorded "+
				"in this audit. Check whether the API assigns it a value when the configuration "+
				"omits it: if it does, an imported resource will never plan clean and the "+
				"attribute needs Computed. Record the answer here either way.", name)
		}
	}
	for name := range known {
		if !contains(found, name) {
			t.Errorf("%s is recorded here as Optional-without-Computed + no_refresh but no longer "+
				"is. If it gained Computed, the post-import problem is solved for it and this "+
				"entry should move to mustBeComputed.", name)
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
