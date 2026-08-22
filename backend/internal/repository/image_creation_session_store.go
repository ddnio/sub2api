package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const imageCreationSessionPrefix = "image_creation:session:"

type imageCreationSessionStore struct {
	redis *redis.Client
}

func NewImageCreationSessionStore(redisClient *redis.Client) service.ImageCreationSessionStore {
	return &imageCreationSessionStore{redis: redisClient}
}

func (s *imageCreationSessionStore) StoreTicket(ctx context.Context, claims service.ImageCreationSessionClaims, ttl time.Duration) (string, error) {
	return s.store(ctx, "ticket:", claims, ttl)
}

func (s *imageCreationSessionStore) ConsumeTicket(ctx context.Context, token string) (service.ImageCreationSessionClaims, bool, error) {
	if s == nil || s.redis == nil {
		return service.ImageCreationSessionClaims{}, false, errors.New("redis is not configured")
	}
	raw, err := s.redis.GetDel(ctx, s.key("ticket:", token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return service.ImageCreationSessionClaims{}, false, nil
	}
	if err != nil {
		return service.ImageCreationSessionClaims{}, false, err
	}
	var claims service.ImageCreationSessionClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return service.ImageCreationSessionClaims{}, false, err
	}
	return claims, true, nil
}

func (s *imageCreationSessionStore) StoreSession(ctx context.Context, claims service.ImageCreationSessionClaims, ttl time.Duration) (string, error) {
	return s.store(ctx, "token:", claims, ttl)
}

func (s *imageCreationSessionStore) GetSession(ctx context.Context, token string) (service.ImageCreationSessionClaims, bool, error) {
	if s == nil || s.redis == nil {
		return service.ImageCreationSessionClaims{}, false, errors.New("redis is not configured")
	}
	raw, err := s.redis.Get(ctx, s.key("token:", token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return service.ImageCreationSessionClaims{}, false, nil
	}
	if err != nil {
		return service.ImageCreationSessionClaims{}, false, err
	}
	var claims service.ImageCreationSessionClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return service.ImageCreationSessionClaims{}, false, err
	}
	return claims, true, nil
}

func (s *imageCreationSessionStore) store(ctx context.Context, kind string, claims service.ImageCreationSessionClaims, ttl time.Duration) (string, error) {
	if s == nil || s.redis == nil {
		return "", errors.New("redis is not configured")
	}
	if ttl <= 0 {
		return "", errors.New("ttl must be positive")
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, s.key(kind, token), payload, ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *imageCreationSessionStore) key(kind, token string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return imageCreationSessionPrefix + kind + hex.EncodeToString(hash[:])
}
