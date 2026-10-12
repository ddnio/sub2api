package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareOpsRequestBodyForQueueEmptyBody(t *testing.T) {
	var raw []byte

	requestBodyJSON, truncated, requestBodyBytes := PrepareOpsRequestBodyForQueue(raw)

	require.Nil(t, requestBodyJSON)
	require.False(t, truncated)
	require.Nil(t, requestBodyBytes)
}

func TestPrepareOpsRequestBodyForQueueInvalidJSONRecordsBytesOnly(t *testing.T) {
	raw := []byte("{invalid-json")

	requestBodyJSON, truncated, requestBodyBytes := PrepareOpsRequestBodyForQueue(raw)

	require.Nil(t, requestBodyJSON)
	require.False(t, truncated)
	require.NotNil(t, requestBodyBytes)
	require.Equal(t, len(raw), *requestBodyBytes)
}

func TestPrepareOpsRequestBodyForQueueRedactsSensitiveFields(t *testing.T) {
	raw := []byte(`{
		"model":"claude-3-5-sonnet-20241022",
		"api_key":"sk-test-123",
		"headers":{"authorization":"Bearer secret-token"},
		"messages":[{"role":"user","content":"hello"}]
	}`)

	requestBodyJSON, truncated, requestBodyBytes := PrepareOpsRequestBodyForQueue(raw)

	require.NotNil(t, requestBodyJSON)
	require.NotNil(t, requestBodyBytes)
	require.False(t, truncated)
	require.Equal(t, len(raw), *requestBodyBytes)

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(*requestBodyJSON), &body))
	require.Equal(t, "[REDACTED]", body["api_key"])
	headers, ok := body["headers"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "[REDACTED]", headers["authorization"])
}

func TestPrepareOpsRequestBodyForQueueLargeBodyTruncated(t *testing.T) {
	largeMsg := strings.Repeat("x", opsMaxStoredRequestBodyBytes*2)
	raw := []byte(`{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":"` + largeMsg + `"}]}`)

	requestBodyJSON, truncated, requestBodyBytes := PrepareOpsRequestBodyForQueue(raw)

	require.NotNil(t, requestBodyJSON)
	require.NotNil(t, requestBodyBytes)
	require.True(t, truncated)
	require.Equal(t, len(raw), *requestBodyBytes)
	require.LessOrEqual(t, len(*requestBodyJSON), opsMaxStoredRequestBodyBytes)
	require.Contains(t, *requestBodyJSON, "payload_truncated")
}

func snapshotBodyOfSize(size int) []byte {
	prefix := []byte(`{"api_key":"sk-snapshot-sentinel","input":[{"role":"user","content":"`)
	suffix := []byte(`"}]}`)
	raw := bytes.Repeat([]byte("x"), size)
	copy(raw, prefix)
	copy(raw[len(raw)-len(suffix):], suffix)
	return raw
}

func TestPrepareOpsRequestBodyForQueueProcessingBoundary(t *testing.T) {
	for _, size := range []int{opsMaxProcessedRequestBodyBytes - 1, opsMaxProcessedRequestBodyBytes, opsMaxProcessedRequestBodyBytes + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			raw := snapshotBodyOfSize(size)
			original := bytes.Clone(raw)
			require.True(t, json.Valid(raw))
			body, truncated, bodyBytes := PrepareOpsRequestBodyForQueue(raw)
			require.NotNil(t, body)
			require.True(t, truncated)
			require.NotNil(t, bodyBytes)
			require.Equal(t, size, *bodyBytes)
			require.True(t, json.Valid([]byte(*body)))
			require.LessOrEqual(t, len(*body), opsMaxStoredRequestBodyBytes)
			require.NotContains(t, *body, "sk-snapshot-sentinel")
			require.Equal(t, original, raw, "snapshot processing must not modify the business request")
			if size > opsMaxProcessedRequestBodyBytes {
				require.JSONEq(t, `{"payload_truncated":true,"snapshot_omitted_reason":"request_body_too_large"}`, *body)
			} else {
				require.NotContains(t, *body, "snapshot_omitted_reason")
			}
		})
	}
}

func TestPrepareOpsRequestBodyForQueueOversizedInvalidJSONDoesNotParse(t *testing.T) {
	raw := bytes.Repeat([]byte("!"), opsMaxProcessedRequestBodyBytes+1)
	body, truncated, bodyBytes := PrepareOpsRequestBodyForQueue(raw)
	require.NotNil(t, body)
	require.True(t, truncated)
	require.Equal(t, len(raw), *bodyBytes)
	require.JSONEq(t, `{"payload_truncated":true,"snapshot_omitted_reason":"request_body_too_large"}`, *body)
}

var snapshotBenchmarkBody *string
var snapshotBenchmarkBytes *int

func BenchmarkPrepareOpsRequestBodyForQueueOversized(b *testing.B) {
	for _, sizeMiB := range []int{59, 85} {
		b.Run(fmt.Sprintf("%dMiB", sizeMiB), func(b *testing.B) {
			raw := snapshotBodyOfSize(sizeMiB * 1024 * 1024)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				snapshotBenchmarkBody, _, snapshotBenchmarkBytes = PrepareOpsRequestBodyForQueue(raw)
			}
		})
	}
}
