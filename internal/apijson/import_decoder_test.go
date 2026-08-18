package apijson

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UnmarshalForImport populates no_refresh fields; Unmarshal must keep skipping
// them.
//
// The no_refresh skip is correct for Read - state already holds those values
// from a previous create or update, and skipping them on refresh is what stops
// server-side normalisation producing a perpetual diff. On import there is no
// prior state to preserve, so the skip protects nothing and leaves nearly every
// writable attribute null: the first plan after `terraform import` shows every
// field changing from null, and `terraform plan -generate-config-out` cannot
// emit config for a required no_refresh attribute at all.

type importModel struct {
	Name        types.String `tfsdk:"name" json:"name,required"`
	DisplayName types.String `tfsdk:"display_name" json:"displayName,optional,no_refresh"`
	MainNode    types.String `tfsdk:"main_node" json:"mainNodeName,required,no_refresh"`
	VersionNum  types.Int64  `tfsdk:"version_num" json:"versionNum,computed,no_refresh"`
	Error       types.String `tfsdk:"error" json:"error,computed"`
}

const importPayload = `{
  "name": "wf",
  "displayName": "My Workflow",
  "mainNodeName": "splitter",
  "versionNum": 3,
  "error": ""
}`

func TestUnmarshalForImport_PopulatesNoRefreshFields(t *testing.T) {
	var model importModel
	if err := UnmarshalForImport([]byte(importPayload), &model); err != nil {
		t.Fatalf("UnmarshalForImport: %v", err)
	}

	if got := model.DisplayName.ValueString(); got != "My Workflow" {
		t.Errorf("display_name = %q, want it populated - a null here is what makes the first "+
			"plan after import show every field changing from null", got)
	}
	if got := model.MainNode.ValueString(); got != "splitter" {
		t.Errorf("main_node = %q, want %q. This one is Required, so a null also breaks "+
			"plan -generate-config-out entirely", got, "splitter")
	}
	if got := model.VersionNum.ValueInt64(); got != 3 {
		t.Errorf("version_num = %d, want 3", got)
	}
}

func TestUnmarshal_StillSkipsNoRefreshFields(t *testing.T) {
	var model importModel
	if err := Unmarshal([]byte(importPayload), &model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !model.DisplayName.IsNull() {
		t.Errorf("display_name = %v on a refresh; no_refresh must still be skipped, or "+
			"server-side normalisation reintroduces perpetual diffs", model.DisplayName)
	}
	if !model.MainNode.IsNull() {
		t.Errorf("main_node = %v on a refresh, want null", model.MainNode)
	}
	// Not tagged no_refresh, so a refresh does update it.
	if model.Error.IsNull() {
		t.Error("error is null; it carries no no_refresh tag and should be refreshed")
	}
}

// The trap this whole change turns on, and the reason ignoreNoRefresh is part of
// decoderEntry.
//
// Decoders are cached in a package-level map keyed by decoderEntry. If the new
// flag were left out of that key, the decoder built for an import would be
// reused for a refresh of the same type - silently making Read stop honouring
// no_refresh, depending only on which ran first in the process. That is the
// cache-poisoning failure class this package has already shipped once
// (BEM-1367), and it would present as "sometimes refresh ignores no_refresh".
//
// Both directions run here, on the same type, in the same process, in both
// orders.
func TestNoRefreshDecoderCacheIsNotShared(t *testing.T) {
	t.Run("import then refresh", func(t *testing.T) {
		var imported importModel
		if err := UnmarshalForImport([]byte(importPayload), &imported); err != nil {
			t.Fatalf("UnmarshalForImport: %v", err)
		}
		if imported.MainNode.IsNull() {
			t.Fatal("import did not populate main_node")
		}

		var refreshed importModel
		if err := Unmarshal([]byte(importPayload), &refreshed); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !refreshed.MainNode.IsNull() {
			t.Errorf("after an import in the same process, a refresh populated main_node (%v). "+
				"The import decoder is being reused for refresh - ignoreNoRefresh is missing from "+
				"decoderEntry.", refreshed.MainNode)
		}
	})

	t.Run("refresh then import", func(t *testing.T) {
		var refreshed importModel
		if err := Unmarshal([]byte(importPayload), &refreshed); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !refreshed.MainNode.IsNull() {
			t.Fatal("refresh populated main_node")
		}

		var imported importModel
		if err := UnmarshalForImport([]byte(importPayload), &imported); err != nil {
			t.Fatalf("UnmarshalForImport: %v", err)
		}
		if imported.MainNode.IsNull() {
			t.Errorf("after a refresh in the same process, an import left main_node null. The " +
				"refresh decoder is being reused for import - ignoreNoRefresh is missing from " +
				"decoderEntry.")
		}
	})
}
