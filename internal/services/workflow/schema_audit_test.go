package workflow

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The bem_workflow half of the BEM-1396 audit. See the bem_function version for
// the full reasoning; in short:
//
//   - `no_refresh` means Read never writes the attribute, so before
//     UnmarshalForImport nothing ever put a server value in state for it. Import
//     now does.
//   - If the API assigns a value the configuration omits, the plan is dirty
//     forever - a new workflow version per apply.
//   - It cannot be repaired in ModifyPlan when the attribute is Optional without
//     Computed: Terraform core rejects a planned value that differs from a
//     non-computed attribute's config value ("planned value ... for a non-computed
//     attribute"), verified by implementing it.
//
// Found server-assigned on bem_workflow, and now Optional+Computed:
//
//	connectors                -> []                        (every workflow)
//	nodes[].function.id       -> resolved from the name
//	nodes[].function.version_num -> resolved from the name
//
// The node ones are the interesting case: a configuration names the function and
// the server resolves the id and version. Before this, an imported workflow
// planned both back to null on every plan, which is also what made
// `ignore_changes = [nodes, edges]` necessary downstream.
func TestAudit_PlainOptionalNoRefreshAttributes(t *testing.T) {
	ctx := context.TODO()
	rschema := ResourceSchema(ctx)

	// `connectors` is Optional+Computed, and it took two regressions to get there.
	//
	// The API returns [] for it, and no_refresh means only an import writes it. So an
	// imported workflow holds [] in state and a created one holds null, with the
	// configuration omitting the attribute either way. Plain Optional breaks the
	// imported case; Computed alone breaks the created case, which is far commoner.
	// It needs Computed *plus* null-prior-state adoption for collections - see
	// preserveServerDefaults, and TestConnectors_BothPriorStateShapesSettle, which
	// asserts both directions because fixing one alone has now broken the other twice.
	if attribute, ok := rschema.Attributes["connectors"]; !ok {
		t.Error("connectors is missing from the schema")
	} else if !attribute.IsComputed() || !attribute.IsOptional() {
		t.Errorf("connectors must be Optional+Computed (optional=%v computed=%v); schema.go "+
			"carries a codegen header, so a regen drops this.",
			attribute.IsOptional(), attribute.IsComputed())
	}

	// Everything still Optional-without-Computed + no_refresh, none of it known to
	// be server-assigned.
	known := map[string]bool{
		"display_name": true,
		"edges":        true,
		"tags":         true,
	}

	noRefresh := map[string]bool{}
	modelType := reflect.TypeOf(WorkflowModel{})
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
		if !known[name] {
			t.Errorf("%s is Optional-without-Computed and carries no_refresh, and is not recorded "+
				"in this audit. Check whether the API assigns it a value when the configuration "+
				"omits it: if it does, an imported workflow will never plan clean and the "+
				"attribute needs Computed. Record the answer here either way.", name)
		}
	}
	for name := range known {
		if !contains(found, name) {
			t.Errorf("%s is recorded here as Optional-without-Computed + no_refresh but no longer "+
				"is. If it gained Computed, the post-import problem is solved for it and this "+
				"entry should move up to the assertion above.", name)
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
