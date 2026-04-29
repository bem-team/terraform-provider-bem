// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"context"

	"github.com/bem-team/bem-go-sdk"
	"github.com/bem-team/bem-go-sdk/packages/param"
	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FunctionDataSourceModel struct {
	ID           types.String                                              `tfsdk:"id" path:"functionName,computed"`
	FunctionName types.String                                              `tfsdk:"function_name" path:"functionName,optional"`
	Function     customfield.NestedObject[FunctionFunctionDataSourceModel] `tfsdk:"function" json:"function,computed"`
	FindOneBy    *FunctionFindOneByDataSourceModel                         `tfsdk:"find_one_by"`
}

func (m *FunctionDataSourceModel) toListParams(_ context.Context) (params bem.FunctionListParams, diags diag.Diagnostics) {
	mFindOneByFunctionIDs := []string{}
	if m.FindOneBy.FunctionIDs != nil {
		for _, item := range *m.FindOneBy.FunctionIDs {
			mFindOneByFunctionIDs = append(mFindOneByFunctionIDs, item.ValueString())
		}
	}
	mFindOneByFunctionNames := []string{}
	if m.FindOneBy.FunctionNames != nil {
		for _, item := range *m.FindOneBy.FunctionNames {
			mFindOneByFunctionNames = append(mFindOneByFunctionNames, item.ValueString())
		}
	}
	mFindOneByTags := []string{}
	if m.FindOneBy.Tags != nil {
		for _, item := range *m.FindOneBy.Tags {
			mFindOneByTags = append(mFindOneByTags, item.ValueString())
		}
	}
	mFindOneByTypes := []bem.FunctionType{}
	if m.FindOneBy.Types != nil {
		for _, item := range *m.FindOneBy.Types {
			mFindOneByTypes = append(mFindOneByTypes, bem.FunctionType(item.ValueString()))
		}
	}
	mFindOneByWorkflowIDs := []string{}
	if m.FindOneBy.WorkflowIDs != nil {
		for _, item := range *m.FindOneBy.WorkflowIDs {
			mFindOneByWorkflowIDs = append(mFindOneByWorkflowIDs, item.ValueString())
		}
	}
	mFindOneByWorkflowNames := []string{}
	if m.FindOneBy.WorkflowNames != nil {
		for _, item := range *m.FindOneBy.WorkflowNames {
			mFindOneByWorkflowNames = append(mFindOneByWorkflowNames, item.ValueString())
		}
	}

	params = bem.FunctionListParams{
		FunctionIDs:   mFindOneByFunctionIDs,
		FunctionNames: mFindOneByFunctionNames,
		Tags:          mFindOneByTags,
		Types:         mFindOneByTypes,
		WorkflowIDs:   mFindOneByWorkflowIDs,
		WorkflowNames: mFindOneByWorkflowNames,
	}

	if !m.FindOneBy.DisplayName.IsNull() {
		params.DisplayName = param.NewOpt(m.FindOneBy.DisplayName.ValueString())
	}
	if !m.FindOneBy.SortOrder.IsNull() {
		params.SortOrder = bem.FunctionListParamsSortOrder(m.FindOneBy.SortOrder.ValueString())
	}

	return
}

type FunctionFunctionDataSourceModel struct {
	EmailAddress            types.String                                                                     `tfsdk:"email_address" json:"emailAddress,computed"`
	FunctionID              types.String                                                                     `tfsdk:"function_id" json:"functionID,computed"`
	FunctionName            types.String                                                                     `tfsdk:"function_name" json:"functionName,computed"`
	OutputSchema            jsontypes.Normalized                                                             `tfsdk:"output_schema" json:"outputSchema,computed"`
	OutputSchemaName        types.String                                                                     `tfsdk:"output_schema_name" json:"outputSchemaName,computed"`
	TabularChunkingEnabled  types.Bool                                                                       `tfsdk:"tabular_chunking_enabled" json:"tabularChunkingEnabled,computed"`
	Type                    types.String                                                                     `tfsdk:"type" json:"type,computed"`
	VersionNum              types.Int64                                                                      `tfsdk:"version_num" json:"versionNum,computed"`
	Audit                   customfield.NestedObject[FunctionFunctionAuditDataSourceModel]                   `tfsdk:"audit" json:"audit,computed"`
	DisplayName             types.String                                                                     `tfsdk:"display_name" json:"displayName,computed"`
	Tags                    customfield.List[types.String]                                                   `tfsdk:"tags" json:"tags,computed"`
	UsedInWorkflows         customfield.NestedObjectList[FunctionFunctionUsedInWorkflowsDataSourceModel]     `tfsdk:"used_in_workflows" json:"usedInWorkflows,computed"`
	EnableBoundingBoxes     types.Bool                                                                       `tfsdk:"enable_bounding_boxes" json:"enableBoundingBoxes,computed"`
	PreCount                types.Bool                                                                       `tfsdk:"pre_count" json:"preCount,computed"`
	Classifications         customfield.NestedObjectList[FunctionFunctionClassificationsDataSourceModel]     `tfsdk:"classifications" json:"classifications,computed"`
	Description             types.String                                                                     `tfsdk:"description" json:"description,computed"`
	DestinationType         types.String                                                                     `tfsdk:"destination_type" json:"destinationType,computed"`
	GoogleDriveFolderID     types.String                                                                     `tfsdk:"google_drive_folder_id" json:"googleDriveFolderId,computed"`
	S3Bucket                types.String                                                                     `tfsdk:"s3_bucket" json:"s3Bucket,computed"`
	S3Prefix                types.String                                                                     `tfsdk:"s3_prefix" json:"s3Prefix,computed"`
	WebhookSigningEnabled   types.Bool                                                                       `tfsdk:"webhook_signing_enabled" json:"webhookSigningEnabled,computed"`
	WebhookURL              types.String                                                                     `tfsdk:"webhook_url" json:"webhookUrl,computed"`
	SplitType               types.String                                                                     `tfsdk:"split_type" json:"splitType,computed"`
	PrintPageSplitConfig    customfield.NestedObject[FunctionFunctionPrintPageSplitConfigDataSourceModel]    `tfsdk:"print_page_split_config" json:"printPageSplitConfig,computed"`
	SemanticPageSplitConfig customfield.NestedObject[FunctionFunctionSemanticPageSplitConfigDataSourceModel] `tfsdk:"semantic_page_split_config" json:"semanticPageSplitConfig,computed"`
	JoinType                types.String                                                                     `tfsdk:"join_type" json:"joinType,computed"`
	ShapingSchema           types.String                                                                     `tfsdk:"shaping_schema" json:"shapingSchema,computed"`
	Config                  customfield.NestedObject[FunctionFunctionConfigDataSourceModel]                  `tfsdk:"config" json:"config,computed"`
	ParseConfig             customfield.NestedObject[FunctionFunctionParseConfigDataSourceModel]             `tfsdk:"parse_config" json:"parseConfig,computed"`
}

type FunctionFunctionAuditDataSourceModel struct {
	FunctionCreatedBy     customfield.NestedObject[FunctionFunctionAuditFunctionCreatedByDataSourceModel]     `tfsdk:"function_created_by" json:"functionCreatedBy,computed"`
	FunctionLastUpdatedBy customfield.NestedObject[FunctionFunctionAuditFunctionLastUpdatedByDataSourceModel] `tfsdk:"function_last_updated_by" json:"functionLastUpdatedBy,computed"`
	VersionCreatedBy      customfield.NestedObject[FunctionFunctionAuditVersionCreatedByDataSourceModel]      `tfsdk:"version_created_by" json:"versionCreatedBy,computed"`
}

type FunctionFunctionAuditFunctionCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionAuditFunctionLastUpdatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionAuditVersionCreatedByDataSourceModel struct {
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at" json:"createdAt,computed" format:"date-time"`
	UserActionID types.String      `tfsdk:"user_action_id" json:"userActionID,computed"`
	APIKeyName   types.String      `tfsdk:"api_key_name" json:"apiKeyName,computed"`
	EmailAddress types.String      `tfsdk:"email_address" json:"emailAddress,computed"`
	UserEmail    types.String      `tfsdk:"user_email" json:"userEmail,computed"`
	UserID       types.String      `tfsdk:"user_id" json:"userID,computed"`
}

type FunctionFunctionUsedInWorkflowsDataSourceModel struct {
	CurrentVersionNum         types.Int64                   `tfsdk:"current_version_num" json:"currentVersionNum,computed"`
	UsedInWorkflowVersionNums customfield.List[types.Int64] `tfsdk:"used_in_workflow_version_nums" json:"usedInWorkflowVersionNums,computed"`
	WorkflowID                types.String                  `tfsdk:"workflow_id" json:"workflowID,computed"`
	WorkflowName              types.String                  `tfsdk:"workflow_name" json:"workflowName,computed"`
}

type FunctionFunctionClassificationsDataSourceModel struct {
	Name            types.String                                                                   `tfsdk:"name" json:"name,computed"`
	Description     types.String                                                                   `tfsdk:"description" json:"description,computed"`
	FunctionID      types.String                                                                   `tfsdk:"function_id" json:"functionID,computed"`
	FunctionName    types.String                                                                   `tfsdk:"function_name" json:"functionName,computed"`
	IsErrorFallback types.Bool                                                                     `tfsdk:"is_error_fallback" json:"isErrorFallback,computed"`
	Origin          customfield.NestedObject[FunctionFunctionClassificationsOriginDataSourceModel] `tfsdk:"origin" json:"origin,computed"`
	Regex           customfield.NestedObject[FunctionFunctionClassificationsRegexDataSourceModel]  `tfsdk:"regex" json:"regex,computed"`
}

type FunctionFunctionClassificationsOriginDataSourceModel struct {
	Email customfield.NestedObject[FunctionFunctionClassificationsOriginEmailDataSourceModel] `tfsdk:"email" json:"email,computed"`
}

type FunctionFunctionClassificationsOriginEmailDataSourceModel struct {
	Patterns customfield.List[types.String] `tfsdk:"patterns" json:"patterns,computed"`
}

type FunctionFunctionClassificationsRegexDataSourceModel struct {
	Patterns customfield.List[types.String] `tfsdk:"patterns" json:"patterns,computed"`
}

type FunctionFunctionPrintPageSplitConfigDataSourceModel struct {
	NextFunctionID types.String `tfsdk:"next_function_id" json:"nextFunctionID,computed"`
}

type FunctionFunctionSemanticPageSplitConfigDataSourceModel struct {
	ItemClasses customfield.NestedObjectList[FunctionFunctionSemanticPageSplitConfigItemClassesDataSourceModel] `tfsdk:"item_classes" json:"itemClasses,computed"`
}

type FunctionFunctionSemanticPageSplitConfigItemClassesDataSourceModel struct {
	Name             types.String `tfsdk:"name" json:"name,computed"`
	Description      types.String `tfsdk:"description" json:"description,computed"`
	NextFunctionID   types.String `tfsdk:"next_function_id" json:"nextFunctionID,computed"`
	NextFunctionName types.String `tfsdk:"next_function_name" json:"nextFunctionName,computed"`
}

type FunctionFunctionConfigDataSourceModel struct {
	Steps customfield.NestedObjectList[FunctionFunctionConfigStepsDataSourceModel] `tfsdk:"steps" json:"steps,computed"`
}

type FunctionFunctionConfigStepsDataSourceModel struct {
	CollectionName        types.String  `tfsdk:"collection_name" json:"collectionName,computed"`
	SourceField           types.String  `tfsdk:"source_field" json:"sourceField,computed"`
	TargetField           types.String  `tfsdk:"target_field" json:"targetField,computed"`
	IncludeScore          types.Bool    `tfsdk:"include_score" json:"includeScore,computed"`
	IncludeSubcollections types.Bool    `tfsdk:"include_subcollections" json:"includeSubcollections,computed"`
	ScoreThreshold        types.Float64 `tfsdk:"score_threshold" json:"scoreThreshold,computed"`
	SearchMode            types.String  `tfsdk:"search_mode" json:"searchMode,computed"`
	TopK                  types.Int64   `tfsdk:"top_k" json:"topK,computed"`
}

type FunctionFunctionParseConfigDataSourceModel struct {
	ExtractEntities     types.Bool           `tfsdk:"extract_entities" json:"extractEntities,computed"`
	LinkAcrossDocuments types.Bool           `tfsdk:"link_across_documents" json:"linkAcrossDocuments,computed"`
	Schema              jsontypes.Normalized `tfsdk:"schema" json:"schema,computed"`
}

type FunctionFindOneByDataSourceModel struct {
	DisplayName   types.String    `tfsdk:"display_name" query:"displayName,optional"`
	FunctionIDs   *[]types.String `tfsdk:"function_ids" query:"functionIDs,optional"`
	FunctionNames *[]types.String `tfsdk:"function_names" query:"functionNames,optional"`
	SortOrder     types.String    `tfsdk:"sort_order" query:"sortOrder,computed_optional"`
	Tags          *[]types.String `tfsdk:"tags" query:"tags,optional"`
	Types         *[]types.String `tfsdk:"types" query:"types,optional"`
	WorkflowIDs   *[]types.String `tfsdk:"workflow_ids" query:"workflowIDs,optional"`
	WorkflowNames *[]types.String `tfsdk:"workflow_names" query:"workflowNames,optional"`
}
