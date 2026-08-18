package function

import "testing"

// Guards the query parameter that makes extraConfig visible at all.
//
// The API omits `extraConfig` unless `includeExtraSettings=true` is requested.
// Codegen emits an empty bem.FunctionGetParams{} for Read and ImportState, so a
// regen of resource.go silently reverts this and `function.extra_config` goes
// back to being null on every refresh - with no test failing anywhere else,
// because every other assertion in the suite is about attributes the API returns
// unconditionally.
func TestFunctionGetParams_RequestsExtraConfig(t *testing.T) {
	params := functionGetParams()

	if !params.IncludeExtraSettings.Valid() {
		t.Fatal("IncludeExtraSettings is unset, so the API will omit extraConfig: " +
			"`function.extra_config` will be null on every refresh, and top-level " +
			"`extra_config` null after every import. If a codegen run reset the call " +
			"sites in resource.go to bem.FunctionGetParams{}, re-apply functionGetParams().")
	}
	if !params.IncludeExtraSettings.Or(false) {
		t.Error("IncludeExtraSettings is false; it must be true to request extraConfig")
	}
}
