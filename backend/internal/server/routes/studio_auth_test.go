package routes

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type routeStudioNonceStore struct {
	mu     sync.Mutex
	claims map[string]struct{}
}

func (s *routeStudioNonceStore) Claim(_ context.Context, clientID, keyID, nonce string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := clientID + ":" + keyID + ":" + nonce
	if _, ok := s.claims[key]; ok {
		return false, nil
	}
	s.claims[key] = struct{}{}
	return true, nil
}

func TestRegisterStudioAuthRoutesRequiresSignedInternalRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Unix(1_800_000_000, 0)
	verifier, err := studioauth.NewVerifier(studioauth.VerifierConfig{
		ClientID: studioauth.ClientID,
		Current: studioauth.SigningKey{
			ID:     "studio-test-key",
			Secret: "studio-test-secret-that-is-at-least-32-bytes",
		},
		Clock: func() time.Time { return now },
	}, &routeStudioNonceStore{claims: map[string]struct{}{}})
	require.NoError(t, err)

	router := gin.New()
	RegisterStudioAuthRoutes(router, &handler.Handlers{
		StudioAuth: handler.NewStudioAuthHandler(nil, nil, nil, nil),
	}, verifier, 1024)

	unsigned := httptest.NewRequest(http.MethodPost, "/internal/v1/studio-auth/login", bytes.NewBufferString(`{}`))
	unsigned.Header.Set("Content-Type", "application/json")
	unsignedResponse := httptest.NewRecorder()
	router.ServeHTTP(unsignedResponse, unsigned)
	require.Equal(t, http.StatusUnauthorized, unsignedResponse.Code)

	signed := signedStudioRouteRequest(now, "00112233445566778899aabbccddeeff", []byte(`{}`))
	signedResponse := httptest.NewRecorder()
	router.ServeHTTP(signedResponse, signed)
	require.Equal(t, http.StatusBadRequest, signedResponse.Code)
}

func TestRegisterStudioAuthRoutesStaysAbsentWhenDisabled(t *testing.T) {
	router := gin.New()
	RegisterStudioAuthRoutes(router, &handler.Handlers{}, nil, 1024)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/internal/v1/studio-auth/login", bytes.NewBufferString(`{}`)))
	require.Equal(t, http.StatusNotFound, response.Code)
}

func signedStudioRouteRequest(now time.Time, nonce string, body []byte) *http.Request {
	path := "/internal/v1/studio-auth/login"
	timestamp := strconv.FormatInt(now.Unix(), 10)
	bodyHash := sha256.Sum256(body)
	canonical := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n%x", path, studioauth.ClientID, timestamp, nonce, bodyHash)
	signature := hmac.New(sha256.New, []byte("studio-test-secret-that-is-at-least-32-bytes"))
	_, _ = signature.Write([]byte(canonical))

	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(studioauth.HeaderClient, studioauth.ClientID)
	request.Header.Set(studioauth.HeaderKeyID, "studio-test-key")
	request.Header.Set(studioauth.HeaderTimestamp, timestamp)
	request.Header.Set(studioauth.HeaderNonce, nonce)
	request.Header.Set(studioauth.HeaderSignature, hex.EncodeToString(signature.Sum(nil)))
	return request
}
