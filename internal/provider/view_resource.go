package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"strings"
)

var _ resource.Resource = &ViewResource{}
var _ resource.ResourceWithImportState = &ViewResource{}

func NewViewResource() resource.Resource {
	return &ViewResource{}
}

type ViewResource struct {
	client *chclient.ClickHouseClient
}

type ViewResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Database string       `tfsdk:"database"`
	Name     string       `tfsdk:"name"`
	FullName types.String `tfsdk:"full_name"`
	Query    string       `tfsdk:"query"`
}

func (r *ViewResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

func (r *ViewResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse view. See: https://clickhouse.com/docs/en/sql-reference/statements/create/view#normal-view",
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "ClickHouse database name",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "View name in ClickHouse database",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "ClickHouse view name in `database.view_name` format",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{fullNamePlanModifier{}},
			},
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{fullNamePlanModifier{}},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "View definition query. It should be a valid SELECT statement.",
				Required:            true,
			},
		},
	}
}

func (r *ViewResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *ViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view := chclient.ClickHouseView{
		Database: model.Database,
		Name:     model.Name,
		Query:    model.Query,
	}

	if err := r.client.CreateView(ctx, view, false); err != nil {
		resp.Diagnostics.AddError("Failed to create a view", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.GetTable(ctx, model.Database, model.Name)
	if err != nil {
		handleNotFoundError(ctx, err, resp, "view", model.FullName.ValueString())
		return
	}
	model.Name = view.Name
	model.Database = view.Database
	model.Query = view.AsSelect

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var stateModel ViewResourceModel
	var planModel ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Database != stateModel.Database || planModel.Name != stateModel.Name {
		resp.Diagnostics.AddError(
			"Cannot change view name or database",
			"View name and database cannot be changed. You should delete the resource and create a new one.",
		)
		return
	}

	view := chclient.ClickHouseView{
		Database: planModel.Database,
		Name:     planModel.Name,
		Query:    planModel.Query,
	}

	if err := r.client.CreateView(ctx, view, true); err != nil {
		resp.Diagnostics.AddError(
			"Cannot replace view",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planModel)...)
}

func (r *ViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	view := chclient.ClickHouseTable{Database: model.Database, Name: model.Name}
	err := r.client.DropTable(ctx, view, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot drop view",
			err.Error(),
		)
		return
	}
}

func (r *ViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")

	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID should be in `database.view_name` format",
		)
		return
	}
	db := parts[0]
	view := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("full_name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("database"), db)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), view)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("query"), "")...)
}
