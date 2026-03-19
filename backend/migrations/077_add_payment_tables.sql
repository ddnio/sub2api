-- 创建支付套餐表
CREATE TABLE IF NOT EXISTS payment_plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    badge VARCHAR(20) DEFAULT NULL,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    duration_days INT NOT NULL CHECK (duration_days > 0),
    price DECIMAL(20,8) NOT NULL DEFAULT 0,
    original_price DECIMAL(20,8) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- 创建支付订单表
CREATE TABLE IF NOT EXISTS payment_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(32) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type VARCHAR(20) NOT NULL,
    plan_id BIGINT DEFAULT NULL REFERENCES payment_plans(id) ON DELETE RESTRICT,
    amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    credit_amount DECIMAL(20,8) DEFAULT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    provider VARCHAR(20) DEFAULT NULL,
    provider_order_no VARCHAR(64) DEFAULT NULL UNIQUE,
    paid_at TIMESTAMPTZ DEFAULT NULL,
    completed_at TIMESTAMPTZ DEFAULT NULL,
    refunded_at TIMESTAMPTZ DEFAULT NULL,
    expired_at TIMESTAMPTZ NOT NULL,
    callback_raw TEXT DEFAULT NULL,
    admin_note TEXT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- payment_plans 索引
CREATE INDEX IF NOT EXISTS idx_payment_plans_group_id ON payment_plans(group_id);
CREATE INDEX IF NOT EXISTS idx_payment_plans_is_active_sort_order ON payment_plans(is_active, sort_order);

-- payment_orders 索引
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_plan_id ON payment_orders(plan_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status_expired_at ON payment_orders(status, expired_at);

COMMENT ON TABLE payment_plans IS '支付套餐';
COMMENT ON TABLE payment_orders IS '支付订单';
COMMENT ON COLUMN payment_orders.type IS '订单类型: plan（订阅套餐）/ topup（余额充值）';
COMMENT ON COLUMN payment_orders.status IS '订单状态: pending / paid / completed / failed / expired / refunded';
COMMENT ON COLUMN payment_orders.provider IS '支付渠道: alipay / wechat';
