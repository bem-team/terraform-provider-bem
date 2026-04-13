// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"context"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigValidators = (*FunctionDataSource)(nil)

func DataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Functions are the core building blocks of data transformation in Bem. Each function type serves a specific purpose:\n\n- **Transform**: Extract structured JSON data from unstructured documents (PDFs, emails, images)\n- **Analyze**: Perform visual analysis on documents to extract layout-aware information\n- **Route**: Direct data to different processing paths based on conditions\n- **Split**: Break multi-page documents into individual pages for parallel processing\n- **Join**: Combine outputs from multiple function calls into a single result\n- **Payload Shaping**: Transform and restructure data using JMESPath expressions\n- **Enrich**: Enhance data with semantic search against collections\n\nUse these endpoints to create, update, list, and manage your functions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"function_name": schema.StringAttribute{
				Optional: true,
			},
			"function": schema.SingleNestedAttribute{
				Description: "A function that delivers workflow outputs to an external destination.\nSend functions receive the output of an upstream workflow node and forward it\nto a webhook, S3 bucket, or Google Drive folder.",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectType[FunctionFunctionDataSourceModel](ctx),
				Attributes: map[string]schema.Attribute{
					"email_address": schema.StringAttribute{
						Description: "Email address automatically created by bem. You can forward emails with or without attachments, to be transformed.",
						Computed:    true,
					},
					"function_id": schema.StringAttribute{
						Description: "Unique identifier of function.",
						Computed:    true,
					},
					"function_name": schema.StringAttribute{
						Description: "Name of function. Must be UNIQUE on a per-environment basis.",
						Computed:    true,
					},
					"output_schema": schema.StringAttribute{
						Description: "Desired output structure defined in standard JSON Schema convention.",
						Computed:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"output_schema_name": schema.StringAttribute{
						Description: "Name of output schema object.",
						Computed:    true,
					},
					"tabular_chunking_enabled": schema.BoolAttribute{
						Description: "Whether tabular chunking is enabled on the pipeline. This processes tables in CSV/Excel in row batches, rather than all rows at once.",
						Computed:    true,
					},
					"type": schema.StringAttribute{
						Description: `Available values: "transform", "analyze", "route", "send", "split", "join", "payload_shaping", "enrich".`,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive(
								"transform",
								"analyze",
								"route",
								"send",
								"split",
								"join",
								"payload_shaping",
								"enrich",
							),
						},
					},
					"version_num": schema.Int64Attribute{
						Description: "Version number of function.",
						Computed:    true,
					},
					"audit": schema.SingleNestedAttribute{
						Description: "Audit trail information for the function.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[FunctionFunctionAuditDataSourceModel](ctx),
						Attributes: map[string]schema.Attribute{
							"function_created_by": schema.SingleNestedAttribute{
								Description: "Information about who created the function.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[FunctionFunctionAuditFunctionCreatedByDataSourceModel](ctx),
								Attributes: map[string]schema.Attribute{
									"created_at": schema.StringAttribute{
										Description: "The date and time the action was created.",
										Computed:    true,
										CustomType:  timetypes.RFC3339Type{},
									},
									"user_action_id": schema.StringAttribute{
										Description: "Unique identifier of the user action.",
										Computed:    true,
									},
									"api_key_name": schema.StringAttribute{
										Description: "API key name. Present for API key-initiated actions.",
										Computed:    true,
									},
									"email_address": schema.StringAttribute{
										Description: "Email address. Present for email-initiated actions.",
										Computed:    true,
									},
									"user_email": schema.StringAttribute{
										Description: "User's email address. Present for user-initiated actions.",
										Computed:    true,
									},
									"user_id": schema.StringAttribute{
										Description: "User's ID. Present for user-initiated actions.",
										Computed:    true,
									},
								},
							},
							"function_last_updated_by": schema.SingleNestedAttribute{
								Description: "Information about who last updated the function.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[FunctionFunctionAuditFunctionLastUpdatedByDataSourceModel](ctx),
								Attributes: map[string]schema.Attribute{
									"created_at": schema.StringAttribute{
										Description: "The date and time the action was created.",
										Computed:    true,
										CustomType:  timetypes.RFC3339Type{},
									},
									"user_action_id": schema.StringAttribute{
										Description: "Unique identifier of the user action.",
										Computed:    true,
									},
									"api_key_name": schema.StringAttribute{
										Description: "API key name. Present for API key-initiated actions.",
										Computed:    true,
									},
									"email_address": schema.StringAttribute{
										Description: "Email address. Present for email-initiated actions.",
										Computed:    true,
									},
									"user_email": schema.StringAttribute{
										Description: "User's email address. Present for user-initiated actions.",
										Computed:    true,
									},
									"user_id": schema.StringAttribute{
										Description: "User's ID. Present for user-initiated actions.",
										Computed:    true,
									},
								},
							},
							"version_created_by": schema.SingleNestedAttribute{
								Description: "Information about who created the current version.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[FunctionFunctionAuditVersionCreatedByDataSourceModel](ctx),
								Attributes: map[string]schema.Attribute{
									"created_at": schema.StringAttribute{
										Description: "The date and time the action was created.",
										Computed:    true,
										CustomType:  timetypes.RFC3339Type{},
									},
									"user_action_id": schema.StringAttribute{
										Description: "Unique identifier of the user action.",
										Computed:    true,
									},
									"api_key_name": schema.StringAttribute{
										Description: "API key name. Present for API key-initiated actions.",
										Computed:    true,
									},
									"email_address": schema.StringAttribute{
										Description: "Email address. Present for email-initiated actions.",
										Computed:    true,
									},
									"user_email": schema.StringAttribute{
										Description: "User's email address. Present for user-initiated actions.",
										Computed:    true,
									},
									"user_id": schema.StringAttribute{
										Description: "User's ID. Present for user-initiated actions.",
										Computed:    true,
									},
								},
							},
						},
					},
					"display_name": schema.StringAttribute{
						Description: "Display name of function. Human-readable name to help you identify the function.",
						Computed:    true,
					},
					"tags": schema.ListAttribute{
						Description: "Array of tags to categorize and organize functions.",
						Computed:    true,
						CustomType:  customfield.NewListType[types.String](ctx),
						ElementType: types.StringType,
					},
					"used_in_workflows": schema.ListNestedAttribute{
						Description: "List of workflows that use this function.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[FunctionFunctionUsedInWorkflowsDataSourceModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"current_version_num": schema.Int64Attribute{
									Description: "Current version number of workflow, provided for reference - compare to usedInWorkflowVersionNums to see whether the current version of the workflow uses this function version.",
									Computed:    true,
								},
								"used_in_workflow_version_nums": schema.ListAttribute{
									Description: "Version numbers of workflows that this function version is used in.",
									Computed:    true,
									CustomType:  customfield.NewListType[types.Int64](ctx),
									ElementType: types.Int64Type,
								},
								"workflow_id": schema.StringAttribute{
									Description: "Unique identifier of workflow.",
									Computed:    true,
								},
								"workflow_name": schema.StringAttribute{
									Description: "Name of workflow.",
									Computed:    true,
								},
							},
						},
					},
					"description": schema.StringAttribute{
						Description: "Description of router. Can be used to provide additional context on router's purpose and expected inputs.",
						Computed:    true,
					},
					"routes": schema.ListNestedAttribute{
						Description: "List of routes.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[FunctionFunctionRoutesDataSourceModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed: true,
								},
								"description": schema.StringAttribute{
									Computed: true,
								},
								"function_id": schema.StringAttribute{
									Computed: true,
								},
								"function_name": schema.StringAttribute{
									Computed: true,
								},
								"is_error_fallback": schema.BoolAttribute{
									Computed: true,
								},
								"origin": schema.SingleNestedAttribute{
									Computed:   true,
									CustomType: customfield.NewNestedObjectType[FunctionFunctionRoutesOriginDataSourceModel](ctx),
									Attributes: map[string]schema.Attribute{
										"email": schema.SingleNestedAttribute{
											Computed:   true,
											CustomType: customfield.NewNestedObjectType[FunctionFunctionRoutesOriginEmailDataSourceModel](ctx),
											Attributes: map[string]schema.Attribute{
												"patterns": schema.ListAttribute{
													Computed:    true,
													CustomType:  customfield.NewListType[types.String](ctx),
													ElementType: types.StringType,
												},
											},
										},
									},
								},
								"regex": schema.SingleNestedAttribute{
									Computed:   true,
									CustomType: customfield.NewNestedObjectType[FunctionFunctionRoutesRegexDataSourceModel](ctx),
									Attributes: map[string]schema.Attribute{
										"patterns": schema.ListAttribute{
											Computed:    true,
											CustomType:  customfield.NewListType[types.String](ctx),
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"destination_type": schema.StringAttribute{
						Description: "Destination type for a Send function.\nAvailable values: \"webhook\", \"s3\", \"google_drive\".",
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive(
								"webhook",
								"s3",
								"google_drive",
							),
						},
					},
					"google_drive_folder_id": schema.StringAttribute{
						Description: "Google Drive folder ID. Present when destinationType is google_drive. Managed via Paragon OAuth.",
						Computed:    true,
					},
					"s3_bucket": schema.StringAttribute{
						Description: "S3 bucket to upload the payload to. Present when destinationType is s3.",
						Computed:    true,
					},
					"s3_prefix": schema.StringAttribute{
						Description: "S3 key prefix (folder path). Optional, present when destinationType is s3.",
						Computed:    true,
					},
					"webhook_signing_enabled": schema.BoolAttribute{
						Description: "Whether webhook payloads are signed with an HMAC-SHA256 `bem-signature` header.",
						Computed:    true,
					},
					"webhook_url": schema.StringAttribute{
						Description: "Webhook URL to POST the payload to. Present when destinationType is webhook.",
						Computed:    true,
					},
					"split_type": schema.StringAttribute{
						Description: "The method used to split pages.\nAvailable values: \"print_page\", \"semantic_page\".",
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive("print_page", "semantic_page"),
						},
					},
					"print_page_split_config": schema.SingleNestedAttribute{
						Description: "Configuration for print page splitting.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[FunctionFunctionPrintPageSplitConfigDataSourceModel](ctx),
						Attributes: map[string]schema.Attribute{
							"next_function_id": schema.StringAttribute{
								Computed: true,
							},
						},
					},
					"semantic_page_split_config": schema.SingleNestedAttribute{
						Description: "Configuration for semantic page splitting.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[FunctionFunctionSemanticPageSplitConfigDataSourceModel](ctx),
						Attributes: map[string]schema.Attribute{
							"item_classes": schema.ListNestedAttribute{
								Computed:   true,
								CustomType: customfield.NewNestedObjectListType[FunctionFunctionSemanticPageSplitConfigItemClassesDataSourceModel](ctx),
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Computed: true,
										},
										"description": schema.StringAttribute{
											Computed: true,
										},
										"next_function_id": schema.StringAttribute{
											Description: "The unique ID of the function you want to use for this item class.",
											Computed:    true,
										},
										"next_function_name": schema.StringAttribute{
											Description: "The unique name of the function you want to use for this item class.",
											Computed:    true,
										},
									},
								},
							},
						},
					},
					"join_type": schema.StringAttribute{
						Description: "The type of join to perform.\nAvailable values: \"standard\".",
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive("standard"),
						},
					},
					"shaping_schema": schema.StringAttribute{
						Description: "JMESPath expression that defines how to transform and customize the input payload structure.\nPayload shaping allows you to extract, reshape, and reorganize data from complex input payloads\ninto a simplified, standardized output format. Use JMESPath syntax to select specific fields,\nperform calculations, and create new data structures tailored to your needs.",
						Computed:    true,
					},
					"config": schema.SingleNestedAttribute{
						Description: "Configuration for enrich function with semantic search steps.\n\n**How Enrich Functions Work:**\n\nEnrich functions use semantic search to augment JSON data with relevant information from collections.\nThey take JSON input (typically from a transform function), extract specified fields, perform vector-based\nsemantic search against collections, and inject the results back into the data.\n\n**Input Requirements:**\n- Must receive JSON input (typically uploaded to S3 from a previous function)\n- Can be chained after transform or other functions that produce JSON output\n\n**Example Use Cases:**\n- Match product descriptions to SKU codes from a product catalog\n- Enrich customer data with account information\n- Link order line items to inventory records\n\n**Configuration:**\n- Define one or more enrichment steps\n- Each step extracts values, searches a collection, and injects results\n- Steps are executed sequentially",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[FunctionFunctionConfigDataSourceModel](ctx),
						Attributes: map[string]schema.Attribute{
							"steps": schema.ListNestedAttribute{
								Description: "Array of enrichment steps to execute sequentially",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectListType[FunctionFunctionConfigStepsDataSourceModel](ctx),
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"collection_name": schema.StringAttribute{
											Description: "Name of the collection to search against. The collection must exist and contain items.\nSupports hierarchical paths when used with `includeSubcollections`.",
											Computed:    true,
										},
										"source_field": schema.StringAttribute{
											Description: "JMESPath expression to extract source data for semantic search.\nCan extract single values or arrays. All extracted values will be used for search.",
											Computed:    true,
										},
										"target_field": schema.StringAttribute{
											Description: "Field path where enriched results should be placed.\nUse simple field names (e.g., \"enriched_products\").\nResults are always injected as an array (list), regardless of topK value.",
											Computed:    true,
										},
										"include_cosine_distance": schema.BoolAttribute{
											Description: "Whether to include cosine distance scores in results.\nCosine distance ranges from 0.0 (perfect match) to 2.0 (completely dissimilar).\nLower scores indicate better semantic similarity.\n\nWhen enabled, each result includes a `cosineDistance` field.",
											Computed:    true,
										},
										"include_subcollections": schema.BoolAttribute{
											Description: "When true, searches all collections under the hierarchical path.\nFor example, \"customers\" will match \"customers\", \"customers.premium\", etc.",
											Computed:    true,
										},
										"score_threshold": schema.Float64Attribute{
											Description: "Maximum cosine distance threshold for filtering results (default: 0.6).\nResults with cosine distance above this threshold are excluded.\n\n**Only applies to `semantic` and `hybrid` search modes.**\nExact search does not use cosine distance and ignores this setting.\n\nCosine distance ranges from 0.0 (identical) to 2.0 (opposite):\n- 0.0 - 0.3: Very similar (strict threshold, high-quality matches only)\n- 0.3 - 0.6: Reasonably similar (moderate threshold)\n- 0.6 - 1.0: Loosely related (lenient threshold)\n- > 1.0: Rarely useful — allows nearly unrelated results\n\nFor most semantic search use cases, good matches typically fall in the 0.2 - 0.5 range.",
											Computed:    true,
											Validators: []validator.Float64{
												float64validator.Between(0, 2),
											},
										},
										"search_mode": schema.StringAttribute{
											Description: "Search mode to use for enrichment (default: \"semantic\").\n\n**semantic**: Vector similarity search using dense embeddings. Best for finding conceptually similar items.\n- Use for: Product descriptions, natural language content\n- Example: \"red sports car\" matches \"crimson convertible automobile\"\n\n**exact**: Exact keyword matching using PostgreSQL text search. Best for exact identifiers.\n- Use for: SKU numbers, routing numbers, account IDs, exact tags\n- Example: \"SKU-12345\" only matches items containing that exact text\n\n**hybrid**: Combined search using 20% semantic + 80% sparse embeddings (keyword-based).\n- Use for: Tags, categories, partial identifiers\n- Example: Balances semantic meaning with exact keyword matching\nAvailable values: \"semantic\", \"exact\", \"hybrid\".",
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.OneOfCaseInsensitive(
													"semantic",
													"exact",
													"hybrid",
												),
											},
										},
										"top_k": schema.Int64Attribute{
											Description: "Number of top matching results to return per query (default: 1).\nResults are always returned as an array (list) and automatically sorted by cosine distance\n(best match = lowest distance first).\n\n- 1: Returns array with single best match: `[{...}]`\n- >1: Returns array with multiple matches: `[{...}, {...}, ...]`",
											Computed:    true,
											Validators: []validator.Int64{
												int64validator.Between(1, 100),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"find_one_by": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{
						Optional: true,
					},
					"function_ids": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"function_names": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"sort_order": schema.StringAttribute{
						Description: `Available values: "asc", "desc".`,
						Computed:    true,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOfCaseInsensitive("asc", "desc"),
						},
					},
					"tags": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"types": schema.ListAttribute{
						Optional: true,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(
								stringvalidator.OneOfCaseInsensitive(
									"transform",
									"route",
									"send",
									"split",
									"join",
									"analyze",
									"payload_shaping",
									"enrich",
								),
							),
						},
						ElementType: types.StringType,
					},
					"workflow_ids": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"workflow_names": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (d *FunctionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceSchema(ctx)
}

func (d *FunctionDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(path.MatchRoot("function_name"), path.MatchRoot("find_one_by")),
	}
}
