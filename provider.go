package main

import (
	"context"
	"net/http"

	"github.com/AdconnectDevOps/terraform-provider-shodan/shodan"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &ShodanProvider{}
)

// ShodanProvider is the provider implementation.
type ShodanProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ShodanProviderModel describes the provider data model.
type ShodanProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *ShodanProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shodan"
	resp.Version = p.version
}

func (p *ShodanProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Shodan API to manage network alerts and monitoring.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Shodan API key for authentication. Can also be set via SHODAN_API_KEY environment variable.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *ShodanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ShodanProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if config.ApiKey.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := &shodan.ShodanClient{
		ApiKey:     config.ApiKey.ValueString(),
		BaseURL:    "https://api.shodan.io",
		HTTPClient: &http.Client{},
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ShodanProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		shodan.NewShodanAlertResource,
	}
}

func (p *ShodanProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		shodan.NewShodanAlertDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShodanProvider{
			version: version,
		}
	}
}
