package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestNewStudioAuthVerifierIsDisabledByDefault(t *testing.T) {
	verifier, err := newStudioAuthVerifier(&config.Config{}, nil)
	require.NoError(t, err)
	require.Nil(t, verifier)
}

func TestNewStudioAuthVerifierBuildsRedisBackedVerifier(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cfg := &config.Config{StudioAuth: config.StudioAuthConfig{
		Enabled:             true,
		CurrentKeyID:        "studio-current",
		CurrentSecret:       "studio-current-secret-that-is-at-least-32-bytes",
		MaxClockSkewSeconds: 60,
		NonceTTLSeconds:     120,
	}}
	verifier, err := newStudioAuthVerifier(cfg, redisClient)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	now := time.Now()
	body := []byte(`{"email":"studio@example.com"}`)
	path := "/internal/v1/studio-auth/login"
	timestamp := strconv.FormatInt(now.Unix(), 10)
	nonce := "00112233445566778899aabbccddeeff"
	bodyHash := sha256.Sum256(body)
	canonical := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n%x", path, studioauth.ClientID, timestamp, nonce, bodyHash)
	signature := hmac.New(sha256.New, []byte(cfg.StudioAuth.CurrentSecret))
	_, _ = signature.Write([]byte(canonical))

	err = verifier.Verify(context.Background(), studioauth.SignedRequest{
		Method: http.MethodPost,
		Path:   path,
		Headers: http.Header{
			studioauth.HeaderClient:    []string{studioauth.ClientID},
			studioauth.HeaderKeyID:     []string{cfg.StudioAuth.CurrentKeyID},
			studioauth.HeaderTimestamp: []string{timestamp},
			studioauth.HeaderNonce:     []string{nonce},
			studioauth.HeaderSignature: []string{hex.EncodeToString(signature.Sum(nil))},
		},
		Body: body,
	})
	require.NoError(t, err)
}
