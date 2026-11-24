package middleware

// NanaFox fork: 鉴权时余额不足的报错补一句中文和充值链接，文案见 rechargehint。

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/rechargehint"
)

const insufficientBalanceMessage = "Insufficient account balance"

func insufficientBalanceText(ctx context.Context) string {
	return rechargehint.Append(ctx, insufficientBalanceMessage)
}
