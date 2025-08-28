package shodan

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource = &ShodanDomainDataSource{}
)

// ShodanDomainDataSource is the data source implementation.
type ShodanDomainDataSource struct {
	client *ShodanClient
}

// ShodanDomainDataSourceModel describes the data source data model.
type ShodanDomainDataSourceModel struct {
	Domain     types.String      `tfsdk:"domain"`
	Tags       []types.String    `tfsdk:"tags"`
	Subdomains []types.String    `tfsdk:"subdomains"`
	Data       []DomainDataModel `tfsdk:"data"`
	More       types.Bool        `tfsdk:"more"`
}

// DomainDataModel represents individual DNS records for a domain
type DomainDataModel struct {
	Subdomain types.String `tfsdk:"subdomain"`
	Type      types.String `tfsdk:"type"`
	Value     types.String `tfsdk:"value"`
	LastSeen  types.String `tfsdk:"last_seen"`
}

func NewShodanDomainDataSource() datasource.DataSource {
	return &ShodanDomainDataSource{}
}

func (d *ShodanDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *ShodanDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve domain information from Shodan including subdomains and DNS records.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to lookup (e.g., 'example.com').",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the domain.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"subdomains": schema.ListAttribute{
				Description: "List of subdomains found for the domain.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"data": schema.ListNestedAttribute{
				Description: "DNS records and other data for the domain.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"subdomain": schema.StringAttribute{
							Description: "The subdomain name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The DNS record type (A, AAAA, MX, NS, etc.).",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value of the DNS record.",
							Computed:    true,
						},
						"last_seen": schema.StringAttribute{
							Description: "When this record was last seen by Shodan.",
							Computed:    true,
						},
					},
				},
			},
			"more": schema.BoolAttribute{
				Description: "Whether there are more results available.",
				Computed:    true,
			},
		},
	}
}

func (d *ShodanDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ShodanClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ShodanClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ShodanDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ShodanDomainDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get domain information from Shodan
	domainInfo, err := d.client.GetDomainInfo(data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain information",
			fmt.Sprintf("Could not read domain information for %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}

	// Convert the response to the data source model
	data.Tags = make([]types.String, len(domainInfo.Tags))
	for i, tag := range domainInfo.Tags {
		data.Tags[i] = types.StringValue(tag)
	}

	data.Subdomains = make([]types.String, len(domainInfo.Subdomains))
	for i, subdomain := range domainInfo.Subdomains {
		data.Subdomains[i] = types.StringValue(subdomain)
	}

	data.Data = make([]DomainDataModel, len(domainInfo.Data))
	for i, record := range domainInfo.Data {
		data.Data[i] = DomainDataModel{
			Subdomain: types.StringValue(record.Subdomain),
			Type:      types.StringValue(record.Type),
			Value:     types.StringValue(record.Value),
			LastSeen:  types.StringValue(record.LastSeen),
		}
	}

	data.More = types.BoolValue(domainInfo.More)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
