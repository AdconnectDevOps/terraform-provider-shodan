package shodan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newDirectClient wires a ShodanClient to a httptest server bypassing the
// 1-second interval floor in NewRateLimitedHTTPClient. Use this when a test
// makes multiple sequential requests or exercises retry behaviour.
func newDirectClient(t *testing.T, h http.HandlerFunc) (*ShodanClient, func()) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := &ShodanClient{
		ApiKey:  "test-key",
		BaseURL: srv.URL,
		HTTPClient: &RateLimitedHTTPClient{
			client:          srv.Client(),
			requestInterval: 0, // no rate limit in tests
		},
	}
	return c, srv.Close
}

// --- AddTrigger / RemoveTrigger ---------------------------------------------

func TestAddTrigger_URLAndMethod(t *testing.T) {
	var gotMethod, gotPath string
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	if err := c.AddTrigger("A1", "new_service"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	want := "/shodan/alert/A1/trigger/new_service"
	if gotPath != want {
		t.Errorf("path = %s, want %s", gotPath, want)
	}
}

// 0.1.16 regression — RemoveTrigger must treat 404 as success (idempotent
// DELETE). Without this, removing a trigger that was deleted out-of-band
// fails the apply.
func TestRemoveTrigger_404IsSuccess(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	if err := c.RemoveTrigger("A1", "uncommon"); err != nil {
		t.Errorf("404 should be success, got error: %v", err)
	}
}

func TestRemoveTrigger_500IsError(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()

	if err := c.RemoveTrigger("A1", "uncommon"); err == nil {
		t.Error("expected error on 500")
	}
}

func TestAddTrigger_EmptyAlertID(t *testing.T) {
	c := &ShodanClient{ApiKey: "k", BaseURL: "http://unreachable", HTTPClient: &RateLimitedHTTPClient{client: &http.Client{}}}
	if err := c.AddTrigger("", "new_service"); err == nil {
		t.Error("expected error on empty alertID")
	}
}

// --- AddNotifier / RemoveNotifier -------------------------------------------

func TestAddNotifier_URLAndMethod(t *testing.T) {
	var gotMethod, gotPath string
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	if err := c.AddNotifier("A1", "default"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	want := "/shodan/alert/A1/notifier/default"
	if gotPath != want {
		t.Errorf("path = %s, want %s", gotPath, want)
	}
}

func TestRemoveNotifier_404IsSuccess(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	if err := c.RemoveNotifier("A1", "default"); err != nil {
		t.Errorf("404 should be success, got error: %v", err)
	}
}

// --- GetAlert ---------------------------------------------------------------

// 0.1.16 regression — Read used to hard-error on 404. The fix moved the
// translation to the resource Read, but the client method still surfaces
// "status 404" in the error string so Read can string-match it. This test
// pins the contract.
func TestGetAlert_404SurfacedAsStatusString(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{"error": "not found"}`)
	})
	defer cleanup()

	_, err := c.GetAlert("A1")
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("error should contain 'status 404' (Read matches on this), got: %v", err)
	}
}

func TestGetAlert_EmptyAlertID(t *testing.T) {
	c := &ShodanClient{ApiKey: "k", BaseURL: "http://unreachable", HTTPClient: &RateLimitedHTTPClient{client: &http.Client{}}}
	if _, err := c.GetAlert(""); err == nil {
		t.Error("expected error on empty alertID")
	}
}

func TestGetAlert_OK(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"id":"A1","name":"test","created":"2026-01-01T00:00:00Z","has_triggers":true,"triggers":{"new_service":{"rule":"new_service"}},"filters":{"ip":["203.0.113.10/32"]}}`)
	})
	defer cleanup()

	alert, err := c.GetAlert("A1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert.ID != "A1" || alert.Name != "test" {
		t.Errorf("alert fields not parsed: %+v", alert)
	}
	if _, ok := alert.Triggers["new_service"]; !ok {
		t.Error("triggers map missing expected key")
	}
}

// --- UpdateAlert ------------------------------------------------------------

func TestUpdateAlert_URLMethodAndBody(t *testing.T) {
	var gotMethod, gotPath, gotCT string
	var gotBody []byte
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	err := c.UpdateAlert("A1", map[string]interface{}{"ip": []string{"203.0.113.10/32"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/shodan/alert/A1" {
		t.Errorf("path = %s", gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", gotCT)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if _, ok := payload["filters"]; !ok {
		t.Errorf("body missing 'filters' key: %s", gotBody)
	}
}

// --- DeleteAlert ------------------------------------------------------------

func TestDeleteAlert_404IsSuccess(t *testing.T) {
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	if err := c.DeleteAlert("A1"); err != nil {
		t.Errorf("404 should be success on delete, got: %v", err)
	}
}

// --- ListAlerts cache -------------------------------------------------------

// 0.1.17 regression — ListAlerts must cache per client process for
// alertsCacheTTL (60s). Without this the recovery path issued N×2 API
// calls and tripped Shodan's rate limit.
func TestListAlerts_Cache(t *testing.T) {
	var hits int32
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, `[{"id":"A1","name":"test"}]`)
	})
	defer cleanup()

	for i := 0; i < 5; i++ {
		if _, err := c.ListAlerts(); err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("expected 1 server hit (cached for %s), got %d", alertsCacheTTL, got)
	}
}

func TestListAlerts_CacheExpiry(t *testing.T) {
	var hits int32
	c, cleanup := newDirectClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, `[]`)
	})
	defer cleanup()

	if _, err := c.ListAlerts(); err != nil {
		t.Fatal(err)
	}
	// Manually expire the cache (simulate alertsCacheTTL elapsing without
	// actually sleeping 60s — we test the TTL boundary check, not the wall
	// clock).
	c.cacheMu.Lock()
	c.alertsCacheAt = time.Now().Add(-2 * alertsCacheTTL)
	c.cacheMu.Unlock()

	if _, err := c.ListAlerts(); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Errorf("expected 2 hits after cache expiry, got %d", got)
	}
}
