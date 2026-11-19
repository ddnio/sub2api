#!/usr/bin/env bash
# Read-only production preflight & post-switch verification for NanaFox Router.
# NEVER modifies production state and NEVER prints secrets/passwords.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

MODE="${1:-}"
shift || true

SSH_HOST="tmhhost"
LOCAL_JS=""
ROUTER_URL="https://router.nanafox.com"
STUDIO_URL="https://studio.nanafox.com"

usage() {
  cat <<'EOF'
Usage:
  verify-prod.sh preflight [--host tmhhost]
  verify-prod.sh post-switch [--host tmhhost] [--local-js <path/to/index-*.js>]

Modes:
  preflight     Read-only: runs check-router-production.py, checks disk/memory/allowlisted env,
                migration count, usage_logs size, and financial/row baseline.
  post-switch   Read-only: re-runs check-router-production.py, checks public endpoints,
                verifies served JS SHA256 against local build, checks restarts/OOM/error logs,
                and prints updated migration & financial baseline.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) SSH_HOST="$2"; shift 2 ;;
    --local-js) LOCAL_JS="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage; exit 2 ;;
  esac
done

if [[ "$MODE" != "preflight" && "$MODE" != "post-switch" ]]; then
  usage
  exit 2
fi

run_remote_check_script() {
  echo "[1] Running deploy/check-router-production.py on $SSH_HOST ..."
  ssh "$SSH_HOST" 'python3 -' < "$REPO_ROOT/deploy/check-router-production.py"
}

collect_remote_readonly_snapshot() {
  echo "[2] Collecting read-only host, container & DB snapshot from $SSH_HOST ..."
  ssh "$SSH_HOST" 'bash -s' <<'REMOTE_EOF'
set -euo pipefail

echo "--- Host Disk ---"
df -h / | awk 'NR==1 || NR==2'

echo "--- Router Container (non-secret) ---"
docker inspect nanafox-router-formal-router --format 'Image={{.Image}}
Status={{.State.Status}} Running={{.State.Running}} Restarts={{.RestartCount}} OOMKilled={{.State.OOMKilled}} StartedAt={{.State.StartedAt}}'
docker stats --no-stream --format 'MemUsage={{.MemUsage}} MemPerc={{.MemPerc}} CPU={{.CPUPerc}}' nanafox-router-formal-router

echo "--- Allowlisted Env ---"
docker inspect nanafox-router-formal-router --format '{{range .Config.Env}}{{println .}}{{end}}' \
  | grep -E '^(GOMEMLIMIT|GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES|TZ|GIN_MODE)=' || true

echo "--- Database Baseline (read-only) ---"
DB_NAME=$(docker exec nanafox-router-formal-postgres sh -c 'psql -U "${POSTGRES_USER:-postgres}" -d postgres -tAc "SELECT datname FROM pg_database WHERE datistemplate = false AND datname NOT IN ('\''postgres'\'') ORDER BY datname LIMIT 1"')
DB_USER=$(docker exec nanafox-router-formal-postgres sh -c 'printf "%s" "${POSTGRES_USER:-postgres}"')
echo "Detected DB: $DB_NAME"

docker exec nanafox-router-formal-postgres psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
SET default_transaction_read_only = on;
SELECT count(*) AS migrations, max(filename) AS last_migration FROM schema_migrations;
SELECT pg_size_pretty(pg_database_size(current_database())) AS db_size;
SELECT
  pg_size_pretty(pg_total_relation_size('usage_logs')) AS usage_logs_total,
  pg_size_pretty(pg_relation_size('usage_logs')) AS usage_logs_heap,
  reltuples::bigint AS usage_logs_est_rows
FROM pg_class WHERE relname = 'usage_logs';
SELECT
  (SELECT count(*) FROM users WHERE deleted_at IS NULL) AS active_users,
  (SELECT coalesce(sum(balance), 0) FROM users WHERE deleted_at IS NULL) AS total_balance,
  (SELECT count(*) FROM accounts WHERE deleted_at IS NULL) AS active_accounts,
  (SELECT count(*) FROM groups WHERE deleted_at IS NULL) AS active_groups,
  (SELECT count(*) FROM api_keys WHERE deleted_at IS NULL) AS active_api_keys;
SQL
REMOTE_EOF
}

if [[ "$MODE" == "preflight" ]]; then
  echo "=== Production Read-Only Preflight ($SSH_HOST) ==="
  run_remote_check_script
  collect_remote_readonly_snapshot
  echo "=== Preflight Complete ==="
  exit 0
fi

# post-switch mode
echo "=== Production Post-Switch Verification ($SSH_HOST) ==="
run_remote_check_script
collect_remote_readonly_snapshot

echo "[3] Checking Router recent logs for errors ..."
ERR_COUNT=$(ssh "$SSH_HOST" "docker logs --since 15m nanafox-router-formal-router 2>&1 | grep -ciE 'panic:|fatal|checksum mismatch|migration failed' || true")
echo "Fatal/panic/migration error lines in last 15m: $ERR_COUNT"
if [[ "$ERR_COUNT" != "0" ]]; then
  echo "FAIL: non-zero error lines in Router logs" >&2
  exit 1
fi

echo "[4] Checking public HTTP endpoints ..."
check_http() {
  local url="$1" expected="$2"
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' "$url")
  if [[ "$code" == "$expected" ]]; then
    echo "[PASS] $url -> $code"
  else
    echo "[FAIL] $url -> $code (expected $expected)" >&2
    exit 1
  fi
}
check_http "${ROUTER_URL}/health" "200"
check_http "${ROUTER_URL}/" "200"
check_http "${ROUTER_URL}/v1/models" "401"
check_http "${STUDIO_URL}/health" "200"

if [[ -z "$LOCAL_JS" ]]; then
  LOCAL_JS=$(find "$REPO_ROOT/frontend/dist/assets" -maxdepth 1 -name 'index-*.js' 2>/dev/null | head -n 1 || true)
fi

if [[ -n "$LOCAL_JS" && -f "$LOCAL_JS" ]]; then
  JS_NAME=$(basename "$LOCAL_JS")
  LOCAL_SHA=$(shasum -a 256 "$LOCAL_JS" | awk '{print $1}')
  REMOTE_SHA=$(curl -fsS "${ROUTER_URL}/assets/${JS_NAME}" | shasum -a 256 | awk '{print $1}')
  if [[ "$LOCAL_SHA" == "$REMOTE_SHA" ]]; then
    echo "[PASS] Served ${JS_NAME} SHA256 matches local build (${LOCAL_SHA:0:12}...)"
  else
    echo "[FAIL] Served ${JS_NAME} SHA256 mismatch: local=$LOCAL_SHA remote=$REMOTE_SHA" >&2
    exit 1
  fi
else
  echo "[WARN] Skipping JS SHA256 verification (pass --local-js <path> to enable)"
fi

echo "=== Post-Switch Verification PASSED ==="
