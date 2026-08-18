// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"fmt"

	"github.com/bem-team/terraform-provider-bem/internal/apijson"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type WorkflowWorkflowEnvelope struct {
	Workflow WorkflowModel `json:"workflow"`
}

type WorkflowModel struct {
	ID   types.String `tfsdk:"id" json:"-,computed"`
	Name types.String `tfsdk:"name" json:"name,required"`
	// mainNodeName, nodes, and edges must all be sent together whenever the
	// workflow DAG changes - the API rejects a partial update
	// ("mainNodeName, nodes, and edges must all be provided together when
	// updating the workflow DAG"). atomic_group=dag makes MarshalJSONForUpdate
	// re-send all three together whenever any one of them changed, even if
	// the others are individually unchanged from state.
	MainNodeName    types.String                                               `tfsdk:"main_node_name" json:"mainNodeName,required,no_refresh,atomic_group=dag"`
	Nodes           *[]*WorkflowNodesModel                                     `tfsdk:"nodes" json:"nodes,required,no_refresh,atomic_group=dag"`
	DisplayName     types.String                                               `tfsdk:"display_name" json:"displayName,optional,no_refresh"`
	Tags            *[]types.String                                            `tfsdk:"tags" json:"tags,optional,no_refresh"`
	Connectors      customfield.NestedObjectList[WorkflowConnectorsModel]      `tfsdk:"connectors" json:"connectors,computed_optional,no_refresh"`
	Edges           *[]*WorkflowEdgesModel                                     `tfsdk:"edges" json:"edges,optional,no_refresh,atomic_group=dag"`
	CreatedAt       timetypes.RFC3339                                          `tfsdk:"created_at" json:"createdAt,computed,no_refresh" format:"date-time"`
	EmailAddress    types.String                                               `tfsdk:"email_address" json:"emailAddress,computed,no_refresh"`
	Error           types.String                                               `tfsdk:"error" json:"error,computed"`
	Restricted      types.Bool                                                 `tfsdk:"restricted" json:"restricted,computed,no_refresh"`
	UpdatedAt       timetypes.RFC3339                                          `tfsdk:"updated_at" json:"updatedAt,computed,no_refresh" format:"date-time"`
	VersionNum      types.Int64                                                `tfsdk:"version_num" json:"versionNum,computed,no_refresh"`
	Audit           customfield.NestedObject[WorkflowAuditModel]               `tfsdk:"audit" json:"audit,computed,no_refresh"`
	ConnectorErrors customfield.NestedObjectList[WorkflowConnectorErrorsModel] `tfsdk:"connector_errors" json:"connectorErrors,computed,no_refresh"`
	Workflow        customfield.NestedObject[WorkflowWorkflowModel]            `tfsdk:"workflow" json:"workflow,computed"`
}

func (m WorkflowModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

func (m WorkflowModel) MarshalJSONForUpdate(state WorkflowModel) (data []byte, err error) {
	data, err = apijson.MarshalForPatch(m, state)
	if err != nil {
		return nil, err
	}
	return dropRedundantNodeFunctionID(data)
}

// dropRedundantNodeFunctionID removes a node's function.id when function.name is
// also present, because the API accepts exactly one:
//
//	400 node 0 function is invalid: function identifier must have either an ID or
//	    a name, but not both
//
// Both can be present without the practitioner asking for it. `id` is
// computed_optional, so a configuration referencing a function by name leaves it
// null, the response decode fills it in, and preserveServerDefaults then preserves
// it into the plan alongside the configured name - this provider's own plan pass is
// part of the trigger. Any DAG edit re-sends the whole node through
// atomic_group=dag, with both keys, and the update fails.
//
// Confirmed against the API rather than inferred: name alone 200, id alone 200, both
// 400. Note the reverse case is safe without help - `name` is plain Optional, not
// Computed, so for an id-based configuration the plan resets it to the config's null
// and only `id` is ever sent.
//
// `name` wins because it is the half a name-based configuration set, and it is not
// filled in from a response for an id-based one - so whichever the practitioner
// chose is the one that survives.
func dropRedundantNodeFunctionID(data []byte) ([]byte, error) {
	nodes := gjson.GetBytes(data, "nodes")
	if !nodes.IsArray() {
		return data, nil
	}

	var err error
	nodes.ForEach(func(index, node gjson.Result) bool {
		fn := node.Get("function")
		if !fn.Exists() || !fn.Get("name").Exists() || !fn.Get("id").Exists() {
			return true
		}
		data, err = sjson.DeleteBytes(data, fmt.Sprintf("nodes.%d.function.id", index.Int()))
		return err == nil
	})
	return data, err
}

type WorkflowNodesModel struct {
	Function *WorkflowNodesFunctionModel `tfsdk:"function" json:"function,required"`
	Metadata jsontypes.Normalized        `tfsdk:"metadata" json:"metadata,optional"`
	Name     types.String                `tfsdk:"name" json:"name,optional"`
}

type WorkflowNodesFunctionModel struct {
	ID         types.String `tfsdk:"id" json:"id,computed_optional"`
	Name       types.String `tfsdk:"name" json:"name,optional"`
	VersionNum types.Int64  `tfsdk:"version_num" json:"versionNum,computed_optional"`
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

type WorkflowAuditModel struct {
	VersionCreatedBy      customfield.NestedObject[WorkflowAuditVersionCreatedByModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
	WorkflowCreatedBy     customfield.NestedObject[WorkflowAuditWorkflowCreatedByModel]     `tfsdk:"workflow_created_by" json:"workflowCreatedBy,computed"`
	WorkflowLastUpdatedBy customfield.NestedObject[WorkflowAuditWorkflowLastUpdatedByModel] `tfsdk:"workflow_last_updated_by" json:"workflowLastUpdatedBy,computed"`
}

type WorkflowAuditVersionCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowAuditWorkflowCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type WorkflowAuditWorkflowLastUpdatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
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
	Restricted   types.Bool                                                    `tfsdk:"restricted" json:"restricted,computed"`
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
