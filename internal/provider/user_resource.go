package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"net"
	"regexp"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *chclient.ClickHouseClient
}

type sha256hash struct {
	Hash string `tfsdk:"hash"`
	Salt string `tfsdk:"salt"`
}

type identifiedWith struct {
	Sha256Hash     *sha256hash `tfsdk:"sha256_hash"`
	Sha256Password *string     `tfsdk:"sha256_password"`
}

type userAllowedHosts struct {
	IP     []string `tfsdk:"ip"`
	Name   []string `tfsdk:"name"`
	Regexp []string `tfsdk:"regexp"`
	Like   []string `tfsdk:"like"`
}

type UserResourceModel struct {
	ID              types.String      `tfsdk:"id"`
	Name            string            `tfsdk:"name"`
	IdentifiedWith  identifiedWith    `tfsdk:"identified_with"`
	Hosts           *userAllowedHosts `tfsdk:"hosts"`
	DefaultDatabase types.String      `tfsdk:"default_database"`
}

func (user UserResourceModel) ToClickHouseClientUser() (chclient.ClickHouseUser, error) {
	var auth chclient.ClickHouseUserAuthType
	if user.IdentifiedWith.Sha256Hash != nil {
		auth = chclient.Sha256HashAuth{
			Hash: user.IdentifiedWith.Sha256Hash.Hash,
			Salt: user.IdentifiedWith.Sha256Hash.Salt,
		}
	} else if user.IdentifiedWith.Sha256Password != nil {
		auth = chclient.Sha256PasswordAuth{
			Password: *user.IdentifiedWith.Sha256Password,
		}
	} else {
		return chclient.ClickHouseUser{}, errors.New(
			"either IdentifiedWith.Sha256Hash or IdentifiedWith.Sha256Password should be non-nil",
		)
	}

	var hosts *chclient.ClickHouseUserHosts

	if user.Hosts != nil {
		var ipHosts []net.IPNet

		for _, cidr := range user.Hosts.IP {
			_, ipNet, err := net.ParseCIDR(cidr)

			if err != nil {
				return chclient.ClickHouseUser{}, err
			}
			ipHosts = append(ipHosts, *ipNet)
		}

		hosts = &chclient.ClickHouseUserHosts{
			Ip:     ipHosts,
			Name:   user.Hosts.Name,
			Regexp: user.Hosts.Regexp,
			Like:   user.Hosts.Like,
		}
	}

	return chclient.ClickHouseUser{
		Name:            user.Name,
		Auth:            auth,
		Hosts:           hosts,
		DefaultDatabase: chclient.DefaultDatabase(user.DefaultDatabase.ValueString()),
	}, nil
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickHouse user",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "ClickHouse user name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("[a-z0-9_]+"),
						"User name should contain only lower case latin letters, digits and _",
					),
				},
			},
			"identified_with": schema.SingleNestedAttribute{
				MarkdownDescription: "User identification method. " +
					"See: https://clickhouse.com/docs/en/sql-reference/statements/create/user#identification ." +
					"Supported methods: `sha256_hash` and `sha256_password`. Only one of them should be set",
				Required: true,
				Attributes: map[string]schema.Attribute{
					"sha256_hash": schema.SingleNestedAttribute{
						MarkdownDescription: "Settings for identification `sha256_hash` identification",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"hash": schema.StringAttribute{
								MarkdownDescription: "SHA256 of the password",
								Required:            true,
								Sensitive:           true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(
										regexp.MustCompile("[a-fA-F0-9]{64}"),
										`SHA256 hash should contain 64 hexadecimal digits (regexp: "[a-f0-9][a-fA-F0-9]{64}")`,
									),
								},
							},
							"salt": schema.StringAttribute{
								MarkdownDescription: "Salt for adding to the password before SHA256-hashing",
								Optional:            true,
								Sensitive:           true,
								Computed:            true,
								Default:             stringdefault.StaticString(""),
							},
						},
					},
					"sha256_password": schema.StringAttribute{
						MarkdownDescription: "Password for `sha256_password` identification. It is better to use " +
							"`sha256_hash` instead of `sha256_password` in order to avoid storing passwords " +
							"in terraform state",
						Optional:  true,
						Sensitive: true,
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(path.Expressions{
								path.MatchRoot("identified_with").AtName("sha256_hash"),
								path.MatchRoot("identified_with").AtName("sha256_password"),
							}...),
						},
					},
				},
			},
			"hosts": schema.SingleNestedAttribute{
				Optional: true,
				MarkdownDescription: "Hosts from which user is allowed to connect to ClickHouse. If unset, " +
					"then ANY host. If set to empty map ({}) - NONE - user won't be able to connect. " +
					"See https://clickhouse.com/docs/en/sql-reference/statements/create/user#user-host",
				Attributes: map[string]schema.Attribute{
					"ip": schema.SetAttribute{
						MarkdownDescription: "Corresponds to `HOST IP 'ip'` expression",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(),
						},
					},
					"name": schema.SetAttribute{
						MarkdownDescription: "Corresponds to `HOST NAME 'fqdn'` expression",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(),
						},
					},
					"regexp": schema.SetAttribute{
						MarkdownDescription: "Corresponds to `HOST REGEXP 'regexp'` expression",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(),
						},
					},
					"like": schema.SetAttribute{
						MarkdownDescription: "Corresponds to `HOST LIKE 'like template'` expression",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(),
						},
					},
				},
			},
			"default_database": schema.StringAttribute{
				MarkdownDescription: "Default database for user",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, err := configureClickHouseClient(ctx, req, resp)
	if err != nil {
		return
	}
	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var userModel UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &userModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userModel.ID = types.StringValue(userModel.Name)
	user, err := userModel.ToClickHouseClientUser()
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot transform user model into client-compatible struct",
			"This usually indicates a bug in the provider. Error: "+err.Error(),
		)
		return
	}

	err = r.client.CreateUser(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create user",
			"Create user query failed: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &userModel)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	receivedUser, err := r.client.GetUser(ctx, model.Name)
	if err != nil {
		handleNotFoundError(ctx, err, resp, "user", model.Name)
		return
	}
	model.Name = receivedUser.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var stateUser *UserResourceModel
	var planUser *UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateUser)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planUser)...)

	if resp.Diagnostics.HasError() {
		return
	}
	planUser.ID = types.StringValue(planUser.Name)

	user, err := planUser.ToClickHouseClientUser()
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot transform user model into client-compatible struct",
			"This usually indicates a bug in the provider. Error: "+err.Error(),
		)
		return
	}

	err = r.client.AlterUser(ctx, stateUser.Name, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot alter user",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planUser)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var user *UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &user)...)

	err := r.client.DropUser(ctx, user.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete user",
			err.Error(),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	user, err := r.client.GetUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot alter user",
			err.Error(),
		)
		return
	}
	var hosts *userAllowedHosts

	if user.Hosts != nil {
		cidrs := make([]string, 0, len(user.Hosts.Ip))
		for _, ipnet := range user.Hosts.Ip {
			cidrs = append(cidrs, ipnet.String())
		}

		hosts = &userAllowedHosts{
			IP:     cidrs,
			Name:   user.Hosts.Name,
			Regexp: user.Hosts.Regexp,
			Like:   user.Hosts.Like,
		}
	}

	emptyPassword := ""
	stateUser := UserResourceModel{
		ID:              types.StringValue(user.Name),
		Name:            user.Name,
		IdentifiedWith:  identifiedWith{Sha256Password: &emptyPassword},
		Hosts:           hosts,
		DefaultDatabase: types.StringValue(string(user.DefaultDatabase)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateUser)...)
}
