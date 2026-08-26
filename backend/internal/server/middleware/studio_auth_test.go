package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/gin-gonic/gin"
)

const (
	middlewareTestKeyID  = "studio-current"
	middlewareTestSecret = "current-secret-with-at-least-32-bytes"
)

type middlewareNonceStore struct {
	claimed map[string]bool
}

func (s *middlewareNonceStore) Claim(_ context.Context, clientID, keyID, nonce string, _ time.Duration) (bool, error) {
	key := clientID + ":" + keyID + ":" + nonce
	if s.claimed[key] {
		return false, nil
	}
	s.claimed[key] = true
	return true, nil
}

func TestStudioAuthMiddlewareVerifiesAndRestoresRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Unix(1787702400, 0)
	verifier := newMiddlewareVerifier(t, now)
	handlerCalled := false
	router := gin.New()
	router.Use(StudioAuth(verifier, 1024))
	router.POST("/internal/v1/studio-auth/login", func(c *gin.Context) {
		handlerCalled = true
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"email":"creator@example.com"}` {
			t.Fatalf("body = %q", body)
		}
		c.JSON(http.StatusOK, gin.H{"accepted": true})
	})

	request := signedMiddlewareRequest(now, []byte(`{"email":"creator@example.com"}`), "00112233445566778899aabbccddeeff")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !handlerCalled {
		t.Fatalf("response = %d %s handlerCalled=%v", response.Code, response.Body.String(), handlerCalled)
	}
}

func TestStudioAuthMiddlewareRejectsInvalidAndReplayedRequestsBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Unix(1787702400, 0)
	verifier := newMiddlewareVerifier(t, now)
	handlerCalls := 0
	router := gin.New()
	router.Use(StudioAuth(verifier, 1024))
	router.POST("/internal/v1/studio-auth/login", func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusNoContent)
	})

	valid := signedMiddlewareRequest(now, []byte(`{"email":"creator@example.com"}`), "00112233445566778899aabbccddeeff")
	first := httptest.NewRecorder()
	router.ServeHTTP(first, valid)
	if first.Code != http.StatusNoContent || handlerCalls != 1 {
		t.Fatalf("first response = %d handlerCalls=%d", first.Code, handlerCalls)
	}

	replay := signedMiddlewareRequest(now, []byte(`{"email":"creator@example.com"}`), "00112233445566778899aabbccddeeff")
	replayResponse := httptest.NewRecorder()
	router.ServeHTTP(replayResponse, replay)
	if replayResponse.Code != http.StatusUnauthorized || handlerCalls != 1 || replayResponse.Body.String() != `{"error":{"code":"unauthorized_request"}}` {
		t.Fatalf("replay response = %d %s handlerCalls=%d", replayResponse.Code, replayResponse.Body.String(), handlerCalls)
	}

	invalid := signedMiddlewareRequest(now, []byte(`{"email":"creator@example.com"}`), "ffeeddccbbaa99887766554433221100")
	invalid.Header.Set(studioauth.HeaderSignature, strings.Repeat("0", sha256.Size*2))
	invalidResponse := httptest.NewRecorder()
	router.ServeHTTP(invalidResponse, invalid)
	if invalidResponse.Code != http.StatusUnauthorized || handlerCalls != 1 || invalidResponse.Body.String() != replayResponse.Body.String() {
		t.Fatalf("invalid response = %d %s handlerCalls=%d", invalidResponse.Code, invalidResponse.Body.String(), handlerCalls)
	}
}

func TestStudioAuthMiddlewareRejectsOversizedBodyBeforeVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Unix(1787702400, 0)
	verifier := newMiddlewareVerifier(t, now)
	handlerCalled := false
	router := gin.New()
	router.Use(StudioAuth(verifier, 32))
	router.POST("/internal/v1/studio-auth/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})

	request := signedMiddlewareRequest(now, bytes.Repeat([]byte("a"), 33), "00112233445566778899aabbccddeeff")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge || handlerCalled || response.Body.String() != `{"error":{"code":"request_too_large"}}` {
		t.Fatalf("response = %d %s handlerCalled=%v", response.Code, response.Body.String(), handlerCalled)
	}
}

func TestStudioAuthMiddlewareFailsClosedWithoutVerifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handlerCalled := false
	router := gin.New()
	router.Use(StudioAuth(nil, 1024))
	router.POST("/internal/v1/studio-auth/login", func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/internal/v1/studio-auth/login", strings.NewReader("{}")))
	if response.Code != http.StatusServiceUnavailable || handlerCalled || response.Body.String() != `{"error":{"code":"service_unavailable"}}` {
		t.Fatalf("response = %d %s handlerCalled=%v", response.Code, response.Body.String(), handlerCalled)
	}
}

func newMiddlewareVerifier(t *testing.T, now time.Time) *studioauth.Verifier {
	t.Helper()
	verifier, err := studioauth.NewVerifier(studioauth.VerifierConfig{
		ClientID: studioauth.ClientID,
		Current:  studioauth.SigningKey{ID: middlewareTestKeyID, Secret: middlewareTestSecret},
		Clock:    func() time.Time { return now },
	}, &middlewareNonceStore{claimed: make(map[string]bool)})
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}

func signedMiddlewareRequest(now time.Time, body []byte, nonce string) *http.Request {
	path := "/internal/v1/studio-auth/login"
	timestamp := strconv.FormatInt(now.Unix(), 10)
	bodyHash := sha256.Sum256(body)
	canonical := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n%x", path, studioauth.ClientID, timestamp, nonce, bodyHash)
	signature := hmac.New(sha256.New, []byte(middlewareTestSecret))
	_, _ = signature.Write([]byte(canonical))
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set(studioauth.HeaderClient, studioauth.ClientID)
	request.Header.Set(studioauth.HeaderKeyID, middlewareTestKeyID)
	request.Header.Set(studioauth.HeaderTimestamp, timestamp)
	request.Header.Set(studioauth.HeaderNonce, nonce)
	request.Header.Set(studioauth.HeaderSignature, hex.EncodeToString(signature.Sum(nil)))
	request.Header.Set("Content-Type", "application/json")
	return request
}
