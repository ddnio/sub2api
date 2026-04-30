-- 095a_subscription_plans_badge.sql
-- Fork patch: add badge column to subscription_plans (UI label: 推荐/热门/限时 etc.)
-- Upstream does not have this field. Must be preserved on every upstream sync.
-- Runs after 095_subscription_plans.sql creates the table.

ALTER TABLE subscription_plans
  ADD COLUMN IF NOT EXISTS badge VARCHAR(20);

COMMENT ON COLUMN subscription_plans.badge IS
  '[fork patch] UI badge label (推荐/热门/限时 etc.). Not in upstream — preserve on each sync.';
