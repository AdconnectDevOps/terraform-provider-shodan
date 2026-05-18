# CLAUDE.md

Notes for contributors (and AI assistants) working on this provider.

## What this is

A Terraform provider for the Shodan Monitor API (network and domain alerts). Built on `terraform-plugin-framework` v1.x. Published to the Terraform Registry as `AdconnectDevOps/shodan`.

## Layout

```
.
├── main.go                            # provider entrypoint (server, debug flag)
├── provider.go                        # ShodanProvider — schema, Configure, resource/datasource registration
├── shodan/                            # provider package
│   ├── client.go                      # REST client + AlertResponse + DomainInfo types
│   ├── rate_limiter.go                # mutex-based request spacer
│   ├── helpers.go                     # syncStringList + list conversion helpers
│   ├── resource_shodan_alert.go       # shodan_alert resource (network monitoring)
│   ├── resource_shodan_domain.go      # shodan_domain resource (domain monitoring)
│   ├── datasource_shodan_alert.go     # data source: read existing alert by id
│   └── datasource_shodan_domain.go    # data source: read domain info (subdomains, DNS records)
├── docs/                              # user-facing docs published to TF registry
├── examples/                          # runnable HCL examples
├── .goreleaser.yml                    # release build matrix
└── .github/workflows/                 # CI (test on PR, release on tag push)
```

## Common commands

```bash
make build      # build for darwin/arm64 → ./terraform-provider-shodan
make install    # build + drop into ~/.terraform.d/plugins/ for local testing
make test       # go test ./...
make fmt        # go fmt ./...
make vet        # go vet ./...
make deps       # go mod tidy + download
make dev        # run with -debug flag (attach Terraform via TF_REATTACH_PROVIDERS)
```

## Release flow

Releases are triggered by pushing a tag matching `v*`:

```bash
git tag v0.1.16
git push origin v0.1.16
```

GitHub Actions (`.github/workflows/release.yml`) runs goreleaser, which builds linux/darwin × amd64/arm64 binaries, signs the checksum file with GPG (key fingerprint from `secrets.GPG_PRIVATE_KEY`), and creates a GitHub release with all artifacts. Terraform Registry picks up new releases automatically (when the repo is registered with it).

Before tagging:
1. Update `CHANGELOG.md` — move the `Unreleased` entries under the new version header with today's date.
2. `make vet && make build` locally — catch the obvious things.
3. Ensure `go.sum` is current (`go mod tidy`).

After pushing the tag:
- Watch goreleaser job: `gh run watch <run-id> --exit-status` (find id via `gh run list --repo AdconnectDevOps/terraform-provider-shodan --limit 3`). Typical runtime ~1m45s.
- Verify Registry pickup (within ~10–15 min of release success):
  ```bash
  curl -s https://registry.terraform.io/v1/providers/AdconnectDevOps/shodan/versions \
    | python3 -c "import json,sys; print(sorted([v['version'] for v in json.load(sys.stdin)['versions']], key=lambda x:[int(p) for p in x.split('.')])[-3:])"
  ```

## Provider Framework conventions used here

### Resource lifecycle methods

Every `resource.Resource` implements `Create`, `Read`, `Update`, `Delete`, `ImportState`. The patterns this provider uses:

- **`Create`** — read plan, call API, write the API-returned ID + computed fields back into the plan struct, `resp.State.Set(ctx, &plan)`.
- **`Read`** — read state, call API, populate computed fields, write back to state. On 404 from the API, call `resp.State.RemoveResource(ctx)` and return (Terraform will recreate on next plan).
- **`Update`** — read **both** plan and state, carry forward computed fields from state to plan (`plan.ID = state.ID` etc.), reconcile API-side resources via diffs, write merged result to state.
- **`Delete`** — read state, call API. Treat 404 as success (already gone).
- **`ImportState`** — `resp.State.SetAttribute(ctx, path.Root("id"), req.ID)`. Refresh then populates the rest.

### Computed `id` / `created_at` must use `UseStateForUnknown`

```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

Without this modifier, Terraform marks the value as `<known after apply>` on every plan, and any Update path that doesn't explicitly reassign `plan.ID = state.ID` will write an unknown/empty value to state. Empty ID then propagates to the next refresh, which issues a URL like `/shodan/alert//info` and 404s. This was the root cause of issue fixed in 0.1.16 — see CHANGELOG.

### Update must carry forward state ID

Even with `UseStateForUnknown`, defensively:

```go
func (r *FooResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan FooResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    // ...

    var state FooResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    // ...

    plan.ID = state.ID
    plan.CreatedAt = state.CreatedAt

    // ... API mutations using plan.ID ...

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

The plan struct is loaded from the plan, which contains user-supplied + framework-computed values. Computed fields without `UseStateForUnknown` arrive as Unknown; assigning from state before any `resp.State.Set` prevents the wipe.

### Reconciling collection attributes (`triggers`, `notifiers`, …)

Use `syncStringList` in `helpers.go`. It:

1. Builds sets from `stateList` and `planList`.
2. Calls `removeFn(alertID, x)` for entries in state but not in plan.
3. Calls `addFn(alertID, x)` for entries in plan but not in state.

Errors are emitted as `Diagnostics.AddWarning` (not hard errors) — Shodan can return transient 4xx, and a partial reconcile is preferable to aborting the apply. The next plan/apply will retry.

### Collections backed by unordered API → `SetAttribute`

Shodan returns `triggers` / `notifiers` / `slack_notifications` as `map[string]interface{}`. `Read` iterates the map (Go map iteration is randomized) → element order in state varies per refresh. If schema is `ListAttribute`, Terraform's order-sensitive list comparison flags every refresh as a diff and `terraform plan` shows eternal `update in-place` churn on unchanged data. Use `schema.SetAttribute` for any collection whose canonical source is an unordered API. Reverting these three to `ListAttribute` = regressing 0.1.19 (see CHANGELOG).

### Schema type changes require version bump + `UpgradeState`

Changing an attribute's TYPE on an existing schema (not just adding/removing attrs) requires all three:

1. Bump `Schema.Version` to `N+1` in `Schema()`.
2. Add `_ resource.ResourceWithUpgradeState = &XResource{}` interface assertion.
3. Implement `UpgradeState(ctx) map[int64]resource.StateUpgrader` returning a `PriorSchema` (FULL prior schema, all attrs) + a `StateUpgrader` func that reads via a v0 model struct and re-encodes into the current model.

Without all three, Terraform errors with type mismatch on first refresh after the upgrade lands. Reference: `resource_shodan_{alert,domain}.go` `UpgradeState` methods (added in 0.1.19 for `List`→`Set` migration).

### Read self-healing for state corruption

For domain alerts, `Read` looks up the alert by its canonical name (`__domain: <domain>` or `__domain: <domain> (<name>)`) when the state ID is empty. This auto-heals state corrupted by older versions of the provider. The same pattern can be applied to other resources if a stable lookup key exists.

If no lookup key exists, `Read` should just call `resp.State.RemoveResource(ctx)` on empty ID — the resource will be recreated.

### Client method conventions

Every method on `ShodanClient` that takes an `alertID`:

1. Guards `if alertID == "" { return ..."alert ID cannot be empty" }` to avoid issuing malformed URLs.
2. Goes through the rate-limited HTTP client (`c.HTTPClient.Do(req)`) — never `http.DefaultClient`.
3. Treats `200 OK` and `404 Not Found` as success for `DELETE` requests (idempotent semantics).
4. Returns `fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, body)` for non-2xx responses so callers can string-match `status 404` if they need to.

## Adding a new resource

1. Create `shodan/resource_shodan_<thing>.go` modelled on the existing resources.
2. Add `NewShodanThingResource` to the slice returned by `provider.Resources()` in `provider.go`.
3. Add user docs at `docs/resources/shodan_<thing>.md`.
4. Add a runnable example at `examples/<thing>_monitoring.tf`.
5. Update `CHANGELOG.md` under the unreleased section.

## Adding a new client method

1. Add the method to `shodan/client.go` next to related ones. Include the empty-ID guard if applicable.
2. If it can be used to reconcile a collection (add/remove pair), expose both halves so callers can use `syncStringList`.
3. If it returns a new struct shape, declare the type next to `AlertResponse` / `DomainInfo`.

## Doc surface area when changing schemas

A resource schema change touches more than the `.go` file. Sync all of:

- `README.md` — argument table (type column) + "Available Trigger Rules" list if relevant.
- `docs/resources/<resource>.md` — Argument Reference (Required/Optional + type label like `Set of String` / `List of String`) + Available Triggers list. This is the canonical published Registry doc — drift here is user-visible.
- `examples/*.tf` — literal HCL types must match the new schema (e.g. scalar `network = "10.0.0.0/24"` is invalid when schema is `list(string)`; wrap in `[]`).
- `CLAUDE.md` "Known gotchas" — refresh the entry if a documented gotcha changed shape.

## Unit tests

**Rule: every functional change in this provider ships with a test in the same MR.** Applies to new client methods, new resource attributes, new helpers, and bug fixes. The recent run of state-corruption bugs (0.1.16-0.1.19) was caught by users in production — each would have been a one-line unit test catch. Test gap = real-world regression risk; closing the gap retroactively is more work than adding the test alongside the change.

For bug fixes specifically: the test that demonstrates the bug goes in first (fails on old code), then the fix is added until the test passes. Reference the originating release in the test comment (e.g. `// 0.1.16 regression — ...`) so future readers see why the case is non-obvious.

### Files

| File | Covers |
|---|---|
| `shodan/helpers_test.go` | `syncStringList` (bidirectional set-diff used by triggers/notifiers/slack), `listToStringSlice`, `setToStringSlice`, warning-not-error contract on transient API failures |
| `shodan/client_test.go` | `AddTrigger`/`RemoveTrigger`/`AddNotifier`/`RemoveNotifier` URL + method + 404-as-success, `GetAlert` 404 surfaced as `status 404` string (Read matches on this), `UpdateAlert` URL/body/Content-Type, `DeleteAlert` idempotent, `ListAlerts` 60s cache + post-TTL re-fetch |
| `shodan/rate_limiter_test.go` | 1-second floor in `NewRateLimitedHTTPClient`, 429 retry-then-success, 429 exhaust-retries surfaces 429, body buffering replays POST/PUT/DELETE bodies on retry |
| `shodan/whitelist_test.go` | `ExtractIgnoredServices` shape variants (port as float64/int/string, malformed entries), `whitelistFromMap`↔`whitelistToMap` round-trip, `syncWhitelist` per-trigger diff, `AddIgnoreService`/`RemoveIgnoreService` URL + 404-as-success + empty-arg guards |
| `provider_test.go` (root) | `SHODAN_API_KEY` env var fallback, config-beats-env precedence, missing-key error |

### Patterns

- **Pure helpers** — table-driven `map[string]struct{in, want}` cases, `reflect.DeepEqual` after sorting both sides (Go map iteration is randomized; the diff helpers and the API both emit unordered output).
- **`sync*` mock funcs** — capture the `(op, ...)` tuple of every call into a shared `[]call` slice, sort both `got` and `want` by deterministic key before comparison. Same shape for `syncStringList` and `syncWhitelist` — copy the patterns when adding the next reconciler.
- **Client methods** — `newDirectClient` (in `client_test.go`) wires a `*ShodanClient` to a `httptest.NewServer` while bypassing the 1-second rate-limiter floor (constructs the `RateLimitedHTTPClient` struct directly with `requestInterval: 0`). Use for tests that exercise multi-request behaviour or retries. For single-request tests, `NewRateLimitedHTTPClient(srv.Client(), 0)` is fine (silently floored to 1s, but the first request goes through immediately because `lastRequest` is zero-time).
- **Rate limiter retry tests** — drive multi-attempt scenarios with an `atomic.Int32` hit counter in the handler; assert both the final response status and the exact hit count (`MaxRetries + 1` for exhausted, `failures + 1` for transient).
- **Provider Configure** — `t.Setenv("SHODAN_API_KEY", ...)` for env var cases; build a real `tfsdk.Config` via `tftypes.NewValue(Object, ...)` shape matching the provider schema. See `provider_test.go` `runConfigure` helper for the boilerplate.

Run with `make test` or `go test -v ./...`. Currently ~50 sub-tests, all <1s total runtime (no real network, no sleeps).

### Not covered yet — extend in follow-up

- **Resource-level Create/Read/Update/Delete lifecycle** — would need `httptest` + `resource.CreateRequest`/`ReadRequest`/etc. scaffolding. The client + helpers carry most of the bug-prone logic, but a full resource-level Update test would close the last gap.
- **`UpgradeState` v0→v1 (List→Set migration)** — needs `resource.UpgradeStateRequest` with a hand-built `tftypes.RawState` matching the v0 prior schema. Mechanically doable, just heavyweight. Add when the next schema version bump is on the roadmap.
- **`shodan_domain` Read recovery path** — the canned `ListAlerts` lookup-by-name flow added in 0.1.16. Test via `newDirectClient` returning a list with one matching `__domain:` entry; verify the resource state gets the recovered ID. Currently only covered by user-reported recovery success.
- **Acceptance tests against real Shodan API** — separate task. Gate on `TF_ACC=1` (upstream convention) so `make test` stays API-key-free.

## Testing locally against a real Terraform configuration

## Testing locally against a real Terraform configuration

```bash
make install                       # writes binary to ~/.terraform.d/plugins/.../darwin_arm64/
cd /path/to/some/tf/config
terraform init -upgrade           # picks up the local override version (see version pin in install target)
terraform plan
```

For breakpoint debugging:

```bash
make dev                          # prints TF_REATTACH_PROVIDERS line — copy to shell of `terraform plan`
```

## Known gotchas

- **API key in URL query string.** Shodan accepts the key only as `?key=<value>`, not as a header. Avoid logging full URLs (they include the key); use the structured error format above so the caller controls what gets surfaced.
- **`AlertResponse.HasTriggers` is approximated as `Enabled`** in `Read`. The Shodan API has no explicit per-alert enabled/disabled field — an alert is functionally inert when it has no triggers. If a real "paused" semantic is ever added, replace this approximation.
- **`Triggers` in `AlertResponse` is `map[string]interface{}`.** Read extracts keys via `for name := range alert.Triggers` (Go map iteration is randomized) and stores them in a `SetAttribute` so plan diffs are order-insensitive. Drift from out-of-band trigger changes is detected correctly; the order changes per refresh but Terraform set semantics ignore it. Do **not** revert to `ListAttribute` — it produces eternal `update in-place` churn (see CHANGELOG 0.1.19).
- **Per-trigger ignore list shape**: each `alert.Triggers[name]` is itself a `map[string]any` containing an `"ignore"` key whose value is `[]any` of `{"ip": string, "port": float64}`. Ports arrive as `float64` (default JSON number unmarshal target) — `ExtractIgnoredServices` in `client.go` switches on the runtime type and formats to `"ip:port"` strings for the `whitelist` attribute. The shape lives only in `ExtractIgnoredServices`; do not duplicate the parsing inline.
- **Whitelist endpoint URL has `ip:port` in the path, not the query string** (`PUT /shodan/alert/{id}/trigger/{trigger}/ignore/{ip}:{port}`). Go's `http.NewRequest` does not URL-escape colons in the path component, so `203.0.113.10:3307` passes through verbatim — matches the API spec. If an IPv6 service ever needs whitelisting, escape the brackets manually before calling `AddIgnoreService`.
- **Rate limiter is single-process.** Parallel `terraform plan` invocations against the same Shodan account can still trip rate limits — the limiter only spaces requests within one provider process.

## Style

- `go fmt` everything (`make fmt` is part of normal flow).
- Comments explain *why*, not *what*. The code already shows what.
- Public methods get a one-line doc comment in standard Go form (`// MethodName does X`).
- No `panic` in non-fatal paths; surface errors through `Diagnostics`.
