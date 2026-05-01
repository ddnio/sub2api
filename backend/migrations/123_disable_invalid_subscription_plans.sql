UPDATE subscription_plans sp
SET for_sale = false,
    updated_at = NOW()
WHERE sp.for_sale = true
  AND NOT EXISTS (
      SELECT 1
      FROM groups g
      WHERE g.id = sp.group_id
        AND g.status = 'active'
        AND g.subscription_type = 'subscription'
  );
