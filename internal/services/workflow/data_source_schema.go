// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"context"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigValidators = (*WorkflowDataSource)(nil)

func DataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Workflows orchestrate one or more functions into a directed acyclic graph (DAG) for document processing.\n\nUse these endpoints to create, update, list, and manage workflows, and to invoke them\nwith file input via `POST /v3/workflows/{workflowName}/call`.\n\nThe call endpoint accepts files as either multipart form data or JSON with base64-encoded\ncontent. In the Bem CLI, use `@path/to/file` inside JSON values to automatically read and\nencode files:\n\n```\nbem workflows call --workflow-name my-workflow \\\n  --input.single-file '{\"inputContent\": \"@file.pdf\", \"inputType\": \"pdf\"}' \\\n  --wait\n```",
		Attributes: map[string]schema.Attribute{
			"workflow_name": schema.StringAttribute{
				Required: true,
			},
			"error": schema.StringAttribute{
				Description: "Error message if the workflow retrieval failed.",
				Computed:    true,
			},
			"workflow": schema.SingleNestedAttribute{
				Description: "V3 read representation of a workflow version.",
				Computed:    true,
				CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowDataSourceModel](ctx),
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Unique identifier of the workflow.",
						Computed:    true,
					},
					"created_at": schema.StringAttribute{
						Description: "The date and time the workflow was created.",
						Computed:    true,
						CustomType:  timetypes.RFC3339Type{},
					},
					"edges": schema.ListNestedAttribute{
						Description: "All directed edges in this workflow version's DAG.",
						Computed:    true,
						CustomType:  customfield.NewNestedObjectListType[WorkflowWorkflowEdgesDataSourceModel](ctx),
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
						CustomType:  customfield.NewNestedObjectListType[WorkflowWorkflowNodesDataSourceModel](ctx),
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"function": schema.SingleNestedAttribute{
									Description: "Function (and version) executing at this call site.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowNodesFunctionDataSourceModel](ctx),
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
						CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditDataSourceModel](ctx),
						Attributes: map[string]schema.Attribute{
							"version_created_by": schema.SingleNestedAttribute{
								Description: "Information about who created the current version.",
								Computed:    true,
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditVersionCreatedByDataSourceModel](ctx),
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
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditWorkflowCreatedByDataSourceModel](ctx),
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
								CustomType:  customfield.NewNestedObjectType[WorkflowWorkflowAuditWorkflowLastUpdatedByDataSourceModel](ctx),
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

func (d *WorkflowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceSchema(ctx)
}

func (d *WorkflowDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}
