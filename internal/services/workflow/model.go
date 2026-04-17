// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"github.com/bem-team/terraform-provider-bem/internal/apijson"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WorkflowModel struct {
	WorkflowName    types.String                                               `tfsdk:"workflow_name" path:"workflowName,optional"`
	MainNodeName    types.String                                               `tfsdk:"main_node_name" json:"mainNodeName,required,no_refresh"`
	Name            types.String                                               `tfsdk:"name" json:"name,required,no_refresh"`
	Nodes           *[]*WorkflowNodesModel                                     `tfsdk:"nodes" json:"nodes,required,no_refresh"`
	DisplayName     types.String                                               `tfsdk:"display_name" json:"displayName,optional,no_refresh"`
	Tags            *[]types.String                                            `tfsdk:"tags" json:"tags,optional,no_refresh"`
	Connectors      *[]*WorkflowConnectorsModel                                `tfsdk:"connectors" json:"connectors,optional,no_refresh"`
	Edges           *[]*WorkflowEdgesModel                                     `tfsdk:"edges" json:"edges,optional,no_refresh"`
	Error           types.String                                               `tfsdk:"error" json:"error,computed"`
	ConnectorErrors customfield.NestedObjectList[WorkflowConnectorErrorsModel] `tfsdk:"connector_errors" json:"connectorErrors,computed,no_refresh"`
	Workflow        customfield.NestedObject[WorkflowWorkflowModel]            `tfsdk:"workflow" json:"workflow,computed"`
}

func (m WorkflowModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

func (m WorkflowModel) MarshalJSONForUpdate(state WorkflowModel) (data []byte, err error) {
	return apijson.MarshalForPatch(m, state)
}

type WorkflowNodesModel struct {
	Function *WorkflowNodesFunctionModel `tfsdk:"function" json:"function,required"`
	Metadata jsontypes.Normalized        `tfsdk:"metadata" json:"metadata,optional"`
	Name     types.String                `tfsdk:"name" json:"name,optional"`
}

type WorkflowNodesFunctionModel struct {
	ID         types.String `tfsdk:"id" json:"id,optional"`
	Name       types.String `tfsdk:"name" json:"name,optional"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,optional"`
}

type WorkflowConnectorsModel struct {
	Name        types.String                    `tfsdk:"name" json:"name,required"`
	Type        types.String                    `tfsdk:"type" json:"type,required"`
	ConnectorID types.String                    `tfsdk:"connector_id" json:"connectorID,optional"`
	Paragon     *WorkflowConnectorsParagonModel `tfsdk:"paragon" json:"paragon,optional"`
}

type WorkflowConnectorsParagonModel struct {
	Configuration jsontypes.Normalized `tfsdk:"configuration" json:"configuration,optional"`
	Integration   types.String         `tfsdk:"integration" json:"integration,optional"`
}

type WorkflowEdgesModel struct {
	DestinationNodeName types.String         `tfsdk:"destination_node_name" json:"destinationNodeName,required"`
	SourceNodeName      types.String         `tfsdk:"source_node_name" json:"sourceNodeName,required"`
	DestinationName     types.String         `tfsdk:"destination_name" json:"destinationName,optional"`
	Metadata            jsontypes.Normalized `tfsdk:"metadata" json:"metadata,optional"`
}

type WorkflowConnectorErrorsModel struct {
	Code        types.String `tfsdk:"code" json:"code,computed"`
	Message     types.String `tfsdk:"message" json:"message,computed"`
	Operation   types.String `tfsdk:"operation" json:"operation,computed"`
	ConnectorID types.String `tfsdk:"connector_id" json:"connectorID,computed"`
	Name        types.String `tfsdk:"name" json:"name,computed"`
}

type WorkflowWorkflowModel struct {
	ID           types.String                                                  `tfsdk:"id" json:"id,computed"`
	Connectors   customfield.NestedObjectList[WorkflowWorkflowConnectorsModel] `tfsdk:"connectors" json:"connectors,computed"`
	CreatedAt    timetypes.RFC3339                                             `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	Edges        customfield.NestedObjectList[WorkflowWorkflowEdgesModel]      `tfsdk:"edges" json:"edges,computed"`
	MainNodeName types.String                                                  `tfsdk:"main_node_name" json:"mainNodeName,computed"`
	Name         types.String                                                  `tfsdk:"name" json:"name,computed"`
	Nodes        customfield.NestedObjectList[WorkflowWorkflowNodesModel]      `tfsdk:"nodes" json:"nodes,computed"`
	UpdatedAt    timetypes.RFC3339                                             `tfsdk:"updated_at" json:"updatedAt,computed" format:"date-time"`
	VersionNum   types.Int64                                                   `tfsdk:"version_num" json:"versionNum,computed"`
	Audit        customfield.NestedObject[WorkflowWorkflowAuditModel]          `tfsdk:"audit" json:"audit,computed"`
	DisplayName  types.String                                                  `tfsdk:"display_name" json:"displayName,computed"`
	EmailAddress types.String                                                  `tfsdk:"email_address" json:"emailAddress,computed"`
	Tags         customfield.List[types.String]                                `tfsdk:"tags" json:"tags,computed"`
}

type WorkflowWorkflowConnectorsModel struct {
	ConnectorID types.String                                                     `tfsdk:"connector_id" json:"connectorID,computed"`
	Name        types.String                                                     `tfsdk:"name" json:"name,computed"`
	Type        types.String                                                     `tfsdk:"type" json:"type,computed"`
	Paragon     customfield.NestedObject[WorkflowWorkflowConnectorsParagonModel] `tfsdk:"paragon" json:"paragon,computed"`
}

type WorkflowWorkflowConnectorsParagonModel struct {
	Configuration jsontypes.Normalized `tfsdk:"configuration" json:"configuration,computed"`
	Integration   types.String         `tfsdk:"integration" json:"integration,computed"`
	SyncID        types.String         `tfsdk:"sync_id" json:"syncID,computed"`
}

type WorkflowWorkflowEdgesModel struct {
	DestinationNodeName types.String         `tfsdk:"destination_node_name" json:"destinationNodeName,computed"`
	SourceNodeName      types.String         `tfsdk:"source_node_name" json:"sourceNodeName,computed"`
	DestinationName     types.String         `tfsdk:"destination_name" json:"destinationName,computed"`
	Metadata            jsontypes.Normalized `tfsdk:"metadata" json:"metadata,computed"`
}

type WorkflowWorkflowNodesModel struct {
	Function customfield.NestedObject[WorkflowWorkflowNodesFunctionModel] `tfsdk:"function" json:"function,computed"`
	Name     types.String                                                 `tfsdk:"name" json:"name,computed"`
	Metadata jsontypes.Normalized                                         `tfsdk:"metadata" json:"metadata,computed"`
}

type WorkflowWorkflowNodesFunctionModel struct {
	ID         types.String `tfsdk:"id" json:"id,computed"`
	Name       types.String `tfsdk:"name" json:"name,computed"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,computed"`
}

type WorkflowWorkflowAuditModel struct {
	VersionCreatedBy      customfield.NestedObject[WorkflowWorkflowAuditVersionCreatedByModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
	WorkflowCreatedBy     customfield.NestedObject[WorkflowWorkflowAuditWorkflowCreatedByModel]     `tfsdk:"workflow_created_by" json:"workflowCreatedBy,computed"`
	WorkflowLastUpdatedBy customfield.NestedObject[WorkflowWorkflowAuditWorkflowLastUpdatedByModel] `tfsdk:"workflow_last_updated_by" json:"workflowLastUpdatedBy,computed"`
}

type WorkflowWorkflowAuditVersionCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowWorkflowAuditWorkflowCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowWorkflowAuditWorkflowLastUpdatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}
