package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ClickHouse/clickhouse-go/v2"
)

var _ provider.Provider = &ClickHouseProvider{}

type ClickHouseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ClickHouseProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
}

func (p *ClickHouseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clickhouse"
	resp.Version = p.version
}

func (p *ClickHouseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{
				Optional:    true,
				Description: "Protocol for connection to ClickHouse. Must be one of `http` or `native`",
				Validators:  []validator.String{stringvalidator.OneOfCaseInsensitive("http", "native")},
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "ClickHouse host, e.g. `localhost`",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "ClickHouse port, e.g. 9000. If not specified, default port will be used (8123 for `http` and 9000 for `native`)",
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "ClickHouse user that have enough permissions to manage databases, users, tables, etc.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for ClickHouse user",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ClickHouseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ClickHouseProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var addr string
	var port int64
	var protocol string

	if data.Protocol.IsNull() {
		protocol = "native"
	} else {
		protocol = strings.ToLower(data.Protocol.ValueString())
	}

	proto := clickhouse.Native
	if protocol == "http" {
		proto = clickhouse.HTTP
	}

	if data.Port.IsNull() {
		if protocol == "http" {
			port = 8123
		} else {
			port = 9000
		}
	} else {
		port = data.Port.ValueInt64()
	}
	addr = fmt.Sprintf("%s:%d", data.Host.ValueString(), port)

	var password string
	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	} else if pwEnv := os.Getenv("CLICKHOUSE_PASSWORD"); pwEnv != "" {
		password = pwEnv
	} else {
		resp.Diagnostics.AddError(
			"ClickHouse password is not set",
			"Either set password attribute or export CLICKHOUSE_PASSWORD environment variable",
		)
		return
	}

	client, err := chclient.NewClickHouseClient(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: data.Username.ValueString(),
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: proto,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot connect to ClickHouse",
			"Cannot connect to ClickHouse: "+err.Error()+
				"\naddr: "+addr,
		)
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ClickHouseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseResource,
		NewUserResource,
		NewTableResource,
		NewRoleResource,
		NewRoleGrantResource,
		NewPrivilegeGrantResource,
		NewViewResource,
	}
}

func (p *ClickHouseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ClickHouseProvider{
			version: version,
		}
	}
}

func configureClickHouseClient(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*chclient.ClickHouseClient, error) {
	if req.ProviderData == nil {
		return nil, errors.New("the provider has not been configured")
	}

	client, ok := req.ProviderData.(*chclient.ClickHouseClient)

	if !ok {
		err := fmt.Sprintf("Expected *chclient.ClickHouseClient, got: %T. Please report this issue to the provider developers.", req.ProviderData)
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", err)

		return nil, errors.New(err)
	}

	return client, nil
}

type dict map[string]interface{}
