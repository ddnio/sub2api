//go:build unit

package middleware

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/rechargehint"
	"github.com/stretchr/testify/require"
)

func TestInsufficientBalanceText(t *testing.T) {
	t.Cleanup(func() { rechargehint.SetURLProvider(nil) })

	rechargehint.SetURLProvider(func(context.Context) string { return "https://router.nanafox.com/purchase" })
	require.Equal(t, "Insufficient account balance. 余额不足，请到 https://router.nanafox.com/purchase 充值后继续使用。", insufficientBalanceText(context.Background()))
}
