-- Backfill payment v2 subscription_plans from legacy payment_plans.
-- Payment v2 subscription orders require plans to bind active subscription
-- groups. Legacy payment_plans allowed any group_id, so skip invalid legacy
-- bindings instead of publishing plans that will fail at checkout.
INSERT INTO subscription_plans (
    group_id,
    name,
    description,
    price,
    original_price,
    validity_days,
    validity_unit,
    features,
    product_name,
    for_sale,
    sort_order,
    created_at,
    updated_at
)
SELECT
    pp.group_id,
    pp.name,
    pp.description,
    pp.price,
    pp.original_price,
    pp.duration_days,
    'day',
    COALESCE(pp.badge, ''),
    pp.name,
    pp.is_active,
    pp.sort_order,
    pp.created_at,
    pp.updated_at
FROM payment_plans pp
JOIN groups g ON g.id = pp.group_id
WHERE pp.deleted_at IS NULL
  AND g.status = 'active'
  AND g.subscription_type = 'subscription'
  AND NOT EXISTS (
      SELECT 1
      FROM subscription_plans sp
      WHERE sp.group_id = pp.group_id
        AND sp.name = pp.name
  );
