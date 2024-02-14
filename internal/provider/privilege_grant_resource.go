package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
)

var _ resource.Resource = &PrivilegeGrantResource{}
var _ resource.ResourceWithImportState = &PrivilegeGrantResource{}

func NewPrivilegeGrantResource() resource.Resource {
	return &PrivilegeGrantResource{}
}

type PrivilegeGrantResource struct {
	client *chclient.ClickHouseClient
}

type GrantRecord struct {
	Database        string   `tfsdk:"database"`
	Table           string   `tfsdk:"table"`
	Columns         []string `tfsdk:"columns"`
	WithGrantOption bool     `tfsdk:"with_grant_option"`
}

type PrivilegeGrantResourceModel struct {
	Grantee    string        `tfsdk:"grantee"`
	AccessType string        `tfsdk:"access_type"`
	Grants     []GrantRecord `tfsdk:"grants"`
}

func (r *PrivilegeGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privilege_grant"
}

func (r *PrivilegeGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse table",
		Attributes: map[string]schema.Attribute{
			"grantee": schema.StringAttribute{
				MarkdownDescription: "User or role to grant the role to",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"access_type": schema.StringAttribute{
				MarkdownDescription: "Name of a table/view/matview/dictionary etc or '*' for all entities",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"grants": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "`TODO`",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"database": schema.StringAttribute{
							MarkdownDescription: "ClickHouse database name",
							Required:            true,
							Validators:          []validator.String{},
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"table": schema.StringAttribute{
							MarkdownDescription: "Name of a table/view/matview/dictionary etc or '*' for all entities",
							Required:            true,
							Validators:          []validator.String{grantEntityValidator{}},
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"columns": schema.ListAttribute{
							Required:            true,
							MarkdownDescription: "Columns of ClickHouse table. If empty or null, it is supposed *all* columns are allowed",
							ElementType:         types.StringType,
							Validators:          []validator.List{listvalidator.All()},
						},
						"with_grant_option": schema.BoolAttribute{
							MarkdownDescription: "Whether to grant privilege with grant option or not",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
						},
					},
				},
			},
		},
	}
}

func (r *PrivilegeGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *PrivilegeGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model PrivilegeGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, grant := range model.Grants {
		g := chclient.PrivilegeGrant{
			Grantee:   model.Grantee,
			Database:  grant.Database,
			Entity:    grant.Table,
			Privilege: model.AccessType,
			Columns:   grant.Columns,
		}
		err := r.client.GrantPrivilege(ctx, g)
		if err != nil {
			resp.Diagnostics.AddError("Failed to grant privilege", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *PrivilegeGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *PrivilegeGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *PrivilegeGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *PrivilegeGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
