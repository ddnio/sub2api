package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubUpstream is a simple HTTPUpstream that returns a fixed response.
type stubUpstream struct {
	resp *http.Response
	err  error
}

func (s *stubUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return s.resp, s.err
}

func (s *stubUpstream) DoWithTLS(_ *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.resp, s.err
}

// newTestOpenAIService creates a minimal OpenAIGatewayService for unit tests.
// cfg has URLAllowlist disabled + HTTP allowed so validateUpstreamBaseURL passes with http://...
func newTestOpenAIService(upstream HTTPUpstream) *OpenAIGatewayService {
	return &OpenAIGatewayService{
		httpUpstream: upstream,
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{
					Enabled:           false,
					AllowInsecureHTTP: true,
				},
			},
		},
	}
}

// --- Helper function unit tests ---

func TestBuildOpenAIChatCompletionsURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://api.example.com", "https://api.example.com/v1/chat/completions"},
		{"https://api.example.com/v1", "https://api.example.com/v1/chat/completions"},
		{"https://api.example.com/v1/", "https://api.example.com/v1/chat/completions"},
		{"https://api.example.com/v1/chat/completions", "https://api.example.com/v1/chat/completions"},
		{"https://api.example.com/v1/responses", "https://api.example.com/v1/chat/completions"},
	}
	for _, tc := range cases {
		got := buildOpenAIChatCompletionsURL(tc.in)
		assert.Equal(t, tc.want, got, "input: %s", tc.in)
	}
}

func TestEnsureStreamIncludeUsage(t *testing.T) {
	t.Run("stream false — no-op", func(t *testing.T) {
		body := []byte(`{"model":"gpt-4","stream":false}`)
		got := ensureStreamIncludeUsage(body)
		assert.Equal(t, string(body), string(got))
	})
	t.Run("stream true, no stream_options — injects true", func(t *testing.T) {
		body := []byte(`{"model":"gpt-4","stream":true}`)
		got := ensureStreamIncludeUsage(body)
		assert.Contains(t, string(got), `"include_usage":true`)
	})
	t.Run("stream true, include_usage already true — no duplicate", func(t *testing.T) {
		body := []byte(`{"model":"gpt-4","stream":true,"stream_options":{"include_usage":true}}`)
		got := ensureStreamIncludeUsage(body)
		assert.Equal(t, string(body), string(got))
	})
	t.Run("stream true, include_usage false — overrides to true", func(t *testing.T) {
		body := []byte(`{"model":"gpt-4","stream":true,"stream_options":{"include_usage":false}}`)
		got := ensureStreamIncludeUsage(body)
		assert.Contains(t, string(got), `"include_usage":true`)
	})
}

func TestShouldFailoverChatCompletionsUpstream(t *testing.T) {
	failover := []int{429, 529}
	noFailover := []int{400, 401, 402, 403, 408, 500, 502, 503, 504, 524}
	for _, code := range failover {
		assert.True(t, shouldFailoverChatCompletionsUpstream(code), "code %d should failover", code)
	}
	for _, code := range noFailover {
		assert.False(t, shouldFailoverChatCompletionsUpstream(code), "code %d should NOT failover", code)
	}
}

func TestParseChatCompletionsUsage_NoPredecrement(t *testing.T) {
	data := []byte(`{
		"usage": {
			"prompt_tokens": 100,
			"completion_tokens": 50,
			"prompt_tokens_details": {"cached_tokens": 30}
		}
	}`)
	u := parseChatCompletionsUsage(data)
	assert.Equal(t, 100, u.InputTokens, "InputTokens must equal prompt_tokens (no pre-decrement)")
	assert.Equal(t, 50, u.OutputTokens)
	assert.Equal(t, 30, u.CacheReadInputTokens)
}

// --- ForwardChatCompletionsUpstream integration tests ---

func makeTestGinContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	return c, w
}

func makeTestAccount(apiKey, baseURL string) *Account {
	return &Account{
		ID:       1,
		Platform: PlatformOpenAI,
		Type:     AccountTypeUpstream,
		Credentials: map[string]any{
			"api_key":  apiKey,
			"base_url": baseURL,
		},
	}
}

func TestForwardChatCompletionsUpstream_NonStream(t *testing.T) {
	responseBody := `{"id":"chatcmpl-1","object":"chat.completion","model":"gpt-4","choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, responseBody)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, w := makeTestGinContext(t)
	account := makeTestAccount("test-key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`)

	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "gpt-4", result.Model)
	assert.Equal(t, 10, result.Usage.InputTokens)
	assert.Equal(t, 5, result.Usage.OutputTokens)
	assert.Equal(t, responseBody, w.Body.String())
}

func TestForwardChatCompletionsUpstream_401Passthrough(t *testing.T) {
	errorBody := `{"error":{"type":"invalid_api_key","message":"Invalid API key"}}`

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, errorBody)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, w := makeTestGinContext(t)
	account := makeTestAccount("bad-key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.Error(t, err, "must return non-nil error for 401 passthrough")
	assert.Nil(t, result)
	assert.True(t, w.Code == http.StatusUnauthorized, "status 401 must be written to client")
	assert.True(t, w.Body.Len() > 0, "error body must be written (Written()==true)")

	var failoverErr *UpstreamFailoverError
	assert.False(t, isUpstreamFailoverError(err, &failoverErr), "401 must NOT be failover")
}

func TestForwardChatCompletionsUpstream_429Failover(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, `{"error":{"message":"rate limit"}}`)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, w := makeTestGinContext(t)
	account := makeTestAccount("key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.Error(t, err)
	assert.Nil(t, result)

	var failoverErr *UpstreamFailoverError
	require.True(t, isUpstreamFailoverError(err, &failoverErr), "429 must be UpstreamFailoverError")
	assert.Equal(t, http.StatusTooManyRequests, failoverErr.StatusCode)
	assert.Equal(t, 0, w.Body.Len(), "429 failover must NOT write to client (Written()==false)")
}

func TestForwardChatCompletionsUpstream_500Passthrough(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"error":{"message":"internal error"}}`)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, w := makeTestGinContext(t)
	account := makeTestAccount("key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","messages":[]}`)

	_, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var failoverErr *UpstreamFailoverError
	assert.False(t, isUpstreamFailoverError(err, &failoverErr), "500 must NOT failover")
}

func TestForwardChatCompletionsUpstream_Stream(t *testing.T) {
	sseBody := "data: {\"id\":\"1\",\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\ndata: {\"usage\":{\"prompt_tokens\":5,\"completion_tokens\":3}}\n\ndata: [DONE]\n\n"

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, sseBody)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, w := makeTestGinContext(t)
	account := makeTestAccount("key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","stream":true,"messages":[]}`)

	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Stream)
	assert.Equal(t, 5, result.Usage.InputTokens)
	assert.Equal(t, 3, result.Usage.OutputTokens)
	assert.Contains(t, w.Body.String(), "[DONE]")
}

func TestForwardChatCompletionsUpstream_ResultModelIsOriginalModel(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"id":"1","object":"chat.completion","model":"gpt-4-mapped","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})
	c, _ := makeTestGinContext(t)

	account := &Account{
		ID:       1,
		Platform: PlatformOpenAI,
		Type:     AccountTypeUpstream,
		Credentials: map[string]any{
			"api_key":       "key",
			"base_url":      fakeServer.URL,
			"model_mapping": map[string]any{"gpt-4": "gpt-4-mapped"},
		},
	}

	body := []byte(`{"model":"gpt-4","messages":[]}`)
	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "gpt-4", result.Model, "Model must be original client-requested model")
	assert.Equal(t, "gpt-4-mapped", result.UpstreamModel, "UpstreamModel must be mapped model")
}

// directHTTPUpstream performs real HTTP using http.DefaultTransport (for httptest servers).
type directHTTPUpstream struct{}

func (d *directHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(req)
}

func (d *directHTTPUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(req)
}

// isUpstreamFailoverError checks if err wraps *UpstreamFailoverError.
func isUpstreamFailoverError(err error, target **UpstreamFailoverError) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	_ = s
	// Use type assertion on err directly (UpstreamFailoverError implements error)
	if fe, ok := err.(*UpstreamFailoverError); ok {
		*target = fe
		return true
	}
	return false
}

// --- convertChatCompletionsSSEToJSON ---

func TestConvertChatCompletionsSSEToJSON(t *testing.T) {
	sse := strings.Join([]string{
		"data: {\"id\":\"c1\",\"model\":\"gpt-4\",\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}",
		"data: {\"choices\":[{\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}",
		"data: {\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":5}}",
		"data: [DONE]",
		"",
	}, "\n")

	out, usage, err := convertChatCompletionsSSEToJSON([]byte(sse), "gpt-4")
	require.NoError(t, err)
	assert.Contains(t, string(out), "Hello world")
	assert.Contains(t, string(out), `"finish_reason":"stop"`)
	assert.Equal(t, 10, usage.InputTokens)
	assert.Equal(t, 5, usage.OutputTokens)
}

func TestConvertChatCompletionsSSEToJSON_MultiChoice(t *testing.T) {
	sse := strings.Join([]string{
		`data: {"id":"c1","model":"gpt-4","choices":[{"index":0,"delta":{"content":"A"}},{"index":1,"delta":{"content":"B"}}]}`,
		`data: {"choices":[{"index":0,"delta":{"content":"1"},"finish_reason":"stop"},{"index":1,"delta":{"content":"2"},"finish_reason":"stop"}]}`,
		`data: [DONE]`,
		``,
	}, "\n")

	out, _, err := convertChatCompletionsSSEToJSON([]byte(sse), "gpt-4")
	require.NoError(t, err)

	outStr := string(out)
	assert.Contains(t, outStr, "A1", "choice 0 content must be aggregated")
	assert.Contains(t, outStr, "B2", "choice 1 content must be aggregated")
}

func TestConvertChatCompletionsSSEToJSON_ToolCalls(t *testing.T) {
	// Two chunks: first carries id/type/name, second carries arguments fragment
	sse := strings.Join([]string{
		`data: {"id":"c1","model":"gpt-4","choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_abc","type":"function","function":{"name":"get_weather","arguments":""}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"location\":\"Beijing\"}"}}]},"finish_reason":"tool_calls"}]}`,
		`data: {"usage":{"prompt_tokens":20,"completion_tokens":10}}`,
		`data: [DONE]`,
		``,
	}, "\n")

	out, usage, err := convertChatCompletionsSSEToJSON([]byte(sse), "gpt-4")
	require.NoError(t, err)

	outStr := string(out)
	assert.Contains(t, outStr, `"tool_calls"`)
	assert.Contains(t, outStr, `"get_weather"`)
	assert.Contains(t, outStr, "Beijing") // nested in JSON-escaped arguments string
	assert.Contains(t, outStr, `"finish_reason":"tool_calls"`)
	// content must be null when tool_calls present and no text content
	assert.Contains(t, outStr, `"content":null`)
	assert.Equal(t, 20, usage.InputTokens)
	assert.Equal(t, 10, usage.OutputTokens)
}

// --- ClientDisconnectDrains ---

// failOnWriteResponseRecorder is an http.ResponseWriter whose Write always fails,
// simulating a disconnected client.
type failOnWriteResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (f *failOnWriteResponseRecorder) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("client disconnected")
}

func TestForwardChatCompletionsUpstream_ClientDisconnectDrains(t *testing.T) {
	sseBody := strings.Join([]string{
		`data: {"id":"1","choices":[{"delta":{"content":"hi"}}]}`,
		`data: {"usage":{"prompt_tokens":5,"completion_tokens":3}}`,
		`data: [DONE]`,
		``,
	}, "\n\n")

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, sseBody)
	}))
	defer fakeServer.Close()

	svc := newTestOpenAIService(&directHTTPUpstream{})

	gin.SetMode(gin.TestMode)
	failWriter := &failOnWriteResponseRecorder{httptest.NewRecorder()}
	c, _ := gin.CreateTestContext(failWriter)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	account := makeTestAccount("key", fakeServer.URL)
	body := []byte(`{"model":"gpt-4","stream":true,"messages":[]}`)

	result, err := svc.ForwardChatCompletionsUpstream(c.Request.Context(), c, account, body, "")
	require.NoError(t, err, "stream drain must succeed even when client write fails")
	require.NotNil(t, result)
	assert.True(t, result.Stream)
	// Usage is still extracted even during drain
	assert.Equal(t, 5, result.Usage.InputTokens)
	assert.Equal(t, 3, result.Usage.OutputTokens)
}
