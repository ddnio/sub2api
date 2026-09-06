---
name: sub2api-ops
description: "Sub2API deployment and operations on the current Router production stack, isolated legacy tests, Docker, Caddy, NAS backups, and release targeting."
---

# Sub2API Ops

User-facing environment name: **Router 生产环境**. Keep existing machine identifiers exact; do not rename live containers just to update terminology.

Project-local skill. First read `AGENTS.md`, `docs/engineering/deployment.md`, and `docs/engineering/git-workflow.md`; use `deploy/check-router-production.py` for a read-only live target check when server work is in scope. Date-stamped migration notes are history, not commands to rerun.

## Current baseline: 2026-09-07

- One production host: Aliyun Singapore Router production, nio@43.106.8.109, hostname iZt4nfdwwipifuqc4q2v01Z, 2 vCPU / 4 GiB.
- router.nanafox.com and fx.nanafox.com share fx-production-router, 127.0.0.1:18080 → 8080.
- studio.nanafox.com uses fx-production-studio, 127.0.0.1:18789 → 8788; studio-fx is a temporary alias.
- fx-production-postgres owns sub2api/router_app and nanafox_studio_prod/studio_app on fx-production-pgdata. Redis is fx-production-redis on fx-production-redisdata. Preserved pre-migration sub2api_tob/PG/Redis are separate preserved data, never production defaults.
- Configuration base /home/nio/recovery/live-production-20260906: actual Router mount router.preview.json is now production content; Studio uses studio.production.env. Preserve actual mounts, trusted auth, JWT/TOTP and payment/generation/R2 keys. Never substitute historical preview settings based on filenames.
- Caddy /etc/caddy/Caddyfile serves the image plugin /tools/image-playground/ from /srv/nanafox/image-playground/prod-current, logo assets from /srv/nanafox/fx-web-assets, Landing from /srv/nanafox-landing/current. These are outside app images and must remain deployed/backed up.
- Old host 108.160.133.141 has stopped Router/Studio production containers (restart=no) and temporary HTTPS forwarding to Router production for stale DNS. Whole host is not decommissioned. Tests remain old-host only; do not deploy tests to Router production by default. pick/newapi are excluded from migration.

## Scope and execution

Use existing user authorization for the specified target and work. Read-only inspection and routine authorized fixes do not require repeated permission. Docs/skill-only work does not trigger application release. Keep origin=ddnio/sub2api and upstream=Wei-Shaw/sub2api; preserve dirty worktrees.

Before replacement, record target, candidate commit/image digest, current config/mount/network/volume/ports, backup status, validation and compatible rollback. The preflight is read-only and is not a deploy tool. `deploy-server.sh prod` is retired; old test script rejects Router production. There is no verified automatic Router production replacement script in this repository: construct the reviewed candidate command from current inspect and the runbook, never guess old defaults.

Do not add preview write guards, disable payment/generation, rewrite menus or alter unrelated services during ordinary deployment. Follow the user's requested scope. Real payment tests need authorization for that transaction. Never output or commit secrets, full env/inspect, production user data or private keys. Authorized backup/config transfer is allowed through private files and verified channels.

## Acceptance

1. Host/container/image/ports/mounts/networks match current target; PostgreSQL app connections use the intended roles and databases.
2. Router health, unsigned Studio Auth 401 (404 means adapter missing), signed Studio auth, Studio ready and formal deep links.
3. Image plugin static files and user/admin APIs, R2 objects, model call and affected payment paths. State what was actually tested; GET 200 is not payment/generation acceptance.
4. Actual formal-domain request marker in Router production logs plus authoritative DNS and old-container stopped state; ping or HTTP 200 alone is insufficient because old DNS may bridge through the source host.
5. Resource/queue/error monitoring and backups. NAS daily pull at 03:35 Beijing stores fx-production/<UTC>/ in nanafox-postgres-backups. Check NAS last-success.json plus MinIO readback; the large Router first run was still transferring at migration handoff. Studio now has an independent 03:45 job, studio-production/<UTC>/ prefix and studio/last-success.json; its 20260906T172952Z snapshot passed all 5 file readbacks. Refresh both success timestamps before claiming current backup freshness.

Rollback an app with a compatible image and current production configuration. Do not re-import a migration snapshot over new production writes. Original destination and source backups remain recovery evidence. The user's no-stop-write exception applied to this cutover, not arbitrary future destructive database work.

Report changed files or runtime actions, evidence, current backup freshness, remaining gaps and rollback. Distinguish local repository updates from remote server checkout updates and actual releases.
