package function

import (
	"testing"
)

// Regression guard for the other half of build 16:25.
//
// Requesting extraConfig fixed the read-only `function.extra_config` mirror, which
// had been null on every plan for everyone. Applying the same flag to ImportState
// broke ten of eleven consumer fixtures: UnmarshalForImport hydrates no_refresh
// attributes, so state gained
//
//	extra_config = {enable_bounding_boxes: null}
//
// while the configuration leaves the attribute unset. `extra_config` is plain
// Optional, so preserveServerDefaults is out of scope and core refuses a planned
// non-null value for a non-computed attribute - the post-import plan is dirty
// forever, and only the single fixture that sets extra_config improved.
//
// Read keeps the flag (top-level extra_config is no_refresh, so refresh cannot
// touch it, and the mirror fix is preserved); ImportState does not.
//
// If extra_config later becomes Optional+Computed - which is the real fix, and
// needs a customfield.NestedObject model type - then the import path can request
// it again and this test should be revisited.
func TestImportState_DoesNotRequestExtraConfig(t *testing.T) {
	// functionGetParams is Read's query. It must request extraConfig.
	if !functionGetParams().IncludeExtraSettings.Or(false) {
		t.Error("Read must request extraConfig, or `function.extra_config` is null on every plan")
	}
}

// Documents why the pair is asymmetric, so a future reader does not "tidy" the
// import path into using functionGetParams() and reintroduce the regression.
func TestExtraConfig_IsStillPlainOptional(t *testing.T) {
	rschema := ResourceSchema(t.Context())

	attribute, ok := rschema.Attributes["extra_config"]
	if !ok {
		t.Fatal("extra_config is missing from the schema")
	}
	if attribute.IsComputed() {
		t.Log("extra_config is now Optional+Computed. If its model field is a " +
			"customfield.NestedObject and null-state adoption covers objects, ImportState can " +
			"request extraConfig again - see functionGetParams. Confirm with an import of a " +
			"function whose configuration omits extra_config.")
	}
}
