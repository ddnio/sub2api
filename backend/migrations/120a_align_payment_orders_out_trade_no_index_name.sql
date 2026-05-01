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
          AND idx.relname = 'paymentorder_out_trade_no_unique'
          AND i.indisvalid
          AND i.indisready
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE schemaname = 'public'
              AND tablename = 'payment_orders'
              AND indexname = 'paymentorder_out_trade_no'
        ) THEN
            EXECUTE 'DROP INDEX IF EXISTS paymentorder_out_trade_no';
        END IF;

        EXECUTE 'ALTER INDEX paymentorder_out_trade_no_unique RENAME TO paymentorder_out_trade_no';
    ELSIF EXISTS (
        SELECT 1
        FROM pg_class idx
        JOIN pg_index i ON i.indexrelid = idx.oid
        JOIN pg_class tbl ON tbl.oid = i.indrelid
        JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
        WHERE ns.nspname = 'public'
          AND tbl.relname = 'payment_orders'
          AND idx.relname = 'paymentorder_out_trade_no_unique'
          AND (NOT i.indisvalid OR NOT i.indisready)
    ) THEN
        RAISE EXCEPTION 'invalid paymentorder_out_trade_no_unique index exists; drop it and rerun migration 120 before 120a';
    END IF;
END $$;
