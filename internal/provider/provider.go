package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/vegassor/terraform-provider-clickhouse/internal/chclient"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Ensure ClickHouseProvider satisfies various provider interfaces.
var _ provider.Provider = &ClickHouseProvider{}

// ClickHouseProvider defines the provider implementation.
type ClickHouseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ClickHouseProviderModel describes the provider data model.
type ClickHouseProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
}

func (p *ClickHouseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clickhouse"
	resp.Version = p.version
}

func (p *ClickHouseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "ClickHouse host for HTTP protocol",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "ClickHouse port for HTTP protocol",
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "ClickHouse user",
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
	if data.Port.IsNull() {
		addr = fmt.Sprintf("%s:9000", data.Host.ValueString())
	} else {
		addr = fmt.Sprintf("%s:%d", data.Host.ValueString(), data.Port.ValueInt64())
	}

	client, err := chclient.NewClickHouseClient(&clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: data.Username.ValueString(),
			Password: data.Password.ValueString(),
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot connect to ClickHouse",
			"Cannot connect to ClickHouse: "+err.Error(),
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
