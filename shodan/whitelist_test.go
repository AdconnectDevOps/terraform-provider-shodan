package shodan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestExtractIgnoredServices covers the shapes the Shodan API actually emits
// (port as float64 from json.Unmarshal default) plus shapes a future API
// version might emit (int, string). Malformed entries are silently skipped —
// that's the forward-compat contract documented in CLAUDE.md.
func TestExtractIgnoredServices(t *testing.T) {
	tests := map[string]struct {
		in   map[string]interface{}
		want map[string][]string
	}{
		"nil input": {
			in:   nil,
			want: map[string][]string{},
		},
		"empty triggers": {
			in:   map[string]interface{}{},
			want: map[string][]string{},
		},
		"port as float64 (real Shodan response)": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"rule": "new_service",
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10", "port": float64(3307)},
					},
				},
			},
			want: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
		},
		"port as int": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10", "port": 3307},
					},
				},
			},
			want: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
		},
		"port as string": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10", "port": "3307"},
					},
				},
			},
			want: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
		},
		"trigger without ignore key": {
			in: map[string]interface{}{
				"vulnerable": map[string]interface{}{"rule": "vulnerable"},
			},
			want: map[string][]string{},
		},
		"trigger not a map (e.g. legacy bare-name shape)": {
			in: map[string]interface{}{
				"malware": "enabled",
			},
			want: map[string][]string{},
		},
		"ignore entry missing ip": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"port": float64(3307)},
					},
				},
			},
			want: map[string][]string{},
		},
		"ignore entry missing port": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10"},
					},
				},
			},
			want: map[string][]string{},
		},
		"multiple triggers, multiple services": {
			in: map[string]interface{}{
				"new_service": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10", "port": float64(3307)},
						map[string]interface{}{"ip": "203.0.113.11", "port": float64(3307)},
					},
				},
				"open_database": map[string]interface{}{
					"ignore": []interface{}{
						map[string]interface{}{"ip": "203.0.113.10", "port": float64(6379)},
					},
				},
				"vulnerable": map[string]interface{}{}, // no ignore — not in output
			},
			want: map[string][]string{
				"new_service":   {"203.0.113.10:3307", "203.0.113.11:3307"},
				"open_database": {"203.0.113.10:6379"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ExtractIgnoredServices(tc.in)
			// Sort both sides — ExtractIgnoredServices iterates a Go map,
			// so per-trigger entry order is non-deterministic.
			for k := range got {
				sort.Strings(got[k])
			}
			for k := range tc.want {
				sort.Strings(tc.want[k])
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

// TestWhitelistMapRoundTrip verifies that whitelistToMap → whitelistFromMap
// is a stable round-trip. Important because the Computed attribute is read
// from API via these helpers on every Read, and any asymmetry produces
// permanent plan drift.
func TestWhitelistMapRoundTrip(t *testing.T) {
	tests := map[string]map[string][]string{
		"empty": {},
		"single trigger, single service": {
			"new_service": {"203.0.113.10:3307"},
		},
		"single trigger, multiple services": {
			"new_service": {"203.0.113.10:3307", "203.0.113.11:3307"},
		},
		"multiple triggers": {
			"new_service":   {"203.0.113.10:3307"},
			"open_database": {"203.0.113.10:6379"},
		},
	}

	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			tfMap := whitelistToMap(in)
			if tfMap.IsNull() {
				t.Fatal("whitelistToMap returned null — empty map breaks Computed attribute (state drift)")
			}
			out := whitelistFromMap(context.Background(), tfMap)
			for k := range in {
				sort.Strings(in[k])
			}
			for k := range out {
				sort.Strings(out[k])
			}
			// Empty input has special semantics: whitelistToMap returns
			// MapValue{} (non-null), whitelistFromMap returns empty map[].
			if len(in) == 0 && len(out) == 0 {
				return
			}
			if !reflect.DeepEqual(out, in) {
				t.Errorf("round-trip mismatch:\n  in:  %#v\n  out: %#v", in, out)
			}
		})
	}
}

func TestWhitelistFromMap_NullInputs(t *testing.T) {
	setType := types.SetType{ElemType: types.StringType}
	ctx := context.Background()

	cases := map[string]types.Map{
		"null":    types.MapNull(setType),
		"unknown": types.MapUnknown(setType),
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			got := whitelistFromMap(ctx, m)
			if len(got) != 0 {
				t.Errorf("expected empty map for %s input, got %#v", name, got)
			}
		})
	}
}

// TestSyncWhitelist exercises the per-trigger diff logic. Mock add/remove
// funcs record what was called so each test asserts the exact API calls
// the real reconciler would issue.
func TestSyncWhitelist(t *testing.T) {
	type call struct {
		op      string // "add" | "remove"
		trigger string
		service string
	}

	tests := map[string]struct {
		state map[string][]string
		plan  map[string][]string
		want  []call
	}{
		"both empty — no calls": {
			state: map[string][]string{},
			plan:  map[string][]string{},
			want:  nil,
		},
		"state empty, plan adds entries": {
			state: map[string][]string{},
			plan: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
			want: []call{{op: "add", trigger: "new_service", service: "203.0.113.10:3307"}},
		},
		"plan empty, state has entries — remove all": {
			state: map[string][]string{
				"new_service": {"203.0.113.10:3307", "203.0.113.11:3307"},
			},
			plan: map[string][]string{},
			want: []call{
				{op: "remove", trigger: "new_service", service: "203.0.113.10:3307"},
				{op: "remove", trigger: "new_service", service: "203.0.113.11:3307"},
			},
		},
		"trigger removed entirely": {
			state: map[string][]string{
				"new_service":   {"203.0.113.10:3307"},
				"open_database": {"203.0.113.10:6379"},
			},
			plan: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
			want: []call{{op: "remove", trigger: "open_database", service: "203.0.113.10:6379"}},
		},
		"trigger gains a service": {
			state: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
			plan: map[string][]string{
				"new_service": {"203.0.113.10:3307", "203.0.113.11:3307"},
			},
			want: []call{{op: "add", trigger: "new_service", service: "203.0.113.11:3307"}},
		},
		"trigger loses a service": {
			state: map[string][]string{
				"new_service": {"203.0.113.10:3307", "203.0.113.11:3307"},
			},
			plan: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
			want: []call{{op: "remove", trigger: "new_service", service: "203.0.113.11:3307"}},
		},
		"no overlap — full swap": {
			state: map[string][]string{
				"new_service": {"203.0.113.10:3307"},
			},
			plan: map[string][]string{
				"new_service": {"203.0.113.11:3307"},
			},
			want: []call{
				{op: "remove", trigger: "new_service", service: "203.0.113.10:3307"},
				{op: "add", trigger: "new_service", service: "203.0.113.11:3307"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			callsPtr := new([]call)
			addFn := func(_, trigger, service string) error {
				*callsPtr = append(*callsPtr, call{op: "add", trigger: trigger, service: service})
				return nil
			}
			removeFn := func(_, trigger, service string) error {
				*callsPtr = append(*callsPtr, call{op: "remove", trigger: trigger, service: service})
				return nil
			}

			diags := diag.Diagnostics{}
			syncWhitelist(context.Background(), "alert-id", tc.state, tc.plan, addFn, removeFn, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected error diagnostics: %v", diags)
			}

			got := *callsPtr
			// Trigger iteration order is non-deterministic; sort both sides
			// by (trigger, service) keeping op stable for the comparison.
			sortCalls := func(c []call) {
				sort.SliceStable(c, func(i, j int) bool {
					if c[i].trigger != c[j].trigger {
						return c[i].trigger < c[j].trigger
					}
					if c[i].op != c[j].op {
						return c[i].op < c[j].op
					}
					return c[i].service < c[j].service
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

// httptest-based tests for the two new client methods. Verify URL shape +
// status-code handling without hitting the real Shodan API.

func newTestClient(t *testing.T, h http.HandlerFunc) (*ShodanClient, func()) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := &ShodanClient{
		ApiKey:     "test-key",
		BaseURL:    srv.URL,
		HTTPClient: NewRateLimitedHTTPClient(srv.Client(), 0),
	}
	return c, srv.Close
}

func TestAddIgnoreService_URLAndMethod(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	c, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	if err := c.AddIgnoreService("AAA111", "new_service", "203.0.113.10:3307"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	wantPath := "/shodan/alert/AAA111/trigger/new_service/ignore/203.0.113.10:3307"
	if gotPath != wantPath {
		t.Errorf("path = %s, want %s", gotPath, wantPath)
	}
	if !strings.Contains(gotQuery, "key=test-key") {
		t.Errorf("query missing api key: %s", gotQuery)
	}
}

func TestAddIgnoreService_EmptyArgGuards(t *testing.T) {
	c := &ShodanClient{ApiKey: "k", BaseURL: "http://unreachable", HTTPClient: NewRateLimitedHTTPClient(&http.Client{}, 0)}
	cases := map[string]struct{ id, trig, svc string }{
		"empty alertID": {"", "new_service", "203.0.113.10:3307"},
		"empty trigger": {"a", "", "203.0.113.10:3307"},
		"empty service": {"a", "new_service", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := c.AddIgnoreService(tc.id, tc.trig, tc.svc); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestRemoveIgnoreService_404IsSuccess(t *testing.T) {
	c, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{"error": "not found"}`)
	})
	defer cleanup()

	if err := c.RemoveIgnoreService("AAA111", "new_service", "203.0.113.10:3307"); err != nil {
		t.Errorf("404 should be treated as success (idempotent), got error: %v", err)
	}
}

func TestRemoveIgnoreService_500IsError(t *testing.T) {
	c, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, `{"error": "boom"}`)
	})
	defer cleanup()

	if err := c.RemoveIgnoreService("AAA111", "new_service", "203.0.113.10:3307"); err == nil {
		t.Error("expected error on 500, got nil")
	}
}
