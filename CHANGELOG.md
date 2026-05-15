# Changelog

All notable changes to this provider are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this provider adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.16] — Unreleased

### Fixed

- **`shodan_domain` Update wiped `id` from state when only `triggers`, `notifiers`, or `slack_notifications` changed.** The previous Update path only handled `domain` changes; for any other diff, it wrote the unknown plan ID back to state, leaving the resource with an empty ID. Subsequent refreshes then issued `GET /shodan/alert//info` (note the double slash) and failed with HTTP 404 `The resource could not be found`. Update now reads both plan and state, carries the computed `id` and `created_at` from state, and reconciles `triggers`/`notifiers`/`slack_notifications` via add+remove diff.
- **`shodan_domain` Update never applied changes to `triggers`, `notifiers`, or `slack_notifications`.** Even when `id` survived (rare), the API was only called inside the `domain`-changed branch. Update now applies these changes on every diff.
- **`shodan_alert` Update only added triggers/notifiers, never removed them.** Comments in the source claimed "Shodan API doesn't support removing triggers" — it does (`DELETE /shodan/alert/{id}/trigger/{trigger}`, same shape for notifiers). Update now performs a proper set-diff: removes entries that left the plan, adds entries that entered the plan.
- **Read functions hard-errored on resources deleted out-of-band.** Refresh on a Shodan alert that was deleted in the UI returned `Error reading alert: API request failed with status 404`. Read now treats 404 as resource-gone and removes from state, letting Terraform recreate on next plan (idiomatic Plugin Framework behaviour).
- **`SHODAN_API_KEY` environment variable was documented but not implemented.** The provider description claimed env var fallback existed; in practice `api_key` was `Required: true` and never read from env. The attribute is now `Optional` and falls back to `SHODAN_API_KEY` when unset. Existing configs that pass `api_key` explicitly are unaffected.

### Added

- **Self-healing recovery for `shodan_domain` state corrupted by pre-0.1.16 Update bug.** When Read finds an empty `id` in state, it calls `GET /shodan/alert/info`, matches by the canonical alert name (`__domain: <domain>` or `__domain: <domain> (<name>)`), and writes the recovered ID back to state. Users do not need to manually `terraform state rm` + `terraform import` to recover.
- **Client method `ListAlerts()`** — wraps `GET /shodan/alert/info`, used by the recovery path above.
- **Client methods `RemoveTrigger(alertID, trigger)`, `RemoveNotifier(alertID, notifierID)`** — wrap the corresponding `DELETE` endpoints. Treat 404 as success (idempotent).
- **`UseStateForUnknown` plan modifiers** on `id` and `created_at` computed attributes across both resources. Prevents these values from being marked `<known after apply>` during in-place updates, eliminating one class of unintended drift.
- **Empty-ID guards** in `GetAlert`, `DeleteAlert`, `UpdateAlert`, `AddTrigger`, `RemoveTrigger`, `AddNotifier`, `RemoveNotifier`. Returns `alert ID cannot be empty` instead of issuing malformed requests like `GET /shodan/alert//info`.

### Upgrade notes

If you are upgrading from 0.1.15 and your state shows refresh errors of the form `Error reading domain alert: API request failed with status 404` with a URL containing `/shodan/alert//info`, the upgrade alone is sufficient. On the next plan, `shodan_domain` Read will detect the empty ID, look up the existing alert via `ListAlerts`, and restore the ID to state. The alert in your Shodan account is not affected and you will not receive duplicate notifications.

If `ListAlerts` does not find an alert matching the expected name pattern (e.g. the alert was deleted in the Shodan UI), the resource is removed from state and recreated on the next apply.

## [0.1.15] and earlier

See git history.
