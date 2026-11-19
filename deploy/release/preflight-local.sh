#!/usr/bin/env bash
# Local environment preflight for upstream merge & release verification.
# Read-only: checks disk space, Go/Node/pnpm/lint/vulncheck versions, Docker/Colima socket, and git remotes.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MIN_FREE_GB="${MIN_FREE_GB:-15}"

FAILURES=0
WARNINGS=0

pass() { printf '[PASS] %s\n' "$1"; }
warn() { printf '[WARN] %s\n' "$1"; WARNINGS=$((WARNINGS + 1)); }
fail() { printf '[FAIL] %s\n' "$1"; FAILURES=$((FAILURES + 1)); }

echo "=== Sub2API Local Release Preflight ==="
echo "Repo: $REPO_ROOT"

# Include GOPATH/bin and nvm Node 24 if present
if command -v go >/dev/null 2>&1; then
  GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin"
  if [[ -d "$GOPATH_BIN" && ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    export PATH="$PATH:$GOPATH_BIN"
  fi
fi

# 1. Disk space check (in GB)
FREE_KB=$(df -k "$REPO_ROOT" | awk 'NR==2 {print $4}')
FREE_GB=$((FREE_KB / 1024 / 1024))
if (( FREE_GB >= MIN_FREE_GB )); then
  pass "Free disk space: ${FREE_GB} GiB (threshold: ${MIN_FREE_GB} GiB)"
else
  fail "Free disk space: ${FREE_GB} GiB < ${MIN_FREE_GB} GiB — clean up caches/Docker before compiling or running integration tests"
fi

# 2. Expected versions from repo files
EXPECTED_GO=$(awk '/^go / {print $2; exit}' "$REPO_ROOT/backend/go.mod" 2>/dev/null || true)
EXPECTED_LINT=$(awk '/golangci-lint-action/{found=1} found && /version:/{gsub(/["'\'' ]/, "", $2); print $2; exit}' "$REPO_ROOT/.github/workflows/backend-ci.yml" 2>/dev/null || true)
EXPECTED_NODE_MAJOR=$(awk '/setup-node/{found=1} found && /node-version:/{gsub(/["'\'' ]/, "", $2); print $2; exit}' "$REPO_ROOT/.github/workflows/backend-ci.yml" 2>/dev/null || true)
EXPECTED_PNPM_MAJOR=$(awk '/pnpm\/action-setup/{found=1} found && /version:/{gsub(/["'\'' ]/, "", $2); print $2; exit}' "$REPO_ROOT/.github/workflows/backend-ci.yml" 2>/dev/null || true)

# 3. Go version
if command -v go >/dev/null 2>&1; then
  ACTUAL_GO=$(go version | awk '{sub(/^go/, "", $3); print $3}')
  if [[ -z "$EXPECTED_GO" || "$ACTUAL_GO" == "$EXPECTED_GO" ]]; then
    pass "Go: $ACTUAL_GO (expected: ${EXPECTED_GO:-any})"
  else
    warn "Go: $ACTUAL_GO (go.mod specifies $EXPECTED_GO)"
  fi
else
  fail "Go not found (go.mod requires ${EXPECTED_GO:-Go})"
fi

# 4. golangci-lint
if command -v golangci-lint >/dev/null 2>&1; then
  ACTUAL_LINT=$(golangci-lint version --short 2>/dev/null || golangci-lint --version 2>/dev/null | head -n1)
  if [[ -n "$EXPECTED_LINT" && "$ACTUAL_LINT" != *"${EXPECTED_LINT#v}"* ]]; then
    warn "golangci-lint: $ACTUAL_LINT (CI pins $EXPECTED_LINT)"
  else
    pass "golangci-lint: $ACTUAL_LINT"
  fi
else
  fail "golangci-lint not found (CI pins ${EXPECTED_LINT:-latest})"
fi

# 5. govulncheck
if command -v govulncheck >/dev/null 2>&1; then
  pass "govulncheck: $(command -v govulncheck)"
else
  warn "govulncheck not found (install: go install golang.org/x/vuln/cmd/govulncheck@latest)"
fi

# 6. Node & pnpm
if command -v node >/dev/null 2>&1; then
  ACTUAL_NODE=$(node -v)
  ACTUAL_NODE_MAJOR="${ACTUAL_NODE#v}"
  ACTUAL_NODE_MAJOR="${ACTUAL_NODE_MAJOR%%.*}"
  if [[ -n "$EXPECTED_NODE_MAJOR" && "$ACTUAL_NODE_MAJOR" != "$EXPECTED_NODE_MAJOR" ]]; then
    warn "Node: $ACTUAL_NODE (CI pins v${EXPECTED_NODE_MAJOR}; consider nvm use ${EXPECTED_NODE_MAJOR})"
  else
    pass "Node: $ACTUAL_NODE"
  fi
else
  fail "Node not found (CI pins v${EXPECTED_NODE_MAJOR:-24})"
fi

if command -v pnpm >/dev/null 2>&1; then
  ACTUAL_PNPM=$(pnpm -v)
  ACTUAL_PNPM_MAJOR="${ACTUAL_PNPM%%.*}"
  if [[ -n "$EXPECTED_PNPM_MAJOR" && "$ACTUAL_PNPM_MAJOR" != "$EXPECTED_PNPM_MAJOR" ]]; then
    warn "pnpm: $ACTUAL_PNPM (CI pins v${EXPECTED_PNPM_MAJOR})"
  else
    pass "pnpm: $ACTUAL_PNPM"
  fi
else
  fail "pnpm not found (CI pins v${EXPECTED_PNPM_MAJOR:-9})"
fi

# 7. Docker / Colima detection
DETECTED_DOCKER_HOST="${DOCKER_HOST:-}"
if [[ -z "$DETECTED_DOCKER_HOST" ]]; then
  for sock in \
    "$HOME/.colima/default/docker.sock" \
    "$HOME/.orbstack/run/docker.sock" \
    "/var/run/docker.sock"; do
    if [[ -S "$sock" ]]; then
      DETECTED_DOCKER_HOST="unix://$sock"
      break
    fi
  done
fi

if command -v docker >/dev/null 2>&1 && [[ -n "$DETECTED_DOCKER_HOST" ]]; then
  if DOCKER_HOST="$DETECTED_DOCKER_HOST" docker info >/dev/null 2>&1; then
    pass "Docker reachable via $DETECTED_DOCKER_HOST"
    if [[ "$DETECTED_DOCKER_HOST" != "unix:///var/run/docker.sock" ]]; then
      echo "       Tip: export DOCKER_HOST='$DETECTED_DOCKER_HOST' TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock"
    fi
  else
    fail "Docker daemon not responding at $DETECTED_DOCKER_HOST (run 'colima start' or start Docker)"
  fi
else
  fail "Docker CLI or socket not found (needed for integration tests & rollback rehearsal)"
fi

# 8. Git remotes & active worktrees
if git -C "$REPO_ROOT" remote get-url origin >/dev/null 2>&1 && git -C "$REPO_ROOT" remote get-url upstream >/dev/null 2>&1; then
  pass "Git remotes present: origin & upstream"
else
  fail "Missing 'origin' or 'upstream' git remote"
fi

echo "--- Active worktrees ---"
git -C "$REPO_ROOT" worktree list

echo "=== Summary: $FAILURES failure(s), $WARNINGS warning(s) ==="
if (( FAILURES > 0 )); then
  exit 1
fi
