#!/usr/bin/env bash
# Run acceptance tests against a real AAP instance.
#
# Usage:
#   1. Create testing/.env with your AAP connection details
#   2. ./testing/setup-env.sh [options]
#
# Options:
#   --run <pattern>   Run only tests matching pattern (e.g. TestAccAAPJob_basic)
#   --skip-setup      Skip collection install and playbook, just run tests
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${REPO_ROOT}/testing/.env"

# ── Load .env ────────────────────────────────────────────────────────────────
if [[ ! -f "${ENV_FILE}" ]]; then
    cat >&2 <<'MSG'
Error: testing/.env not found.

Create it with your AAP connection details:

    cat > testing/.env <<'EOF'
    AAP_HOSTNAME=https://aap.example.com
    AAP_USERNAME=admin
    AAP_PASSWORD=changeme
    AAP_INSECURE_SKIP_VERIFY=true
    EOF

MSG
    exit 1
fi

set -a
# shellcheck source=/dev/null
source "${ENV_FILE}"

# Derive Ansible collection vars from AAP vars
CONTROLLER_HOST="${AAP_HOSTNAME}"
CONTROLLER_USERNAME="${AAP_USERNAME}"
CONTROLLER_PASSWORD="${AAP_PASSWORD}"
CONTROLLER_VERIFY_SSL="false"
EDA_CONTROLLER_HOST="${AAP_HOSTNAME}"
EDA_CONTROLLER_USERNAME="${AAP_USERNAME}"
EDA_CONTROLLER_PASSWORD="${AAP_PASSWORD}"
EDA_CONTROLLER_VERIFY_SSL="${CONTROLLER_VERIFY_SSL}"
set +a

# ── Parse args ───────────────────────────────────────────────────────────────
RUN_PATTERN=""
SKIP_SETUP=false
while [[ $# -gt 0 ]]; do
    case "$1" in
        --run)        RUN_PATTERN="$2"; shift 2 ;;
        --skip-setup) SKIP_SETUP=true; shift ;;
        *)            echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# ── Ensure testing/ansible.cfg exists ──────────────────────────────────────
ANSIBLE_CFG="${REPO_ROOT}/testing/ansible.cfg"
if [[ ! -f "${ANSIBLE_CFG}" ]]; then
    if [[ -z "${AAP_AUTOMATION_HUB_TOKEN:-}" ]]; then
        cat >&2 <<'MSG'
Error: testing/ansible.cfg not found and AAP_AUTOMATION_HUB_TOKEN is not set.

The required collections (ansible.controller, ansible.platform, ansible.eda)
are hosted on Red Hat Automation Hub, not public Galaxy.

Add your Automation Hub token to testing/.env:

    AAP_AUTOMATION_HUB_TOKEN=<your-token-here>

Get your token at: https://console.redhat.com/ansible/automation-hub/token

The script will generate testing/ansible.cfg automatically.
MSG
        exit 1
    fi
    echo "==> Generating testing/ansible.cfg from AAP_AUTOMATION_HUB_TOKEN..."
    cat > "${ANSIBLE_CFG}" <<EOF
[galaxy]
server_list = automation_hub

[galaxy_server.automation_hub]
url=https://console.redhat.com/api/automation-hub/content/published/
auth_url=https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token
token=${AAP_AUTOMATION_HUB_TOKEN}
EOF
fi
export ANSIBLE_CONFIG="${ANSIBLE_CFG}"

# ── Collection check helper ────────────────────────────────────────────────
collections_installed() {
    local missing=false
    while IFS= read -r name; do
        [[ -z "${name}" ]] && continue
        if ! ansible-galaxy collection list "${name}" 2>/dev/null | grep -q "${name}"; then
            missing=true
            echo "  missing: ${name}" >&2
        fi
    done < <(grep '^\s*- name:' "${REPO_ROOT}/testing/requirements.yml" | awk '{print $NF}')
    [[ "${missing}" == "false" ]]
}

# ── Setup ────────────────────────────────────────────────────────────────────
if [[ "${SKIP_SETUP}" == "false" ]]; then
    echo "==> Checking Ansible collections..."
    if collections_installed; then
        echo "    All required collections already installed — skipping install."
    else
        echo "==> Installing Ansible collections..."
        if ! ansible-galaxy collection install -r "${REPO_ROOT}/testing/requirements.yml" 2>&1; then
            cat >&2 <<'MSG'

Error: Collection install failed.

The required collections (ansible.controller, ansible.platform, ansible.eda)
are hosted on Red Hat Automation Hub, not public Galaxy.

Verify your Automation Hub token is correct in testing/.env:

    AAP_AUTOMATION_HUB_TOKEN=<your-token-here>

Get your token at: https://console.redhat.com/ansible/automation-hub/token

MSG
            exit 1
        fi
    fi

    echo "==> Provisioning AAP test resources..."
    ansible-playbook "${REPO_ROOT}/testing/playbook.yml"
fi

# ── Source generated resource IDs ────────────────────────────────────────────
VARS_FILE="${REPO_ROOT}/testing/acceptance_test_vars.env"
if [[ ! -f "${VARS_FILE}" ]]; then
    echo "Error: ${VARS_FILE} not found. Run without --skip-setup first." >&2
    exit 1
fi
set -a
# shellcheck source=/dev/null
source "${VARS_FILE}"
set +a

# ── Run tests ────────────────────────────────────────────────────────────────
echo "==> Running acceptance tests..."
cd "${REPO_ROOT}"
if [[ -n "${RUN_PATTERN}" ]]; then
    TF_ACC=1 go test -count=1 -v -run "${RUN_PATTERN}" ./... -timeout 30m
else
    TF_ACC=1 go test -count=1 -v ./... -timeout 30m
fi
