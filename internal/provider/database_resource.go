package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	client *chclient.ClickHouseClient
}

type DatabaseResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Engine  types.String `tfsdk:"engine"`
	Comment types.String `tfsdk:"comment"`
	//	TODO: Add database UUID?
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse database",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("[a-z0-9_]+"),
						"Database name should contain only lower case latin letters, digits and _",
					),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "Database engine. Currently supported only `Atomic` and `Memory`. https://clickhouse.com/docs/en/engines/database-engines",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Atomic"),
				Validators:          []validator.String{stringvalidator.OneOf("Atomic", "Memory")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for database",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			// TODO: engine args and settings
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateDatabase(
		ctx,
		chclient.ClickHouseDatabase{
			Name:    data.Name.ValueString(),
			Engine:  chclient.DatabaseEngineFromString(data.Engine.ValueString()),
			Comment: data.Comment.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create database",
			"Create database query failed: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created a clickhouse_database resource", dict{"name": data.Name.ValueString()})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var db *DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &db)...)

	if resp.Diagnostics.HasError() {
		return
	}

	receivedDb, err := r.client.GetDatabase(ctx, db.Name.ValueString())
	if err != nil {
		handleNotFoundError(ctx, err, resp, "database", db.Name.ValueString())
		return
	}

	db.Name = types.StringValue(receivedDb.Name)
	db.Comment = types.StringValue(receivedDb.Comment)
	db.Engine = types.StringValue(receivedDb.Engine.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &db)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// ClickHouse does not support ALTER DATABASE
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DropDatabase(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete database",
			"Cannot drop database "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
