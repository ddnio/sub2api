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
