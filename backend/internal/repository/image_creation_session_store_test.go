package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestImageCreationSessionStoreHashesTokensAndConsumesTicketsOnce(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	store := NewImageCreationSessionStore(client)
	claims := service.ImageCreationSessionClaims{UserID: 9, TokenVersion: 7, Scope: service.ImageCreationScopeUser}

	ticket, err := store.StoreTicket(context.Background(), claims, time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, ticket)
	for _, key := range mini.Keys() {
		require.False(t, strings.Contains(key, ticket))
	}

	got, ok, err := store.ConsumeTicket(context.Background(), ticket)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, claims, got)

	_, ok, err = store.ConsumeTicket(context.Background(), ticket)
	require.NoError(t, err)
	require.False(t, ok)
}
