package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestConsumePasswordResetTokenIsSingleUse(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewEmailCache(client)
	ctx := context.Background()

	require.NoError(t, cache.SetPasswordResetToken(ctx, "reset@example.com", &service.PasswordResetTokenData{
		Token:     "secret-token",
		CreatedAt: time.Now(),
	}, 2*time.Minute))

	consumed, err := cache.ConsumePasswordResetToken(ctx, "reset@example.com", "wrong-token")
	require.NoError(t, err)
	require.False(t, consumed)

	consumed, err = cache.ConsumePasswordResetToken(ctx, "reset@example.com", "secret-token")
	require.NoError(t, err)
	require.True(t, consumed)

	consumed, err = cache.ConsumePasswordResetToken(ctx, "reset@example.com", "secret-token")
	require.NoError(t, err)
	require.False(t, consumed)
}
