// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"context"

	"github.com/bem-team/bem-go-sdk"
	"github.com/bem-team/bem-go-sdk/packages/param"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkflowsWorkflowsListDataSourceEnvelope struct {
	Workflows customfield.NestedObjectList[WorkflowsItemsDataSourceModel] `json:"workflows,computed"`
}

type WorkflowsDataSourceModel struct {
	DisplayName   types.String                                                `tfsdk:"display_name" query:"displayName,optional"`
	FunctionIDs   *[]types.String                                             `tfsdk:"function_ids" query:"functionIDs,optional"`
	FunctionNames *[]types.String                                             `tfsdk:"function_names" query:"functionNames,optional"`
	Tags          *[]types.String                                             `tfsdk:"tags" query:"tags,optional"`
	WorkflowIDs   *[]types.String                                             `tfsdk:"workflow_ids" query:"workflowIDs,optional"`
	WorkflowNames *[]types.String                                             `tfsdk:"workflow_names" query:"workflowNames,optional"`
	SortOrder     types.String                                                `tfsdk:"sort_order" query:"sortOrder,computed_optional"`
	MaxItems      types.Int64                                                 `tfsdk:"max_items"`
	Items         customfield.NestedObjectList[WorkflowsItemsDataSourceModel] `tfsdk:"items"`
}

func (m *WorkflowsDataSourceModel) toListParams(_ context.Context) (params bem.WorkflowListParams, diags diag.Diagnostics) {
	mFunctionIDs := []string{}
	if m.FunctionIDs != nil {
		for _, item := range *m.FunctionIDs {
			mFunctionIDs = append(mFunctionIDs, item.ValueString())
		}
	}
	mFunctionNames := []string{}
	if m.FunctionNames != nil {
		for _, item := range *m.FunctionNames {
			mFunctionNames = append(mFunctionNames, item.ValueString())
		}
	}
	mTags := []string{}
	if m.Tags != nil {
		for _, item := range *m.Tags {
			mTags = append(mTags, item.ValueString())
		}
	}
	mWorkflowIDs := []string{}
	if m.WorkflowIDs != nil {
		for _, item := range *m.WorkflowIDs {
			mWorkflowIDs = append(mWorkflowIDs, item.ValueString())
		}
	}
	mWorkflowNames := []string{}
	if m.WorkflowNames != nil {
		for _, item := range *m.WorkflowNames {
			mWorkflowNames = append(mWorkflowNames, item.ValueString())
		}
	}

	params = bem.WorkflowListParams{
		FunctionIDs:   mFunctionIDs,
		FunctionNames: mFunctionNames,
		Tags:          mTags,
		WorkflowIDs:   mWorkflowIDs,
		WorkflowNames: mWorkflowNames,
	}

	if !m.DisplayName.IsNull() {
		params.DisplayName = param.NewOpt(m.DisplayName.ValueString())
	}
	if !m.SortOrder.IsNull() {
		params.SortOrder = bem.WorkflowListParamsSortOrder(m.SortOrder.ValueString())
	}

	return
}

type WorkflowsItemsDataSourceModel struct {
	ID           types.String                                                `tfsdk:"id" json:"id,computed"`
	CreatedAt    timetypes.RFC3339                                           `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	Edges        customfield.NestedObjectList[WorkflowsEdgesDataSourceModel] `tfsdk:"edges" json:"edges,computed"`
	MainNodeName types.String                                                `tfsdk:"main_node_name" json:"mainNodeName,computed"`
	Name         types.String                                                `tfsdk:"name" json:"name,computed"`
	Nodes        customfield.NestedObjectList[WorkflowsNodesDataSourceModel] `tfsdk:"nodes" json:"nodes,computed"`
	UpdatedAt    timetypes.RFC3339                                           `tfsdk:"updated_at" json:"updatedAt,computed" format:"date-time"`
	VersionNum   types.Int64                                                 `tfsdk:"version_num" json:"versionNum,computed"`
	Audit        customfield.NestedObject[WorkflowsAuditDataSourceModel]     `tfsdk:"audit" json:"audit,computed"`
	DisplayName  types.String                                                `tfsdk:"display_name" json:"displayName,computed"`
	EmailAddress types.String                                                `tfsdk:"email_address" json:"emailAddress,computed"`
	Tags         customfield.List[types.String]                              `tfsdk:"tags" json:"tags,computed"`
}

type WorkflowsEdgesDataSourceModel struct {
	DestinationNodeName types.String `tfsdk:"destination_node_name" json:"destinationNodeName,computed"`
	SourceNodeName      types.String `tfsdk:"source_node_name" json:"sourceNodeName,computed"`
	DestinationName     types.String `tfsdk:"destination_name" json:"destinationName,computed"`
}

type WorkflowsNodesDataSourceModel struct {
	Function customfield.NestedObject[WorkflowsNodesFunctionDataSourceModel] `tfsdk:"function" json:"function,computed"`
	Name     types.String                                                    `tfsdk:"name" json:"name,computed"`
}

type WorkflowsNodesFunctionDataSourceModel struct {
	ID         types.String `tfsdk:"id" json:"id,computed"`
	Name       types.String `tfsdk:"name" json:"name,computed"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,computed"`
}

type WorkflowsAuditDataSourceModel struct {
	VersionCreatedBy      customfield.NestedObject[WorkflowsAuditVersionCreatedByDataSourceModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
	WorkflowCreatedBy     customfield.NestedObject[WorkflowsAuditWorkflowCreatedByDataSourceModel]     `tfsdk:"workflow_created_by" json:"workflowCreatedBy,computed"`
	WorkflowLastUpdatedBy customfield.NestedObject[WorkflowsAuditWorkflowLastUpdatedByDataSourceModel] `tfsdk:"workflow_last_updated_by" json:"workflowLastUpdatedBy,computed"`
}

type WorkflowsAuditVersionCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowsAuditWorkflowCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowsAuditWorkflowLastUpdatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}
