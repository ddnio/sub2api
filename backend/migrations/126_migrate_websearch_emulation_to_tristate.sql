-- Convert legacy boolean web_search_emulation values to tri-state mode.
-- true becomes "enabled"; false is removed so the account follows channel default.
UPDATE accounts
SET extra = (extra - 'web_search_emulation') || jsonb_build_object('web_search_emulation', 'enabled')
WHERE extra ? 'web_search_emulation'
  AND extra->>'web_search_emulation' = 'true';

UPDATE accounts
SET extra = extra - 'web_search_emulation'
WHERE extra ? 'web_search_emulation'
  AND extra->>'web_search_emulation' = 'false';
