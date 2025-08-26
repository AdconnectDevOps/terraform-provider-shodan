package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource = &ShodanAlertDataSource{}
)

// ShodanAlertDataSource is the data source implementation.
type ShodanAlertDataSource struct {
	client *ShodanClient
}

// ShodanAlertDataSourceModel describes the data source data model.
type ShodanAlertDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Network     types.String `tfsdk:"network"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Triggers    types.List   `tfsdk:"triggers"`
	Notifiers   types.List   `tfsdk:"notifiers"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewShodanAlertDataSource() datasource.DataSource {
	return &ShodanAlertDataSource{}
}

// Metadata returns the data source type name.
func (d *ShodanAlertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

// Schema defines the schema for the data source.
func (d *ShodanAlertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Shodan network alert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the Shodan alert.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Shodan alert.",
				Computed:    true,
			},
			"network": schema.StringAttribute{
				Description: "The IP network range being monitored.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the alert.",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the alert.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert is enabled.",
				Computed:    true,
			},
			"triggers": schema.ListAttribute{
				Description: "List of trigger rules enabled for the alert.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"notifiers": schema.ListAttribute{
				Description: "List of notifier IDs associated with the alert.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the alert was created.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ShodanAlertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ShodanClient)
}

// Read refreshes the Terraform state with the latest data.
func (d *ShodanAlertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ShodanAlertDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the alert from Shodan API
	alert, err := d.client.GetAlert(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Shodan alert",
			fmt.Sprintf("Could not read alert %s, unexpected error: %s", config.ID.ValueString(), err.Error()),
		)
		return
	}

	// Set computed values
	config.Name = types.StringValue(alert.Name)
	config.CreatedAt = types.StringValue(alert.Created)
	config.Enabled = types.BoolValue(alert.HasTriggers)

	// Extract network from filters if available
	if ipFilters, ok := alert.Filters["ip"]; ok {
		if ipList, ok := ipFilters.([]interface{}); ok && len(ipList) > 0 {
			if network, ok := ipList[0].(string); ok {
				config.Network = types.StringValue(network)
			}
		}
	}

	// Set state
	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}
