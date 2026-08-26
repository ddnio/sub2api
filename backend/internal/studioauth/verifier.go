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
	"strings"
	"time"
)

const (
	ClientID        = "nanafox-studio"
	HeaderClient    = "X-NanaFox-Client"
	HeaderKeyID     = "X-NanaFox-Key-ID"
	HeaderTimestamp = "X-NanaFox-Timestamp"
	HeaderNonce     = "X-NanaFox-Nonce"
	HeaderSignature = "X-NanaFox-Signature"
)

var (
	ErrInvalidConfig = errors.New("invalid Studio auth verifier config")
	ErrUnauthorized  = errors.New("unauthorized Studio auth request")
)

type NonceStore interface {
	Claim(ctx context.Context, clientID, keyID, nonce string, ttl time.Duration) (bool, error)
}

type SigningKey struct {
	ID     string
	Secret string
}

type VerifierConfig struct {
	ClientID     string
	Current      SigningKey
	Previous     *SigningKey
	MaxClockSkew time.Duration
	NonceTTL     time.Duration
	Clock        func() time.Time
}

type SignedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

type Verifier struct {
	clientID     string
	keys         map[string][]byte
	maxClockSkew time.Duration
	nonceTTL     time.Duration
	clock        func() time.Time
	nonces       NonceStore
}

func NewVerifier(config VerifierConfig, nonces NonceStore) (*Verifier, error) {
	if config.ClientID != ClientID || nonces == nil || !validSigningKey(config.Current) {
		return nil, ErrInvalidConfig
	}
	keys := map[string][]byte{config.Current.ID: []byte(config.Current.Secret)}
	if config.Previous != nil {
		if !validSigningKey(*config.Previous) || config.Previous.ID == config.Current.ID {
			return nil, ErrInvalidConfig
		}
		keys[config.Previous.ID] = []byte(config.Previous.Secret)
	}
	maxClockSkew := config.MaxClockSkew
	if maxClockSkew <= 0 {
		maxClockSkew = time.Minute
	}
	nonceTTL := config.NonceTTL
	if nonceTTL <= 0 {
		nonceTTL = 2 * time.Minute
	}
	clock := config.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Verifier{
		clientID:     config.ClientID,
		keys:         keys,
		maxClockSkew: maxClockSkew,
		nonceTTL:     nonceTTL,
		clock:        clock,
		nonces:       nonces,
	}, nil
}

func (v *Verifier) Verify(ctx context.Context, request SignedRequest) error {
	clientID := request.Headers.Get(HeaderClient)
	keyID := request.Headers.Get(HeaderKeyID)
	timestampText := request.Headers.Get(HeaderTimestamp)
	nonce := request.Headers.Get(HeaderNonce)
	signatureText := request.Headers.Get(HeaderSignature)
	secret, ok := v.keys[keyID]
	if request.Method != http.MethodPost || !strings.HasPrefix(request.Path, "/internal/v1/studio-auth/") || clientID != v.clientID || !ok {
		return ErrUnauthorized
	}
	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return ErrUnauthorized
	}
	age := v.clock().Sub(time.Unix(timestamp, 0))
	if age < -v.maxClockSkew || age > v.maxClockSkew {
		return ErrUnauthorized
	}
	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil || len(nonceBytes) != 16 {
		return ErrUnauthorized
	}
	signatureBytes, err := hex.DecodeString(signatureText)
	if err != nil || len(signatureBytes) != sha256.Size {
		return ErrUnauthorized
	}
	bodyHash := sha256.Sum256(request.Body)
	canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%x", request.Method, request.Path, clientID, timestampText, nonce, bodyHash)
	expected := hmac.New(sha256.New, secret)
	_, _ = expected.Write([]byte(canonical))
	if !hmac.Equal(signatureBytes, expected.Sum(nil)) {
		return ErrUnauthorized
	}
	claimed, err := v.nonces.Claim(ctx, clientID, keyID, nonce, v.nonceTTL)
	if err != nil || !claimed {
		return ErrUnauthorized
	}
	return nil
}

func validSigningKey(key SigningKey) bool {
	if len(key.Secret) < 32 || key.ID == "" || len(key.ID) > 64 {
		return false
	}
	for _, char := range key.ID {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}
