package service

// NanaFox fork: 余额不足时 API 报错里附带的充值链接，取后台「低余额提醒」里配置的充值页地址。
// 只接受 http(s) 地址；结果缓存 1 分钟，余额为 0 的客户端反复请求也不会压到数据库。

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const rechargeURLCacheTTL = time.Minute

type cachedRechargeURL struct {
	value     string
	expiresAt int64
}

var rechargeURLCache atomic.Pointer[cachedRechargeURL]

// GetRechargeURL 返回后台配置的充值页地址；没配或不是 http(s) 地址时返回空串。
func (s *SettingService) GetRechargeURL(ctx context.Context) string {
	now := time.Now()
	if cached := rechargeURLCache.Load(); cached != nil && now.UnixNano() < cached.expiresAt {
		return cached.value
	}
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	value := ""
	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyBalanceLowNotifyRechargeURL)
	switch {
	case err == nil:
		value = normalizeRechargeURL(raw)
	case !errors.Is(err, ErrSettingNotFound):
		// 数据库临时出错：沿用上一次的地址
		if cached := rechargeURLCache.Load(); cached != nil {
			value = cached.value
		}
	}
	rechargeURLCache.Store(&cachedRechargeURL{value: value, expiresAt: now.Add(rechargeURLCacheTTL).UnixNano()})
	return value
}

func normalizeRechargeURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	u, err := url.Parse(trimmed)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return trimmed
}
