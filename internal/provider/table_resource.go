package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"regexp"
	"slices"
	"strings"
)

var _ resource.Resource = &TableResource{}
var _ resource.ResourceWithImportState = &TableResource{}
var enginesRequiresReplaceIfSettingsChanges = []string{"RabbitMQ"}

func NewTableResource() resource.Resource {
	return &TableResource{}
}

type TableResource struct {
	client *chclient.ClickHouseClient
}

type ColumnModel struct {
	Name     string `tfsdk:"name"`
	Type     string `tfsdk:"type"`
	Nullable bool   `tfsdk:"nullable"`
	Comment  string `tfsdk:"comment"`
}

type TableResourceModel struct {
	Database string       `tfsdk:"database"`
	Name     string       `tfsdk:"name"`
	FullName types.String `tfsdk:"full_name"`
	ID       types.String `tfsdk:"id"`

	Columns []ColumnModel `tfsdk:"columns"`

	Engine           string     `tfsdk:"engine"`
	EngineParameters types.List `tfsdk:"engine_parameters"`

	PartitionBy types.String `tfsdk:"partition_by"`
	OrderBy     []string     `tfsdk:"order_by"`
	PrimaryKey  types.List   `tfsdk:"primary_key"`
	Settings    types.Map    `tfsdk:"settings"`

	Comment string `tfsdk:"comment"`
}

func (r *TableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *TableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse table",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{NewCompositePlanModifierFromStr([]string{"database", "name"}, ".")},
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "ClickHouse database name",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "ClickHouse table name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("[a-z0-9_]+"),
						"Table name should contain only lower case latin letters, digits and _",
					),
				},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "ClickHouse table name in `database.table` format",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{NewCompositePlanModifierFromStr([]string{"database", "name"}, ".")},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for the table",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "ClickHouse table engine. See: https://clickhouse.com/docs/en/engines/table-engines",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"engine_parameters": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Parameters for engine. Will be transformed to `engine(param1, param2, ...)`",
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, make([]attr.Value, 0))),
				PlanModifiers:       []planmodifier.List{partitionByPlanModifier{}, listplanmodifier.RequiresReplace()},
			},
			"partition_by": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Expression to fill `PARTITION BY` clause.",
				Default:             stringdefault.StaticString(""),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"order_by": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Values to fill `ORDER BY` clause.",
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, make([]attr.Value, 0))),
				PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"primary_key": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Values to fill `PRIMARY KEY` clause.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Values to fill `SETTINGS` clause.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
					mapplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
							var engine string
							resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("engine"), &engine)...)
							if resp.Diagnostics.HasError() {
								return
							}

							if slices.Contains(enginesRequiresReplaceIfSettingsChanges, engine) {
								resp.RequiresReplace = true
							}
						},
						"settings",
						"",
					),
				},
			},
			"columns": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Columns of ClickHouse table",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Column name in ClickHouse table",
							Validators:          []validator.String{clickHouseIdentifierValidator},
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Column type. See: https://clickhouse.com/docs/en/sql-reference/data-types",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"UInt8",
									"UInt16",
									"UInt32",
									"UInt64",
									"UInt128",
									"UInt256",
									"Int8",
									"Int16",
									"Int32",
									"Int64",
									"Int128",
									"Int256",
									"Float32",
									"Float64",
									"Decimal",
									"Boolean",
									"String",
									"FixedString",
									"Date",
									"Date32",
									"DateTime",
									"DateTime64",
									"JSON",
									"UUID",
									"LowCardinality",
									"SimpleAggregateFunction",
									"AggregateFunction",
									"IPv4",
									"IPv6",
								),
							},
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment for a column",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"nullable": schema.BoolAttribute{
							MarkdownDescription: "Whether a column can contain NULL values",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.List{listvalidator.SizeAtLeast(1)},
			},
		},
	}
}

func (r *TableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *TableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var tableModel TableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tableModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	table, diags := toChClientTable(ctx, tableModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateTable(ctx, table)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create table",
			"Create table query failed: "+err.Error(),
		)
		return
	}
	createdTableInfo, err := r.client.GetTable(ctx, table.Database, table.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot read table info",
			"Create table query failed: "+err.Error(),
		)
		return
	}
	createdTableModel, diags := fromChClientTableInfo(ctx, createdTableInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, createdTableModel)...)
}

func (r *TableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateTableModel TableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateTableModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	receivedTableInfo, err := r.client.GetTable(ctx, stateTableModel.Database, stateTableModel.Name)
	if err != nil {
		handleNotFoundError(ctx, err, resp, "table", stateTableModel.FullName.ValueString())
		return
	}

	table, diags := fromChClientTableInfo(ctx, receivedTableInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, table)...)
}

func (r *TableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var stateTable TableResourceModel
	var planTable TableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateTable)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planTable)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if stateTable.Database != planTable.Database || stateTable.Comment != planTable.Comment {
		resp.Diagnostics.AddError(
			"Cannot change database or comment of the table",
			"Table should be recreated if database or comment is changed. It is a bug in the provider."+
				"If you see this error, please report it to the provider maintainers",
		)
		return
	}

	table, diags := toChClientTable(ctx, planTable)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.AlterTable(ctx, stateTable.Name, table)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot alter table",
			err.Error(),
		)
		return
	}

	updatedTableInfo, err := r.client.GetTable(ctx, table.Database, table.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot read table info",
			"Update table query failed: "+err.Error(),
		)
		return
	}

	updatedTableModel, diags := fromChClientTableInfo(ctx, updatedTableInfo)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updatedTableModel)...)
}

func (r *TableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model TableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	table, diags := toChClientTable(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkIfTableEmpty := true
	if slices.Contains(enginesRequiresReplaceIfSettingsChanges, model.Engine) {
		checkIfTableEmpty = false
	}

	err := r.client.DropTable(ctx, table, checkIfTableEmpty)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete table",
			err.Error(),
		)
		return
	}
}

func (r *TableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")

	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID should be in `database.table` format",
		)
		return
	}
	db := parts[0]
	table := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("database"), db)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), table)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("engine"), "")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("comment"), "")...)
}

func toChClientTable(ctx context.Context, table TableResourceModel) (chclient.ClickHouseTable, diag.Diagnostics) {
	var diags diag.Diagnostics

	cols := make([]chclient.ClickHouseColumn, 0, len(table.Columns))
	for _, col := range table.Columns {
		cols = append(cols, chclient.ClickHouseColumn{
			Name:     col.Name,
			Type:     col.Type,
			Comment:  col.Comment,
			Nullable: col.Nullable,
		})
	}

	var pk []string
	if !table.PrimaryKey.IsUnknown() {
		val, ds := table.PrimaryKey.ToListValue(ctx)
		diags.Append(ds...)
		diags.Append(val.ElementsAs(ctx, &pk, true)...)
	}

	var engineParams []string
	if !table.EngineParameters.IsUnknown() {
		val, ds := table.EngineParameters.ToListValue(ctx)
		diags.Append(ds...)
		diags.Append(val.ElementsAs(ctx, &engineParams, true)...)
	} else {
		engineParams = make([]string, 0)
	}

	var settings map[string]string
	if !table.Settings.IsUnknown() {
		val, ds := table.Settings.ToMapValue(ctx)
		diags.Append(ds...)
		diags.Append(val.ElementsAs(ctx, &settings, false)...)
	}

	return chclient.ClickHouseTable{
		Database:      table.Database,
		Name:          table.Name,
		Comment:       table.Comment,
		Engine:        table.Engine,
		EngineParams:  engineParams,
		PartitionBy:   table.PartitionBy.ValueString(),
		OrderBy:       table.OrderBy,
		PrimaryKeyArr: pk,
		Settings:      settings,
		Columns:       cols,
	}, diags
}

func fromChClientTableInfo(ctx context.Context, table chclient.ClickHouseTableFullInfo) (TableResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	cols := make([]ColumnModel, 0, len(table.Columns))
	for _, col := range table.Columns {
		cols = append(cols, ColumnModel{
			Name:     col.Name,
			Type:     col.Type,
			Comment:  col.Comment,
			Nullable: col.Nullable,
		})
	}

	pk, ds := basetypes.NewListValueFrom(ctx, types.StringType, table.PrimaryKeyArr)
	diags.Append(ds...)

	engineParams, ds := basetypes.NewListValueFrom(ctx, types.StringType, table.EngineParams)
	diags.Append(ds...)

	settings, ds := basetypes.NewMapValueFrom(ctx, types.StringType, table.Settings)
	diags.Append(ds...)

	return TableResourceModel{
		ID:               types.StringValue(table.Database + "." + table.Name),
		Database:         table.Database,
		Name:             table.Name,
		FullName:         types.StringValue(table.Database + "." + table.Name),
		Comment:          table.Comment,
		Engine:           table.Engine,
		EngineParameters: engineParams,
		PartitionBy:      types.StringValue(table.PartitionBy),
		OrderBy:          table.OrderBy,
		PrimaryKey:       pk,
		Settings:         settings,
		Columns:          cols,
	}, diags
}

func handleNotFoundError(ctx context.Context, err error, resp *resource.ReadResponse, entity string, name string) {
	var notFoundError *chclient.NotFoundError
	ok := errors.As(err, &notFoundError)
	if !ok {
		resp.Diagnostics.AddError(
			"Cannot find "+entity,
			err.Error(),
		)
	}

	resp.Diagnostics.AddWarning(
		"Cannot find "+entity,
		entity+" "+name+" not found in ClickHouse. Removing the resource from state",
	)
	resp.State.RemoveResource(ctx)
}
