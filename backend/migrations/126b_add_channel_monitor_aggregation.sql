-- Migration: 126_add_channel_monitor_aggregation
-- 渠道监控日聚合：把 channel_monitor_histories 的明细按天预聚合到 rollup 表，
-- 作为未来长窗口/降级查询的兜底。当前读路径直接使用 30 天 raw histories，
-- rollup 同样保留 30 天并由 ops cleanup 任务每天维护。
--
-- 设计要点：
--   - channel_monitor_histories 在本迁移先加 deleted_at；后续 127b 会移除该列，
--     最终策略是 raw histories 物理保留 30 天，避免 read-side 过滤软删除状态。
--   - channel_monitor_daily_rollups 按 (monitor_id, model, bucket_date) 唯一，
--     用 ON CONFLICT DO UPDATE 实现幂等回填，状态分布和延迟分子分母都保留，
--     方便后续按窗口任意求加权可用率和均值。
--   - watermark 表只有一行（id=1），记录最近一次聚合到达的日期，避免重启后重复
--     扫全表。
--   - rollup 上 (bucket_date) 索引服务清理任务的 DELETE WHERE bucket_date < cutoff。

-- 1) 给历史明细表加软删除字段
ALTER TABLE channel_monitor_histories
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_channel_monitor_histories_deleted_at
    ON channel_monitor_histories (deleted_at);

-- 2) 创建日聚合表
CREATE TABLE IF NOT EXISTS channel_monitor_daily_rollups (
    id                    BIGSERIAL PRIMARY KEY,
    monitor_id            BIGINT       NOT NULL REFERENCES channel_monitors(id) ON DELETE CASCADE,
    model                 VARCHAR(200) NOT NULL,
    bucket_date           DATE         NOT NULL,
    total_checks          INT          NOT NULL DEFAULT 0,
    ok_count              INT          NOT NULL DEFAULT 0,
    operational_count     INT          NOT NULL DEFAULT 0,
    degraded_count        INT          NOT NULL DEFAULT 0,
    failed_count          INT          NOT NULL DEFAULT 0,
    error_count           INT          NOT NULL DEFAULT 0,
    sum_latency_ms        BIGINT       NOT NULL DEFAULT 0,
    count_latency         INT          NOT NULL DEFAULT 0,
    sum_ping_latency_ms   BIGINT       NOT NULL DEFAULT 0,
    count_ping_latency    INT          NOT NULL DEFAULT 0,
    computed_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_monitor_daily_rollups_unique
    ON channel_monitor_daily_rollups (monitor_id, model, bucket_date);
CREATE INDEX IF NOT EXISTS idx_channel_monitor_daily_rollups_bucket
    ON channel_monitor_daily_rollups (bucket_date);
CREATE INDEX IF NOT EXISTS idx_channel_monitor_daily_rollups_deleted_at
    ON channel_monitor_daily_rollups (deleted_at);

-- 3) 创建 watermark 表（单行：id=1）
CREATE TABLE IF NOT EXISTS channel_monitor_aggregation_watermark (
    id                   INT          PRIMARY KEY DEFAULT 1,
    last_aggregated_date DATE,
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT channel_monitor_aggregation_watermark_singleton CHECK (id = 1)
);

INSERT INTO channel_monitor_aggregation_watermark (id, last_aggregated_date, updated_at)
VALUES (1, NULL, NOW())
ON CONFLICT (id) DO NOTHING;
