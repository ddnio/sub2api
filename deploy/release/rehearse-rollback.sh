#!/usr/bin/env bash
# Local 4-phase rollback rehearsal (zero production impact):
#   Phase 1: old binary AUTO_SETUP init -> seed test user (balance 123.45)
#   Phase 2: new binary upgrade
#   Phase 3: old binary reopen upgraded DB (rollback verification)
#   Phase 4: new binary restart
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

OLD_REF=""
NEW_BIN=""
PG_IMAGE="${REHEARSAL_PG_IMAGE:-postgres:18-alpine}"
REDIS_IMAGE="${REHEARSAL_REDIS_IMAGE:-redis:8-alpine}"
PG_PORT="${REHEARSAL_PG_PORT:-55432}"
REDIS_PORT="${REHEARSAL_REDIS_PORT:-56379}"
APP_PORT="${REHEARSAL_APP_PORT:-58080}"
KEEP_TMP=0

COMPLIANCE_PHRASE="I have read, understood, and agree to the Sub2API Deployment and Operation Compliance Commitment"
ADMIN_EMAIL="admin@rehearsal.local"
ADMIN_PASS="RehearsalAdmin123!"

usage() {
  cat <<'EOF'
Usage: rehearse-rollback.sh --old-ref <git-ref> [--new-bin <path>] [--keep-tmp]

Options:
  --old-ref <ref>   Git ref/commit of currently deployed production version (get from prod image revision label)
  --new-bin <path>  Path to candidate binary (if omitted, builds ./backend/cmd/server from current repo)
  --keep-tmp        Keep temporary directory on exit for debugging
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --old-ref) OLD_REF="$2"; shift 2 ;;
    --new-bin) NEW_BIN="$2"; shift 2 ;;
    --keep-tmp) KEEP_TMP=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage; exit 2 ;;
  esac
done

if [[ -z "$OLD_REF" ]]; then
  echo "Error: --old-ref is required" >&2
  usage
  exit 2
fi

# Auto-detect Docker socket (Colima / OrbStack / default)
if [[ -z "${DOCKER_HOST:-}" ]]; then
  for sock in \
    "$HOME/.colima/default/docker.sock" \
    "$HOME/.orbstack/run/docker.sock" \
    "/var/run/docker.sock"; do
    if [[ -S "$sock" ]]; then
      export DOCKER_HOST="unix://$sock"
      break
    fi
  done
fi

RUN_ID="rh-$$"
PG_CONTAINER="${RUN_ID}-pg"
REDIS_CONTAINER="${RUN_ID}-redis"
TMP_DIR="$(mktemp -d "/tmp/sub2api-rehearsal-XXXXXX")"
OLD_WT="$TMP_DIR/old-wt"
APP_PID=""

cleanup() {
  if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true
  fi
  docker rm -f "$PG_CONTAINER" "$REDIS_CONTAINER" >/dev/null 2>&1 || true
  if [[ -d "$OLD_WT" ]]; then
    git -C "$REPO_ROOT" worktree remove --force "$OLD_WT" >/dev/null 2>&1 || true
  fi
  if (( KEEP_TMP == 0 )); then
    rm -rf "$TMP_DIR"
  else
    echo "Kept temp dir: $TMP_DIR"
  fi
}
trap cleanup EXIT

echo "=== Sub2API Local Rollback Rehearsal ==="
echo "Old ref: $OLD_REF | Docker: ${DOCKER_HOST:-default} | Temp: $TMP_DIR"

# 1. Build old & new binaries
echo "[1/4] Building old binary from $OLD_REF ..."
git -C "$REPO_ROOT" worktree add --detach "$OLD_WT" "$OLD_REF" >/dev/null
OLD_BIN="$TMP_DIR/sub2api-old"
(cd "$OLD_WT/backend" && go build -o "$OLD_BIN" ./cmd/server)

if [[ -z "$NEW_BIN" ]]; then
  echo "[1/4] Building new candidate binary from current worktree ..."
  NEW_BIN="$TMP_DIR/sub2api-new"
  (cd "$REPO_ROOT/backend" && go build -o "$NEW_BIN" ./cmd/server)
fi

# 2. Start ephemeral PG & Redis
echo "[2/4] Starting ephemeral Postgres ($PG_IMAGE) & Redis ($REDIS_IMAGE) ..."
docker run -d --rm --name "$PG_CONTAINER" \
  -e POSTGRES_USER=sub2api \
  -e POSTGRES_PASSWORD=rehearsal \
  -e POSTGRES_DB=sub2api \
  -p "127.0.0.1:${PG_PORT}:5432" \
  "$PG_IMAGE" >/dev/null

docker run -d --rm --name "$REDIS_CONTAINER" \
  -p "127.0.0.1:${REDIS_PORT}:6379" \
  "$REDIS_IMAGE" >/dev/null

PG_READY=0
for _ in $(seq 1 30); do
  # Check TCP (-h 127.0.0.1) so we don't match the initdb temporary unix-socket server
  if docker exec "$PG_CONTAINER" pg_isready -h 127.0.0.1 -U sub2api -d sub2api >/dev/null 2>&1; then
    PG_READY=1
    break
  fi
  sleep 1
done
if (( PG_READY == 0 )); then
  echo "FAIL: Postgres container did not become ready on TCP within 30s" >&2
  exit 1
fi

DATA_DIR="$TMP_DIR/data"
mkdir -p "$DATA_DIR"

run_phase() {
  local phase_name="$1"
  local bin_path="$2"
  local is_init="$3"
  local log_file="$TMP_DIR/${phase_name}.log"

  echo "--- Running phase: $phase_name ---"
  if [[ "$is_init" == "1" ]]; then
    # Phase 1 uses AUTO_SETUP=true so AutoSetupFromEnv() creates the admin user and writes DATA_DIR/config.yaml
    AUTO_SETUP="true" \
    DATA_DIR="$DATA_DIR" \
    SERVER_HOST="127.0.0.1" \
    SERVER_PORT="$APP_PORT" \
    SERVER_MODE="release" \
    DATABASE_HOST="127.0.0.1" \
    DATABASE_PORT="$PG_PORT" \
    DATABASE_USER="sub2api" \
    DATABASE_PASSWORD="rehearsal" \
    DATABASE_DBNAME="sub2api" \
    DATABASE_SSLMODE="disable" \
    REDIS_HOST="127.0.0.1" \
    REDIS_PORT="$REDIS_PORT" \
    JWT_SECRET="rehearsal-only-jwt-secret-not-for-production-use-32chars" \
    ADMIN_EMAIL="$ADMIN_EMAIL" \
    ADMIN_PASSWORD="$ADMIN_PASS" \
    GOMEMLIMIT="900MiB" \
    GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES="268435456" \
    "$bin_path" >"$log_file" 2>&1 &
  else
    GOMEMLIMIT="900MiB" \
    GATEWAY_REQUEST_BODY_MEMORY_BUDGET_BYTES="268435456" \
    DATA_DIR="$DATA_DIR" \
    "$bin_path" >"$log_file" 2>&1 &
  fi
  APP_PID=$!

  local healthy=0
  for _ in $(seq 1 60); do
    if curl -fsS --max-time 5 "http://127.0.0.1:${APP_PORT}/health" >/dev/null 2>&1; then
      healthy=1
      break
    fi
    if ! kill -0 "$APP_PID" 2>/dev/null; then
      echo "FAIL [$phase_name]: server exited prematurely. Tail of log:" >&2
      tail -n 40 "$log_file" >&2
      exit 1
    fi
    sleep 1
  done
  if (( healthy == 0 )); then
    echo "FAIL [$phase_name]: /health did not become ready in 60s" >&2
    tail -n 40 "$log_file" >&2
    exit 1
  fi

  # Unauthenticated check -> 401
  local unauth_code
  unauth_code=$(curl -s --max-time 10 -o /dev/null -w '%{http_code}' "http://127.0.0.1:${APP_PORT}/v1/models")
  [[ "$unauth_code" == "401" ]] || { echo "FAIL [$phase_name]: /v1/models returned $unauth_code (expected 401)" >&2; exit 1; }

  # Admin login
  local token
  token=$(curl -fsS --max-time 10 -X POST -H 'Content-Type: application/json' \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASS}\"}" \
    "http://127.0.0.1:${APP_PORT}/api/v1/auth/login" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d.get("data") or {}).get("access_token",""))')
  [[ -n "$token" ]] || { echo "FAIL [$phase_name]: admin login failed" >&2; exit 1; }

  # Accept compliance commitment (avoids 423 on admin endpoints)
  curl -s --max-time 10 -o /dev/null -X POST \
    -H "Authorization: Bearer $token" \
    -H 'Content-Type: application/json' \
    -d "{\"phrase\":\"${COMPLIANCE_PHRASE}\",\"language\":\"en\"}" \
    "http://127.0.0.1:${APP_PORT}/api/v1/admin/compliance/accept" || true

  # Verify core admin endpoints return 200
  for ep in "/api/v1/admin/groups?page=1&page_size=20" "/api/v1/admin/users?page=1&page_size=20" "/api/v1/admin/accounts?page=1&page_size=20" "/api/v1/admin/settings"; do
    local code
    code=$(curl -s --max-time 10 -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $token" "http://127.0.0.1:${APP_PORT}${ep}")
    [[ "$code" == "200" ]] || { echo "FAIL [$phase_name]: $ep returned $code (expected 200)" >&2; exit 1; }
  done

  # Seed test user on phase 1 (uses partial unique index predicate WHERE deleted_at IS NULL)
  if [[ "$is_init" == "1" ]]; then
    docker exec "$PG_CONTAINER" psql -U sub2api -d sub2api -v ON_ERROR_STOP=1 -c \
      "INSERT INTO users (email, password_hash, role, balance, status, created_at, updated_at)
       VALUES ('seed@rehearsal.local', 'hash', 'user', 123.45, 'active', NOW(), NOW())
       ON CONFLICT (email) WHERE deleted_at IS NULL DO NOTHING;" >/dev/null
  fi

  # Query migrations & seeded balance
  local mig_info balance fatal_cnt
  mig_info=$(docker exec "$PG_CONTAINER" psql -U sub2api -d sub2api -tAc \
    "SELECT count(*) || ' (' || max(filename) || ')' FROM schema_migrations;")
  balance=$(docker exec "$PG_CONTAINER" psql -U sub2api -d sub2api -tAc \
    "SELECT balance::numeric(10,2) FROM users WHERE email='seed@rehearsal.local' AND deleted_at IS NULL;")
  [[ "$balance" == "123.45" ]] || { echo "FAIL [$phase_name]: seed balance=$balance (expected 123.45)" >&2; exit 1; }

  # Stop server gracefully
  kill "$APP_PID" 2>/dev/null || true
  wait "$APP_PID" 2>/dev/null || true
  APP_PID=""

  fatal_cnt=$(grep -ciE '"level"\s*:\s*"(fatal|panic)"|\bpanic:|checksum mismatch|migration failed' "$log_file" || true)
  [[ "$fatal_cnt" == "0" ]] || { echo "FAIL [$phase_name]: found $fatal_cnt fatal/panic lines in $log_file" >&2; exit 1; }

  printf '| %s | 200 | 401 | 200x4 | %s | %s | 0 |\n' "$phase_name" "$mig_info" "$balance" >> "$TMP_DIR/summary.md"
}

cat > "$TMP_DIR/summary.md" <<'EOF'
| Phase | /health | /v1/models (no auth) | Admin APIs | Migrations | Seed Balance | Fatal Logs |
| --- | --- | --- | --- | --- | --- | --- |
EOF

echo "[3/4] Running 4 rehearsal phases ..."
run_phase "1-old-init" "$OLD_BIN" "1"
run_phase "2-new-upgrade" "$NEW_BIN" "0"
run_phase "3-old-rollback" "$OLD_BIN" "0"
run_phase "4-new-restart" "$NEW_BIN" "0"

echo ""
echo "=== [4/4] Rollback Rehearsal PASSED ==="
cat "$TMP_DIR/summary.md"
