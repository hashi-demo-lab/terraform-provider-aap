# 012 — Acceptance Test Run and Gitignore Update

**Status:** Complete
**Date:** 2026-03-10

## Summary

Ran acceptance tests against a live AAP 2.5 instance, diagnosed authentication failures, and added acceptance test generated files to `.gitignore`.

## Changes

### 1. Gitignore Update

Added acceptance test generated files to `.gitignore` to prevent accidental commits of sensitive credentials and environment-specific data:

- `testing/.env` — AAP connection details (hostname, username, password, Automation Hub token)
- `testing/ansible.cfg` — Auto-generated Automation Hub config with auth token
- `testing/acceptance_test_vars.env` — Generated resource IDs and API token from setup playbook

### 2. Acceptance Test Execution

Attempted full acceptance test run via `./testing/setup-env.sh`:

1. **Automation Hub token expired** — The offline JWT token (issued 2025-05-26) had its server-side session revoked. Diagnosed by decoding the JWT payload and testing against the Red Hat SSO endpoint (`HTTP 400: Offline user session not found`). User regenerated the token.

2. **AAP credentials invalid** — After token refresh, the setup playbook failed with `HTTP 401: Unauthorized` on the first task. User updated the password.

3. **SSL certificate expired on AAP instance** — Setup playbook provisioned all 17 resources successfully, but `ansible.platform.token` creation failed due to an expired SSL certificate (`SSL: CERTIFICATE_VERIFY_FAILED certificate has expired`). The playbook fell back to a controller token.

4. **Controller token rejected by EDA endpoint** — All acceptance tests failed because the provider's `NewClient` initialization hits `/api/eda/` during endpoint discovery, and the controller token (fallback) is rejected by the EDA gateway endpoint with `401: Invalid token`. This blocks all tests, not just EDA ones.

5. **Workaround: basic auth** — Removed `AAP_TOKEN` from `acceptance_test_vars.env` and re-ran with `--skip-setup` to use basic auth (username/password) instead of token auth. Tests pending.

## Key Findings

- The provider's `readAPIEndpoint()` in `internal/provider/client.go` performs EDA endpoint discovery during client initialization. If the EDA endpoint returns an error (e.g., 401 with an invalid token), it fails all tests — not just EDA-related ones.
- Offline JWT tokens (Red Hat SSO) have no `exp` field but can be revoked server-side, causing `invalid_grant` errors.
- The setup playbook's `block/rescue` pattern for platform vs controller token fallback works for provisioning, but the resulting controller token may not be sufficient for the provider's endpoint discovery.

## Files Changed

| File | Change |
|---|---|
| `.gitignore` | Added `testing/.env`, `testing/ansible.cfg`, `testing/acceptance_test_vars.env` |
