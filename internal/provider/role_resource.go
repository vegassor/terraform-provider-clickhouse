package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"regexp"
)

var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	client *chclient.ClickHouseClient
}

type RoleResourceModel struct {
	Name string `tfsdk:"name"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse user",
		Attributes: map[string]schema.Attribute{
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
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
