//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRechargeURL(t *testing.T) {
	require.Equal(t, "https://router.nanafox.com/purchase", normalizeRechargeURL("  https://router.nanafox.com/purchase "))
	require.Equal(t, "http://example.com", normalizeRechargeURL("http://example.com"))
	for _, raw := range []string{"", "javascript:alert(1)", "/purchase", "https://", "ftp://example.com"} {
		require.Empty(t, normalizeRechargeURL(raw), raw)
	}
}

func TestGetRechargeURLCachesLookup(t *testing.T) {
	rechargeURLCache.Store(nil)
	t.Cleanup(func() { rechargeURLCache.Store(nil) })

	repo := &settingRepoStub{values: map[string]string{SettingKeyBalanceLowNotifyRechargeURL: "https://router.nanafox.com/purchase"}}
	svc := &SettingService{settingRepo: repo}
	require.Equal(t, "https://router.nanafox.com/purchase", svc.GetRechargeURL(context.Background()))
	require.Equal(t, "https://router.nanafox.com/purchase", svc.GetRechargeURL(context.Background()))
	require.Equal(t, 1, repo.getValueCalls)
}

func TestGetRechargeURLKeepsLastValueOnDBError(t *testing.T) {
	rechargeURLCache.Store(&cachedRechargeURL{value: "https://router.nanafox.com/purchase"})
	t.Cleanup(func() { rechargeURLCache.Store(nil) })

	svc := &SettingService{settingRepo: &settingRepoStub{err: errors.New("db down")}}
	require.Equal(t, "https://router.nanafox.com/purchase", svc.GetRechargeURL(context.Background()))
}

func TestGetRechargeURLMissingSetting(t *testing.T) {
	rechargeURLCache.Store(nil)
	t.Cleanup(func() { rechargeURLCache.Store(nil) })

	svc := &SettingService{settingRepo: &settingRepoStub{values: map[string]string{}}}
	require.Empty(t, svc.GetRechargeURL(context.Background()))
}
