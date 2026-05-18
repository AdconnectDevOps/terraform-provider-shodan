package shodan

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSyncStringList is the canonical test for the bidirectional set-diff
// reconciler used by triggers / notifiers / slack_notifications. The 0.1.16
// bug ("Update only added triggers, never removed") would have been caught
// by the "state has, plan empty — remove all" case below.
func TestSyncStringList(t *testing.T) {
	type call struct {
		op    string // "add" | "remove"
		value string
	}

	toStringSlice := func(in []string) []types.String {
		out := make([]types.String, 0, len(in))
		for _, v := range in {
			out = append(out, types.StringValue(v))
		}
		return out
	}

	tests := map[string]struct {
		state []string
		plan  []string
		want  []call
	}{
		"both empty — no calls": {
			state: nil,
			plan:  nil,
			want:  nil,
		},
		"state empty, plan adds": {
			state: nil,
			plan:  []string{"new_service", "vulnerable"},
			want: []call{
				{op: "add", value: "new_service"},
				{op: "add", value: "vulnerable"},
			},
		},
		"plan empty, state had entries — remove all (0.1.16 regression)": {
			state: []string{"uncommon", "uncommon_plus"},
			plan:  nil,
			want: []call{
				{op: "remove", value: "uncommon"},
				{op: "remove", value: "uncommon_plus"},
			},
		},
		"overlap — only diff": {
			state: []string{"new_service", "vulnerable"},
			plan:  []string{"new_service", "malware"},
			want: []call{
				{op: "remove", value: "vulnerable"},
				{op: "add", value: "malware"},
			},
		},
		"identical — no calls": {
			state: []string{"new_service", "vulnerable"},
			plan:  []string{"vulnerable", "new_service"}, // order differs, set semantics
			want:  nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			callsPtr := new([]call)
			addFn := func(_, v string) error {
				*callsPtr = append(*callsPtr, call{op: "add", value: v})
				return nil
			}
			removeFn := func(_, v string) error {
				*callsPtr = append(*callsPtr, call{op: "remove", value: v})
				return nil
			}

			diags := diag.Diagnostics{}
			syncStringList(context.Background(), "alert-id",
				toStringSlice(tc.state),
				toStringSlice(tc.plan),
				addFn, removeFn,
				"trigger", &diags)
			if diags.HasError() {
				t.Fatalf("unexpected error diagnostics: %v", diags)
			}

			got := *callsPtr
			sortCalls := func(c []call) {
				sort.SliceStable(c, func(i, j int) bool {
					if c[i].op != c[j].op {
						return c[i].op < c[j].op
					}
					return c[i].value < c[j].value
				})
			}
			sortCalls(got)
			want := append([]call(nil), tc.want...)
			sortCalls(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("call sequence mismatch:\n  got:  %#v\n  want: %#v", got, want)
			}
		})
	}
}

// TestSyncStringList_FailuresSurfaceAsWarnings verifies that an add/remove
// failure is recorded as a warning (not a hard error) — matches the
// documented "Shodan can return transient 4xx; partial reconcile is
// preferable to aborting apply" contract in helpers.go.
func TestSyncStringList_FailuresSurfaceAsWarnings(t *testing.T) {
	addFn := func(_, _ string) error { return errors.New("boom-add") }
	removeFn := func(_, _ string) error { return errors.New("boom-remove") }

	diags := diag.Diagnostics{}
	syncStringList(context.Background(), "alert-id",
		[]types.String{types.StringValue("uncommon")},   // state: will be removed → fails
		[]types.String{types.StringValue("new_service")}, // plan: will be added → fails
		addFn, removeFn, "trigger", &diags)

	if diags.HasError() {
		t.Fatal("transient API failures must not raise hard errors — break the apply unnecessarily")
	}
	if len(diags.Warnings()) != 2 {
		t.Errorf("expected 2 warnings (one per failed op), got %d", len(diags.Warnings()))
	}
}

func TestListToStringSlice(t *testing.T) {
	ctx := context.Background()
	mkAttrs := func(vs ...string) []attr.Value {
		out := make([]attr.Value, 0, len(vs))
		for _, v := range vs {
			out = append(out, types.StringValue(v))
		}
		return out
	}

	t.Run("null", func(t *testing.T) {
		if got := listToStringSlice(ctx, types.ListNull(types.StringType)); got != nil {
			t.Errorf("expected nil for null list, got %v", got)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		if got := listToStringSlice(ctx, types.ListUnknown(types.StringType)); got != nil {
			t.Errorf("expected nil for unknown list, got %v", got)
		}
	})
	t.Run("populated", func(t *testing.T) {
		l := types.ListValueMust(types.StringType, mkAttrs("a", "b"))
		got := listToStringSlice(ctx, l)
		if len(got) != 2 || got[0].ValueString() != "a" || got[1].ValueString() != "b" {
			t.Errorf("unexpected output: %v", got)
		}
	})
}

func TestSetToStringSlice(t *testing.T) {
	ctx := context.Background()
	mkAttrs := func(vs ...string) []attr.Value {
		out := make([]attr.Value, 0, len(vs))
		for _, v := range vs {
			out = append(out, types.StringValue(v))
		}
		return out
	}

	t.Run("null", func(t *testing.T) {
		if got := setToStringSlice(ctx, types.SetNull(types.StringType)); got != nil {
			t.Errorf("expected nil for null set, got %v", got)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		if got := setToStringSlice(ctx, types.SetUnknown(types.StringType)); got != nil {
			t.Errorf("expected nil for unknown set, got %v", got)
		}
	})
	t.Run("populated", func(t *testing.T) {
		s := types.SetValueMust(types.StringType, mkAttrs("a", "b"))
		got := setToStringSlice(ctx, s)
		if len(got) != 2 {
			t.Errorf("expected 2 elements, got %d", len(got))
		}
	})
}
