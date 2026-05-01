DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class idx
        JOIN pg_index i ON i.indexrelid = idx.oid
        JOIN pg_class tbl ON tbl.oid = i.indrelid
        JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
        WHERE ns.nspname = 'public'
          AND tbl.relname = 'payment_orders'
          AND idx.relname IN ('paymentorder_out_trade_no', 'paymentorder_out_trade_no_unique')
          AND (NOT i.indisvalid OR NOT i.indisready)
    ) THEN
        RAISE EXCEPTION 'invalid payment order out_trade_no index exists; drop the invalid index and rerun migrations';
    END IF;
END $$;
