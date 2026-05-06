package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAllowedMonitorGroupNamesTrimsAndSkipsEmptyNames(t *testing.T) {
	allowed := allowedMonitorGroupNames([]service.Group{
		{Name: " public "},
		{Name: ""},
		{Name: "   "},
		{Name: "private"},
	})

	require.Equal(t, map[string]struct{}{
		"public":  {},
		"private": {},
	}, allowed)
}
