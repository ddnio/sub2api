---
name: sub2api-upstream-release
description: "Use for the full Sub2API upstream release workflow when upstream has a new version/tag or the user asks to merge, review, promote, or production-deploy an upstream update. Covers planning the upstream merge, direct-batch code merge and conflict repair, code review, an explicit environment validation gate, then one Router production rollout and main promotion."
---

# Sub2API Upstream Release

Run the complete upstream release lane for this repository: plan the merge, integrate upstream, review and fix, wait for an explicitly authorized validation target, then promote to production.

This is a project-local skill for `/Users/nio/project/nanafox/sub2api`. Use `sub2api-ops` for live deployment details and keep `origin=ddnio/sub2api`, `upstream=Wei-Shaw/sub2api`.

## Operating Mode

- Execute merge, verification, review, and release-branch push directly when the user asks to "合并 upstream", "upstream 更新到 X", or "按流程执行".
- If OMX runtime is attached and `$ralplan` / `$ralph` is available, use them for planning/execution loops. In Codex App outside tmux, run the same phases directly.
- Use the direct-batch merge model by default: create an isolated scratch worktree, merge the upstream release tag, capture conflicts/protected deletions/migration drift, repair and verify, then promote or replay into the release branch.
- Do not default to per-PR or per-commit cherry-pick planning unless the direct merge proves a subfeature must be split out.
- Treat production deploys as live operations: only run them when the user explicitly asks for production/fx deployment or has explicitly authorized the end-to-end release after test validation.

## Phase 0: Refresh Facts

Start every run by refreshing local source of truth:

```bash
git status --short --branch
git remote -v
git fetch origin
git fetch upstream --tags
sed -n '1,140p' AGENTS.md
sed -n '1,160p' docs/engineering/git-workflow.md
sed -n '1,220p' docs/engineering/deployment.md
sed -n '1,220p' .codex/skills/sub2api-ops/SKILL.md
```

Identify:

- target upstream tag/version
- current `origin/main`
- latest deployed/tested release branch, if any
- whether `main` already contains the release
- whether there are dirty local changes that must not be overwritten

## Phase 1: Plan The Upstream Merge

Create a concise plan before changing code. Prefer a plan artifact under `.omx/plans/` or `docs/plans/` when the merge is substantial.

Planning checklist:

- Compare `origin/main` against the target upstream tag.
- Summarize feature groups and likely risk areas.
- List migrations added upstream and existing local migration numbers; plan renumbering if needed.
- Identify protected local behavior that must survive, especially gateway fixes, auth/OAuth flows, billing/quota, deployment scripts, and prior hotfixes.
- Define verification: backend tests, frontend lint/typecheck, targeted Vitest, `make build`, conflict marker scan, and deployment smoke checks.
- Define stop conditions: unresolved conflict, destructive migration ambiguity, failing tests without a clear fix, or production validation failure.

The planning output should clearly say whether to proceed with one direct release merge or split out a risky subfeature.

## Phase 2: Merge And Repair

Use isolated worktrees. Do not merge directly in a dirty main checkout.

Typical sequence:

```bash
git fetch origin
git fetch upstream --tags
git worktree add .claude/worktrees/release-vX.Y.Z-merge -b release/vX.Y.Z origin/main
cd .claude/worktrees/release-vX.Y.Z-merge
git merge upstream/vX.Y.Z
```

When conflicts occur:

- Resolve in favor of preserving local production behavior unless upstream intentionally fixes the same behavior better.
- Keep local migration history immutable; renumber upstream migrations upward instead of changing already-applied migration files.
- Search for conflict markers with a precise pattern after every conflict pass:

```bash
rg -n "^(<<<<<<< .+|=======|>>>>>>> .+)$" .
```

After repair:

```bash
git diff --check
go test ./...
pnpm --dir frontend run lint:check
pnpm --dir frontend run typecheck
pnpm --dir frontend exec vitest run <targeted changed-area specs>
make build
```

If generated code or lockfiles are expected to change, regenerate using the repo's existing commands and keep those changes in the release branch.

Commit with the repository's Lore commit protocol. If a hook hangs for known tooling reasons, use `--no-verify` only after independent verification passes, and mention the hook issue in the handoff.

Push the release branch to `origin`:

```bash
git push origin release/vX.Y.Z
```

## Phase 3: Code Review Gate

Review before any live deployment.

Be explicit about which review path was completed:

- A manual focused inspection is useful, but it is not the same as the formal `$code-review` gate.
- If the user asks whether code review was done, answer separately for "manual inspection" and "formal code-review workflow".
- In Codex App sessions where sub-agent review lanes are unavailable or not authorized, run the same checklist locally and label it as a local code-review gate, not as delegated `code-reviewer` / `architect` output.
- Do not advance to test deployment until the review result is unambiguous: blocking findings fixed, or no blocking findings found.

Review stance:

- Findings first, ordered by severity with file/line references.
- Focus on behavioral regressions, migration safety, auth/gateway compatibility, billing/quota correctness, deployment blast radius, and missing tests.
- If an issue is real and low-risk to fix, fix it immediately, rerun targeted verification, commit, and push the release branch.
- If no blocking issues remain, say that explicitly and list residual risks.

Minimum review checks:

```bash
git diff --stat origin/main...HEAD
git diff --name-only origin/main...HEAD
git diff --check
rg -n "^(<<<<<<< .+|=======|>>>>>>> .+)$" .
```

Also inspect high-risk paths touched by the merge:

- `backend/migrations/`
- `backend/internal/service/`
- `backend/internal/handler/`
- `backend/internal/repository/`
- `backend/internal/server/`
- `frontend/src/api/`
- `frontend/src/views/admin/`
- `frontend/src/views/auth/`
- `frontend/src/components/`
- `deploy/`

Do not proceed to test deployment until blocking review findings are fixed or explicitly accepted.

## Phase 4: Wait For An Explicit Validation Target

There is currently no active Sub2API test environment. The old host `108.160.133.141`, its `sub2api-test` container, `/etc/sub2api/test.yaml`, port `8081`, and `router-test.nanafox.com` are legacy state, not deployment targets.

After the code-review gate:

- Report the release branch, commit, local verification, review result, database changes, and a web validation checklist.
- Stop before any SSH, remote checkout, Docker build, database migration, or deployment.
- Continue only when the user explicitly identifies and authorizes a current validation or production target.
- Never infer a test environment from a legacy container, config file, DNS record, prior deployment history, or the words "执行合并".

If a new test environment is introduced, update `docs/engineering/deployment.md`, `sub2api-ops`, and this section with its verified host, domain, data boundary, deployment command, validation, and rollback before using it.

## Phase 5: Deploy One Router production Production Candidate

After the user-authorized validation gate, follow `sub2api-ops` and `docs/engineering/deployment.md`. Freeze the release commit and verified image digest. Run `deploy/check-router-production.py` on Router production; capture current actual mounts/configuration, database roles, networks, resource limits and rollback privately.

Production is one `fx-production-router` on Router production at 127.0.0.1:18080. router.nanafox.com and fx.nanafox.com are aliases, not independent 迁移前双生产 deployments. Studio remains a separate release. Do not use `deploy-server.sh prod`, revive old-host production, replace old Router production database containers, or run the historical three-surface rollout.

Replace only the authorized app using the current runbook. Require Router health, unsigned trusted auth 401, signed Studio auth, both domain routes and affected model/payment/plugin checks. A 404 trusted-auth route is a configuration regression. Preserve secrets and active capabilities; never add migration restrictions. Tests and code review do not authorize an unrelated app or DNS release.

## Phase 6: Promote Verified Release To Main

After production verification, merge the corresponding release into main according to `docs/engineering/git-workflow.md` and push only to origin. Do not reverse this order by merging an unverified application release into production main first. Record the exact tested and deployed digest and commit; merging main does not trigger a second rollout to Router production or the old host.

## Completion Report

Use this shape:

```text
Target:
- upstream tag:
- release branch:
- main commit:
- deployed surfaces:

Merge:
- conflicts:
- migrations:
- protected local behavior:

Review:
- findings:
- fixes:

Verification:
- local tests:
- test deployment:
- production deployment:
- fx alias verification:

Safety:
- secrets exposed: no
- upstream pushed: no
- remaining risks:
```

Never print secrets from `/etc/sub2api/*.yaml`, env files, tokens, private keys, payment keys, TOTP keys, or JWT secrets.
