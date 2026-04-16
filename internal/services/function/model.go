// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"github.com/bem-team/terraform-provider-bem/internal/apijson"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/sjson"
)

type FunctionModel struct {
	ID                      types.String                                    `tfsdk:"id" json:"-,computed"`
	PathFunctionName        types.String                                    `tfsdk:"path_function_name" path:"functionName,optional"`
	FunctionName            types.String                                    `tfsdk:"function_name" json:"functionName,required,no_refresh"`
	Type                    types.String                                    `tfsdk:"type" json:"type,required,no_refresh"`
	Description             types.String                                    `tfsdk:"description" json:"description,optional,no_refresh"`
	DestinationType         types.String                                    `tfsdk:"destination_type" json:"destinationType,optional,no_refresh"`
	DisplayName             types.String                                    `tfsdk:"display_name" json:"displayName,optional,no_refresh"`
	EnableBoundingBoxes     types.Bool                                      `tfsdk:"enable_bounding_boxes" json:"enableBoundingBoxes,optional,no_refresh"`
	GoogleDriveFolderID     types.String                                    `tfsdk:"google_drive_folder_id" json:"googleDriveFolderId,optional,no_refresh"`
	JoinType                types.String                                    `tfsdk:"join_type" json:"joinType,optional,no_refresh"`
	OutputSchemaName        types.String                                    `tfsdk:"output_schema_name" json:"outputSchemaName,optional,no_refresh"`
	PreCount                types.Bool                                      `tfsdk:"pre_count" json:"preCount,optional,no_refresh"`
	S3Bucket                types.String                                    `tfsdk:"s3_bucket" json:"s3Bucket,optional,no_refresh"`
	S3Prefix                types.String                                    `tfsdk:"s3_prefix" json:"s3Prefix,optional,no_refresh"`
	ShapingSchema           types.String                                    `tfsdk:"shaping_schema" json:"shapingSchema,optional,no_refresh"`
	SplitType               types.String                                    `tfsdk:"split_type" json:"splitType,optional,no_refresh"`
	TabularChunkingEnabled  types.Bool                                      `tfsdk:"tabular_chunking_enabled" json:"tabularChunkingEnabled,optional,no_refresh"`
	WebhookSigningEnabled   types.Bool                                      `tfsdk:"webhook_signing_enabled" json:"webhookSigningEnabled,optional,no_refresh"`
	WebhookURL              types.String                                    `tfsdk:"webhook_url" json:"webhookUrl,optional,no_refresh"`
	Tags                    *[]types.String                                 `tfsdk:"tags" json:"tags,optional,no_refresh"`
	Config                  *FunctionConfigModel                            `tfsdk:"config" json:"config,optional,no_refresh"`
	PrintPageSplitConfig    *FunctionPrintPageSplitConfigModel              `tfsdk:"print_page_split_config" json:"printPageSplitConfig,optional,no_refresh"`
	Routes                  *[]*FunctionRoutesModel                         `tfsdk:"routes" json:"routes,optional,no_refresh"`
	SemanticPageSplitConfig *FunctionSemanticPageSplitConfigModel           `tfsdk:"semantic_page_split_config" json:"semanticPageSplitConfig,optional,no_refresh"`
	OutputSchema            jsontypes.Normalized                            `tfsdk:"output_schema" json:"outputSchema,optional,no_refresh"`
	Function                customfield.NestedObject[FunctionFunctionModel] `tfsdk:"function" json:"function,computed"`
}

func (m FunctionModel) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(m)
}

func (m FunctionModel) MarshalJSONForUpdate(state FunctionModel) (data []byte, err error) {
	data, err = apijson.MarshalForPatch(m, state)
	if err != nil {
		return nil, err
	}
	return sjson.SetBytes(data, "functionName", m.FunctionName.ValueString())
}

type FunctionConfigModel struct {
	Steps *[]*FunctionConfigStepsModel `tfsdk:"steps" json:"steps,required"`
}

type FunctionConfigStepsModel struct {
	CollectionName        types.String  `tfsdk:"collection_name" json:"collectionName,required"`
	SourceField           types.String  `tfsdk:"source_field" json:"sourceField,required"`
	TargetField           types.String  `tfsdk:"target_field" json:"targetField,required"`
	IncludeCosineDistance types.Bool    `tfsdk:"include_cosine_distance" json:"includeCosineDistance,optional"`
	IncludeSubcollections types.Bool    `tfsdk:"include_subcollections" json:"includeSubcollections,optional"`
	ScoreThreshold        types.Float64 `tfsdk:"score_threshold" json:"scoreThreshold,computed_optional"`
	SearchMode            types.String  `tfsdk:"search_mode" json:"searchMode,computed_optional"`
	TopK                  types.Int64   `tfsdk:"top_k" json:"topK,computed_optional"`
}

type FunctionPrintPageSplitConfigModel struct {
	NextFunctionID   types.String `tfsdk:"next_function_id" json:"nextFunctionID,optional"`
	NextFunctionName types.String `tfsdk:"next_function_name" json:"nextFunctionName,optional"`
}

type FunctionRoutesModel struct {
	Name            types.String               `tfsdk:"name" json:"name,required"`
	Description     types.String               `tfsdk:"description" json:"description,optional"`
	FunctionID      types.String               `tfsdk:"function_id" json:"functionID,optional"`
	FunctionName    types.String               `tfsdk:"function_name" json:"functionName,optional"`
	IsErrorFallback types.Bool                 `tfsdk:"is_error_fallback" json:"isErrorFallback,optional"`
	Origin          *FunctionRoutesOriginModel `tfsdk:"origin" json:"origin,optional"`
	Regex           *FunctionRoutesRegexModel  `tfsdk:"regex" json:"regex,optional"`
}

type FunctionRoutesOriginModel struct {
	Email *FunctionRoutesOriginEmailModel `tfsdk:"email" json:"email,optional"`
}

type FunctionRoutesOriginEmailModel struct {
	Patterns *[]types.String `tfsdk:"patterns" json:"patterns,optional"`
}

type FunctionRoutesRegexModel struct {
	Patterns *[]types.String `tfsdk:"patterns" json:"patterns,optional"`
}

type FunctionSemanticPageSplitConfigModel struct {
	ItemClasses *[]*FunctionSemanticPageSplitConfigItemClassesModel `tfsdk:"item_classes" json:"itemClasses,optional"`
}

type FunctionSemanticPageSplitConfigItemClassesModel struct {
	Name             types.String `tfsdk:"name" json:"name,required"`
	Description      types.String `tfsdk:"description" json:"description,optional"`
	NextFunctionID   types.String `tfsdk:"next_function_id" json:"nextFunctionID,optional"`
	NextFunctionName types.String `tfsdk:"next_function_name" json:"nextFunctionName,optional"`
}

type FunctionFunctionModel struct {
	EmailAddress            types.String                                                           `tfsdk:"email_address" json:"emailAddress,computed"`
	FunctionID              types.String                                                           `tfsdk:"function_id" json:"functionID,computed"`
	FunctionName            types.String                                                           `tfsdk:"function_name" json:"functionName,computed"`
	OutputSchema            jsontypes.Normalized                                                   `tfsdk:"output_schema" json:"outputSchema,computed"`
	OutputSchemaName        types.String                                                           `tfsdk:"output_schema_name" json:"outputSchemaName,computed"`
	TabularChunkingEnabled  types.Bool                                                             `tfsdk:"tabular_chunking_enabled" json:"tabularChunkingEnabled,computed"`
	Type                    types.String                                                           `tfsdk:"type" json:"type,computed"`
	VersionNum              types.Int64                                                            `tfsdk:"version_num" json:"versionNum,computed"`
	Audit                   customfield.NestedObject[FunctionFunctionAuditModel]                   `tfsdk:"audit" json:"audit,computed"`
	DisplayName             types.String                                                           `tfsdk:"display_name" json:"displayName,computed"`
	Tags                    customfield.List[types.String]                                         `tfsdk:"tags" json:"tags,computed"`
	UsedInWorkflows         customfield.NestedObjectList[FunctionFunctionUsedInWorkflowsModel]     `tfsdk:"used_in_workflows" json:"usedInWorkflows,computed"`
	EnableBoundingBoxes     types.Bool                                                             `tfsdk:"enable_bounding_boxes" json:"enableBoundingBoxes,computed"`
	PreCount                types.Bool                                                             `tfsdk:"pre_count" json:"preCount,computed"`
	Description             types.String                                                           `tfsdk:"description" json:"description,computed"`
	Routes                  customfield.NestedObjectList[FunctionFunctionRoutesModel]              `tfsdk:"routes" json:"routes,computed"`
	DestinationType         types.String                                                           `tfsdk:"destination_type" json:"destinationType,computed"`
	GoogleDriveFolderID     types.String                                                           `tfsdk:"google_drive_folder_id" json:"googleDriveFolderId,computed"`
	S3Bucket                types.String                                                           `tfsdk:"s3_bucket" json:"s3Bucket,computed"`
	S3Prefix                types.String                                                           `tfsdk:"s3_prefix" json:"s3Prefix,computed"`
	WebhookSigningEnabled   types.Bool                                                             `tfsdk:"webhook_signing_enabled" json:"webhookSigningEnabled,computed"`
	WebhookURL              types.String                                                           `tfsdk:"webhook_url" json:"webhookUrl,computed"`
	SplitType               types.String                                                           `tfsdk:"split_type" json:"splitType,computed"`
	PrintPageSplitConfig    customfield.NestedObject[FunctionFunctionPrintPageSplitConfigModel]    `tfsdk:"print_page_split_config" json:"printPageSplitConfig,computed"`
	SemanticPageSplitConfig customfield.NestedObject[FunctionFunctionSemanticPageSplitConfigModel] `tfsdk:"semantic_page_split_config" json:"semanticPageSplitConfig,computed"`
	JoinType                types.String                                                           `tfsdk:"join_type" json:"joinType,computed"`
	ShapingSchema           types.String                                                           `tfsdk:"shaping_schema" json:"shapingSchema,computed"`
	Config                  customfield.NestedObject[FunctionFunctionConfigModel]                  `tfsdk:"config" json:"config,computed"`
}

type FunctionFunctionAuditModel struct {
	FunctionCreatedBy     customfield.NestedObject[FunctionFunctionAuditFunctionCreatedByModel]     `tfsdk:"function_created_by" json:"functionCreatedBy,computed"`
	FunctionLastUpdatedBy customfield.NestedObject[FunctionFunctionAuditFunctionLastUpdatedByModel] `tfsdk:"function_last_updated_by" json:"functionLastUpdatedBy,computed"`
	VersionCreatedBy      customfield.NestedObject[FunctionFunctionAuditVersionCreatedByModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
}

type FunctionFunctionAuditFunctionCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionAuditFunctionLastUpdatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionAuditVersionCreatedByModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionUsedInWorkflowsModel struct {
	CurrentVersionNum         types.Int64                   `tfsdk:"current_version_num" json:"currentVersionNum,computed"`
	UsedInWorkflowVersionNums customfield.List[types.Int64] `tfsdk:"used_in_workflow_version_nums" json:"usedInWorkflowVersionNums,computed"`
	WorkflowID                types.String                  `tfsdk:"workflow_id" json:"workflowID,computed"`
	WorkflowName              types.String                  `tfsdk:"workflow_name" json:"workflowName,computed"`
}

type FunctionFunctionRoutesModel struct {
	Name            types.String                                                `tfsdk:"name" json:"name,computed"`
	Description     types.String                                                `tfsdk:"description" json:"description,computed"`
	FunctionID      types.String                                                `tfsdk:"function_id" json:"functionID,computed"`
	FunctionName    types.String                                                `tfsdk:"function_name" json:"functionName,computed"`
	IsErrorFallback types.Bool                                                  `tfsdk:"is_error_fallback" json:"isErrorFallback,computed"`
	Origin          customfield.NestedObject[FunctionFunctionRoutesOriginModel] `tfsdk:"origin" json:"origin,computed"`
	Regex           customfield.NestedObject[FunctionFunctionRoutesRegexModel]  `tfsdk:"regex" json:"regex,computed"`
}

type FunctionFunctionRoutesOriginModel struct {
	Email customfield.NestedObject[FunctionFunctionRoutesOriginEmailModel] `tfsdk:"email" json:"email,computed"`
}

type FunctionFunctionRoutesOriginEmailModel struct {
	Patterns customfield.List[types.String] `tfsdk:"patterns" json:"patterns,computed"`
}

type FunctionFunctionRoutesRegexModel struct {
	Patterns customfield.List[types.String] `tfsdk:"patterns" json:"patterns,computed"`
}

type FunctionFunctionPrintPageSplitConfigModel struct {
	NextFunctionID types.String `tfsdk:"next_function_id" json:"nextFunctionID,computed"`
}

type FunctionFunctionSemanticPageSplitConfigModel struct {
	ItemClasses customfield.NestedObjectList[FunctionFunctionSemanticPageSplitConfigItemClassesModel] `tfsdk:"item_classes" json:"itemClasses,computed"`
}

type FunctionFunctionSemanticPageSplitConfigItemClassesModel struct {
	Name             types.String `tfsdk:"name" json:"name,computed"`
	Description      types.String `tfsdk:"description" json:"description,computed"`
	NextFunctionID   types.String `tfsdk:"next_function_id" json:"nextFunctionID,computed"`
	NextFunctionName types.String `tfsdk:"next_function_name" json:"nextFunctionName,computed"`
}

type FunctionFunctionConfigModel struct {
	Steps customfield.NestedObjectList[FunctionFunctionConfigStepsModel] `tfsdk:"steps" json:"steps,computed"`
}

type FunctionFunctionConfigStepsModel struct {
	CollectionName        types.String  `tfsdk:"collection_name" json:"collectionName,computed"`
	SourceField           types.String  `tfsdk:"source_field" json:"sourceField,computed"`
	TargetField           types.String  `tfsdk:"target_field" json:"targetField,computed"`
	IncludeCosineDistance types.Bool    `tfsdk:"include_cosine_distance" json:"includeCosineDistance,computed"`
	IncludeSubcollections types.Bool    `tfsdk:"include_subcollections" json:"includeSubcollections,computed"`
	ScoreThreshold        types.Float64 `tfsdk:"score_threshold" json:"scoreThreshold,computed"`
	SearchMode            types.String  `tfsdk:"search_mode" json:"searchMode,computed"`
	TopK                  types.Int64   `tfsdk:"top_k" json:"topK,computed"`
}
