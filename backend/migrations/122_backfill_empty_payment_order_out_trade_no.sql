DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'payment_orders'
          AND column_name = 'out_trade_no'
    ) THEN
        RAISE NOTICE '122: payment_orders.out_trade_no not found, skipping';
        RETURN;
    END IF;

    UPDATE payment_orders
    SET out_trade_no = 'sub2_' || id::text
    WHERE out_trade_no = '';
END $$;
