# 011 — Add Provider Docs Skill and Acceptance Test Setup Improvements

**Status:** Complete
**Date:** 2026-03-10

## Summary

Added the `provider-docs` skill for Terraform Registry documentation workflows, improved the acceptance test setup script to handle Automation Hub collection installation gracefully, and updated `testing/README.md` with full details from the 010 history spec.

## Changes

### 1. Provider Docs Skill

Downloaded and installed the `provider-docs` skill from the `agent-skills` repo into `.claude/skills/provider-docs/`.

| File | Purpose |
|---|---|
| `SKILL.md` | 7-step workflow for creating, updating, and reviewing Terraform provider docs using HashiCorp-recommended patterns and `tfplugindocs` |
| `agents/openai.yaml` | OpenAI agent interface config |
| `references/hashicorp-provider-docs.md` | Source-backed reference with template paths, generation workflow, release constraints, and canonical links |

The skill covers: schema descriptions, template files in `docs/`, `tfplugindocs` generation, validation, Registry publication rules, and troubleshooting.

### 2. Acceptance Test Setup Script Improvements

Updated `testing/setup-env.sh` to handle the Automation Hub dependency:

- **Collection pre-check** — Added `collections_installed()` helper that checks each collection from `requirements.yml` via `ansible-galaxy collection list`. Skips install entirely if all collections are already present.
- **Auto-detect `testing/ansible.cfg`** — If `testing/ansible.cfg` exists, exports `ANSIBLE_CONFIG` so both `ansible-galaxy` and `ansible-playbook` use it for Automation Hub authentication.
- **Actionable error on install failure** — If `ansible-galaxy collection install` fails, prints a clear message explaining the collections are hosted on Red Hat Automation Hub (not public Galaxy) with an example `ansible.cfg` and token URL.

### 3. Testing README Update

Updated `testing/README.md` with all details from the 010 history spec:

- **Automation Hub Setup** section with `ansible.cfg` example and token URL
- **Authentication** section documenting basic auth vs token auth modes
- **Expanded resource table** — all 18 provisioned resources (added Default org, Demo Credential, Instance Group)
- **Generated Environment Variables** — full table of 14 env vars
- **Test Inventory** — 39 tests across 14 files with test function names and counts
- **Files table** — added `ansible.cfg` entry
- **Troubleshooting** — added token creation, `testAccPreCheck` defaults, and updated collection install guidance

### 4. Provider Test Patterns Skill Update

Updated `provider-test-patterns` skill to include ephemeral resource testing with the `echoprovider` package:

- Added `references/ephemeral.md` reference file
- Updated `SKILL.md` description and references section

## Files Changed

| File | Change |
|---|---|
| `.claude/skills/provider-docs/SKILL.md` | Created |
| `.claude/skills/provider-docs/agents/openai.yaml` | Created |
| `.claude/skills/provider-docs/references/hashicorp-provider-docs.md` | Created |
| `.claude/skills/provider-test-patterns/SKILL.md` | Updated — added ephemeral resource testing |
| `.claude/skills/provider-test-patterns/references/ephemeral.md` | Created |
| `testing/setup-env.sh` | Updated — collection check, Automation Hub config, error handling |
| `testing/README.md` | Updated — full details from 010 history spec |
| `agent-skills` | Submodule updated |
