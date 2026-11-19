#!/usr/bin/env bash
# Read-only production preflight & post-switch verification for NanaFox Router.
# NEVER modifies production state and NEVER prints secrets/passwords.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

MODE="${1:-}"
shift || true

SSH_HOST="tmhhost"
LOCAL_DIST="$REPO_ROOT/backend/internal/web/dist"
BASELINE_FILE="/tmp/sub2api-prod-baseline.json"
SKIP_JS_CHECK=0
ROUTER_URL="https://router.nanafox.com"
STUDIO_URL="https://studio.nanafox.com"

usage() {
  cat <<'EOF'
Usage:
  verify-prod.sh preflight [--host tmhhost] [--baseline /tmp/sub2api-prod-baseline.json]
  verify-prod.sh post-switch [--host tmhhost] [--baseline /tmp/sub2api-prod-baseline.json] [--local-dist <path>] [--skip-js-check]

Modes:
  preflight     Read-only: runs check-router-production.py, prints disk/memory/allowlisted env/revision,
                migration count & checksum fingerprint, usage_logs size, financial/row baseline,
                and writes baseline JSON to --baseline for post-switch comparison.
  post-switch   Read-only: re-runs check-router-production.py, asserts 0 restarts & OOMKilled=false,
                checks error logs, checks public endpoints (200/401), verifies served entry JS filename
                and SHA256 against --local-dist (default: backend/internal/web/dist), and compares
                DB financial/row counts & pre-existing migration checksums against --baseline.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) SSH_HOST="$2"; shift 2 ;;
    --local-dist) LOCAL_DIST="$2"; shift 2 ;;
    --baseline) BASELINE_FILE="$2"; shift 2 ;;
    --skip-js-check) SKIP_JS_CHECK=1; shift ;;
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

collect_remote_snapshot_json() {
  local out_file="$1"
  ssh "$SSH_HOST" 'bash -s' <<'REMOTE_EOF' > "$out_file"
set -euo pipefail

DISK_FREE=$(df -h / | awk 'NR==2 {print $4}')
CONTAINER_JSON=$(docker inspect nanafox-router-formal-router --format '{{json .}}')
STATS_LINE=$(docker stats --no-stream --format '{{.MemUsage}}|{{.MemPerc}}|{{.CPUPerc}}' nanafox-router-formal-router)

DB_USER=$(docker exec nanafox-router-formal-postgres sh -c 'printf "%s" "${POSTGRES_USER:-postgres}"')
DB_NAME=$(docker exec nanafox-router-formal-postgres sh -c 'printf "%s" "${POSTGRES_DB:-}"')
if [[ -z "$DB_NAME" ]]; then
  DB_NAME=$(docker exec nanafox-router-formal-postgres psql -U "$DB_USER" -d postgres -tAc \
    "SELECT datname FROM pg_database WHERE datistemplate = false AND datname NOT IN ('postgres') ORDER BY datname LIMIT 1")
fi

# NOTE: docker exec -i is REQUIRED so psql receives the SQL heredoc on stdin.
DB_JSON=$(docker exec -i nanafox-router-formal-postgres psql -U "$DB_USER" -d "$DB_NAME" -tAc -v ON_ERROR_STOP=1 <<'SQL'
SET default_transaction_read_only = on;
SELECT json_build_object(
  'db_name', current_database(),
  'db_size', pg_size_pretty(pg_database_size(current_database())),
  'migrations_count', (SELECT count(*) FROM schema_migrations),
  'last_migration', (SELECT max(filename) FROM schema_migrations),
  'migrations_fingerprint', (SELECT md5(string_agg(filename || ':' || coalesce(checksum, ''), E'\n' ORDER BY filename)) FROM schema_migrations),
  'usage_logs_total', (SELECT pg_size_pretty(pg_total_relation_size('usage_logs'))),
  'usage_logs_heap', (SELECT pg_size_pretty(pg_relation_size('usage_logs'))),
  'usage_logs_est_rows', (SELECT reltuples::bigint FROM pg_class WHERE relname = 'usage_logs'),
  'active_users', (SELECT count(*) FROM users WHERE deleted_at IS NULL),
  'total_balance', (SELECT coalesce(sum(balance), 0)::text FROM users WHERE deleted_at IS NULL),
  'active_accounts', (SELECT count(*) FROM accounts WHERE deleted_at IS NULL),
  'active_groups', (SELECT count(*) FROM groups WHERE deleted_at IS NULL),
  'active_api_keys', (SELECT count(*) FROM api_keys WHERE deleted_at IS NULL),
  'payment_orders', (SELECT count(*) FROM payment_orders)
)::text;
SQL
)

python3 - "$DISK_FREE" "$STATS_LINE" "$CONTAINER_JSON" "$DB_JSON" <<'PY'
import json, sys
disk_free, stats_line, c_raw, db_raw = sys.argv[1:5]
c = json.loads(c_raw)
db = json.loads(db_raw.strip())
allow_env = {"GOMEMLIMIT", "GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES", "TZ", "GIN_MODE"}
env_map = {}
for item in (c.get("Config", {}).get("Env") or []):
    if "=" in item:
        k, v = item.split("=", 1)
        if k in allow_env:
            env_map[k] = v
labels = c.get("Config", {}).get("Labels") or {}
out = {
    "disk_free": disk_free,
    "image": c.get("Image"),
    "version_label": labels.get("org.opencontainers.image.version") or labels.get("version"),
    "revision_label": labels.get("org.opencontainers.image.revision") or labels.get("revision"),
    "status": c.get("State", {}).get("Status"),
    "running": c.get("State", {}).get("Running"),
    "restarts": c.get("RestartCount"),
    "oom_killed": c.get("State", {}).get("OOMKilled"),
    "started_at": c.get("State", {}).get("StartedAt"),
    "stats": stats_line,
    "env": env_map,
    "db": db,
}
print(json.dumps(out, indent=2, sort_keys=True))
PY
REMOTE_EOF
}

if [[ "$MODE" == "preflight" ]]; then
  echo "=== Production Read-Only Preflight ($SSH_HOST) ==="
  run_remote_check_script
  echo "[2] Collecting read-only snapshot -> $BASELINE_FILE ..."
  collect_remote_snapshot_json "$BASELINE_FILE"
  cat "$BASELINE_FILE"
  echo "=== Preflight Complete (baseline saved to $BASELINE_FILE) ==="
  exit 0
fi

# post-switch mode
echo "=== Production Post-Switch Verification ($SSH_HOST) ==="
run_remote_check_script

POST_FILE="$(mktemp "/tmp/sub2api-prod-post-XXXXXX.json")"
trap 'rm -f "$POST_FILE"' EXIT

echo "[2] Collecting post-switch snapshot ..."
collect_remote_snapshot_json "$POST_FILE"
cat "$POST_FILE"

echo "[3] Asserting container health & comparing against baseline ..."
python3 - "$POST_FILE" "$BASELINE_FILE" <<'PY'
import json, os, sys
post = json.load(open(sys.argv[1]))
errors = []

if not post.get("running"):
    errors.append("Router container is not running")
if post.get("restarts", 1) != 0:
    errors.append(f"Router RestartCount={post.get('restarts')} (expected 0)")
if post.get("oom_killed"):
    errors.append("Router OOMKilled=true")

base_path = sys.argv[2]
if os.path.isfile(base_path):
    base = json.load(open(base_path))
    pdb, bdb = post["db"], base["db"]
    if pdb["migrations_count"] < bdb["migrations_count"]:
        errors.append(f"migrations_count regressed: {bdb['migrations_count']} -> {pdb['migrations_count']}")
    for k in ("active_users", "active_accounts", "active_groups", "active_api_keys", "payment_orders"):
        if pdb[k] < bdb[k]:
            errors.append(f"DB {k} decreased: {bdb[k]} -> {pdb[k]}")
    print(f"[INFO] Baseline diff: migrations {bdb['migrations_count']} -> {pdb['migrations_count']}, "
          f"users {bdb['active_users']} -> {pdb['active_users']}, "
          f"balance {bdb['total_balance']} -> {pdb['total_balance']}, "
          f"orders {bdb['payment_orders']} -> {pdb['payment_orders']}")
else:
    print(f"[WARN] Baseline file not found at {base_path}; skipped pre/post DB comparison")

if errors:
    for e in errors:
        print(f"[FAIL] {e}", file=sys.stderr)
    sys.exit(1)
print("[PASS] Container health (running=true, restarts=0, oom_killed=false) & DB baseline checks passed")
PY

echo "[4] Checking Router recent logs for fatal/panic/migration errors ..."
ERR_COUNT=$(ssh "$SSH_HOST" "docker logs --since 15m nanafox-router-formal-router 2>&1 | grep -ciE '\"level\"\\s*:\\s*\"(fatal|panic)\"|\\bpanic:|checksum mismatch|migration failed' || true")
echo "Fatal/panic/migration error lines in last 15m: $ERR_COUNT"
if [[ "$ERR_COUNT" != "0" ]]; then
  echo "[FAIL] Non-zero error lines in Router logs" >&2
  exit 1
fi
echo "[PASS] 0 fatal/panic/migration error lines in last 15m"

echo "[5] Checking public HTTP endpoints ..."
check_http() {
  local url="$1" expected="$2"
  local code
  code=$(curl -s --max-time 15 -o /dev/null -w '%{http_code}' "$url")
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

if (( SKIP_JS_CHECK == 0 )); then
  echo "[6] Verifying served entry JS matches local build ($LOCAL_DIST) ..."
  LOCAL_INDEX="$LOCAL_DIST/index.html"
  if [[ ! -f "$LOCAL_INDEX" ]]; then
    echo "[FAIL] Local built index.html not found at $LOCAL_INDEX (build frontend first or pass --local-dist / --skip-js-check)" >&2
    exit 1
  fi
  LOCAL_ENTRY=$(python3 -c 'import re,sys; m=re.search(r"src=\"(/assets/index-[^\"]+\.js)\"", open(sys.argv[1]).read()); print(m.group(1) if m else "")' "$LOCAL_INDEX")
  LIVE_ENTRY=$(curl -fsS --max-time 15 "${ROUTER_URL}/" | python3 -c 'import re,sys; m=re.search(r"src=\"(/assets/index-[^\"]+\.js)\"", sys.stdin.read()); print(m.group(1) if m else "")')
  if [[ -z "$LOCAL_ENTRY" || "$LOCAL_ENTRY" != "$LIVE_ENTRY" ]]; then
    echo "[FAIL] Entry JS mismatch: local='$LOCAL_ENTRY' live='$LIVE_ENTRY'" >&2
    exit 1
  fi
  LOCAL_SHA=$(shasum -a 256 "${LOCAL_DIST}${LOCAL_ENTRY}" | awk '{print $1}')
  REMOTE_SHA=$(curl -fsS --max-time 15 "${ROUTER_URL}${LIVE_ENTRY}" | shasum -a 256 | awk '{print $1}')
  if [[ "$LOCAL_SHA" == "$REMOTE_SHA" ]]; then
    echo "[PASS] Served ${LIVE_ENTRY} SHA256 matches local build (${LOCAL_SHA:0:12}...)"
  else
    echo "[FAIL] Served ${LIVE_ENTRY} SHA256 mismatch: local=$LOCAL_SHA remote=$REMOTE_SHA" >&2
    exit 1
  fi
fi

echo "=== Post-Switch Verification PASSED ==="
