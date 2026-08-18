// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bem-team/bem-go-sdk"
	"github.com/bem-team/bem-go-sdk/option"
	"github.com/bem-team/terraform-provider-bem/internal/apijson"
	"github.com/bem-team/terraform-provider-bem/internal/importpath"
	"github.com/bem-team/terraform-provider-bem/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*FunctionResource)(nil)
var _ resource.ResourceWithModifyPlan = (*FunctionResource)(nil)
var _ resource.ResourceWithImportState = (*FunctionResource)(nil)

func NewResource() resource.Resource {
	return &FunctionResource{}
}

// FunctionResource defines the resource implementation.
type FunctionResource struct {
	client *bem.Client
}

func (r *FunctionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *FunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*bem.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf("Expected *bem.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FunctionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dataBytes, err := data.MarshalJSON()
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize http request", err.Error())
		return
	}
	res := new(http.Response)
	_, err = r.client.Functions.New(
		ctx,
		bem.FunctionNewParams{},
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	// Two passes - see hydrateFromResponse in envelope.go. Decoding straight
	// into the model leaves `config` unhydrated, which is what produced a
	// perpetual in-place update on enrich functions.
	err = hydrateFromResponse(bytes, data, apijson.UnmarshalComputed)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.FunctionName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FunctionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state *FunctionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := bem.FunctionUpdateParams{}

	dataBytes, err := data.MarshalJSONForUpdate(*state)
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize http request", err.Error())
		return
	}
	res := new(http.Response)
	_, err = r.client.Functions.Update(
		ctx,
		state.FunctionName.ValueString(),
		params,
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	// Two passes - see hydrateFromResponse in envelope.go. Decoding straight
	// into the model leaves `config` unhydrated, which is what produced a
	// perpetual in-place update on enrich functions.
	err = hydrateFromResponse(bytes, data, apijson.UnmarshalComputed)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.FunctionName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FunctionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res := new(http.Response)
	_, err := r.client.Functions.Get(
		ctx,
		data.FunctionName.ValueString(),
		functionGetParams(),
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if res != nil && res.StatusCode == 404 {
		resp.Diagnostics.AddWarning("Resource not found", "The resource was not found on the server and will be removed from state.")
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	err = apijson.Unmarshal(bytes, &data)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.FunctionName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FunctionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Functions.Delete(
		ctx,
		data.FunctionName.ValueString(),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data = new(FunctionModel)

	path := ""
	diags := importpath.ParseImportID(
		req.ID,
		"<function_name>",
		&path,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.FunctionName = types.StringValue(path)

	res := new(http.Response)
	_, err := r.client.Functions.Get(
		ctx,
		path,
		// Deliberately NOT functionGetParams(): see the comment there. Requesting
		// extraConfig on the import path hydrates it into state as a non-null
		// object, against a configuration that leaves it unset - and it is plain
		// Optional, so nothing can reconcile that.
		bem.FunctionGetParams{},
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	// UnmarshalForImport, not Unmarshal: import has no prior state, so the
	// no_refresh skip that protects Read from server-side normalisation instead
	// leaves nearly every writable attribute null. See UnmarshalForImport.
	//
	// It must go through the envelope. Once no_refresh fields are no longer
	// skipped, a direct decode actively *nulls* them, because their keys live
	// under the response's "function" wrapper rather than at the root - which
	// broke import outright with "missing required functionName parameter".
	err = hydrateFromResponse(bytes, data, apijson.UnmarshalForImport)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	// The import ID is authoritative for the identifier, whatever the response
	// carried.
	data.FunctionName = types.StringValue(path)
	data.ID = data.FunctionName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plan: nothing to modify.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Create plan: id is already unknown, nothing to do.
	if req.State.Raw.IsNull() {
		return
	}

	var plan, state FunctionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// id always mirrors function_name (see Create/Update/Read/ImportState).
	// The id attribute's UseNonNullStateForUnknown plan modifier otherwise
	// keeps it pinned to the prior value even when function_name is renamed
	// in place, which Terraform's plan/apply consistency check then rejects
	// once Update() actually returns the new id.
	if !plan.FunctionName.Equal(state.FunctionName) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
	}

	// BEM-1396 finding 2, in two passes that must run in this order.
	//
	// 1. Optional+Computed leaves the configuration omits (enrich's
	//    config.steps[].top_k and friends) come back as null from core while
	//    prior state holds the server's default, so `config` differs on every
	//    plan. Restore prior state for those first.
	// 2. Only then is it true that nothing configured changed, which lets the
	//    no-op collapse deal with the purely-computed `function` mirror.
	//
	// Running the collapse first would find `config` differing and correctly
	// decline, which is exactly what happened before this pass existed.
	rschema := ResourceSchema(ctx)

	preserved, err := preserveServerDefaults(ctx, rschema, req.Config, resp.Plan, req.State)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Modifying Planned State",
			"Could not preserve server-assigned defaults for optional attributes. "+
				"This is always a problem with the provider; please report it:\n\n"+err.Error(),
		)
		return
	}
	resp.Plan.Raw = preserved

	// Float64 plan values are quantised by the framework while prior state
	// carries the exact decimal from the API, so an untouched score_threshold
	// reads as changed. Adopt state's representation where the two agree at
	// float64 precision. Root cause of BEM-1396 finding 2.
	normalized, err := normalizeNumberRepresentation(ctx, resp.Plan.Raw, req.State.Raw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Modifying Planned State",
			"Could not normalise numeric plan values. This is always a problem with the "+
				"provider; please report it:\n\n"+err.Error(),
		)
		return
	}
	resp.Plan.Raw = normalized

	resp.Diagnostics.Append(collapseNoOpPlan(ctx, rschema, resp.Plan, req.State, &resp.Plan)...)
}

// functionGetParams is the query for Read, and ONLY for Read.
//
// extraConfig is gated server-side: the API omits it unless
// includeExtraSettings=true is requested. Passing the empty FunctionGetParams
// that codegen emits therefore left the read-only `function.extra_config`
// mirror null on every single plan, for every practitioner - the mirror is
// plain `computed`, so Read does refresh it. Requesting the flag here fixes
// that, and cannot disturb top-level `extra_config`, which is no_refresh and so
// skipped on refresh.
//
// ImportState must NOT use this. It decodes with UnmarshalForImport, which
// deliberately hydrates no_refresh attributes, so requesting extraConfig there
// writes a non-null object into state:
//
//	state: extra_config = {enable_bounding_boxes: null}
//	plan:  - extra_config = {} -> null
//
// against a configuration that leaves the attribute unset. `extra_config` is
// plain Optional, so preserveServerDefaults cannot absorb it and core will not
// accept a planned non-null value for it either - the plan stays dirty forever.
// That regressed ten of eleven consumer fixtures on build 16:25; only the one
// that sets extra_config improved. Import leaving it null is the lesser evil
// until extra_config is Optional+Computed (see the audit test).
//
// The parameter only became expressible in bem-go-sdk 0.28.0; before that
// Functions.Get took no params at all.
//
// Kept as a function rather than inlined because resource.go carries the
// codegen header: a regen reverts the call site to bem.FunctionGetParams{},
// and TestFunctionGetParams_RequestsExtraConfig then fails loudly.
func functionGetParams() bem.FunctionGetParams {
	return bem.FunctionGetParams{IncludeExtraSettings: bem.Bool(true)}
}
