// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package workflow

import (
	"context"

	"github.com/bem-team/terraform-provider-bem/internal/customfield"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigValidators = (*WorkflowsDataSource)(nil)

func ListDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Workflow operations",
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
				CustomType:  customfield.NewNestedObjectListType[WorkflowsItemsDataSourceModel](ctx),
				NestedObject: schema.NestedAttributeObject{
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
							CustomType:  customfield.NewNestedObjectListType[WorkflowsEdgesDataSourceModel](ctx),
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
							CustomType:  customfield.NewNestedObjectListType[WorkflowsNodesDataSourceModel](ctx),
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"function": schema.SingleNestedAttribute{
										Description: "Function (and version) executing at this call site.",
										Computed:    true,
										CustomType:  customfield.NewNestedObjectType[WorkflowsNodesFunctionDataSourceModel](ctx),
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
							CustomType:  customfield.NewNestedObjectType[WorkflowsAuditDataSourceModel](ctx),
							Attributes: map[string]schema.Attribute{
								"version_created_by": schema.SingleNestedAttribute{
									Description: "Information about who created the current version.",
									Computed:    true,
									CustomType:  customfield.NewNestedObjectType[WorkflowsAuditVersionCreatedByDataSourceModel](ctx),
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
									CustomType:  customfield.NewNestedObjectType[WorkflowsAuditWorkflowCreatedByDataSourceModel](ctx),
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
									CustomType:  customfield.NewNestedObjectType[WorkflowsAuditWorkflowLastUpdatedByDataSourceModel](ctx),
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
		},
	}
}

func (d *WorkflowsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ListDataSourceSchema(ctx)
}

func (d *WorkflowsDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}
