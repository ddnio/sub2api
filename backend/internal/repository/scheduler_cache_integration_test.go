//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSchedulerCacheSnapshotUsesSlimMetadataButKeepsFullAccount(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)
	cache := NewSchedulerCache(rdb)

	bucket := service.SchedulerBucket{GroupID: 2, Platform: service.PlatformGemini, Mode: service.SchedulerModeSingle}
	now := time.Now().UTC().Truncate(time.Second)
	limitReset := now.Add(10 * time.Minute)
	overloadUntil := now.Add(2 * time.Minute)
	tempUnschedUntil := now.Add(3 * time.Minute)
	windowEnd := now.Add(5 * time.Hour)

	account := service.Account{
		ID:          101,
		Name:        "gemini-heavy",
		Platform:    service.PlatformGemini,
		Type:        service.AccountTypeOAuth,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 3,
		Priority:    7,
		LastUsedAt:  &now,
		Credentials: map[string]any{
			"api_key":       "gemini-api-key",
			"access_token":  "secret-access-token",
			"project_id":    "proj-1",
			"oauth_type":    "ai_studio",
			"model_mapping": map[string]any{"gemini-2.5-pro": "gemini-2.5-pro"},
			"huge_blob":     strings.Repeat("x", 4096),
		},
		Extra: map[string]any{
			"mixed_scheduling":             true,
			"window_cost_limit":            12.5,
			"window_cost_sticky_reserve":   8.0,
			"max_sessions":                 4,
			"session_idle_timeout_minutes": 11,
			"unused_large_field":           strings.Repeat("y", 4096),
		},
		RateLimitResetAt:       &limitReset,
		OverloadUntil:          &overloadUntil,
		TempUnschedulableUntil: &tempUnschedUntil,
		SessionWindowStart:     &now,
		SessionWindowEnd:       &windowEnd,
		SessionWindowStatus:    "active",
	}

	require.NoError(t, cache.SetSnapshot(ctx, bucket, []service.Account{account}))

	snapshot, hit, err := cache.GetSnapshot(ctx, bucket)
	require.NoError(t, err)
	require.True(t, hit)
	require.Len(t, snapshot, 1)

	got := snapshot[0]
	require.NotNil(t, got)
	require.Equal(t, "gemini-api-key", got.GetCredential("api_key"))
	require.Equal(t, "proj-1", got.GetCredential("project_id"))
	require.Equal(t, "ai_studio", got.GetCredential("oauth_type"))
	require.NotEmpty(t, got.GetModelMapping())
	require.Empty(t, got.GetCredential("access_token"))
	require.Empty(t, got.GetCredential("huge_blob"))
	require.Equal(t, true, got.Extra["mixed_scheduling"])
	require.Equal(t, 12.5, got.GetWindowCostLimit())
	require.Equal(t, 8.0, got.GetWindowCostStickyReserve())
	require.Equal(t, 4, got.GetMaxSessions())
	require.Equal(t, 11, got.GetSessionIdleTimeoutMinutes())
	require.Nil(t, got.Extra["unused_large_field"])

	full, err := cache.GetAccount(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, full)
	require.Equal(t, "secret-access-token", full.GetCredential("access_token"))
	require.Equal(t, strings.Repeat("x", 4096), full.GetCredential("huge_blob"))
}

func TestSchedulerCacheSetSnapshotDoesNotRollBackActiveVersion(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)
	cache := NewSchedulerCache(rdb)

	bucket := service.SchedulerBucket{GroupID: 2, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeSingle}
	activeKey := schedulerBucketKey(schedulerActivePrefix, bucket)
	require.NoError(t, rdb.Set(ctx, activeKey, "999", 0).Err())

	account := service.Account{
		ID:          201,
		Name:        "stale-writer",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
	}
	require.NoError(t, cache.SetSnapshot(ctx, bucket, []service.Account{account}))

	active, err := rdb.Get(ctx, activeKey).Result()
	require.NoError(t, err)
	require.Equal(t, "999", active)

	exists, err := rdb.Exists(ctx, schedulerSnapshotKey(bucket, "1")).Result()
	require.NoError(t, err)
	require.Zero(t, exists)
}

func TestSchedulerCacheSetSnapshotKeepsOldSnapshotDuringGracePeriod(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)
	cache := NewSchedulerCache(rdb)

	bucket := service.SchedulerBucket{GroupID: 3, Platform: service.PlatformGemini, Mode: service.SchedulerModeSingle}
	first := service.Account{
		ID:          301,
		Name:        "first",
		Platform:    service.PlatformGemini,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
	}
	second := first
	second.ID = 302
	second.Name = "second"

	require.NoError(t, cache.SetSnapshot(ctx, bucket, []service.Account{first}))
	require.NoError(t, cache.SetSnapshot(ctx, bucket, []service.Account{second}))

	active, err := rdb.Get(ctx, schedulerBucketKey(schedulerActivePrefix, bucket)).Result()
	require.NoError(t, err)
	require.Equal(t, "2", active)

	oldMembers, err := rdb.ZRange(ctx, schedulerSnapshotKey(bucket, "1"), 0, -1).Result()
	require.NoError(t, err)
	require.Equal(t, []string{"301"}, oldMembers)

	ttl, err := rdb.TTL(ctx, schedulerSnapshotKey(bucket, "1")).Result()
	require.NoError(t, err)
	require.Positive(t, ttl)
	require.LessOrEqual(t, ttl, time.Duration(snapshotGraceTTLSeconds)*time.Second)

	snapshot, hit, err := cache.GetSnapshot(ctx, bucket)
	require.NoError(t, err)
	require.True(t, hit)
	require.Len(t, snapshot, 1)
	require.Equal(t, int64(302), snapshot[0].ID)
}

func TestSchedulerCacheUnlockBucketReleasesLock(t *testing.T) {
	ctx := context.Background()
	rdb := testRedis(t)
	cache := NewSchedulerCache(rdb)

	bucket := service.SchedulerBucket{GroupID: 4, Platform: service.PlatformAnthropic, Mode: service.SchedulerModeSingle}
	ok, err := cache.TryLockBucket(ctx, bucket, time.Minute)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = cache.TryLockBucket(ctx, bucket, time.Minute)
	require.NoError(t, err)
	require.False(t, ok)

	require.NoError(t, cache.UnlockBucket(ctx, bucket))
	ok, err = cache.TryLockBucket(ctx, bucket, time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
}
