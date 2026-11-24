// Package rechargehint 是 NanaFox fork 的余额不足提示：在原英文报错后面补一句中文和充值链接。
// 在 Claude Code / Codex 里用的人只看得到报错文本，这句话告诉他们去哪儿充值。
// 原错误码和英文开头不动，按错误文本判断的客户端和运维统计不受影响。
package rechargehint

import (
	"context"
	"sync/atomic"
)

var urlProvider atomic.Pointer[func(context.Context) string]

// SetURLProvider 设置充值页地址的来源（后台低余额提醒里配置的地址），启动时调用一次；传 nil 清除。
func SetURLProvider(provider func(context.Context) string) {
	if provider == nil {
		urlProvider.Store(nil)
		return
	}
	urlProvider.Store(&provider)
}

// Append 在英文报错后面接上中文充值提示。
func Append(ctx context.Context, message string) string {
	hint := "余额不足，请登录控制台充值后继续使用。"
	if provider := urlProvider.Load(); provider != nil {
		if rechargeURL := (*provider)(ctx); rechargeURL != "" {
			hint = "余额不足，请到 " + rechargeURL + " 充值后继续使用。"
		}
	}
	return message + ". " + hint
}
