package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	Grantee         string `tfsdk:"grantee"`
	Role            string `tfsdk:"role"`
	WithAdminOption bool   `tfsdk:"with_admin_option"`
}

func (r *RoleGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_grant"
}

func (r *RoleGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Grant of ClickHouse role to user or another role",
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				MarkdownDescription: "Role to grant",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"grantee": schema.StringAttribute{
				MarkdownDescription: "User or role to grant the role to",
				Required:            true,
				Validators:          []validator.String{clickHouseIdentifierValidator},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"with_admin_option": schema.BoolAttribute{
				MarkdownDescription: "Whether to grant role with admin option or not",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
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
		name := fmt.Sprintf("(role=%s, grantee=%s)", model.Role, model.Grantee)
		handleNotFoundError(ctx, err, resp, "role grant", name)
		return
	}

	model = RoleGrantResourceModel{
		Role:            receivedGrant.Role,
		Grantee:         receivedGrant.Grantee,
		WithAdminOption: receivedGrant.WithAdminOption,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *RoleGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Cannot update role grant",
		"Role grant should always be recreated. "+
			"If you see this error - it is a bug in the provider.",
	)
}

func (r *RoleGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model RoleGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RevokeRole(ctx, model.Role, model.Grantee)
	if err == nil {
		return
	}

	var notFoundError *chclient.NotFoundError
	if errors.As(err, &notFoundError) {
		tflog.Info(ctx, "Role grant already deleted", dict{"err": err.Error()})
		return
	}

	resp.Diagnostics.AddError(
		"Cannot revoke role grant",
		err.Error(),
	)
}

func (r *RoleGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
