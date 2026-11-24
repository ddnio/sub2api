package middleware

// NanaFox fork: 余额不足的报错补一句中文和充值链接——在 Claude Code / Codex 里用的人只看得到这条报错。
// 保留原来的错误码和英文开头，按错误文本判断的客户端和运维统计不受影响。

import (
	"context"
	"sync/atomic"
)

const insufficientBalanceMessage = "Insufficient account balance"

var rechargeURLProvider atomic.Pointer[func(context.Context) string]

// SetRechargeURLProvider 设置充值页地址的来源（后台低余额提醒里配置的地址），启动时调用一次。
func SetRechargeURLProvider(provider func(context.Context) string) {
	if provider == nil {
		rechargeURLProvider.Store(nil)
		return
	}
	rechargeURLProvider.Store(&provider)
}

func insufficientBalanceText(ctx context.Context) string {
	hint := "余额不足，请登录控制台充值后继续使用。"
	if provider := rechargeURLProvider.Load(); provider != nil {
		if rechargeURL := (*provider)(ctx); rechargeURL != "" {
			hint = "余额不足，请到 " + rechargeURL + " 充值后继续使用。"
		}
	}
	return insufficientBalanceMessage + ". " + hint
}
