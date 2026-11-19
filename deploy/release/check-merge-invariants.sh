#!/usr/bin/env bash
# Deterministic post-merge invariant checks for NanaFox fork of Sub2API.
# Verifies migration numbering, historical migration immutability, hardcoded SQL filename references,
# VERSION alignment, forbidden secret tracking, and lists both-sides-modified files for targeted review.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

BASE_REF="main"
UPSTREAM_TAG=""

usage() {
  cat <<'EOF'
Usage: check-merge-invariants.sh [--base <git-ref>] [--upstream <tag>]

Options:
  --base <ref>       Base branch/commit before merge (default: main)
  --upstream <tag>   Upstream tag merged (e.g. v0.2.23) for VERSION & both-sides diff checks
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base) BASE_REF="$2"; shift 2 ;;
    --upstream) UPSTREAM_TAG="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage; exit 2 ;;
  esac
done

FAILURES=0
WARNINGS=0

pass() { printf '[PASS] %s\n' "$1"; }
warn() { printf '[WARN] %s\n' "$1"; WARNINGS=$((WARNINGS + 1)); }
fail() { printf '[FAIL] %s\n' "$1"; FAILURES=$((FAILURES + 1)); }

echo "=== Sub2API Merge Invariant Checks ==="
echo "Repo: $REPO_ROOT | Base: $BASE_REF | Upstream: ${UPSTREAM_TAG:-<none>}"

# 1a. Check historical migrations in --base were NOT modified, deleted, or renamed (protects checksums)
MIG_DIR="$REPO_ROOT/backend/migrations"
MUTATED_MIGS=$(git -C "$REPO_ROOT" diff --name-only --diff-filter=MDR "$BASE_REF" -- 'backend/migrations/*.sql' || true)
if [[ -z "$MUTATED_MIGS" ]]; then
  pass "No historical migration SQL files modified, deleted, or renamed since $BASE_REF"
else
  while IFS= read -r m; do
    [[ -z "$m" ]] && continue
    fail "Historical migration mutated ($m) — breaks production schema_migrations checksum unless compatibility rule added"
  done <<< "$MUTATED_MIGS"
fi

# 1b. Check newly added migration numbering against base max
BASE_MAX_NUM=$(git -C "$REPO_ROOT" ls-tree --name-only "$BASE_REF" -- backend/migrations/ \
  | sed -n 's|^backend/migrations/\([0-9][0-9]*\)_.*\.sql$|\1|p' \
  | sort -n | tail -n 1 || echo "0")
BASE_MAX_NUM="${BASE_MAX_NUM:-0}"

NEW_MIG_BAD=0
while IFS= read -r relpath; do
  [[ -z "$relpath" ]] && continue
  fname=$(basename "$relpath")
  if [[ "$fname" =~ ^([0-9]+)_.*\.sql$ ]]; then
    num=$((10#${BASH_REMATCH[1]}))
    if (( num <= 10#$BASE_MAX_NUM )); then
      fail "New migration '$fname' has number $num <= base max $BASE_MAX_NUM (must be renumbered after $BASE_MAX_NUM)"
      NEW_MIG_BAD=$((NEW_MIG_BAD + 1))
    fi
  fi
done < <(git -C "$REPO_ROOT" diff --name-only --diff-filter=A "$BASE_REF" -- 'backend/migrations/*.sql')

if (( NEW_MIG_BAD == 0 )); then
  CURRENT_COUNT=$(find "$MIG_DIR" -maxdepth 1 -name '*.sql' | wc -l | tr -d ' ')
  pass "New migration numbering OK (base max: $BASE_MAX_NUM, current total SQL files: $CURRENT_COUNT)"
fi

# 2. Check hardcoded NNN_*.sql references introduced since --base still exist in backend/migrations/
# NOTE: uses python3 -c (NOT heredoc) so git diff pipe reaches sys.stdin properly.
MISSING_REFS=0
while IFS= read -r ref_line; do
  [[ -z "$ref_line" ]] && continue
  go_file="${ref_line%%:*}"
  sql_name="${ref_line#*:}"
  if [[ ! -f "$MIG_DIR/$sql_name" ]]; then
    if [[ "$go_file" == *"_test.go" && "$sql_name" =~ ^(000|001|002|010|999)_ ]]; then
      continue
    fi
    fail "Newly referenced SQL '$sql_name' in $go_file does not exist in backend/migrations/ (forgot to update after renumbering?)"
    MISSING_REFS=$((MISSING_REFS + 1))
  fi
done < <(
  git -C "$REPO_ROOT" diff -U0 "$BASE_REF" -- 'backend/**/*.go' 'backend/*.go' | python3 -c '
import re, sys
pat = re.compile(r"\"(\d{3}_[a-zA-Z0-9_]+\.sql)\"")
cur_file = ""
seen = set()
for line in sys.stdin:
    if line.startswith("+++ b/"):
        cur_file = line[6:].strip()
    elif line.startswith("+") and not line.startswith("+++"):
        for m in pat.findall(line):
            key = f"{cur_file}:{m}"
            if key not in seen:
                seen.add(key)
                print(key)
'
)
if (( MISSING_REFS == 0 )); then
  pass "All newly added/modified SQL filename references in backend Go code resolve to existing files"
fi

# 3. Check VERSION file if --upstream provided
if [[ -n "$UPSTREAM_TAG" ]]; then
  EXPECTED_VER="${UPSTREAM_TAG#v}"
  ACTUAL_VER=$(tr -d '[:space:]' < "$REPO_ROOT/backend/cmd/server/VERSION")
  if [[ "$ACTUAL_VER" == "$EXPECTED_VER" ]]; then
    pass "backend/cmd/server/VERSION matches $EXPECTED_VER"
  else
    fail "backend/cmd/server/VERSION is '$ACTUAL_VER', expected '$EXPECTED_VER'"
  fi
fi

# 4. Forbidden secret files must not be tracked
TRACKED_SECRETS=$(git -C "$REPO_ROOT" ls-files -- backend/config.yaml deploy/config.yaml 'backend/config/*.pem' 'backend/config/*.key' || true)
if [[ -z "$TRACKED_SECRETS" ]]; then
  pass "No forbidden config/key files tracked in git"
else
  fail "Forbidden secret files tracked in git: $TRACKED_SECRETS"
fi

# 5. Both-sides-modified file inventory (when --upstream is given)
if [[ -n "$UPSTREAM_TAG" ]]; then
  if MERGE_BASE=$(git -C "$REPO_ROOT" merge-base "$BASE_REF" "$UPSTREAM_TAG" 2>/dev/null); then
    BOTH_COUNT=$(comm -12 \
      <(git -C "$REPO_ROOT" diff --name-only "$MERGE_BASE" "$BASE_REF" | sort) \
      <(git -C "$REPO_ROOT" diff --name-only "$MERGE_BASE" "$UPSTREAM_TAG" | sort) | wc -l | tr -d ' ')
    echo "--- Both-sides-modified files ($BOTH_COUNT total since merge-base ${MERGE_BASE:0:9}) ---"
    SENSITIVE=$(comm -12 \
      <(git -C "$REPO_ROOT" diff --name-only "$MERGE_BASE" "$BASE_REF" | sort) \
      <(git -C "$REPO_ROOT" diff --name-only "$MERGE_BASE" "$UPSTREAM_TAG" | sort) \
      | grep -E 'wxpay|payment|studio|referral|image_creation|async_image|wire_gen|auth_service|gateway_claude_oauth_body|openai_gateway_handler|migrations_runner' || true)
    if [[ -n "$SENSITIVE" ]]; then
      warn "High-risk fork files modified on both sides (review carefully):"
      printf '  - %s\n' $SENSITIVE
    else
      pass "No high-risk fork core files in both-sides-modified set"
    fi
  else
    warn "Could not compute merge-base between $BASE_REF and $UPSTREAM_TAG"
  fi
fi

echo "=== Summary: $FAILURES failure(s), $WARNINGS warning(s) ==="
if (( FAILURES > 0 )); then
  exit 1
fi
