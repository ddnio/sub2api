package studioauth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisNonceStoreClaimsOnceWithBoundedTTL(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	store := NewRedisNonceStore(client, "studio-auth:nonce")

	claimed, err := store.Claim(context.Background(), ClientID, testCurrentKeyID, "00112233445566778899aabbccddeeff", 2*time.Minute)
	if err != nil || !claimed {
		t.Fatalf("first Claim() = (%v, %v), want (true, nil)", claimed, err)
	}
	claimed, err = store.Claim(context.Background(), ClientID, testCurrentKeyID, "00112233445566778899aabbccddeeff", 2*time.Minute)
	if err != nil || claimed {
		t.Fatalf("second Claim() = (%v, %v), want (false, nil)", claimed, err)
	}

	key := "studio-auth:nonce:" + ClientID + ":" + testCurrentKeyID + ":00112233445566778899aabbccddeeff"
	if ttl := server.TTL(key); ttl != 2*time.Minute {
		t.Fatalf("TTL = %v, want %v", ttl, 2*time.Minute)
	}
	claimed, err = store.Claim(context.Background(), ClientID, testPreviousKeyID, "00112233445566778899aabbccddeeff", 2*time.Minute)
	if err != nil || !claimed {
		t.Fatalf("other key Claim() = (%v, %v), want (true, nil)", claimed, err)
	}
}

func TestRedisNonceStoreFailsClosedWithoutRedis(t *testing.T) {
	store := NewRedisNonceStore(redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 10 * time.Millisecond,
		ReadTimeout: 10 * time.Millisecond,
	}), "")
	claimed, err := store.Claim(context.Background(), ClientID, testCurrentKeyID, "00112233445566778899aabbccddeeff", 2*time.Minute)
	if err == nil || claimed {
		t.Fatalf("Claim() = (%v, %v), want (false, error)", claimed, err)
	}
}

func TestRedisNonceStoreRejectsInvalidInput(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	store := NewRedisNonceStore(client, "")
	for _, test := range []struct {
		clientID string
		keyID    string
		nonce    string
		ttl      time.Duration
	}{
		{keyID: testCurrentKeyID, nonce: "nonce", ttl: time.Minute},
		{clientID: ClientID, nonce: "nonce", ttl: time.Minute},
		{clientID: ClientID, keyID: testCurrentKeyID, ttl: time.Minute},
		{clientID: ClientID, keyID: testCurrentKeyID, nonce: "nonce"},
	} {
		if claimed, err := store.Claim(context.Background(), test.clientID, test.keyID, test.nonce, test.ttl); err == nil || claimed {
			t.Fatalf("Claim(%#v) = (%v, %v), want (false, error)", test, claimed, err)
		}
	}
}
