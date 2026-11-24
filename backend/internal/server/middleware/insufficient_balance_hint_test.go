//go:build unit

package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsufficientBalanceText(t *testing.T) {
	t.Cleanup(func() { SetRechargeURLProvider(nil) })

	SetRechargeURLProvider(nil)
	require.Equal(t, "Insufficient account balance. 余额不足，请登录控制台充值后继续使用。", insufficientBalanceText(context.Background()))

	SetRechargeURLProvider(func(context.Context) string { return "" })
	require.Equal(t, "Insufficient account balance. 余额不足，请登录控制台充值后继续使用。", insufficientBalanceText(context.Background()))

	SetRechargeURLProvider(func(context.Context) string { return "https://router.nanafox.com/purchase" })
	require.Equal(t, "Insufficient account balance. 余额不足，请到 https://router.nanafox.com/purchase 充值后继续使用。", insufficientBalanceText(context.Background()))
}
