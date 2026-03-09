# Acceptance Tests

Run the provider's acceptance tests against a real AAP 2.5+ instance.

## Prerequisites

- AAP 2.5+ instance with admin access (2.4 partially supported — EDA tests will be skipped)
- `ansible-galaxy` and `ansible-playbook` CLI
- Go toolchain
- Access to Ansible Automation Hub for collection install (configure token in `ansible.cfg` if needed)

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

1. Install required Ansible collections (`ansible.controller`, `ansible.platform`, `ansible.eda`)
2. Run the setup playbook to provision test resources in AAP
3. Source the generated resource IDs and API token
4. Execute all acceptance tests

## Options

```bash
./testing/setup-env.sh --run TestAccAAPJob_basic   # run a single test
./testing/setup-env.sh --skip-setup                # skip provisioning, just run tests
./testing/setup-env.sh --skip-setup --run TestAcc  # combine options
```

## What the Setup Playbook Creates

The playbook (`playbook.yml`) provisions these resources in your AAP instance:

| Resource | Name | Used By |
|---|---|---|
| Organization | Non-Default | Inventory tests |
| Job Template | Test Job Template - hello_world.yml | Job resource tests |
| Job Template | ...with Inventory Prompt | Prompt-on-launch tests |
| Job Template | ...with All Fields on Prompt | All-fields validation tests |
| Job Template | Test Job Template - sleep.yml | Host delete-retry tests |
| Job Template | Test Job Template - fail.yml | Failure handling tests |
| Workflow Job Template | Demo Workflow Job Template | Workflow tests |
| Workflow Job Template | Demo FailureWorkflow Job Template | Workflow failure tests |
| Workflow Job Template | Workflow with Inventory | Workflow inventory tests |
| Inventory | Inventory for Workflow | Workflow tests |
| Project | Test Playbooks | sleep.yml / fail.yml source |
| Label | Test Label | Launch with all fields |
| EDA Credential | Test Credential | Event stream tests (2.5+) |
| EDA Event Stream | Test Event Stream | Event stream tests (2.5+) |
| API Token | Platform or Controller token | Token auth tests |

Resource IDs are written to `testing/acceptance_test_vars.env` (gitignored).

## Files

| File | Purpose |
|---|---|
| `setup-env.sh` | Single script — provisions AAP and runs tests |
| `.env` | Your AAP connection details (gitignored) |
| `playbook.yml` | Ansible playbook that provisions test resources |
| `requirements.yml` | Ansible collection dependencies |
| `templates/acceptance_test_vars.env.j2` | Jinja2 template for generated env vars |
| `acceptance_test_vars.env` | Generated file with resource IDs (gitignored) |

## Troubleshooting

**Collection install fails** — Ensure you have access to `console.redhat.com/ansible/automation-hub`. You may need to configure an authentication token in your `ansible.cfg`.

**EDA tasks fail** — Expected on AAP 2.4. The playbook handles this gracefully, but the 4 EDA acceptance tests will fail without an event stream.

**Tests time out** — Some tests involve long-running jobs (sleep.yml). The script sets a 30-minute timeout by default.

**Re-running after failure** — Use `--skip-setup` to avoid re-provisioning resources that already exist.
