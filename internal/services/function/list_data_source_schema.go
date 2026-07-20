// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"context"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigValidators = (*FunctionsDataSource)(nil)

func ListDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
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
							"extract",
							"route",
							"classify",
							"send",
							"split",
							"join",
							"analyze",
							"payload_shaping",
							"enrich",
							"parse",
							"render",
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
			"sort_order": schema.StringAttribute{
				Description: `Available values: "asc", "desc".`,
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("asc", "desc"),
				},
			},
			"max_items": schema.Int64Attribute{
				Description: "Max items to fetch, default: 1000",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"items": schema.ListNestedAttribute{
				Description: "The items returned by the data source",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectListType[FunctionsItemsDataSourceModel](ctx),
				NestedObject: schema.NestedAttributeObject{
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
							Description: `Available values: "transform", "extract", "analyze", "classify", "send", "split", "join", "payload_shaping", "enrich", "parse", "render".`,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive(
									"transform",
									"extract",
									"analyze",
									"classify",
									"send",
									"split",
									"join",
									"payload_shaping",
									"enrich",
									"parse",
									"render",
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
							CustomType:  customfield.NewNestedObjectType[FunctionsAuditDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"function_created_by": schema.SingleNestedAttribute{
									Description: "Information about who created the function.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[FunctionsAuditFunctionCreatedByDataSourceModel](ctx),
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
									CustomType:  customfield.NewNestedObjectType[FunctionsAuditFunctionLastUpdatedByDataSourceModel](ctx),
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
									CustomType:  customfield.NewNestedObjectType[FunctionsAuditVersionCreatedByDataSourceModel](ctx),
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
							CustomType:  customfield.NewNestedObjectListType[FunctionsUsedInWorkflowsDataSourceModel](ctx),
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
						"enable_bounding_boxes": schema.BoolAttribute{
							Description: "Whether bounding box extraction is enabled. Applies to vision input types\n(pdf, png, jpeg, heic, heif, webp) that dispatch through the analyze path.\nWhen true, the function returns the document regions (page, coordinates) from which each\nfield was extracted.",
							Computed:    true,
						},
						"pre_count": schema.BoolAttribute{
							Description: "Reducing the risk of the model stopping early on long documents.\nTrade-off: Increases total latency.",
							Computed:    true,
						},
						"classifications": schema.ListNestedAttribute{
							Description: "List of classifications a classify function can produce. Shares the underlying route list shape.",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectListType[FunctionsClassificationsDataSourceModel](ctx),
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
										CustomType: customfield.NewNestedObjectType[FunctionsClassificationsOriginDataSourceModel](ctx),
										Attributes: map[string]schema.Attribute{
											"email": schema.SingleNestedAttribute{
												Computed:   true,
												CustomType: customfield.NewNestedObjectType[FunctionsClassificationsOriginEmailDataSourceModel](ctx),
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
										CustomType: customfield.NewNestedObjectType[FunctionsClassificationsRegexDataSourceModel](ctx),
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
						"description": schema.StringAttribute{
							Description: "Description of classifier. Can be used to provide additional context on classifier's purpose and expected inputs.",
							Computed:    true,
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
							CustomType:  customfield.NewNestedObjectType[FunctionsPrintPageSplitConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"next_function_id": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"semantic_page_split_config": schema.SingleNestedAttribute{
							Description: "Configuration for semantic page splitting.",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectType[FunctionsSemanticPageSplitConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"item_classes": schema.ListNestedAttribute{
									Computed:   true,
									CustomType: customfield.NewNestedObjectListType[FunctionsSemanticPageSplitConfigItemClassesDataSourceModel](ctx),
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
							Description: "Configuration for an enrich function.\n\n**How Enrich Functions Work:**\n\nEnrich functions augment JSON input with data from external sources. They take JSON input\n(typically from a previous function), extract specified fields, fetch or search for matching\ndata, and inject the results back into the JSON.\n\n**Data Sources:**\n- **Collections** (`source: \"collection\"`): Vector/keyword search against a BEM collection.\nBest for semantic matching against pre-indexed documents.\n- **Endpoints** (`source: \"endpoint\"`): HTTP call to any user-provided REST API.\nBest for looking up live data from CRMs, ERPs, or other external systems.\nOptionally uses LLM agent reasoning to rank candidates returned by the endpoint.\n\n**Input Requirements:**\n- Must receive JSON input (typically from a previous function's output)\n\n**Example Use Cases:**\n- Match product descriptions to SKU codes from a product catalog collection\n- Enrich customer data with account details from a CRM endpoint\n- Use LLM agent reasoning to fuzzy-match line item descriptions to catalog products\n\n**Configuration:**\n- Define named endpoints (for endpoint-source steps)\n- Define one or more enrichment steps; steps are executed sequentially",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectType[FunctionsConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"steps": schema.ListNestedAttribute{
									Description: "Array of enrichment steps to execute sequentially.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectListType[FunctionsConfigStepsDataSourceModel](ctx),
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"source_field": schema.StringAttribute{
												Description: "JMESPath expression to extract source data.\nCan extract a single value or an array. Each extracted value is looked up independently.",
												Computed:    true,
											},
											"target_field": schema.StringAttribute{
												Description: "Field path where enriched results should be placed.\nUse simple field names (e.g., \"enriched_products\").\nResults are always injected as an array (list), regardless of topK value.",
												Computed:    true,
											},
											"collection_name": schema.StringAttribute{
												Description: "Name of the collection to search against.\nRequired when `source` is `\"collection\"`. The collection must exist and contain items.\nSupports hierarchical paths when used with `includeSubcollections`.",
												Computed:    true,
											},
											"endpoint_name": schema.StringAttribute{
												Description: "Name of an endpoint defined in `enrichConfig.endpoints`.\nRequired when `source` is `\"endpoint\"`.",
												Computed:    true,
											},
											"include_score": schema.BoolAttribute{
												Description: "Whether to include cosine distance scores in results.\nCosine distance ranges from 0.0 (perfect match) to 2.0 (completely dissimilar).\nLower scores indicate better semantic similarity.\n\nWhen enabled, each result includes a `cosine_distance` field (semantic mode)\nor a `hybrid_score` field (hybrid mode).",
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
											"source": schema.StringAttribute{
												Description: "Where to fetch enrichment data from (default: `\"collection\"`).\n\n- `\"collection\"`: Vector/keyword search against a BEM collection. Requires `collectionName`.\n- `\"endpoint\"`: HTTP call to a named endpoint defined in `enrichConfig.endpoints`. Requires `endpointName`.\nAvailable values: \"collection\", \"endpoint\".",
												Computed:    true,
												Validators: []validator.String{
													stringvalidator.OneOfCaseInsensitive("collection", "endpoint"),
												},
											},
											"top_k": schema.Int64Attribute{
												Description: "Number of top matching results to return per query (default: 1).\nResults are always returned as an array (list) and automatically sorted by cosine distance\n(best match = lowest distance first).\n\n- 1: Returns array with single best match: `[{...}]`\n- >1: Returns array with multiple matches: `[{...}, {...}, ...]`\n\nWhen re-ranking is on (the default for `semantic`/`hybrid`), `topK` is still the\nnumber of results returned — re-ranking changes their order, not the count. The\ncandidate pool the LLM chooses from is widened internally to at least 5, so even\n`topK: 1` re-ranks a real pool and returns the single best match.",
												Computed:    true,
												Validators: []validator.Int64{
													int64validator.Between(1, 100),
												},
											},
										},
									},
								},
								"endpoints": schema.ListNestedAttribute{
									Description: "Named HTTP endpoints available to endpoint-source steps.\nEach endpoint must have a unique `name` referenced by the step's `endpointName`.\nRequired when any step uses `source: \"endpoint\"`.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectListType[FunctionsConfigEndpointsDataSourceModel](ctx),
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"method": schema.StringAttribute{
												Description: "HTTP method to use.\nAvailable values: \"GET\", \"POST\".",
												Computed:    true,
												Validators: []validator.String{
													stringvalidator.OneOfCaseInsensitive("GET", "POST"),
												},
											},
											"name": schema.StringAttribute{
												Description: "Unique name for this endpoint, referenced by enrichStep.endpointName.",
												Computed:    true,
											},
											"url": schema.StringAttribute{
												Description: "Full URL of the endpoint (must be http:// or https://).",
												Computed:    true,
											},
											"body_template": schema.StringAttribute{
												Description: "JSON body template for POST requests.\n**Required for POST endpoints.** Must contain the `{value}` placeholder, which is replaced\nwith the extracted source value at runtime.\n\nExample: `bodyTemplate: \"{\\\"query\\\": \\\"{value}\\\", \\\"limit\\\": 10}\"`",
												Computed:    true,
											},
											"headers": schema.StringAttribute{
												Description: "Additional HTTP headers to include in every request (e.g. `Authorization: Bearer <token>`).",
												Computed:    true,
												CustomType:  jsontypes.NormalizedType{},
											},
											"match_instructions": schema.StringAttribute{
												Description: "Natural-language instructions for LLM agent reasoning.\n\nWhen set, the candidates fetched from the endpoint are passed to an LLM with these\ninstructions, which selects the best match(es) and returns them ranked best-first.\nEach injected result has the shape `{ data, rank, confidence, reasoning? }` (rank is\n1-based, 1 = best).\n\nWhen omitted, the raw fetched value is injected without any LLM involvement.",
												Computed:    true,
											},
											"match_top_k": schema.Int64Attribute{
												Description: "Maximum number of ranked matches to return per source value when `matchInstructions`\nis set (default: 1). Ignored when `matchInstructions` is empty.",
												Computed:    true,
												Validators: []validator.Int64{
													int64validator.Between(1, 100),
												},
											},
											"max_candidates": schema.Int64Attribute{
												Description: "LLM batch size during agent reasoning (default: 50). All candidates — across all\nfetched pages — are scored in batches of this size. Smaller values reduce per-call\ntoken usage; larger values mean fewer LLM calls. Ignored when `matchInstructions`\nis empty.",
												Computed:    true,
												Validators: []validator.Int64{
													int64validator.AtLeast(1),
												},
											},
											"max_pages": schema.Int64Attribute{
												Description: "Maximum number of pages to fetch (default: 10). Acts as a safety cap against\ninfinite pagination loops when the server never returns an empty cursor.",
												Computed:    true,
												Validators: []validator.Int64{
													int64validator.AtLeast(1),
												},
											},
											"next_page_param": schema.StringAttribute{
												Description: "Query parameter name used to pass the cursor on subsequent GET requests, or the\n`{placeholder}` name used in the POST `bodyTemplate` (e.g. `\"cursor\"`,\n`\"pageToken\"`, `\"offset\"`).\n\nMust be set together with `nextPagePath`.",
												Computed:    true,
											},
											"next_page_path": schema.StringAttribute{
												Description: "JMESPath expression applied to each raw response to extract the cursor or token\nfor the next page (e.g. `\"nextCursor\"`, `\"pagination.nextToken\"`). An absent,\nnull, or empty-string result stops pagination. Both string and numeric values are\nsupported — numbers are converted to their decimal string representation before\nbeing forwarded as a query parameter.\n\nMust be set together with `nextPageParam`.\n\n**Supported pagination styles:**\n- **Cursor/token-based** — server returns an opaque token in the response body\n(e.g. `{\"nextCursor\": \"abc123\"}`). Set `nextPagePath: \"nextCursor\"` and the\nplatform forwards it verbatim on the next request.\n- **Server-computed offset/page** — server echoes back the next offset or page\nnumber in the response body (e.g. `{\"nextOffset\": 50}` or `{\"nextPage\": 2}`).\nSet `nextPagePath: \"nextOffset\"` and the platform forwards the value as-is.\n\n**Not supported:**\n- **Client-computed offset** — APIs where the client must compute `offset += limit`\nitself (e.g. `?offset=0&limit=50` with no next-offset in the response). Workaround:\nask the API provider to return the next offset in the response body, or bake a\nfixed page size into the URL and use a server-side cursor instead.\n- **Client-computed page number** — APIs where the client increments `?page=N`\nitself with no next-page value in the response. Same workaround applies.\n- **Link header** — `Link: <url>; rel=\"next\"` in HTTP response headers. The\nplatform only inspects the response body.",
												Computed:    true,
											},
											"query_param": schema.StringAttribute{
												Description: "Query parameter name used to pass the extracted source value.\n**Required for GET endpoints.** The value is URL-encoded and appended as `?{queryParam}={sourceValue}`.\n\nExample: `queryParam: \"q\"` → `GET /products?q=blue+widget`",
												Computed:    true,
											},
											"response_path": schema.StringAttribute{
												Description: "JMESPath expression applied to the response body to extract the enrichment value.\nOmit to use the entire response body as the result.\n\n**For agent reasoning:** use a wildcard projection (e.g. `items[*]` or `results[*].data`)\nso the endpoint's list of candidates is flattened into an array before being passed to the LLM.\nA non-wildcard path (e.g. `data.product`) extracts a single value treated as one candidate.\n\n**Response size:** the platform reads at most 50 MB of the response body before decoding,\nregardless of the Content-Length header.",
												Computed:    true,
											},
										},
									},
								},
							},
						},
						"extra_config": schema.SingleNestedAttribute{
							Description: "Cross-cutting toggles for Parse functions. Mirrors the `extraConfig`\nsurface on Extract / Join — separated from `parseConfig` so the per-call\nParse output shape stays distinct from operator-level execution flags.",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectType[FunctionsExtraConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"enable_bounding_boxes": schema.BoolAttribute{
									Description: "When true, return per-section and per-entity-mention coordinates in\nthe parse event's `fieldBoundingBoxes` map (same shape as Extract:\nJSON Pointer key → array of `{page, left, top, width, height}` with\ncoordinates normalized to [0, 1]). Keys are `/sections/{N}` and\n`/entities/{N}/occurrences/{M}` into the parse output. Only applies\nto the open-ended discovery path (no `schema`) and to vision input\ntypes. Bedrock-backed parse functions silently return an empty map\n(no native bbox support). Defaults to false.",
									Computed:    true,
								},
							},
						},
						"parse_config": schema.SingleNestedAttribute{
							Description: "Per-version configuration for a Parse function.\n\nParse renders document pages (PDF, image) via vision LLM and emits\nstructured JSON. The two toggles below independently control entity\nextraction (a per-call output concern) and cross-document memory\nlinking (an environment-wide concern).",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectType[FunctionsParseConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"default_bucket": schema.StringAttribute{
									Description: "Optional bucket NAME that parse-extracted entities land in when no\ncall-level bucket is supplied. Lower precedence than a call-level bucket,\nhigher than the account+environment default.",
									Computed:    true,
								},
								"extract_entities": schema.BoolAttribute{
									Description: "When true, extract named entities (people, organizations, products,\nstudies, identifiers, etc.) and the relationships between them, and\ndedupe by canonical name within the document. When false, only\n`sections[]` is extracted; `entities[]` and `relationships[]` come\nback empty in the parse output. Defaults to true.",
									Computed:    true,
								},
								"link_across_documents": schema.BoolAttribute{
									Description: "When true, link this document's entities to entities seen in earlier\ndocuments in this environment, building one canonical record per\nreal-world thing across the corpus. Visible in the Memory tab and\nqueryable via `POST /v3/fs` (op=find / open / xref). Doesn't change\nthis call's parse output. Requires `extractEntities=true`. Defaults\nto true.",
									Computed:    true,
								},
								"schema": schema.StringAttribute{
									Description: "Optional JSONSchema. When provided, each chunk performs schema-guided\nextraction. When absent, chunks perform open-ended discovery and\nreturn sections, entities, and relationships per the discovery\nschema.",
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
							},
						},
						"render_config": schema.SingleNestedAttribute{
							Description: "Per-version configuration for a Render function.\n\nRender emits a `.docx` from schema-typed JSON by composing the JSON into a\n`.docx` template. The template document is stored server-side; this response\nexposes only the contract derived from it. Schema validation runs internally\nin the ML service against the bundled core schema; no customer-supplied\nschema rides this surface.",
							Computed:    true,
							CustomType:  customfield.NewNestedObjectType[FunctionsRenderConfigDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"template": schema.SingleNestedAttribute{
									Description: "The uploaded template: its filename, a short-lived presigned download URL,\nand the placeholder/style contract derived from it. Absent on configs\ncreated before template capture existed.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[FunctionsRenderConfigTemplateDataSourceModel](ctx),
									Attributes: map[string]schema.Attribute{
										"download_url": schema.StringAttribute{
											Description: "Short-lived presigned URL to download the stored `.docx`. The private\nstorage location is never exposed.",
											Computed:    true,
										},
										"list_kinds": schema.ListAttribute{
											Description: "Supported list kinds (`decimal`, `bullet`) the template's `numbering.xml`\ndefines an `abstractNum` for. Empty means the template can hold no list, so\nany list primitive will fail at render.",
											Computed:    true,
											Validators: []validator.List{
												listvalidator.ValueStringsAre(
													stringvalidator.OneOfCaseInsensitive("decimal", "bullet"),
												),
											},
											CustomType:  customfield.NewListType[types.String](ctx),
											ElementType: types.StringType,
										},
										"name": schema.StringAttribute{
											Description: "Original filename of the uploaded template (e.g. `contract.docx`), echoed\nback for display. Absent on templates uploaded before the filename was\ncaptured.",
											Computed:    true,
										},
										"placeholders": schema.SingleNestedAttribute{
											Description: "The placeholder contract a Render template declares, grouped by how each\nplaceholder is filled. Derived from the template at create/update time by\nscanning its `docxtpl` tags; not user-supplied.\n\n- `stringKeys`: bare string placeholders (`{{ key }}`) filled with a single\nvalue.\n- `blockKeys`: wrapped-primitive placeholders (`{{p key }}`) — bind one core\nprimitive (paragraph, table, image, or list). The placeholder's own\nparagraph dissolves and is replaced by the rendered subdocument's blocks,\nrather than substituting text inline.",
											Computed:    true,
											CustomType:  customfield.NewNestedObjectType[FunctionsRenderConfigTemplatePlaceholdersDataSourceModel](ctx),
											Attributes: map[string]schema.Attribute{
												"block_keys": schema.ListAttribute{
													Computed:    true,
													CustomType:  customfield.NewListType[types.String](ctx),
													ElementType: types.StringType,
												},
												"string_keys": schema.ListAttribute{
													Computed:    true,
													CustomType:  customfield.NewListType[types.String](ctx),
													ElementType: types.StringType,
												},
											},
										},
										"style_ids": schema.ListAttribute{
											Description: "Paragraph/character style IDs the uploaded template defines and the\nrendered output can reference. Derived from the template's `styles.xml` at\ncreate/update time.",
											Computed:    true,
											CustomType:  customfield.NewListType[types.String](ctx),
											ElementType: types.StringType,
										},
										"table_style_ids": schema.ListAttribute{
											Description: "Style IDs whose type is table — the styles a `table` primitive's required\n`styleId` can name. Empty means the template defines no table style, so any\ntable primitive will fail at render.",
											Computed:    true,
											CustomType:  customfield.NewListType[types.String](ctx),
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *FunctionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ListDataSourceSchema(ctx)
}

func (d *FunctionsDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}
