package shodan

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &ShodanDomainResource{}
	_ resource.ResourceWithConfigure   = &ShodanDomainResource{}
	_ resource.ResourceWithImportState = &ShodanDomainResource{}
)

// ShodanDomainResource is the resource implementation.
type ShodanDomainResource struct {
	client *ShodanClient
}

// ShodanDomainResourceModel describes the resource data model.
type ShodanDomainResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Domain             types.String   `tfsdk:"domain"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	Enabled            types.Bool     `tfsdk:"enabled"`
	Triggers           []types.String `tfsdk:"triggers"`
	Notifiers          []types.String `tfsdk:"notifiers"`
	SlackNotifications []types.String `tfsdk:"slack_notifications"`
	CreatedAt          types.String   `tfsdk:"created_at"`
}

func NewShodanDomainResource() resource.Resource {
	return &ShodanDomainResource{}
}

func (r *ShodanDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *ShodanDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Monitor a domain for security threats using Shodan alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the Shodan domain alert.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The domain name to monitor (e.g., 'example.com').",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Optional custom name for the alert. If not provided, will use '__domain: {domain}' format.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the domain monitoring alert.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the domain monitoring alert is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			"triggers": schema.ListAttribute{
				Description: "List of trigger rules to enable for domain monitoring.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"notifiers": schema.ListAttribute{
				Description: "List of notifier IDs to associate with the domain alert.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"slack_notifications": schema.ListAttribute{
				Description: "List of Slack notification IDs to associate with the domain alert.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the domain alert was created.",
				Computed:    true,
			},
		},
	}
}

func (r *ShodanDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ShodanClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ShodanClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ShodanDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ShodanDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values
	if data.Enabled.IsNull() {
		data.Enabled = types.BoolValue(true)
	}

	// Convert triggers to string slice
	var triggers []string
	if len(data.Triggers) > 0 {
		for _, trigger := range data.Triggers {
			triggers = append(triggers, trigger.ValueString())
		}
	}

	// Create domain alert without triggers first
	alertResp, err := r.client.CreateDomainAlert(data.Name.ValueString(), data.Domain.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain alert",
			fmt.Sprintf("Could not create domain alert for %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}

	// Set the ID and created timestamp
	data.ID = types.StringValue(alertResp.ID)
	data.CreatedAt = types.StringValue(alertResp.Created)

	// Add triggers if specified
	if len(triggers) > 0 {
		for _, trigger := range triggers {
			err := r.client.AddTrigger(alertResp.ID, trigger)
			if err != nil {
				resp.Diagnostics.AddWarning(
					"Warning adding trigger",
					fmt.Sprintf("Could not add trigger %s: %s", trigger, err.Error()),
				)
			}
		}
	}

	// Add notifiers if specified (after triggers are set)
	if len(data.Notifiers) > 0 {
		for _, notifier := range data.Notifiers {
			err := r.client.AddNotifier(alertResp.ID, notifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddWarning(
					"Warning adding notifier",
					fmt.Sprintf("Could not add notifier %s: %s", notifier.ValueString(), err.Error()),
				)
			}
		}
	}

	// Add Slack notifications if specified (after triggers are set)
	if len(data.SlackNotifications) > 0 {
		for _, slackNotifier := range data.SlackNotifications {
			err := r.client.AddNotifier(alertResp.ID, slackNotifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddWarning(
					"Warning adding Slack notifier",
					fmt.Sprintf("Could not add Slack notifier %s: %s", slackNotifier.ValueString(), err.Error()),
				)
			}
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShodanDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShodanDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the alert information
	alert, err := r.client.GetAlert(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain alert",
			fmt.Sprintf("Could not read domain alert %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update the model with the current state
	data.CreatedAt = types.StringValue(alert.Created)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShodanDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ShodanDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For domain alerts, we need to recreate if the domain changes
	// since the IP addresses might have changed
	var oldData ShodanDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If domain changed, we need to recreate the alert
	if oldData.Domain.ValueString() != data.Domain.ValueString() {
		// Delete the old alert
		err := r.client.DeleteAlert(oldData.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Warning deleting old alert",
				fmt.Sprintf("Could not delete old alert %s: %s", oldData.ID.ValueString(), err.Error()),
			)
		}

		// Create new alert
		alertResp, err := r.client.CreateDomainAlert(data.Name.ValueString(), data.Domain.ValueString(), nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating new domain alert",
				fmt.Sprintf("Could not create new domain alert for %s: %s", data.Domain.ValueString(), err.Error()),
			)
			return
		}

		data.ID = types.StringValue(alertResp.ID)
		data.CreatedAt = types.StringValue(alertResp.Created)

		// Add triggers if specified
		if len(data.Triggers) > 0 {
			for _, trigger := range data.Triggers {
				err := r.client.AddTrigger(alertResp.ID, trigger.ValueString())
				if err != nil {
					resp.Diagnostics.AddWarning(
						"Warning adding trigger",
						fmt.Sprintf("Could not add trigger %s: %s", trigger.ValueString(), err.Error()),
					)
				}
			}
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShodanDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShodanDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the alert
	err := r.client.DeleteAlert(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain alert",
			fmt.Sprintf("Could not delete domain alert %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}
}

func (r *ShodanDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by alert ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
