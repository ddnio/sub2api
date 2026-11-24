//go:build unit

package rechargehint

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	t.Cleanup(func() { SetURLProvider(nil) })

	SetURLProvider(nil)
	require.Equal(t, "insufficient balance. 余额不足，请登录控制台充值后继续使用。", Append(context.Background(), "insufficient balance"))

	SetURLProvider(func(context.Context) string { return "" })
	require.Equal(t, "insufficient balance. 余额不足，请登录控制台充值后继续使用。", Append(context.Background(), "insufficient balance"))

	SetURLProvider(func(context.Context) string { return "https://router.nanafox.com/purchase" })
	require.Equal(t, "Insufficient account balance. 余额不足，请到 https://router.nanafox.com/purchase 充值后继续使用。", Append(context.Background(), "Insufficient account balance"))
}
