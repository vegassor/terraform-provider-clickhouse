package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vegassor/terraform-provider-clickhouse/internal/clickhouse_client"
	"regexp"
)

var _ resource.Resource = &TableResource{}
var _ resource.ResourceWithImportState = &TableResource{}

func NewTableResource() resource.Resource {
	return &TableResource{}
}

type TableResource struct {
	client *clickhouse_client.ClickHouseClient
}

type ColumnModel struct {
	Name string `tfsdk:"name"`
	Type string `tfsdk:"type"`
	//Nullable bool   `tfsdk:"nullable"`
	Comment string `tfsdk:"comment"`
}

type TableResourceModel struct {
	Database string        `tfsdk:"database"`
	Name     string        `tfsdk:"name"`
	Engine   string        `tfsdk:"engine"`
	Columns  []ColumnModel `tfsdk:"columns"`
	Comment  string        `tfsdk:"comment"`
}

func (r *TableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *TableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse user",
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "ClickHouse database name",
				Required:            true,
				Validators:          []validator.String{ClickHouseIdentifierValidator},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "ClickHouse table name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("[a-z0-9_]+"),
						"User name should contain only lower case latin letters, digits and _",
					),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for the table",
				Optional:            true,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "ClickHouse table engine. See: https://clickhouse.com/docs/en/engines/table-engines",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Memory",
						"Buffer",

						"MergeTree",
						"ReplacingMergeTree",
						"SummingMergeTree",
						"AggregatingMergeTree",
						"CollapsingMergeTree",
						"VersionedCollapsingMergeTree",
						"GraphiteMergeTree",

						"Log",
						"TinyLog",
						"StripeLog",
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
							Validators:          []validator.String{ClickHouseIdentifierValidator},
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "See: https://clickhouse.com/docs/en/sql-reference/data-types",
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
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
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
	tflog.Info(ctx, "ABOBUS", map[string]interface{}{"a": req.Plan.Raw})
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tableModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var table clickhouse_client.ClickHouseTable

	table.Name = tableModel.Name
	table.Engine = tableModel.Engine
	table.Comment = tableModel.Comment
	table.Columns = make([]clickhouse_client.ClickHouseColumn, 0, len(tableModel.Columns))
	for _, col := range tableModel.Columns {
		table.Columns = append(table.Columns, clickhouse_client.ClickHouseColumn{
			Name:    col.Name,
			Type:    col.Type,
			Comment: col.Comment,
		})
	}

	err := r.client.CreateTable(ctx, tableModel.Database, table)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create user",
			"Create user query failed: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tableModel)...)
}

func (r *TableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *TableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *TableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *TableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
