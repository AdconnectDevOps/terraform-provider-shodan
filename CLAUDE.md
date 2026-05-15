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
- **Rate limiter is single-process.** Parallel `terraform plan` invocations against the same Shodan account can still trip rate limits — the limiter only spaces requests within one provider process.

## Style

- `go fmt` everything (`make fmt` is part of normal flow).
- Comments explain *why*, not *what*. The code already shows what.
- Public methods get a one-line doc comment in standard Go form (`// MethodName does X`).
- No `panic` in non-fatal paths; surface errors through `Diagnostics`.
