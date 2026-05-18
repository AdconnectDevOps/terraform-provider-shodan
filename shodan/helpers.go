package shodan

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// syncStringList reconciles a Shodan API resource collection by adding entries
// present in plan but missing in state, then removing entries present in state
// but missing in plan. Failures surface as warnings (Shodan API can return
// transient 4xx — operator can retry). Both addFn and removeFn take (alertID,
// value) and return error.
func syncStringList(
	ctx context.Context,
	alertID string,
	stateList, planList []types.String,
	addFn func(string, string) error,
	removeFn func(string, string) error,
	kind string,
	diags *diag.Diagnostics,
) {
	stateSet := make(map[string]bool, len(stateList))
	for _, v := range stateList {
		stateSet[v.ValueString()] = true
	}
	planSet := make(map[string]bool, len(planList))
	for _, v := range planList {
		planSet[v.ValueString()] = true
	}

	// Remove entries that left the plan.
	for v := range stateSet {
		if planSet[v] {
			continue
		}
		if err := removeFn(alertID, v); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to remove %s %s from alert %s: %s", kind, v, alertID, err.Error()))
			diags.AddWarning(
				fmt.Sprintf("Warning removing %s", kind),
				fmt.Sprintf("Could not remove %s %s: %s", kind, v, err.Error()),
			)
		}
	}

	// Add entries that entered the plan.
	for v := range planSet {
		if stateSet[v] {
			continue
		}
		if err := addFn(alertID, v); err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Failed to add %s %s to alert %s: %s", kind, v, alertID, err.Error()))
			diags.AddWarning(
				fmt.Sprintf("Warning adding %s", kind),
				fmt.Sprintf("Could not add %s %s: %s", kind, v, err.Error()),
			)
		}
	}
}

// listToStringSlice extracts a []types.String from a types.List value, returning
// nil if the list is null or unknown.
func listToStringSlice(ctx context.Context, list types.List) []types.String {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []types.String
	list.ElementsAs(ctx, &out, false)
	return out
}

// setToStringSlice extracts a []types.String from a types.Set value, returning
// nil if the set is null or unknown.
func setToStringSlice(ctx context.Context, set types.Set) []types.String {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var out []types.String
	set.ElementsAs(ctx, &out, false)
	return out
}

// whitelistFromMap converts a types.Map of types.Set values into a flat
// map[trigger][]service slice — convenient shape for diffing and for the
// Shodan ignore API. Null/unknown inputs return an empty map.
func whitelistFromMap(ctx context.Context, m types.Map) map[string][]string {
	out := map[string][]string{}
	if m.IsNull() || m.IsUnknown() {
		return out
	}
	for trigger, raw := range m.Elements() {
		set, ok := raw.(types.Set)
		if !ok || set.IsNull() || set.IsUnknown() {
			continue
		}
		var services []string
		set.ElementsAs(ctx, &services, false)
		if len(services) > 0 {
			out[trigger] = services
		}
	}
	return out
}

// whitelistToMap converts a map[trigger][]service back into a types.Map for
// state assignment. Empty input produces an empty (non-null) map so the
// Computed attribute is always known and Terraform doesn't show drift between
// "absent" and "empty".
func whitelistToMap(in map[string][]string) types.Map {
	setType := types.SetType{ElemType: types.StringType}
	elements := make(map[string]attr.Value, len(in))
	for trigger, services := range in {
		sort.Strings(services)
		values := make([]attr.Value, 0, len(services))
		for _, svc := range services {
			values = append(values, types.StringValue(svc))
		}
		elements[trigger] = types.SetValueMust(types.StringType, values)
	}
	return types.MapValueMust(setType, elements)
}

// syncWhitelist reconciles per-trigger ignore lists against the Shodan API.
// For each trigger, computes set-diff(state, plan) and calls removeFn for
// state-only entries + addFn for plan-only entries. Triggers that left the
// plan have all their services removed. Failures surface as warnings so the
// next plan retries — matches syncStringList semantics.
func syncWhitelist(
	ctx context.Context,
	alertID string,
	stateMap, planMap map[string][]string,
	addFn func(string, string, string) error,
	removeFn func(string, string, string) error,
	diags *diag.Diagnostics,
) {
	triggers := map[string]struct{}{}
	for t := range stateMap {
		triggers[t] = struct{}{}
	}
	for t := range planMap {
		triggers[t] = struct{}{}
	}

	for trigger := range triggers {
		stateSet := make(map[string]bool, len(stateMap[trigger]))
		for _, v := range stateMap[trigger] {
			stateSet[v] = true
		}
		planSet := make(map[string]bool, len(planMap[trigger]))
		for _, v := range planMap[trigger] {
			planSet[v] = true
		}

		for v := range stateSet {
			if planSet[v] {
				continue
			}
			if err := removeFn(alertID, trigger, v); err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to remove whitelist %s for trigger %s on alert %s: %s", v, trigger, alertID, err.Error()))
				diags.AddWarning(
					"Warning removing whitelist entry",
					fmt.Sprintf("Could not remove %s from trigger %s: %s", v, trigger, err.Error()),
				)
			}
		}
		for v := range planSet {
			if stateSet[v] {
				continue
			}
			if err := addFn(alertID, trigger, v); err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to add whitelist %s for trigger %s on alert %s: %s", v, trigger, alertID, err.Error()))
				diags.AddWarning(
					"Warning adding whitelist entry",
					fmt.Sprintf("Could not add %s to trigger %s: %s", v, trigger, err.Error()),
				)
			}
		}
	}
}
