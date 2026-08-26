package studioauth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrInvalidNonceClaim = errors.New("invalid Studio auth nonce claim")

type RedisNonceStore struct {
	client *redis.Client
	prefix string
}

func NewRedisNonceStore(client *redis.Client, prefix string) *RedisNonceStore {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "studio-auth:nonce"
	}
	return &RedisNonceStore{client: client, prefix: strings.TrimSuffix(prefix, ":") + ":"}
}

func (s *RedisNonceStore) Claim(ctx context.Context, clientID, keyID, nonce string, ttl time.Duration) (bool, error) {
	clientID = strings.TrimSpace(clientID)
	keyID = strings.TrimSpace(keyID)
	nonce = strings.TrimSpace(nonce)
	if s == nil || s.client == nil || clientID == "" || keyID == "" || nonce == "" || ttl <= 0 || strings.Contains(clientID, ":") || strings.Contains(keyID, ":") || strings.Contains(nonce, ":") {
		return false, ErrInvalidNonceClaim
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.client.SetNX(ctx, s.prefix+clientID+":"+keyID+":"+nonce, "1", ttl).Result()
}
