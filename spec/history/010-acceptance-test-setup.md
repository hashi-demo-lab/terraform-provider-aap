# 010 — Acceptance Test Setup Guide

**Status:** Complete
**Date:** 2026-03-09

## Overview

This document captures the full setup and execution process for running the terraform-provider-aap acceptance tests (`TestAcc*`) against a real AAP instance. The provider has 39 acceptance tests across 14 test files covering resources, data sources, and actions.

A single script (`testing/setup-env.sh`) handles collection install, AAP provisioning, and test execution. Connection details are provided via a `testing/.env` file.

## Prerequisites

- A running AAP 2.5+ instance (2.4 is partially supported with fallbacks)
- Admin credentials (username/password) for the AAP instance
- `ansible-galaxy` and `ansible-playbook` CLI tools installed
- Go toolchain

## Usage

Create `testing/.env` with your AAP connection details:

```bash
cat > testing/.env <<'EOF'
AAP_HOSTNAME=https://aap.example.com
AAP_USERNAME=admin
AAP_PASSWORD=changeme
AAP_INSECURE_SKIP_VERIFY=true
EOF
```

Run everything:

```bash
./testing/setup-env.sh
```

The script will:

1. Load connection details from `testing/.env`
2. Derive `CONTROLLER_*` and `EDA_CONTROLLER_*` vars for the Ansible collections
3. Install required Ansible collections (`ansible.controller`, `ansible.platform`, `ansible.eda`)
4. Run the setup playbook to provision test resources in AAP
5. Source the generated resource IDs and API token from `testing/acceptance_test_vars.env`
6. Execute all acceptance tests with `TF_ACC=1`

### Options

```bash
./testing/setup-env.sh --run TestAccAAPJob_basic   # run a single test
./testing/setup-env.sh --skip-setup                # skip provisioning, just run tests
./testing/setup-env.sh --skip-setup --run TestAcc  # combine options
```

## Authentication

The provider supports two authentication modes (see `internal/provider/provider_test.go:30-58`):

1. **Basic auth** — `AAP_USERNAME` + `AAP_PASSWORD` (used when `AAP_TOKEN` is not set)
2. **Token auth** — `AAP_TOKEN` takes precedence; username/password are ignored when token is present

The setup playbook generates an API token automatically and writes it to `testing/acceptance_test_vars.env`, so after first run both modes are available.

## What the Setup Playbook Creates

The playbook (`testing/playbook.yml`) provisions these resources in AAP:

| Resource | Name | Purpose |
|---|---|---|
| Organization | Default (description updated) | Consistent description across 2.4/2.5 |
| Organization | Non-Default | Tests for non-default org inventories |
| Job Template | Test Job Template - hello_world.yml | Basic job template tests |
| Job Template | ...with Inventory Prompt | Tests inventory prompt-on-launch behavior |
| Job Template | ...with All Fields on Prompt | Tests all promptable fields validation |
| Job Template | Test Job Template - sleep.yml | Long-running job for host delete-retry tests |
| Job Template | Test Job Template - fail.yml | Failing job template for error handling tests |
| Workflow Job Template | Demo Workflow Job Template | Basic workflow tests |
| Workflow Job Template | Demo FailureWorkflow Job Template | Workflow with failure node (fail.yml) |
| Workflow Job Template | Workflow with Inventory | Workflow with pre-configured inventory |
| Inventory | Inventory for Workflow | Dedicated inventory for workflow testing |
| Project | Test Playbooks | Git project from `ansible/test-playbooks` repo |
| Credential | Demo Credential (looked up) | Machine credential for launch tests |
| Credential | Test Credential (EDA) | Basic Event Stream credential (2.5+ only) |
| Event Stream | Test Event Stream (EDA) | EDA event stream for post-URL tests (2.5+ only) |
| Label | Test Label | Label for all-fields-on-prompt tests |
| Instance Group | default (looked up) | Instance group for all-fields-on-prompt tests |
| Token | Platform or Controller token | API token for token-auth testing |

## Generated Environment Variables

The playbook writes `testing/acceptance_test_vars.env` (gitignored) with 14 resource IDs plus an API token:

| Variable | Purpose |
|---|---|
| `AAP_TEST_JOB_TEMPLATE_ID` | Basic test job template (hello_world.yml) |
| `AAP_TEST_JOB_TEMPLATE_INVENTORY_PROMPT_ID` | Job template with inventory prompt-on-launch |
| `AAP_TEST_JOB_TEMPLATE_ALL_FIELDS_PROMPT_ID` | Job template with all promptable fields |
| `AAP_TEST_WORKFLOW_JOB_TEMPLATE_ID` | Basic workflow job template |
| `AAP_TEST_WORKFLOW_JOB_TEMPLATE_FAIL_ID` | Workflow with failure node |
| `AAP_TEST_ORGANIZATION_ID` | Non-Default organization |
| `AAP_TEST_WORKFLOW_INVENTORY_ID` | Workflow with pre-configured inventory |
| `AAP_TEST_INVENTORY_FOR_WF_ID` | Inventory for workflow testing |
| `AAP_TEST_JOB_FOR_HOST_RETRY_ID` | Long-running job template (sleep.yml) |
| `AAP_TEST_JOB_TEMPLATE_FAIL_ID` | Failing job template (fail.yml) |
| `AAP_TEST_DEMO_CREDENTIAL_ID` | Demo Machine credential |
| `AAP_TEST_LABEL_ID` | Test label |
| `AAP_TEST_DEFAULT_INSTANCE_GROUP_ID` | Default instance group |
| `AAP_TOKEN` | Generated API token (platform token on 2.5+, controller token on 2.4) |

## Test Inventory

39 acceptance tests across 14 files:

| File | Tests | Count |
|---|---|---|
| `job_resource_test.go` | TestAccAAPJob_{basic, UpdateWithSameParameters, UpdateWithNewInventoryIdPromptOnLaunch, UpdateWithTrigger, WaitForCompletion, disappears, AllFieldsOnPrompt, AllFieldsOnPrompt_MissingRequired} | 8 |
| `job_launch_action_test.go` | TestAccAAPJobAction_{basic, fail, failIgnore, AllFieldsOnPrompt, AllFieldsOnPrompt_MissingRequired} | 5 |
| `job_template_data_source_test.go` | TestAccJobTemplateDataSource | 1 |
| `workflow_job_resource_test.go` | TestAccAAPWorkflowJob_{Basic, WithNoInventoryID, UpdateWithSameParameters, UpdateWithNewInventoryIdPromptOnLaunch, UpdateWithTrigger, waitForCompletionWithFailure, Disappears} | 7 |
| `workflow_job_launch_action_test.go` | TestAccAAPWorkflowJobAction_{Basic, fail, failIgnore} | 3 |
| `workflow_job_template_data_source_test.go` | TestAccWorkflowJobTemplateDataSource | 1 |
| `organization_data_source_test.go` | TestAccOrganizationDataSource{, BadConfig, WithIdAndName, NonExistentValues} | 4 |
| `organization_acceptance_test.go` | TestAccInventoryResourceWithOrganizationDataSource | 1 |
| `inventory_data_source_test.go` | TestAccInventoryDataSource | 1 |
| `inventory_resource_test.go` | TestAccInventoryResource | 1 |
| `host_resource_test.go` | TestAccHostResource, TestAccHostResourceDeleteWithRetry | 2 |
| `group_resource_test.go` | TestAccGroupResource | 1 |
| `eda_eventstream_datasource_test.go` | TestAccEDAEventStreamDataSourceRetrievesPostURL | 1 |
| `eda_eventstream_post_action_test.go` | TestAccEDAEventStream{AfterCreateAction, UnrelatedActionDoesNotTrigger, AfterUpdateAction} | 3 |

## Key Files

| File | Purpose |
|---|---|
| `testing/setup-env.sh` | Single script — loads `.env`, provisions AAP, runs tests |
| `testing/.env` | AAP connection details (gitignored, user-created) |
| `testing/playbook.yml` | Ansible playbook that provisions AAP test resources |
| `testing/requirements.yml` | Ansible collection dependencies |
| `testing/templates/acceptance_test_vars.env.j2` | Jinja2 template for generated env file |
| `testing/acceptance_test_vars.env` | Generated file with resource IDs (gitignored) |
| `internal/provider/provider_test.go` | `testAccPreCheck()` and provider test factories |
| `Makefile` | `testacc` and `testacccov` targets |

## Troubleshooting

### EDA tests fail on AAP 2.4
The EDA credential and event stream tasks are wrapped in `block/rescue` — they gracefully skip on 2.4. However, the EDA acceptance tests will fail if no event stream exists. These tests are 2.5+ only.

### Token creation fails
The playbook tries `ansible.platform.token` first (2.5+), falling back to `ansible.controller.token` (2.4). If both fail, check that your user has permission to create tokens.

### `testAccPreCheck` defaults
If `AAP_HOSTNAME` is not set, the pre-check defaults to `https://localhost:8043`. If `AAP_INSECURE_SKIP_VERIFY` is not set, it defaults to `true`. Username and password have no defaults and will cause test failure if missing (unless `AAP_TOKEN` is set).

### Running a single test
```bash
./testing/setup-env.sh --skip-setup --run TestAccAAPJob_basic
```
