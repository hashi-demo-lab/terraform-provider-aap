# Future Provider Resources

Resources needed to close the gap between what acceptance tests require (provisioned by Ansible playbook) and what the provider can manage natively.

## Current Provider Coverage

| Type | Resources |
|------|-----------|
| **Resources** | `aap_inventory`, `aap_job`, `aap_workflow_job`, `aap_group`, `aap_host` |
| **Data sources** | `aap_inventory`, `aap_job_template`, `aap_workflow_job_template`, `aap_organization`, `aap_eda_event_stream` |
| **Actions** | `aap_job_launch`, `aap_workflow_job_launch`, `aap_eda_eventstream_post` |

## Target Resources

Priority based on: test unblock count, community demand (upstream issues), and dependency chain.

### P1 — Unblocks acceptance tests + high community demand

| Resource | API Endpoint | Upstream Issue | Tests Unblocked |
|----------|-------------|----------------|-----------------|
| `aap_project` | `/api/v2/projects/` | #79, #35 | Dependency for job templates |
| `aap_job_template` | `/api/v2/job_templates/` | #173 | 12 (all job + action tests) |
| `aap_workflow_job_template` | `/api/v2/workflow_job_templates/` | #173 | 7 (all workflow tests) |
| `aap_organization` | `/api/v2/organizations/` | #35 | 1 (`TestAccInventoryResource`) |
| `aap_credential` | `/api/v2/credentials/` | #125 | 2 (all-fields-on-prompt tests) |

**Dependency chain**: Project must exist before job templates can reference it.

### P2 — Completes job launch coverage

| Resource | API Endpoint | Upstream Issue | Purpose |
|----------|-------------|----------------|---------|
| `aap_label` | `/api/v2/labels/` | — | Prompt-on-launch field |
| `aap_workflow_job_template_node` | `/api/v2/workflow_job_template_nodes/` | — | Workflow topology |
| `aap_inventory_source` | `/api/v2/inventory_sources/` | #52 | Dynamic inventory sync |

### P3 — Full platform coverage

| Resource | API Endpoint | Upstream Issue | Purpose |
|----------|-------------|----------------|---------|
| `aap_user` | `/api/v2/users/` | #79 | RBAC management |
| `aap_team` | `/api/v2/teams/` | #79 | RBAC management |
| `aap_role` | `/api/v2/roles/` | #79 | Permission grants |
| `aap_schedule` | `/api/v2/schedules/` | #79 | Recurring job execution |
| `aap_execution_environment` | `/api/v2/execution_environments/` | — | Custom EE management |
| `aap_instance_group` | `/api/v2/instance_groups/` | — | Compute topology |
| `aap_notification_template` | `/api/v2/notification_templates/` | — | Alerting |

### Data sources to add alongside resources

| Data source | API Endpoint | Purpose |
|-------------|-------------|---------|
| `data.aap_project` | `/api/v2/projects/` | Reference existing projects |
| `data.aap_credential` | `/api/v2/credentials/` | Reference existing credentials |
| `data.aap_execution_environment` | `/api/v2/execution_environments/` | Reference existing EEs |

## Implementation Order

```
P1 chain: organization → project → credential → job_template → workflow_job_template
```

Each P1 resource eliminates Ansible playbook dependencies from acceptance tests. Once P1 is complete, `testing/setup-env.sh` playbook becomes optional — tests can self-provision all required resources via Terraform.
