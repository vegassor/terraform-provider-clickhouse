package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
			"grants": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "`TODO`",
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown(), setplanmodifier.RequiresReplace()},
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
							Optional:            true,
							MarkdownDescription: "Columns of ClickHouse table. If empty or null, it is supposed *all* columns are allowed",
							PlanModifiers:       []planmodifier.List{listplanmodifier.RequiresReplace(), listplanmodifier.UseStateForUnknown()},
							ElementType:         types.StringType,
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
			Grantee:    model.Grantee,
			Database:   grant.Database,
			Table:      grant.Table,
			AccessType: model.AccessType,
			Columns:    grant.Columns,
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
	var stateModel PrivilegeGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	receivedGrants, err := r.client.GetPrivilegeGrants(ctx, stateModel.Grantee, stateModel.AccessType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot find privilege grant",
			err.Error(),
		)
		return
	}

	grants := make([]GrantRecord, 0, len(receivedGrants))
	for _, grant := range receivedGrants {
		if len(grant.Columns) == 0 {
			grant.Columns = nil
		}
		grants = append(grants, GrantRecord{
			Database:        grant.Database,
			Table:           grant.Table,
			Columns:         grant.Columns,
			WithGrantOption: grant.GrantOption,
		})
	}

	readModel := PrivilegeGrantResourceModel{
		Grantee:    stateModel.Grantee,
		AccessType: stateModel.AccessType,
		Grants:     grants,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *PrivilegeGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update is not supported",
		"Update is not supported for privilege_grant resource. "+
			"You should never see this message, because every change must cause the "+
			"resource to be re-created. Update may be implemented in future releases.",
	)
}

func (r *PrivilegeGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model PrivilegeGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant := chclient.PrivilegeGrant{
		Grantee:    model.Grantee,
		AccessType: model.AccessType,
		Database:   "*",
		Table:      "*",
		Columns:    make([]string, 0),
	}
	err := r.client.RevokePrivilege(ctx, grant)
	if err != nil {
		resp.Diagnostics.AddError("Failed to revoke privilege", err.Error())
		return
	}
}

func (r *PrivilegeGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
