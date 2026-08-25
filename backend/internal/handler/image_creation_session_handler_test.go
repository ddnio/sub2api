package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestImageCreationSessionAPIKeyEligibility(t *testing.T) {
	allowed := &service.APIKey{Status: service.StatusActive, Group: &service.Group{AllowImageGeneration: true}}
	blocked := &service.APIKey{Status: service.StatusActive, Group: &service.Group{AllowImageGeneration: false}}

	require.True(t, isEligibleImageCreationAPIKey(allowed))
	require.False(t, isEligibleImageCreationAPIKey(blocked))
}
