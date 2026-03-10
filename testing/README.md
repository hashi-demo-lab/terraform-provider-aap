# Acceptance Tests

Run the provider's acceptance tests against a real AAP 2.5+ instance.

## Prerequisites

- AAP 2.5+ instance with admin access (2.4 partially supported — EDA tests will be skipped)
- `ansible-galaxy` and `ansible-playbook` CLI
- Go toolchain
- Access to Red Hat Automation Hub for collection install (see [Automation Hub Setup](#automation-hub-setup))

## Quick Start

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
3. Check if required collections are already installed (skip install if so)
4. Install required Ansible collections (`ansible.controller`, `ansible.platform`, `ansible.eda`)
5. Run the setup playbook to provision test resources in AAP
6. Source the generated resource IDs and API token from `testing/acceptance_test_vars.env`
7. Execute all acceptance tests with `TF_ACC=1`

## Options

```bash
./testing/setup-env.sh --run TestAccAAPJob_basic   # run a single test
./testing/setup-env.sh --skip-setup                # skip provisioning, just run tests
./testing/setup-env.sh --skip-setup --run TestAcc  # combine options
```

## Automation Hub Setup

The required collections (`ansible.controller`, `ansible.platform`, `ansible.eda`) are hosted on Red Hat Automation Hub, not public Galaxy. The script auto-detects `testing/ansible.cfg` if present.

Create `testing/ansible.cfg`:

```ini
[galaxy]
server_list = automation_hub

[galaxy_server.automation_hub]
url=https://console.redhat.com/api/automation-hub/content/published/
auth_url=https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token
token=<YOUR_AUTOMATION_HUB_TOKEN>
```

Get your token at: https://console.redhat.com/ansible/automation-hub/token

If collections are already installed, the script skips the install step entirely.

## Authentication

The provider supports two authentication modes (see `internal/provider/provider_test.go`):

1. **Basic auth** — `AAP_USERNAME` + `AAP_PASSWORD` (used when `AAP_TOKEN` is not set)
2. **Token auth** — `AAP_TOKEN` takes precedence; username/password are ignored when token is present

The setup playbook generates an API token automatically and writes it to `testing/acceptance_test_vars.env`, so after first run both modes are available.

## What the Setup Playbook Creates

The playbook (`playbook.yml`) provisions these resources in your AAP instance:

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

The playbook writes `testing/acceptance_test_vars.env` (gitignored) with resource IDs plus an API token:

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

## Files

| File | Purpose |
|---|---|
| `setup-env.sh` | Single script — loads `.env`, provisions AAP, runs tests |
| `.env` | AAP connection details (gitignored, user-created) |
| `ansible.cfg` | Automation Hub config (gitignored, user-created) |
| `playbook.yml` | Ansible playbook that provisions test resources |
| `requirements.yml` | Ansible collection dependencies |
| `templates/acceptance_test_vars.env.j2` | Jinja2 template for generated env vars |
| `acceptance_test_vars.env` | Generated file with resource IDs (gitignored) |

## Troubleshooting

**Collection install fails** — The collections are on Red Hat Automation Hub, not public Galaxy. See [Automation Hub Setup](#automation-hub-setup). If collections are already installed the script skips this step automatically.

**EDA tasks fail** — Expected on AAP 2.4. The playbook handles this gracefully with `block/rescue`, but the 4 EDA acceptance tests will fail without an event stream. These tests are 2.5+ only.

**Token creation fails** — The playbook tries `ansible.platform.token` first (2.5+), falling back to `ansible.controller.token` (2.4). If both fail, check that your user has permission to create tokens.

**`testAccPreCheck` defaults** — If `AAP_HOSTNAME` is not set, the pre-check defaults to `https://localhost:8043`. If `AAP_INSECURE_SKIP_VERIFY` is not set, it defaults to `true`. Username and password have no defaults and will cause test failure if missing (unless `AAP_TOKEN` is set).

**Tests time out** — Some tests involve long-running jobs (sleep.yml). The script sets a 30-minute timeout by default.

**Re-running after failure** — Use `--skip-setup` to avoid re-provisioning resources that already exist.
