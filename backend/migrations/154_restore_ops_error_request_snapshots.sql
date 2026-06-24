ALTER TABLE ops_error_logs
  ADD COLUMN IF NOT EXISTS request_body JSONB,
  ADD COLUMN IF NOT EXISTS request_headers JSONB,
  ADD COLUMN IF NOT EXISTS request_body_truncated BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS request_body_bytes INT;

COMMENT ON COLUMN ops_error_logs.request_body IS 'Sanitized and capped inbound request body snapshot for error diagnostics only.';
COMMENT ON COLUMN ops_error_logs.request_headers IS 'Whitelisted inbound request headers snapshot for error diagnostics only.';
COMMENT ON COLUMN ops_error_logs.request_body_truncated IS 'Whether request_body was truncated before persistence.';
COMMENT ON COLUMN ops_error_logs.request_body_bytes IS 'Original inbound request body byte length before sanitization/truncation.';
