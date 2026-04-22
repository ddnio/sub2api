package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount_IsOpenAIUpstream(t *testing.T) {
	assert.True(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeUpstream}).IsOpenAIUpstream())
	assert.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}).IsOpenAIUpstream())
	assert.False(t, (&Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}).IsOpenAIUpstream())
	assert.False(t, (&Account{Platform: PlatformAnthropic, Type: AccountTypeUpstream}).IsOpenAIUpstream())
}

func TestAccount_GetOpenAIBaseURL_Upstream(t *testing.T) {
	t.Run("upstream with base_url returns it", func(t *testing.T) {
		a := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeUpstream,
			Credentials: map[string]any{
				"base_url": "https://custom.example.com/v1",
			},
		}
		assert.Equal(t, "https://custom.example.com/v1", a.GetOpenAIBaseURL())
	})

	t.Run("upstream without base_url falls back to default", func(t *testing.T) {
		a := &Account{
			Platform:    PlatformOpenAI,
			Type:        AccountTypeUpstream,
			Credentials: map[string]any{},
		}
		assert.Equal(t, "https://api.openai.com", a.GetOpenAIBaseURL())
	})

	t.Run("apikey with base_url still works", func(t *testing.T) {
		a := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"base_url": "https://other.example.com",
			},
		}
		assert.Equal(t, "https://other.example.com", a.GetOpenAIBaseURL())
	})

	t.Run("non-openai returns empty", func(t *testing.T) {
		a := &Account{Platform: PlatformAnthropic, Type: AccountTypeUpstream}
		assert.Equal(t, "", a.GetOpenAIBaseURL())
	})
}

func TestAccount_ShouldUseDirectChatCompletionsUpstream(t *testing.T) {
	mkApikey := func(passthrough bool, baseURL string) *Account {
		extra := map[string]any{}
		if passthrough {
			extra["openai_passthrough"] = true
		}
		creds := map[string]any{"api_key": "sk-test"}
		if baseURL != "" {
			creds["base_url"] = baseURL
		}
		return &Account{
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Credentials: creds,
			Extra:       extra,
		}
	}

	cases := []struct {
		name    string
		account *Account
		want    bool
	}{
		{
			name:    "nil account",
			account: nil,
			want:    false,
		},
		{
			name:    "type=upstream always true (openai)",
			account: &Account{Platform: PlatformOpenAI, Type: AccountTypeUpstream, Credentials: map[string]any{"base_url": "https://api.kimi.com/coding/v1"}},
			want:    true,
		},
		{
			name:    "apikey + passthrough + kimi base_url → true",
			account: mkApikey(true, "https://api.kimi.com/coding/v1"),
			want:    true,
		},
		{
			name:    "apikey + passthrough + deepseek base_url → true",
			account: mkApikey(true, "https://api.deepseek.com/v1"),
			want:    true,
		},
		{
			name:    "apikey + passthrough + azure openai → true (Azure not treated as official)",
			account: mkApikey(true, "https://myresource.openai.azure.com/openai"),
			want:    true,
		},
		{
			name:    "apikey + passthrough + official api.openai.com → true",
			account: mkApikey(true, "https://api.openai.com/v1"),
			want:    true,
		},
		{
			name:    "apikey + passthrough + platform.openai.com → true",
			account: mkApikey(true, "https://platform.openai.com"),
			want:    true,
		},
		{
			name:    "apikey + passthrough + empty base_url → true",
			account: mkApikey(true, ""),
			want:    true,
		},
		{
			name:    "apikey + passthrough=false + custom base_url → false",
			account: mkApikey(false, "https://api.kimi.com/coding/v1"),
			want:    false,
		},
		{
			name: "oauth (codex) + passthrough + custom base_url → false",
			account: &Account{
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				Credentials: map[string]any{"base_url": "https://api.kimi.com/coding/v1"},
				Extra:       map[string]any{"openai_passthrough": true},
			},
			want: false,
		},
		{
			name: "anthropic upstream account → false (only openai covered here)",
			account: &Account{
				Platform:    PlatformAnthropic,
				Type:        AccountTypeUpstream,
				Credentials: map[string]any{"base_url": "https://api.kimi.com/coding/v1"},
			},
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, c.account.ShouldUseDirectChatCompletionsUpstream())
		})
	}
}

func TestIsOpenAIOfficialBaseURL(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://api.openai.com", true},
		{"https://api.openai.com/v1", true},
		{"http://api.openai.com:443", true},
		{"https://platform.openai.com", true},
		{"https://platform.openai.com/foo", true},
		{"https://openai.com", true},
		{"https://myresource.openai.azure.com/openai", false},
		{"https://api.kimi.com/coding/v1", false},
		{"https://api.deepseek.com/v1", false},
		{"https://api.OPENAI.com/v1", true},
		{"", false},
		{"not-a-url", false},
		{"https://api.openai.com.evil.example", false},
		// FQDN trailing dot variants — must still be detected as official
		{"https://api.openai.com./v1", true},
		{"https://platform.openai.com.", true},
		{"https://openai.com.", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			assert.Equal(t, c.want, isOpenAIOfficialBaseURL(c.in))
		})
	}
}
