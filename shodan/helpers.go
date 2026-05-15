package shodan

import (
	"context"
	"fmt"

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
