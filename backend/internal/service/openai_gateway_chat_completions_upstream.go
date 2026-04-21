package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

// ForwardChatCompletionsUpstream forwards chat completions to an upstream OpenAI-compatible
// service (AccountTypeUpstream). Passes body through directly without Responses API conversion.
func (s *OpenAIGatewayService) ForwardChatCompletionsUpstream(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	defaultMappedModel string,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	// 1. 凭证 (D1)
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	baseURL := strings.TrimSpace(account.GetOpenAIBaseURL())
	if apiKey == "" {
		return nil, errors.New("upstream openai account missing api_key")
	}
	if baseURL == "" {
		return nil, errors.New("upstream openai account missing base_url")
	}

	// 2. URL 校验 + 构建
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream base_url: %w", err)
	}
	upstreamURL := buildOpenAIChatCompletionsURL(validatedURL)

	// 3. 请求元信息
	reqModel := gjson.GetBytes(body, "model").String()
	reqStream := gjson.GetBytes(body, "stream").Bool()
	originalModel := reqModel

	// 4. 模型映射
	billingModel := resolveOpenAIForwardModel(account, originalModel, defaultMappedModel)
	upstreamModel := billingModel
	if upstreamModel != originalModel {
		if newBody, perr := sjson.SetBytes(body, "model", upstreamModel); perr == nil {
			body = newBody
		}
	}

	// 5. stream_options.include_usage 注入
	if reqStream {
		body = ensureStreamIncludeUsage(body)
	}

	setOpsUpstreamRequestBody(c, body)

	// 6. context detach for streaming (D3)
	upstreamCtx, releaseUpstreamCtx := detachStreamUpstreamContext(ctx, reqStream)
	defer releaseUpstreamCtx()

	// 7. 构建请求
	req, rerr := http.NewRequestWithContext(upstreamCtx, http.MethodPost, upstreamURL, bytes.NewReader(body))
	if rerr != nil {
		return nil, fmt.Errorf("create upstream request: %w", rerr)
	}
	req.Header.Set("Content-Type", "application/json")

	allowTimeoutHeaders := s.isOpenAIPassthroughTimeoutHeadersAllowed()
	if c != nil && c.Request != nil {
		for key, values := range c.Request.Header {
			lower := strings.ToLower(strings.TrimSpace(key))
			if !isOpenAIPassthroughAllowedRequestHeader(lower, allowTimeoutHeaders) {
				continue
			}
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}
	req.Header.Del("authorization")
	req.Header.Del("x-api-key")
	req.Header.Del("x-goog-api-key")
	req.Header.Set("authorization", "Bearer "+apiKey)

	if ua := account.GetOpenAIUserAgent(); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	// 8. 发送
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	resp, httpErr := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if httpErr != nil {
		safeErr := sanitizeUpstreamErrorMessage(httpErr.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
			Kind: "request_error", Message: safeErr,
		})
		return nil, fmt.Errorf("upstream chat completions request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	// 9. 错误处理
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		upstreamMsg := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(respBody))
		if s.rateLimitService != nil {
			s.rateLimitService.HandleUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody)
		}

		// 9a. Failover (D4: 只 429/529)
		if shouldFailoverChatCompletionsUpstream(resp.StatusCode) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
				UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
				Kind: "failover", Message: upstreamMsg,
			})
			return nil, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: false,
			}
		}

		// 9b. 非 failover：透传后返回 non-nil error (D7)
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
			UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
			Kind: "http_error", Message: upstreamMsg,
		})
		writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
		c.Status(resp.StatusCode)
		_, _ = c.Writer.Write(respBody)
		return nil, fmt.Errorf("upstream chat completions error: status=%d msg=%s", resp.StatusCode, upstreamMsg)
	}

	// 10. 成功响应
	var usage OpenAIUsage
	var firstTokenMs *int

	if reqStream {
		writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
		c.Status(resp.StatusCode)
		var streamErr error
		usage, firstTokenMs, streamErr = s.streamChatCompletionsUpstream(resp, c, account, startTime)
		if streamErr != nil {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
				UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
				Kind: "stream_error", Message: sanitizeUpstreamErrorMessage(streamErr.Error()),
			})
			return nil, streamErr
		}
	} else {
		respBody, parsedUsage, readErr := s.bufferedChatCompletionsUpstream(resp, account)
		if readErr != nil {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
				UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
				Kind: "http_error", Message: sanitizeUpstreamErrorMessage(readErr.Error()),
			})
			return nil, readErr
		}

		if isEventStreamResponse(resp.Header) {
			logger.L().With(
				zap.String("component", "service.openai_gateway"),
				zap.Int64("account_id", account.ID),
				zap.String("base_url", baseURL),
			).Info("upstream_chat_completions.sse_to_json_conversion")
			convertedBody, convertedUsage, convErr := convertChatCompletionsSSEToJSON(respBody, upstreamModel)
			if convErr != nil {
				appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
					Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
					UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
					Kind: "http_error", Message: "sse_to_json_convert_failed: " + convErr.Error(),
				})
				return nil, fmt.Errorf("convert upstream SSE to JSON: %w", convErr)
			}
			writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Status(resp.StatusCode)
			_, _ = c.Writer.Write(convertedBody)
			usage = convertedUsage
		} else {
			writeOpenAIPassthroughResponseHeaders(c.Writer.Header(), resp.Header, s.responseHeaderFilter)
			c.Status(resp.StatusCode)
			_, _ = c.Writer.Write(respBody)
			usage = parsedUsage
		}
	}

	// 11. 零 usage 监控 (D5: 仅日志)
	if usage.InputTokens == 0 && usage.OutputTokens == 0 {
		logger.L().With(
			zap.String("component", "service.openai_gateway"),
			zap.Int64("account_id", account.ID),
			zap.String("base_url", baseURL),
			zap.String("model", upstreamModel),
			zap.Bool("stream", reqStream),
		).Warn("upstream_chat_completions.zero_usage")
	}

	// v8 补齐: ServiceTier / ReasoningEffort 透传
	serviceTier := extractOpenAIServiceTierFromBody(body)
	reasoningEffort := extractOpenAIReasoningEffortFromBody(body, originalModel)

	return &OpenAIForwardResult{
		RequestID:       resp.Header.Get("x-request-id"),
		Usage:           usage,
		Model:           originalModel,
		BillingModel:    billingModel,
		UpstreamModel:   upstreamModel,
		ServiceTier:     serviceTier,
		ReasoningEffort: reasoningEffort,
		Stream:          reqStream,
		Duration:        time.Since(startTime),
		FirstTokenMs:    firstTokenMs,
	}, nil
}

// buildOpenAIChatCompletionsURL normalizes a base URL to the chat completions endpoint.
func buildOpenAIChatCompletionsURL(base string) string {
	normalized := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(normalized, "/chat/completions") {
		return normalized
	}
	if strings.HasSuffix(normalized, "/responses") {
		normalized = strings.TrimSuffix(normalized, "/responses")
	}
	if strings.HasSuffix(normalized, "/v1") {
		return normalized + "/chat/completions"
	}
	return normalized + "/v1/chat/completions"
}

// ensureStreamIncludeUsage injects stream_options.include_usage=true when stream=true.
// Forces override even if user explicitly set false to prevent billing gaps.
func ensureStreamIncludeUsage(body []byte) []byte {
	if !gjson.GetBytes(body, "stream").Bool() {
		return body
	}
	if gjson.GetBytes(body, "stream_options.include_usage").Bool() {
		return body
	}
	out, err := sjson.SetBytes(body, "stream_options.include_usage", true)
	if err != nil {
		return body
	}
	return out
}

// shouldFailoverChatCompletionsUpstream returns true only for rate-limit codes (D4).
func shouldFailoverChatCompletionsUpstream(statusCode int) bool {
	switch statusCode {
	case 429, 529:
		return true
	}
	return false
}

// parseChatCompletionsUsage extracts usage from a chat completions JSON response.
// Returns raw prompt_tokens without pre-decrementing cache; RecordUsage handles the subtraction.
func parseChatCompletionsUsage(data []byte) OpenAIUsage {
	return OpenAIUsage{
		InputTokens:          int(gjson.GetBytes(data, "usage.prompt_tokens").Int()),
		OutputTokens:         int(gjson.GetBytes(data, "usage.completion_tokens").Int()),
		CacheReadInputTokens: int(gjson.GetBytes(data, "usage.prompt_tokens_details.cached_tokens").Int()),
	}
}

// streamChatCompletionsUpstream pipes SSE from upstream to client, accumulates usage,
// and returns an error if the stream terminates without [DONE] or hits a scanner error.
func (s *OpenAIGatewayService) streamChatCompletionsUpstream(
	resp *http.Response, c *gin.Context, account *Account, startTime time.Time,
) (OpenAIUsage, *int, error) {
	flusher, _ := c.Writer.(http.Flusher)

	scannerBuf := getSSEScannerBuf64K()
	defer putSSEScannerBuf64K(scannerBuf)

	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer((*scannerBuf)[:0], maxLineSize)

	var usage OpenAIUsage
	var firstTokenMs *int
	clientDisconnected := false
	sawDone := false

	for scanner.Scan() {
		line := scanner.Text()

		if !clientDisconnected {
			if _, werr := fmt.Fprintln(c.Writer, line); werr != nil {
				clientDisconnected = true
				logger.L().With(
					zap.String("component", "service.openai_gateway"),
					zap.Int64("account_id", account.ID),
				).Info("upstream_chat_completions.client_disconnected_drain_continuing")
			} else if flusher != nil {
				flusher.Flush()
			}
		}

		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimSpace(line[6:])
			if payload == "" {
				continue
			}
			if payload == "[DONE]" {
				sawDone = true
				continue
			}
			if firstTokenMs == nil {
				ms := int(time.Since(startTime).Milliseconds())
				firstTokenMs = &ms
			}
			u := parseChatCompletionsUsage([]byte(payload))
			if u.InputTokens > 0 || u.OutputTokens > 0 {
				usage = u
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.L().With(
			zap.String("component", "service.openai_gateway"),
			zap.Int64("account_id", account.ID),
			zap.Error(err),
		).Warn("upstream_chat_completions.stream_scanner_error")
		if clientDisconnected {
			return usage, firstTokenMs, fmt.Errorf("stream drain aborted after client disconnect: %w", err)
		}
		return usage, firstTokenMs, fmt.Errorf("stream read failed: %w", err)
	}
	if !sawDone {
		logger.L().With(
			zap.String("component", "service.openai_gateway"),
			zap.Int64("account_id", account.ID),
			zap.Bool("client_disconnected", clientDisconnected),
		).Warn("upstream_chat_completions.stream_missing_done")
		return usage, firstTokenMs, errors.New("upstream stream ended without [DONE] terminal event")
	}

	return usage, firstTokenMs, nil
}

// toolCallAcc accumulates streaming tool_call deltas by index.
type toolCallAcc struct {
	id          string
	typ         string
	name        string
	argsBuilder strings.Builder
}

// convertChatCompletionsSSEToJSON aggregates a chat completions SSE stream into a single JSON response.
// Used when upstream returns text/event-stream despite stream=false (cascaded sub2api scenario).
func convertChatCompletionsSSEToJSON(body []byte, upstreamModel string) ([]byte, OpenAIUsage, error) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), defaultMaxLineSize)

	var id, model, finishReason string
	var contentBuilder strings.Builder
	var toolCallsMap map[int]*toolCallAcc
	var usage OpenAIUsage
	var lastPayload []byte

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(line[6:])
		if payload == "" || payload == "[DONE]" {
			continue
		}

		pb := []byte(payload)
		if id == "" {
			id = gjson.GetBytes(pb, "id").String()
		}
		if model == "" {
			model = gjson.GetBytes(pb, "model").String()
		}
		if fr := gjson.GetBytes(pb, "choices.0.finish_reason").String(); fr != "" {
			finishReason = fr
		}
		if delta := gjson.GetBytes(pb, "choices.0.delta.content").String(); delta != "" {
			contentBuilder.WriteString(delta)
		}
		gjson.GetBytes(pb, "choices.0.delta.tool_calls").ForEach(func(_, v gjson.Result) bool {
			idx := int(v.Get("index").Int())
			if toolCallsMap == nil {
				toolCallsMap = make(map[int]*toolCallAcc)
			}
			if _, ok := toolCallsMap[idx]; !ok {
				toolCallsMap[idx] = &toolCallAcc{}
			}
			tc := toolCallsMap[idx]
			if s := v.Get("id").String(); s != "" {
				tc.id = s
			}
			if s := v.Get("type").String(); s != "" {
				tc.typ = s
			}
			if s := v.Get("function.name").String(); s != "" {
				tc.name = s
			}
			if s := v.Get("function.arguments").String(); s != "" {
				tc.argsBuilder.WriteString(s)
			}
			return true
		})
		if u := parseChatCompletionsUsage(pb); u.InputTokens > 0 || u.OutputTokens > 0 {
			usage = u
		}
		lastPayload = pb
	}
	if err := scanner.Err(); err != nil {
		return nil, OpenAIUsage{}, fmt.Errorf("scan sse: %w", err)
	}
	if lastPayload == nil {
		return nil, OpenAIUsage{}, errors.New("empty sse stream")
	}

	if model == "" {
		model = upstreamModel
	}

	msg := map[string]any{"role": "assistant"}
	if len(toolCallsMap) > 0 {
		indices := make([]int, 0, len(toolCallsMap))
		for idx := range toolCallsMap {
			indices = append(indices, idx)
		}
		sort.Ints(indices)
		toolCalls := make([]map[string]any, 0, len(indices))
		for _, idx := range indices {
			tc := toolCallsMap[idx]
			toolCalls = append(toolCalls, map[string]any{
				"id":   tc.id,
				"type": tc.typ,
				"function": map[string]any{
					"name":      tc.name,
					"arguments": tc.argsBuilder.String(),
				},
			})
		}
		msg["tool_calls"] = toolCalls
		if contentBuilder.Len() > 0 {
			msg["content"] = contentBuilder.String()
		} else {
			msg["content"] = nil
		}
	} else {
		msg["content"] = contentBuilder.String()
	}

	result := map[string]any{
		"id":     id,
		"object": "chat.completion",
		"model":  model,
		"choices": []map[string]any{{
			"index":         0,
			"message":       msg,
			"finish_reason": finishReason,
		}},
	}
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		usageMap := map[string]any{
			"prompt_tokens":     usage.InputTokens,
			"completion_tokens": usage.OutputTokens,
			"total_tokens":      usage.InputTokens + usage.OutputTokens,
		}
		if usage.CacheReadInputTokens > 0 {
			usageMap["prompt_tokens_details"] = map[string]any{
				"cached_tokens": usage.CacheReadInputTokens,
			}
		}
		result["usage"] = usageMap
	}
	out, err := json.Marshal(result)
	if err != nil {
		return nil, OpenAIUsage{}, err
	}
	return out, usage, nil
}

// bufferedChatCompletionsUpstream reads the non-streaming response body with size protection.
func (s *OpenAIGatewayService) bufferedChatCompletionsUpstream(
	resp *http.Response, account *Account,
) ([]byte, OpenAIUsage, error) {
	limit := resolveUpstreamResponseReadLimit(s.cfg)
	body, err := readUpstreamResponseBodyLimited(resp.Body, limit)
	if err != nil {
		logger.L().With(
			zap.String("component", "service.openai_gateway"),
			zap.Int64("account_id", account.ID),
			zap.Int64("limit", limit),
			zap.Error(err),
		).Warn("upstream_chat_completions.body_read_failed")
		return nil, OpenAIUsage{}, fmt.Errorf("read upstream body: %w", err)
	}
	return body, parseChatCompletionsUsage(body), nil
}
