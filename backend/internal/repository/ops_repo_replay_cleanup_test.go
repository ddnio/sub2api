package repository

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestOpsErrorLogInsertPersistsRequestSnapshotFields(t *testing.T) {
	requiredColumns := []string{
		"request_body",
		"request_headers",
		"request_body_truncated",
		"request_body_bytes",
	}

	insertSQL := strings.ToLower(insertOpsErrorLogSQL)
	for _, column := range requiredColumns {
		if !strings.Contains(insertSQL, column) {
			t.Fatalf("ops error log insert does not reference restored request snapshot column %q", column)
		}
	}

	inputType := reflect.TypeOf(service.OpsInsertErrorLogInput{})
	requiredFields := []string{
		"RequestBodyJSON",
		"RequestBodyTruncated",
		"RequestBodyBytes",
		"RequestHeadersJSON",
	}
	for _, field := range requiredFields {
		if _, ok := inputType.FieldByName(field); !ok {
			t.Fatalf("OpsInsertErrorLogInput does not carry restored request snapshot field %q", field)
		}
	}
}

func TestOpsErrorLogInsertDoesNotRestoreRetryReplayExecutionFields(t *testing.T) {
	disallowedColumns := []string{
		"is_retryable",
		"retry_count",
		"resolved_retry_id",
	}

	insertSQL := strings.ToLower(insertOpsErrorLogSQL)
	for _, column := range disallowedColumns {
		if strings.Contains(insertSQL, column) {
			t.Fatalf("ops error log insert still references dropped replay column %q", column)
		}
	}

	inputType := reflect.TypeOf(service.OpsInsertErrorLogInput{})
	disallowedFields := []string{
		"IsRetryable",
		"RetryCount",
		"ResolvedRetryID",
	}
	for _, field := range disallowedFields {
		if _, ok := inputType.FieldByName(field); ok {
			t.Fatalf("OpsInsertErrorLogInput still carries replay field %q", field)
		}
	}
}
