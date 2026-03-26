ALTER TABLE payment_orders ADD COLUMN refund_no VARCHAR(64) DEFAULT NULL;
COMMENT ON COLUMN payment_orders.refund_no IS '商户退款单号';
