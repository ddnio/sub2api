package studioauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const (
	testClientID       = "nanafox-studio"
	testCurrentKeyID   = "studio-current"
	testPreviousKeyID  = "studio-previous"
	testCurrentSecret  = "current-secret-with-at-least-32-bytes"
	testPreviousSecret = "previous-secret-with-at-least-32-bytes"
)

type fakeNonceStore struct {
	claimed map[string]bool
	err     error
}

func (s *fakeNonceStore) Claim(_ context.Context, clientID, keyID, nonce string, _ time.Duration) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	key := clientID + ":" + keyID + ":" + nonce
	if s.claimed[key] {
		return false, nil
	}
	s.claimed[key] = true
	return true, nil
}

func TestVerifierAcceptsCurrentAndPreviousKeysOnce(t *testing.T) {
	now := time.Unix(1787702400, 0)
	nonces := &fakeNonceStore{claimed: make(map[string]bool)}
	verifier, err := NewVerifier(VerifierConfig{
		ClientID: testClientID,
		Current:  SigningKey{ID: testCurrentKeyID, Secret: testCurrentSecret},
		Previous: &SigningKey{ID: testPreviousKeyID, Secret: testPreviousSecret},
		Clock:    func() time.Time { return now },
	}, nonces)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		keyID  string
		secret string
		nonce  string
	}{
		{name: "current", keyID: testCurrentKeyID, secret: testCurrentSecret, nonce: "00112233445566778899aabbccddeeff"},
		{name: "previous", keyID: testPreviousKeyID, secret: testPreviousSecret, nonce: "ffeeddccbbaa99887766554433221100"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := signedRequest(now, testClientID, test.keyID, test.secret, test.nonce, []byte(`{"email":"creator@example.com"}`))
			if err := verifier.Verify(context.Background(), request); err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
		})
	}
}

func TestVerifierRejectsInvalidOrReplayedRequests(t *testing.T) {
	now := time.Unix(1787702400, 0)
	tests := []struct {
		name   string
		mutate func(*SignedRequest)
	}{
		{name: "wrong client", mutate: func(request *SignedRequest) { request.Headers.Set(HeaderClient, "other-client") }},
		{name: "unknown key", mutate: func(request *SignedRequest) { request.Headers.Set(HeaderKeyID, "unknown") }},
		{name: "expired timestamp", mutate: func(request *SignedRequest) {
			request.Headers.Set(HeaderTimestamp, strconv.FormatInt(now.Add(-61*time.Second).Unix(), 10))
		}},
		{name: "future timestamp", mutate: func(request *SignedRequest) {
			request.Headers.Set(HeaderTimestamp, strconv.FormatInt(now.Add(61*time.Second).Unix(), 10))
		}},
		{name: "invalid nonce", mutate: func(request *SignedRequest) { request.Headers.Set(HeaderNonce, "not-hex") }},
		{name: "tampered body", mutate: func(request *SignedRequest) { request.Body = []byte(`{"email":"attacker@example.com"}`) }},
		{name: "tampered path", mutate: func(request *SignedRequest) { request.Path = "/internal/v1/studio-auth/register" }},
		{name: "tampered method", mutate: func(request *SignedRequest) { request.Method = http.MethodGet }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nonces := &fakeNonceStore{claimed: make(map[string]bool)}
			verifier, err := NewVerifier(VerifierConfig{
				ClientID: testClientID,
				Current:  SigningKey{ID: testCurrentKeyID, Secret: testCurrentSecret},
				Clock:    func() time.Time { return now },
			}, nonces)
			if err != nil {
				t.Fatal(err)
			}
			request := signedRequest(now, testClientID, testCurrentKeyID, testCurrentSecret, "00112233445566778899aabbccddeeff", []byte(`{"email":"creator@example.com"}`))
			test.mutate(&request)
			if err := verifier.Verify(context.Background(), request); !errors.Is(err, ErrUnauthorized) {
				t.Fatalf("Verify() error = %v, want %v", err, ErrUnauthorized)
			}
		})
	}

	nonces := &fakeNonceStore{claimed: make(map[string]bool)}
	verifier, err := NewVerifier(VerifierConfig{
		ClientID: testClientID,
		Current:  SigningKey{ID: testCurrentKeyID, Secret: testCurrentSecret},
		Clock:    func() time.Time { return now },
	}, nonces)
	if err != nil {
		t.Fatal(err)
	}
	request := signedRequest(now, testClientID, testCurrentKeyID, testCurrentSecret, "00112233445566778899aabbccddeeff", []byte(`{"email":"creator@example.com"}`))
	if err := verifier.Verify(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if err := verifier.Verify(context.Background(), request); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("replay error = %v, want %v", err, ErrUnauthorized)
	}
}

func TestVerifierFailsClosedWhenNonceStoreIsUnavailable(t *testing.T) {
	now := time.Unix(1787702400, 0)
	verifier, err := NewVerifier(VerifierConfig{
		ClientID: testClientID,
		Current:  SigningKey{ID: testCurrentKeyID, Secret: testCurrentSecret},
		Clock:    func() time.Time { return now },
	}, &fakeNonceStore{claimed: make(map[string]bool), err: errors.New("redis unavailable")})
	if err != nil {
		t.Fatal(err)
	}
	request := signedRequest(now, testClientID, testCurrentKeyID, testCurrentSecret, "00112233445566778899aabbccddeeff", []byte(`{"email":"creator@example.com"}`))
	if err := verifier.Verify(context.Background(), request); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrUnauthorized)
	}
}

func TestNewVerifierRejectsUnsafeConfiguration(t *testing.T) {
	nonces := &fakeNonceStore{claimed: make(map[string]bool)}
	for _, config := range []VerifierConfig{
		{},
		{ClientID: testClientID, Current: SigningKey{ID: testCurrentKeyID, Secret: "short"}},
		{ClientID: testClientID, Current: SigningKey{ID: "invalid key", Secret: testCurrentSecret}},
		{ClientID: "other-client", Current: SigningKey{ID: testCurrentKeyID, Secret: testCurrentSecret}},
	} {
		if _, err := NewVerifier(config, nonces); !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("NewVerifier(%#v) error = %v, want %v", config, err, ErrInvalidConfig)
		}
	}
}

func signedRequest(now time.Time, clientID, keyID, secret, nonce string, body []byte) SignedRequest {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	bodyHash := sha256.Sum256(body)
	path := "/internal/v1/studio-auth/login"
	canonical := fmt.Sprintf("POST\n%s\n%s\n%s\n%s\n%x", path, clientID, timestamp, nonce, bodyHash)
	signature := hmac.New(sha256.New, []byte(secret))
	_, _ = signature.Write([]byte(canonical))
	headers := make(http.Header)
	headers.Set(HeaderClient, clientID)
	headers.Set(HeaderKeyID, keyID)
	headers.Set(HeaderTimestamp, timestamp)
	headers.Set(HeaderNonce, nonce)
	headers.Set(HeaderSignature, hex.EncodeToString(signature.Sum(nil)))
	return SignedRequest{Method: http.MethodPost, Path: path, Headers: headers, Body: body}
}
