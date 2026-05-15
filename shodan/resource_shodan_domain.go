package shodan

import (
	"context"
	"fmt"
	"strings"

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ShodanDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Enabled.IsNull() {
		data.Enabled = types.BoolValue(true)
	}

	alertResp, err := r.client.CreateDomainAlert(data.Name.ValueString(), data.Domain.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain alert",
			fmt.Sprintf("Could not create domain alert for %s: %s", data.Domain.ValueString(), err.Error()),
		)
		return
	}

	data.ID = types.StringValue(alertResp.ID)
	data.CreatedAt = types.StringValue(alertResp.Created)

	for _, trigger := range data.Triggers {
		if err := r.client.AddTrigger(alertResp.ID, trigger.ValueString()); err != nil {
			resp.Diagnostics.AddWarning(
				"Warning adding trigger",
				fmt.Sprintf("Could not add trigger %s: %s", trigger.ValueString(), err.Error()),
			)
		}
	}

	for _, notifier := range data.Notifiers {
		if err := r.client.AddNotifier(alertResp.ID, notifier.ValueString()); err != nil {
			resp.Diagnostics.AddWarning(
				"Warning adding notifier",
				fmt.Sprintf("Could not add notifier %s: %s", notifier.ValueString(), err.Error()),
			)
		}
	}

	for _, slackNotifier := range data.SlackNotifications {
		if err := r.client.AddNotifier(alertResp.ID, slackNotifier.ValueString()); err != nil {
			resp.Diagnostics.AddWarning(
				"Warning adding Slack notifier",
				fmt.Sprintf("Could not add Slack notifier %s: %s", slackNotifier.ValueString(), err.Error()),
			)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShodanDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShodanDomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Recovery path: state may carry an empty ID due to a pre-0.1.16 Update bug
	// that wiped the computed ID when only triggers/notifiers changed. Look up
	// the alert by its expected name (`__domain: <domain>` or `__domain: <domain> (<name>)`).
	// ListAlerts is client-side cached, so 13 concurrent recoveries share one API call.
	if data.ID.IsNull() || data.ID.ValueString() == "" {
		expectedName := fmt.Sprintf("__domain: %s", data.Domain.ValueString())
		if !data.Name.IsNull() && data.Name.ValueString() != "" {
			expectedName = fmt.Sprintf("__domain: %s (%s)", data.Domain.ValueString(), data.Name.ValueString())
		}

		alerts, err := r.client.ListAlerts()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error recovering domain alert ID",
				fmt.Sprintf("State carries empty ID for domain %s and ListAlerts failed: %s", data.Domain.ValueString(), err.Error()),
			)
			return
		}

		for _, alert := range alerts {
			if alert.Name == expectedName {
				data.ID = types.StringValue(alert.ID)
				data.CreatedAt = types.StringValue(alert.Created)
				tflog.Info(ctx, fmt.Sprintf("Recovered ID %s for domain %s", alert.ID, data.Domain.ValueString()))
				// ListAlerts already returned the full alert record — skip the
				// redundant GetAlert and save state directly.
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}

		tflog.Warn(ctx, fmt.Sprintf("Domain alert for %s not found in Shodan, removing from state", data.Domain.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	alert, err := r.client.GetAlert(data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "status 404") {
			tflog.Warn(ctx, fmt.Sprintf("Domain alert %s returned 404, removing from state", data.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading domain alert",
			fmt.Sprintf("Could not read domain alert %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.CreatedAt = types.StringValue(alert.Created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShodanDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ShodanDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ShodanDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Carry forward computed fields from state to plan so they survive any
	// code path that doesn't reassign them.
	plan.ID = state.ID
	plan.CreatedAt = state.CreatedAt

	// If domain changed → destroy + recreate the alert (filters are domain-bound).
	if state.Domain.ValueString() != plan.Domain.ValueString() {
		if err := r.client.DeleteAlert(state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddWarning(
				"Warning deleting old alert",
				fmt.Sprintf("Could not delete old alert %s: %s", state.ID.ValueString(), err.Error()),
			)
		}

		alertResp, err := r.client.CreateDomainAlert(plan.Name.ValueString(), plan.Domain.ValueString(), nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating new domain alert",
				fmt.Sprintf("Could not create new domain alert for %s: %s", plan.Domain.ValueString(), err.Error()),
			)
			return
		}

		plan.ID = types.StringValue(alertResp.ID)
		plan.CreatedAt = types.StringValue(alertResp.Created)

		// On recreate every plan trigger / notifier is added fresh; state diffs are irrelevant.
		state.Triggers = nil
		state.Notifiers = nil
		state.SlackNotifications = nil
	}

	if plan.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Cannot update domain alert",
			fmt.Sprintf("Alert ID is empty for domain %s", plan.Domain.ValueString()),
		)
		return
	}

	syncStringList(ctx, plan.ID.ValueString(), state.Triggers, plan.Triggers,
		r.client.AddTrigger, r.client.RemoveTrigger,
		"trigger", &resp.Diagnostics)

	syncStringList(ctx, plan.ID.ValueString(), state.Notifiers, plan.Notifiers,
		r.client.AddNotifier, r.client.RemoveNotifier,
		"notifier", &resp.Diagnostics)

	syncStringList(ctx, plan.ID.ValueString(), state.SlackNotifications, plan.SlackNotifications,
		r.client.AddNotifier, r.client.RemoveNotifier,
		"slack notifier", &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ShodanDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShodanDomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.ValueString() == "" {
		return
	}

	if err := r.client.DeleteAlert(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain alert",
			fmt.Sprintf("Could not delete domain alert %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}
}

func (r *ShodanDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
