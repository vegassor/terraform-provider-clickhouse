package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
)

var _ resource.Resource = &RoleGrantResource{}
var _ resource.ResourceWithImportState = &RoleGrantResource{}

func NewRoleGrantResource() resource.Resource {
	return &RoleGrantResource{}
}

type RoleGrantResource struct {
	client *chclient.ClickHouseClient
}

type RoleGrantResourceModel struct {
	Grantee         string       `tfsdk:"grantee"`
	Role            string       `tfsdk:"role"`
	WithAdminOption bool         `tfsdk:"with_admin_option"`
	GrantType       types.String `tfsdk:"grant_type"`
}

func (r *RoleGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_grant"
}

func (r *RoleGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse user",
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to grant",
				Required:            true,
				Validators:          []validator.String{ClickHouseIdentifierValidator},
			},
			"grantee": schema.StringAttribute{
				MarkdownDescription: "User or role to grant the role to",
				Required:            true,
				Validators:          []validator.String{ClickHouseIdentifierValidator},
			},
			"with_admin_option": schema.BoolAttribute{
				MarkdownDescription: "Whether to grant role with admin option or not",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"grant_type": schema.StringAttribute{
				MarkdownDescription: "Whether grant given to user or role",
				Computed:            true,
			},
		},
	}
}

func (r *RoleGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *RoleGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model RoleGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.GrantRole(ctx, model.Role, model.Grantee, model.WithAdminOption)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot grant role",
			"GRANT query failed: "+err.Error(),
		)
		return
	}

	model.GrantType = types.StringValue("user")
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *RoleGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model RoleGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	receivedGrant, err := r.client.GetRoleGrant(ctx, model.Role, model.Grantee)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot find role grant",
			err.Error(),
		)
		return
	}

	model = RoleGrantResourceModel{
		Role:            receivedGrant.Role,
		Grantee:         receivedGrant.Grantee,
		WithAdminOption: receivedGrant.WithAdminOption,
		GrantType:       types.StringValue(string(receivedGrant.RoleGrantType)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *RoleGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *RoleGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *RoleGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
