// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"context"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigValidators = (*WorkflowResource)(nil)

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Workflows orchestrate one or more functions into a directed acyclic graph (DAG) for document processing.\n\nUse these endpoints to create, update, list, and manage workflows, and to invoke them\nwith file input via `POST /v3/workflows/{workflowName}/call`.\n\nThe call endpoint accepts files as either multipart form data or JSON with base64-encoded\ncontent. In the Bem CLI, use `@path/to/file` inside JSON values to automatically read and\nencode files:\n\n```\nbem workflows call --workflow-name my-workflow \\\n  --input.single-file '{\"inputContent\": \"@file.pdf\", \"inputType\": \"pdf\"}' \\\n  --wait\n```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Unique name for the workflow. Must match `^[a-zA-Z0-9_-]{1,128}$`.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Description:   "Unique name for the workflow. Must match `^[a-zA-Z0-9_-]{1,128}$`.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"main_node_name": schema.StringAttribute{
				Description: "Name of the entry-point node. Must not be a destination of any edge.",
				Required:    true,
			},
			"nodes": schema.ListNestedAttribute{
				Description: "Call-site nodes in the DAG. At least one is required.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"function": schema.SingleNestedAttribute{
							Description: "The function (and version) to execute at this call site.",
							Required:    true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Unique identifier of function. Provide either id or name, not both.",
									Optional:    true,
								},
								"name": schema.StringAttribute{
									Description: "Name of function. Must be UNIQUE on a per-environment basis. Provide either id or name, not both.",
									Optional:    true,
								},
								"version_num": schema.Int64Attribute{
									Description: "Version number of function.",
									Optional:    true,
								},
							},
						},
						"metadata": schema.StringAttribute{
							Description: "Opaque free-form JSON object attached to this node. Stored and returned\nverbatim; the server does not interpret it. Intended for client-side\nconcerns such as canvas display properties (position, color, collapsed\nstate, etc.).",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"name": schema.StringAttribute{
							Description: "Name for this call site. Must be unique within the workflow version.\nDefaults to the function's own name when omitted.",
							Optional:    true,
						},
					},
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags to categorize and organize the workflow.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"connectors": schema.ListNestedAttribute{
				Description: "Connectors to attach to the workflow at creation. If any entry fails to\nprovision, the entire workflow creation is rolled back.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Human-friendly connector name.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Discriminator for a workflow connector. V3 supports `paragon` only.\nAvailable values: \"paragon\".",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("paragon"),
							},
						},
						"connector_id": schema.StringAttribute{
							Description: "Present → update. Absent → create.",
							Optional:    true,
						},
						"paragon": schema.SingleNestedAttribute{
							Description: "Request-side config block for a Paragon connector. Fields absent on update are unchanged.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"configuration": schema.StringAttribute{
									Description: "Opaque per-integration configuration. Required on create.",
									Optional:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								"integration": schema.StringAttribute{
									Description: "Paragon integration key. Required on create.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
			"edges": schema.ListNestedAttribute{
				Description: "Directed edges between nodes. Omit or leave empty for single-node workflows.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"destination_node_name": schema.StringAttribute{
							Description: "Name of the destination node.",
							Required:    true,
						},
						"source_node_name": schema.StringAttribute{
							Description: "Name of the source node.",
							Required:    true,
						},
						"destination_name": schema.StringAttribute{
							Description: "Labelled outlet on the source node that activates this edge.\nOmit for the default (unlabelled) outlet.",
							Optional:    true,
						},
						"metadata": schema.StringAttribute{
							Description: "Opaque free-form JSON object attached to this edge. Stored and returned\nverbatim; the server does not interpret it.",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The date and time the workflow was created.",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
			"email_address": schema.StringAttribute{
				Description: "Inbound email address associated with the workflow, if any.",
				Computed:    true,
			},
			"error": schema.StringAttribute{
				Description: "Error message if the workflow retrieval failed.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The date and time the workflow was last updated.",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
			"version_num": schema.Int64Attribute{
				Description: "Version number of this workflow version.",
				Computed:    true,
			},
			"audit": schema.SingleNestedAttribute{
				Description: "Audit trail information.",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectType[WorkflowAuditModel](ctx),
				Attributes: map[string]schema.Attribute{
					"version_created_by": schema.SingleNestedAttribute{
						Description: "Information about who created the current version.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[WorkflowAuditVersionCreatedByModel](ctx),
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
					"workflow_created_by": schema.SingleNestedAttribute{
						Description: "Information about who created the workflow.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[WorkflowAuditWorkflowCreatedByModel](ctx),
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
					"workflow_last_updated_by": schema.SingleNestedAttribute{
						Description: "Information about who last updated the workflow.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[WorkflowAuditWorkflowLastUpdatedByModel](ctx),
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
			"connector_errors": schema.ListNestedAttribute{
				Description: "Per-connector failures from the diff/apply phase. Empty or omitted when all\noperations succeeded.",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectListType[WorkflowConnectorErrorsModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Description: "Machine-readable error code.",
							Computed:    true,
						},
						"message": schema.StringAttribute{
							Description: "Human-readable error message.",
							Computed:    true,
						},
						"operation": schema.StringAttribute{
							Description: "Which diff operation was attempted.\nAvailable values: \"create\", \"update\", \"delete\".",
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive(
									"create",
									"update",
									"delete",
								),
							},
						},
						"connector_id": schema.StringAttribute{
							Description: "Populated for update/delete failures.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Populated for create failures.",
							Computed:    true,
						},
					},
				},
			},
			"workflow": schema.SingleNestedAttribute{
				Description: "V3 read representation of a workflow version.",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowModel](ctx),
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Unique identifier of the workflow.",
						Computed:    true,
					},
					"connectors": schema.ListNestedAttribute{
						Description: "Connectors currently attached to this workflow. For version-scoped reads\n(`/versions/{n}`) this is always empty — connectors are current-state and\nnot part of version history.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[WorkflowWorkflowConnectorsModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"connector_id": schema.StringAttribute{
									Description: "Unique connector API ID.",
									Computed:    true,
								},
								"name": schema.StringAttribute{
									Description: "Human-friendly connector name.",
									Computed:    true,
								},
								"type": schema.StringAttribute{
									Description: "Discriminator for a workflow connector. V3 supports `paragon` only.\nAvailable values: \"paragon\".",
									Computed:    true,
									Validators: []validator.String{
										stringvalidator.OneOfCaseInsensitive("paragon"),
									},
								},
								"paragon": schema.SingleNestedAttribute{
									Description: "Paragon-integration configuration on a workflow connector.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowConnectorsParagonModel](ctx),
									Attributes: map[string]schema.Attribute{
										"configuration": schema.StringAttribute{
											Description: "Opaque per-integration configuration (e.g. `{\"folderId\": \"...\"}`).",
											Computed:    true,
											CustomType:  jsontypes.NormalizedType{},
										},
										"integration": schema.StringAttribute{
											Description: `Paragon integration key (e.g. "googledrive").`,
											Computed:    true,
										},
										"sync_id": schema.StringAttribute{
											Description: "Paragon sync ID managed by the server. Read-only.",
											Computed:    true,
										},
									},
								},
							},
						},
					},
					"created_at": schema.StringAttribute{
						Description: "The date and time the workflow was created.",
						Computed:    true,
						CustomType:  timetypes.RFC3339Type{},
					},
					"edges": schema.ListNestedAttribute{
						Description: "All directed edges in this workflow version's DAG.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[WorkflowWorkflowEdgesModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"destination_node_name": schema.StringAttribute{
									Description: "Name of the destination node.",
									Computed:    true,
								},
								"source_node_name": schema.StringAttribute{
									Description: "Name of the source node.",
									Computed:    true,
								},
								"destination_name": schema.StringAttribute{
									Description: "Labelled outlet on the source node, if any.",
									Computed:    true,
								},
								"metadata": schema.StringAttribute{
									Description: "Opaque free-form JSON object attached to this edge on create/update.\nReturned verbatim; never interpreted by the server.",
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
							},
						},
					},
					"main_node_name": schema.StringAttribute{
						Description: "Name of the entry-point call-site node.",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "Unique name of the workflow within the environment.",
						Computed:    true,
					},
					"nodes": schema.ListNestedAttribute{
						Description: "All call-site nodes in this workflow version's DAG.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[WorkflowWorkflowNodesModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"function": schema.SingleNestedAttribute{
									Description: "Function (and version) executing at this call site.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowNodesFunctionModel](ctx),
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Unique identifier of function. Provide either id or name, not both.",
											Computed:    true,
										},
										"name": schema.StringAttribute{
											Description: "Name of function. Must be UNIQUE on a per-environment basis. Provide either id or name, not both.",
											Computed:    true,
										},
										"version_num": schema.Int64Attribute{
											Description: "Version number of function.",
											Computed:    true,
										},
									},
								},
								"name": schema.StringAttribute{
									Description: "Name of this call site, unique within the workflow version.",
									Computed:    true,
								},
								"metadata": schema.StringAttribute{
									Description: "Opaque free-form JSON object attached to this node on create/update.\nReturned verbatim; never interpreted by the server.",
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
							},
						},
					},
					"updated_at": schema.StringAttribute{
						Description: "The date and time the workflow was last updated.",
						Computed:    true,
						CustomType:  timetypes.RFC3339Type{},
					},
					"version_num": schema.Int64Attribute{
						Description: "Version number of this workflow version.",
						Computed:    true,
					},
					"audit": schema.SingleNestedAttribute{
						Description: "Audit trail information.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditModel](ctx),
						Attributes: map[string]schema.Attribute{
							"version_created_by": schema.SingleNestedAttribute{
								Description: "Information about who created the current version.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditVersionCreatedByModel](ctx),
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
							"workflow_created_by": schema.SingleNestedAttribute{
								Description: "Information about who created the workflow.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditWorkflowCreatedByModel](ctx),
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
							"workflow_last_updated_by": schema.SingleNestedAttribute{
								Description: "Information about who last updated the workflow.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditWorkflowLastUpdatedByModel](ctx),
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
						Description: "Human-readable display name.",
						Computed:    true,
					},
					"email_address": schema.StringAttribute{
						Description: "Inbound email address associated with the workflow, if any.",
						Computed:    true,
					},
					"tags": schema.ListAttribute{
						Description: "Tags associated with the workflow.",
						Computed:    true,
						CustomType:  customfield.NewListType[types.String](ctx),
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *WorkflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *WorkflowResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}
