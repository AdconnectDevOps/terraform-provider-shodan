package shodan

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                  = &ShodanAlertResource{}
	_ resource.ResourceWithConfigure     = &ShodanAlertResource{}
	_ resource.ResourceWithImportState   = &ShodanAlertResource{}
	_ resource.ResourceWithUpgradeState  = &ShodanAlertResource{}
)

// ShodanAlertResource is the resource implementation.
type ShodanAlertResource struct {
	client *ShodanClient
}

// ShodanAlertResourceModel describes the resource data model.
type ShodanAlertResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Network            types.List   `tfsdk:"network"`
	Description        types.String `tfsdk:"description"`
	Tags               types.List   `tfsdk:"tags"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Triggers           types.Set    `tfsdk:"triggers"`
	Notifiers          types.Set    `tfsdk:"notifiers"`
	SlackNotifications types.Set    `tfsdk:"slack_notifications"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func NewShodanAlertResource() resource.Resource {
	return &ShodanAlertResource{}
}

// Metadata returns the resource type name.
func (r *ShodanAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

// Schema defines the schema for the resource.
func (r *ShodanAlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a Shodan network alert for monitoring specific IP ranges.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the Shodan alert.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Shodan alert.",
				Required:    true,
			},
			"network": schema.ListAttribute{
				Description: "List of IP network ranges to monitor (e.g., ['192.168.1.0/24', '5.6.7.8/32']).",
				ElementType: types.StringType,
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the alert.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags to associate with the alert.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"triggers": schema.SetAttribute{
				Description: "Set of trigger rules to enable for the alert. Order-insensitive; matches Shodan's unordered API.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"notifiers": schema.SetAttribute{
				Description: "Set of notifier IDs to associate with the alert.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"slack_notifications": schema.SetAttribute{
				Description: "Set of Slack notifier IDs to associate with the alert. Use the notifier ID from your Shodan account settings.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the alert was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ShodanAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ShodanClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *ShodanAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ShodanAlertResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the alert via Shodan API
	var networks []string
	plan.Network.ElementsAs(ctx, &networks, false)

	filters := map[string]interface{}{
		"ip": networks,
	}

	alert, err := r.client.CreateAlert(plan.Name.ValueString(), filters)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Shodan alert",
			fmt.Sprintf("Could not create alert, unexpected error: %s", err.Error()),
		)
		return
	}

	// Add triggers if specified
	if !plan.Triggers.IsNull() {
		var triggers []types.String
		plan.Triggers.ElementsAs(ctx, &triggers, false)
		for _, trigger := range triggers {
			if err := r.client.AddTrigger(alert.ID, trigger.ValueString()); err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to add trigger %s: %s", trigger.ValueString(), err.Error()))
			}
		}
	}

	// Add notifiers if specified
	if !plan.Notifiers.IsNull() {
		var notifiers []types.String
		plan.Notifiers.ElementsAs(ctx, &notifiers, false)
		for _, notifier := range notifiers {
			if err := r.client.AddNotifier(alert.ID, notifier.ValueString()); err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to add notifier %s: %s", notifier.ValueString(), err.Error()))
			}
		}
	}

	// Add Slack notifications if specified
	if !plan.SlackNotifications.IsNull() {
		var slackChannels []types.String
		plan.SlackNotifications.ElementsAs(ctx, &slackChannels, false)
		for _, channel := range slackChannels {
			if err := r.client.AddSlackNotifier(alert.ID, channel.ValueString()); err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to add Slack notification for channel %s: %s", channel.ValueString(), err.Error()))
			}
		}
	}

	// Set computed values
	plan.ID = types.StringValue(alert.ID)
	plan.CreatedAt = types.StringValue(alert.Created)
	if plan.Enabled.IsNull() {
		plan.Enabled = types.BoolValue(true)
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ShodanAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ShodanAlertResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.ValueString() == "" {
		tflog.Warn(ctx, "Alert has empty ID in state, removing")
		resp.State.RemoveResource(ctx)
		return
	}

	// Get the alert from Shodan API
	alert, err := r.client.GetAlert(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "status 404") {
			tflog.Warn(ctx, fmt.Sprintf("Alert %s returned 404, removing from state", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Shodan alert",
			fmt.Sprintf("Could not read alert %s, unexpected error: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update state with latest values
	state.Name = types.StringValue(alert.Name)
	state.CreatedAt = types.StringValue(alert.Created)
	state.Enabled = types.BoolValue(alert.HasTriggers)

	// Extract networks from filters
	if alert.Filters != nil {
		if ipFilters, ok := alert.Filters["ip"]; ok {
			if ipList, ok := ipFilters.([]interface{}); ok {
				var networks []attr.Value
				for _, ip := range ipList {
					if ipStr, ok := ip.(string); ok {
						networks = append(networks, types.StringValue(ipStr))
					}
				}
				if len(networks) > 0 {
					state.Network = types.ListValueMust(types.StringType, networks)
				}
			}
		}
	}

	// Reflect the API's current triggers in state. Without this, a previous
	// apply that only ran AddTrigger (legacy SDK behaviour) leaves state with
	// the planned-but-not-fully-applied trigger set, hiding drift on
	// subsequent plans.
	if alert.Triggers != nil {
		triggerKeys := make([]attr.Value, 0, len(alert.Triggers))
		for name := range alert.Triggers {
			triggerKeys = append(triggerKeys, types.StringValue(name))
		}
		state.Triggers = types.SetValueMust(types.StringType, triggerKeys)
	} else {
		state.Triggers = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ShodanAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ShodanAlertResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ShodanAlertResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update network filters if changed
	if !plan.Network.Equal(state.Network) {
		var networks []string
		plan.Network.ElementsAs(ctx, &networks, false)

		filters := map[string]interface{}{
			"ip": networks,
		}

		// Use state.ID instead of plan.ID since plan.ID might be empty during updates
		alertID := state.ID.ValueString()
		if alertID == "" {
			resp.Diagnostics.AddError(
				"Error updating Shodan alert network",
				"Alert ID is empty, cannot update network filters",
			)
			return
		}

		if err := r.client.UpdateAlert(alertID, filters); err != nil {
			resp.Diagnostics.AddError(
				"Error updating Shodan alert network",
				fmt.Sprintf("Could not update alert network filters, unexpected error: %s", err.Error()),
			)
			return
		}
	}

	// Reconcile triggers / notifiers / slack notifications via add+remove diff.
	if !plan.Triggers.Equal(state.Triggers) {
		syncStringList(ctx, state.ID.ValueString(),
			setToStringSlice(ctx, state.Triggers),
			setToStringSlice(ctx, plan.Triggers),
			r.client.AddTrigger, r.client.RemoveTrigger,
			"trigger", &resp.Diagnostics)
	}

	if !plan.Notifiers.Equal(state.Notifiers) {
		syncStringList(ctx, state.ID.ValueString(),
			setToStringSlice(ctx, state.Notifiers),
			setToStringSlice(ctx, plan.Notifiers),
			r.client.AddNotifier, r.client.RemoveNotifier,
			"notifier", &resp.Diagnostics)
	}

	if !plan.SlackNotifications.Equal(state.SlackNotifications) {
		syncStringList(ctx, state.ID.ValueString(),
			setToStringSlice(ctx, state.SlackNotifications),
			setToStringSlice(ctx, plan.SlackNotifications),
			r.client.AddSlackNotifier, r.client.RemoveNotifier,
			"slack notifier", &resp.Diagnostics)
	}

	// After all updates, read the current state from the API to ensure computed fields are set correctly
	updatedAlert, err := r.client.GetAlert(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated Shodan alert",
			fmt.Sprintf("Could not read updated alert %s, unexpected error: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update the plan with the latest values from the API
	plan.ID = types.StringValue(updatedAlert.ID)
	plan.CreatedAt = types.StringValue(updatedAlert.Created)
	plan.Enabled = types.BoolValue(updatedAlert.HasTriggers)

	// Extract networks from filters if available
	if updatedAlert.Filters != nil {
		if ipFilters, ok := updatedAlert.Filters["ip"]; ok {
			if ipList, ok := ipFilters.([]interface{}); ok {
				var networks []attr.Value
				for _, ip := range ipList {
					if ipStr, ok := ip.(string); ok {
						networks = append(networks, types.StringValue(ipStr))
					}
				}
				if len(networks) > 0 {
					plan.Network = types.ListValueMust(types.StringType, networks)
				}
			}
		}
	}

	// Set state with updated values
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ShodanAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ShodanAlertResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the alert via Shodan API
	err := r.client.DeleteAlert(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Shodan alert",
			fmt.Sprintf("Could not delete alert %s, unexpected error: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform state.
func (r *ShodanAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by alert ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// UpgradeState migrates v0 state (triggers/notifiers/slack_notifications as
// ordered lists) to v1 state (same fields as unordered sets). Without this
// upgrader Terraform errors with a type mismatch on the first plan after the
// schema change.
func (r *ShodanAlertResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                  schema.StringAttribute{Computed: true},
					"name":                schema.StringAttribute{Required: true},
					"network":             schema.ListAttribute{ElementType: types.StringType, Required: true},
					"description":         schema.StringAttribute{Optional: true},
					"tags":                schema.ListAttribute{ElementType: types.StringType, Optional: true},
					"enabled":             schema.BoolAttribute{Optional: true, Computed: true},
					"triggers":            schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true},
					"notifiers":           schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true},
					"slack_notifications": schema.ListAttribute{ElementType: types.StringType, Optional: true},
					"created_at":          schema.StringAttribute{Computed: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				type alertV0 struct {
					ID                 types.String `tfsdk:"id"`
					Name               types.String `tfsdk:"name"`
					Network            types.List   `tfsdk:"network"`
					Description        types.String `tfsdk:"description"`
					Tags               types.List   `tfsdk:"tags"`
					Enabled            types.Bool   `tfsdk:"enabled"`
					Triggers           types.List   `tfsdk:"triggers"`
					Notifiers          types.List   `tfsdk:"notifiers"`
					SlackNotifications types.List   `tfsdk:"slack_notifications"`
					CreatedAt          types.String `tfsdk:"created_at"`
				}
				var prior alertV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				listToSet := func(l types.List) types.Set {
					if l.IsNull() {
						return types.SetNull(types.StringType)
					}
					if l.IsUnknown() {
						return types.SetUnknown(types.StringType)
					}
					return types.SetValueMust(types.StringType, l.Elements())
				}
				upgraded := ShodanAlertResourceModel{
					ID:                 prior.ID,
					Name:               prior.Name,
					Network:            prior.Network,
					Description:        prior.Description,
					Tags:               prior.Tags,
					Enabled:            prior.Enabled,
					Triggers:           listToSet(prior.Triggers),
					Notifiers:          listToSet(prior.Notifiers),
					SlackNotifications: listToSet(prior.SlackNotifications),
					CreatedAt:          prior.CreatedAt,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}
