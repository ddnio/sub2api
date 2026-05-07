//go:build unit

package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyReferralRegistrationHooksDisabled(t *testing.T) {
	content, err := os.ReadFile("auth_service.go")
	require.NoError(t, err)

	source := string(content)
	require.NotContains(t, source, ".GenerateReferralCode(ctx, user.ID)")
	require.NotContains(t, source, ".ProcessRegistrationReferral(ctx, user.ID")
}

func TestLegacyReferralRechargeRewardDisabled(t *testing.T) {
	content, err := os.ReadFile("redeem_service.go")
	require.NoError(t, err)

	source := string(content)
	require.NotContains(t, source, ".GrantFirstRechargeReward(ctx, userID)")
	require.Contains(t, source, "tryAccrueAffiliateRebateForRedeem")
}
