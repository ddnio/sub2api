-- Migration: 132_update_claude_code_monitor_template
-- 129 号迁移在 v0.1.117 已部署过，ON CONFLICT DO NOTHING 不会覆盖现有环境中的
-- 「Claude Code 伪装」模板。这里补一条前向迁移，把旧 2.1.114 头升级到当前运行时
-- 伪装版本 2.1.92，同时只改请求头字段，不覆盖用户可能已调整的 body_override。

WITH constants AS (
    SELECT
        'claude-cli/2.1.114 (external, sdk-cli)'::text AS old_user_agent,
        'claude-code-20250219,interleaved-thinking-2025-05-14,context-management-2025-06-27,prompt-caching-scope-2026-01-05,advisor-tool-2026-03-01'::text AS old_beta,
        'claude-cli/2.1.92 (external, cli)'::text AS new_user_agent,
        'claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,prompt-caching-scope-2026-01-05,effort-2025-11-24,redact-thinking-2026-02-12,context-management-2025-06-27,extended-cache-ttl-2025-04-11'::text AS new_beta,
        '完整模拟 Claude Code 2.1.92 客户端：UA + anthropic-beta + system + metadata.user_id 全部对齐，绕过 Anthropic 上游 ''Claude Code only'' 限制（如 Max 套餐）。'::text AS new_description
),
stale_template AS (
    SELECT t.id
    FROM channel_monitor_request_templates t, constants c
    WHERE t.provider = 'anthropic'
      AND t.name = 'Claude Code 伪装'
      AND (
          t.description LIKE '%2.1.114%'
          OR t.extra_headers->>'User-Agent' = c.old_user_agent
          OR t.extra_headers->>'anthropic-beta' = c.old_beta
          OR COALESCE(t.extra_headers->>'anthropic-beta', '') LIKE '%advisor-tool-2026-03-01%'
      )
),
updated_template AS (
    UPDATE channel_monitor_request_templates t
    SET
        description = c.new_description,
        extra_headers = jsonb_set(
            jsonb_set(
                jsonb_set(
                    jsonb_set(
                        jsonb_set(
                            COALESCE(t.extra_headers, '{}'::jsonb),
                            '{User-Agent}', to_jsonb(c.new_user_agent), true
                        ),
                        '{X-App}', to_jsonb('cli'::text), true
                    ),
                    '{anthropic-version}', to_jsonb('2023-06-01'::text), true
                ),
                '{anthropic-beta}', to_jsonb(c.new_beta), true
            ),
            '{anthropic-dangerous-direct-browser-access}', to_jsonb('true'::text), true
        ),
        updated_at = NOW()
    FROM constants c
    WHERE t.id IN (SELECT id FROM stale_template)
    RETURNING t.id
)
UPDATE channel_monitors m
SET extra_headers = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    COALESCE(m.extra_headers, '{}'::jsonb),
                    '{User-Agent}', to_jsonb(c.new_user_agent), true
                ),
                '{X-App}', to_jsonb('cli'::text), true
            ),
            '{anthropic-version}', to_jsonb('2023-06-01'::text), true
        ),
        '{anthropic-beta}', to_jsonb(c.new_beta), true
    ),
    '{anthropic-dangerous-direct-browser-access}', to_jsonb('true'::text), true
)
FROM constants c
WHERE m.provider = 'anthropic'
  AND (
      m.extra_headers->>'User-Agent' = c.old_user_agent
      OR m.extra_headers->>'anthropic-beta' = c.old_beta
      OR COALESCE(m.extra_headers->>'anthropic-beta', '') LIKE '%advisor-tool-2026-03-01%'
  );
