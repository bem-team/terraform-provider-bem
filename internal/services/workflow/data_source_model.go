// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkflowDataSourceModel struct {
	WorkflowName types.String                                              `tfsdk:"workflow_name" path:"workflowName,required"`
	Error        types.String                                              `tfsdk:"error" json:"error,computed"`
	Workflow     customfield.NestedObject[WorkflowWorkflowDataSourceModel] `tfsdk:"workflow" json:"workflow,computed"`
}

type WorkflowWorkflowDataSourceModel struct {
	ID           types.String                                                            `tfsdk:"id" json:"id,computed"`
	Connectors   customfield.NestedObjectList[WorkflowWorkflowConnectorsDataSourceModel] `tfsdk:"connectors" json:"connectors,computed"`
	CreatedAt    timetypes.RFC3339                                                       `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	Edges        customfield.NestedObjectList[WorkflowWorkflowEdgesDataSourceModel]      `tfsdk:"edges" json:"edges,computed"`
	MainNodeName types.String                                                            `tfsdk:"main_node_name" json:"mainNodeName,computed"`
	Name         types.String                                                            `tfsdk:"name" json:"name,computed"`
	Nodes        customfield.NestedObjectList[WorkflowWorkflowNodesDataSourceModel]      `tfsdk:"nodes" json:"nodes,computed"`
	UpdatedAt    timetypes.RFC3339                                                       `tfsdk:"updated_at" json:"updatedAt,computed" format:"date-time"`
	VersionNum   types.Int64                                                             `tfsdk:"version_num" json:"versionNum,computed"`
	Audit        customfield.NestedObject[WorkflowWorkflowAuditDataSourceModel]          `tfsdk:"audit" json:"audit,computed"`
	DisplayName  types.String                                                            `tfsdk:"display_name" json:"displayName,computed"`
	EmailAddress types.String                                                            `tfsdk:"email_address" json:"emailAddress,computed"`
	Tags         customfield.List[types.String]                                          `tfsdk:"tags" json:"tags,computed"`
}

type WorkflowWorkflowConnectorsDataSourceModel struct {
	ConnectorID types.String                                                               `tfsdk:"connector_id" json:"connectorID,computed"`
	Name        types.String                                                               `tfsdk:"name" json:"name,computed"`
	Type        types.String                                                               `tfsdk:"type" json:"type,computed"`
	Paragon     customfield.NestedObject[WorkflowWorkflowConnectorsParagonDataSourceModel] `tfsdk:"paragon" json:"paragon,computed"`
}

type WorkflowWorkflowConnectorsParagonDataSourceModel struct {
	Configuration jsontypes.Normalized `tfsdk:"configuration" json:"configuration,computed"`
	Integration   types.String         `tfsdk:"integration" json:"integration,computed"`
	SyncID        types.String         `tfsdk:"sync_id" json:"syncID,computed"`
}

type WorkflowWorkflowEdgesDataSourceModel struct {
	DestinationNodeName types.String         `tfsdk:"destination_node_name" json:"destinationNodeName,computed"`
	SourceNodeName      types.String         `tfsdk:"source_node_name" json:"sourceNodeName,computed"`
	DestinationName     types.String         `tfsdk:"destination_name" json:"destinationName,computed"`
	Metadata            jsontypes.Normalized `tfsdk:"metadata" json:"metadata,computed"`
}

type WorkflowWorkflowNodesDataSourceModel struct {
	Function customfield.NestedObject[WorkflowWorkflowNodesFunctionDataSourceModel] `tfsdk:"function" json:"function,computed"`
	Name     types.String                                                           `tfsdk:"name" json:"name,computed"`
	Metadata jsontypes.Normalized                                                   `tfsdk:"metadata" json:"metadata,computed"`
}

type WorkflowWorkflowNodesFunctionDataSourceModel struct {
	ID         types.String `tfsdk:"id" json:"id,computed"`
	Name       types.String `tfsdk:"name" json:"name,computed"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,computed"`
}

type WorkflowWorkflowAuditDataSourceModel struct {
	VersionCreatedBy      customfield.NestedObject[WorkflowWorkflowAuditVersionCreatedByDataSourceModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
	WorkflowCreatedBy     customfield.NestedObject[WorkflowWorkflowAuditWorkflowCreatedByDataSourceModel]     `tfsdk:"workflow_created_by" json:"workflowCreatedBy,computed"`
	WorkflowLastUpdatedBy customfield.NestedObject[WorkflowWorkflowAuditWorkflowLastUpdatedByDataSourceModel] `tfsdk:"workflow_last_updated_by" json:"workflowLastUpdatedBy,computed"`
}

type WorkflowWorkflowAuditVersionCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowWorkflowAuditWorkflowCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowWorkflowAuditWorkflowLastUpdatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}
