// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.ResourceWithConfigure = (*WorkflowResource)(nil)
var _ resource.ResourceWithModifyPlan = (*WorkflowResource)(nil)
var _ resource.ResourceWithImportState = (*WorkflowResource)(nil)

func NewResource() resource.Resource {
	return &WorkflowResource{}
}

// WorkflowResource defines the resource implementation.
type WorkflowResource struct {
	client *bem.Client
}

func (r *WorkflowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *WorkflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkflowModel

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
	env := WorkflowWorkflowEnvelope{*data}
	_, err = r.client.Workflows.New(
		ctx,
		bem.WorkflowNewParams{},
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	err = apijson.UnmarshalComputed(bytes, &env)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data = &env.Workflow
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkflowModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state *WorkflowModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dataBytes, err := data.MarshalJSONForUpdate(*state)
	if err != nil {
		resp.Diagnostics.AddError("failed to serialize http request", err.Error())
		return
	}
	res := new(http.Response)
	_, err = r.client.Workflows.Update(
		ctx,
		data.Name.ValueString(),
		bem.WorkflowUpdateParams{},
		option.WithRequestBody("application/json", dataBytes),
		option.WithResponseBodyInto(&res),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	// The API wraps its response as {"workflow": {...}}, so it has to be
	// decoded through the envelope - exactly as Create, Read and ImportState
	// already do. Decoding straight into the model instead makes the response's
	// "workflow" key land on WorkflowModel.Workflow (the computed nested
	// attribute, which also claims json:"workflow") and leaves every top-level
	// computed attribute - version_num, created_at, updated_at, email_address,
	// restricted, audit - at null, because their keys live one level down
	// inside the envelope.
	//
	// The two paths therefore populated disjoint halves of the model, and each
	// apply flipped which half was set: Create wrote the top-level attributes,
	// Update nulled them and wrote the nested object, so the next plan saw
	// nulls and proposed another update, forever. Confirmed against a real
	// response - the inner object has no nested "workflow" key of its own.
	err = hydrateWorkflowResponse(bytes, data, apijson.UnmarshalComputed)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkflowModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res := new(http.Response)
	_, err := r.client.Workflows.Get(
		ctx,
		data.Name.ValueString(),
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
	// Two passes, same as Create and Update. An envelope-only decode refreshes
	// the top-level attributes and nulls the read-only `workflow` mirror, because
	// the mirror is tagged json:"workflow" and so matches the wrapper the envelope
	// consumes. Read runs on every plan, so that made the mirror null in state
	// almost all of the time - it survived only until the next refresh - and
	// configurations read workflow.version_num from it. Caught by
	// TestAccWorkflowResource_ImportThenUpdate's ImportStateCheck, since
	// `terraform import` performs a refresh immediately after ImportState.
	err = hydrateWorkflowResponse(bytes, data, apijson.Unmarshal)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkflowModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.Delete(
		ctx,
		data.Name.ValueString(),
		option.WithMiddleware(logging.Middleware(ctx)),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to make http request", err.Error())
		return
	}
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data = new(WorkflowModel)

	path := ""
	diags := importpath.ParseImportID(
		req.ID,
		"<workflow_name>",
		&path,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Name = types.StringValue(path)

	res := new(http.Response)
	_, err := r.client.Workflows.Get(
		ctx,
		path,
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
	// Via hydrateWorkflowResponse rather than the envelope alone, so the
	// read-only `workflow` mirror is populated too - an envelope-only decode
	// leaves it null, and configurations read workflow.version_num from it.
	err = hydrateWorkflowResponse(bytes, data, apijson.UnmarshalForImport)
	if err != nil {
		resp.Diagnostics.AddError("failed to deserialize http request", err.Error())
		return
	}
	// The import ID is authoritative for the identifier, whatever the response
	// carried.
	data.Name = types.StringValue(path)
	data.ID = data.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkflowResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plan: nothing to modify.
	if req.Plan.Raw.IsNull() {
		return
	}
	// Create plan: everything is new, nothing to preserve.
	if req.State.Raw.IsNull() {
		return
	}

	// Same treatment as bem_function - see collapseNoOpPlan in
	// function/noop_plan.go for the full explanation. bem_workflow has the same
	// shape of exposure: `workflow` and `audit` are Computed-only nested
	// attributes that core proposes as null, so a plan touching nothing
	// configured still wants an update.
	//
	// The guard is what keeps a real update working: when a referenced
	// function's id or version_num is unknown because that function is
	// changing, `nodes` differs from state and the plan is left alone, so
	// BEM-1392's DAG propagation is unaffected.
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
