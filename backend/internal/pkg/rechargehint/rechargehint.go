// Package rechargehint 是 NanaFox fork 的余额不足提示：在原英文报错后面补一句中文和充值链接。
// 在 Claude Code / Codex 里用的人只看得到报错文本，这句话告诉他们去哪儿充值。
// 原错误码和英文开头不动，按错误文本判断的客户端和运维统计不受影响。
package rechargehint

import (
	"context"
	"strings"
	"sync/atomic"
	"time"
)

// lookupTimeout 限制读取充值地址的时间：计费路径可能没有请求上下文，读不到就退回通用提示。
const lookupTimeout = 2 * time.Second

var urlProvider atomic.Pointer[func(context.Context) string]

// SetURLProvider 设置充值页地址的来源（后台低余额提醒里配置的地址），启动时调用一次；传 nil 清除。
func SetURLProvider(provider func(context.Context) string) {
	if provider == nil {
		urlProvider.Store(nil)
		return
	}
	urlProvider.Store(&provider)
}

// Append 在英文报错后面接上中文充值提示；英文为空时只返回中文提示。
func Append(ctx context.Context, message string) string {
	hint := "余额不足，请登录控制台充值后继续使用。"
	if provider := urlProvider.Load(); provider != nil {
		lookupCtx, cancel := context.WithTimeout(ctx, lookupTimeout)
		rechargeURL := strings.TrimSpace((*provider)(lookupCtx))
		cancel()
		if rechargeURL != "" {
			hint = "余额不足，请到 " + rechargeURL + " 充值后继续使用。"
		}
	}
	message = strings.TrimRight(strings.TrimSpace(message), ". ")
	if message == "" {
		return hint
	}
	return message + ". " + hint
}
