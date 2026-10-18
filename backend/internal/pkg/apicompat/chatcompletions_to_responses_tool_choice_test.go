package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Chat Completions nests a named choice under "function"; the Responses API
// expects the name at the top level and rejects the nested shape.
func TestChatCompletionsToResponses_ToolChoiceShape(t *testing.T) {
	tests := []struct {
		name   string
		choice string
		want   string
	}{
		{name: "named chat function is flattened", choice: `{"type":"function","function":{"name":"get_weather"}}`, want: `{"type":"function","name":"get_weather"}`},
		{name: "flat responses function is kept", choice: `{"type":"function","name":"get_weather"}`, want: `{"type":"function","name":"get_weather"}`},
		{name: "auto string is kept", choice: `"auto"`, want: `"auto"`},
		{name: "required string is kept", choice: `"required"`, want: `"required"`},
		{name: "none string is kept", choice: `"none"`, want: `"none"`},
		{name: "allowed_tools policy is kept", choice: `{"type":"allowed_tools","mode":"auto","tools":[{"type":"function","name":"get_weather"}]}`, want: `{"type":"allowed_tools","mode":"auto","tools":[{"type":"function","name":"get_weather"}]}`},
		{name: "function without name is kept", choice: `{"type":"function","function":{}}`, want: `{"type":"function","function":{}}`},
		{name: "nested name is trimmed", choice: `{"type":"function","function":{"name":" get_weather "}}`, want: `{"type":"function","name":"get_weather"}`},
		{name: "explicit name wins over nested name", choice: `{"type":"function","name":"get_weather","function":{"name":"other"}}`, want: `{"type":"function","name":"get_weather"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ChatCompletionsRequest{
				Model:    "gpt-4o",
				Messages: []ChatMessage{{Role: "user", Content: json.RawMessage(`"weather?"`)}},
				Tools: []ChatTool{{
					Type:     "function",
					Function: &ChatFunction{Name: "get_weather", Parameters: json.RawMessage(`{"type":"object"}`)},
				}},
				ToolChoice: json.RawMessage(tt.choice),
			}

			resp, err := ChatCompletionsToResponses(req)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(resp.ToolChoice))
		})
	}
}

func TestChatCompletionsToResponses_MixedLegacyAndModernToolHistory(t *testing.T) {
	req := &ChatCompletionsRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call_legacy_1", Type: "function", Function: ChatFunctionCall{Name: "modern", Arguments: `{}`}}}},
			{Role: "tool", ToolCallID: "call_legacy_1", Content: json.RawMessage(`"modern result"`)},
			{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "legacy", Arguments: `{}`}},
			{Role: "function", Name: "legacy", Content: json.RawMessage(`"legacy result"`)},
		},
	}
	out, err := ChatCompletionsToResponses(req)
	require.NoError(t, err)
	var items []ResponsesInputItem
	require.NoError(t, json.Unmarshal(out.Input, &items))
	require.Len(t, items, 4)
	require.Equal(t, "modern", items[1].Name, "modern tool output keeps the local call-name mapping")
	require.Equal(t, items[0].CallID, items[1].CallID)
	require.NotEqual(t, items[0].CallID, items[2].CallID, "legacy IDs cannot collide with modern calls")
	require.Equal(t, items[2].CallID, items[3].CallID)
}
