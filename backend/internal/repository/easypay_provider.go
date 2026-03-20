package repository

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// httpClient with explicit timeout to avoid hanging on slow providers.
var easypayHTTPClient = &http.Client{Timeout: 15 * time.Second}

type easypayProvider struct {
	cfg config.PaymentConfig
}

func NewEasyPayProvider(cfg config.PaymentConfig) service.PaymentProvider {
	return &easypayProvider{cfg: cfg}
}

// sign 计算易支付标准 MD5 签名。
// 规则：参数按 key 字母升序排列，拼接为 k=v&k=v，追加密钥，取 MD5 小写。
// 排除 sign、sign_type 字段，排除值为空的字段。
func (p *easypayProvider) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params[k])
	}
	sb.WriteString(p.cfg.EasyPaySignKey)

	h := md5.Sum([]byte(sb.String()))
	return fmt.Sprintf("%x", h)
}

// mapiResponse 是易支付 mapi.php 返回的 JSON 结构。
// mazfu.com 的实现只返回 trade_no，不含 qrcode/payurl；
// 支付页面 URL 需要拼接为 {scheme}://{host}/pay/{trade_no}。
type mapiResponse struct {
	Code    int    `json:"code"`     // 1=成功
	Msg     string `json:"msg"`
	TradeNo string `json:"trade_no"` // 平台订单号
}

// CreatePayment 调用易支付 mapi.php 创建订单，返回支付跳转 URL。
// mazfu.com 特有：mapi.php 返回 trade_no，支付 URL 为 {base}/pay/{trade_no}。
func (p *easypayProvider) CreatePayment(ctx context.Context, req service.PaymentRequest) (*service.PaymentResult, error) {
	// 确定支付类型（仅支持 wxpay）
	payType := "wxpay"
	if req.Provider == "alipay" {
		payType = "alipay"
	}

	notifyURL := fmt.Sprintf("%s/api/v1/payment/callback/easypay", p.cfg.CallbackBaseURL)

	params := map[string]string{
		"pid":          p.cfg.EasyPayAppID,
		"type":         payType,
		"out_trade_no": req.OrderNo,
		"notify_url":   notifyURL,
		"name":         req.Subject,
		"money":        fmt.Sprintf("%.2f", req.Amount),
	}
	params["sign"] = p.sign(params)
	params["sign_type"] = "MD5"

	// 构造请求 URL
	mapiURL := strings.TrimRight(p.cfg.EasyPayBaseURL, "/") + "/mapi.php"
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	fullURL := mapiURL + "?" + query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("easypay: build request: %w", err)
	}

	resp, err := easypayHTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("easypay: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("easypay: read response: %w", err)
	}

	var result mapiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("easypay: parse response: %w (body: %s)", err, string(body))
	}

	if result.Code != 1 {
		return nil, fmt.Errorf("easypay: order creation failed: %s", result.Msg)
	}

	if result.TradeNo == "" {
		return nil, fmt.Errorf("easypay: empty trade_no in response")
	}

	// mazfu.com：从 EasyPayBaseURL 提取 scheme://host，构造支付页面 URL
	parsedBase, err := url.Parse(p.cfg.EasyPayBaseURL)
	if err != nil {
		return nil, fmt.Errorf("easypay: invalid base URL config: %w", err)
	}
	paymentPageURL := fmt.Sprintf("%s://%s/pay/%s", parsedBase.Scheme, parsedBase.Host, result.TradeNo)

	return &service.PaymentResult{QRCodeURL: paymentPageURL}, nil
}

// ParseCallback 解析并验证易支付异步回调通知。
// 回调参数通过 GET 或 POST 传入，trade_status=TRADE_SUCCESS 表示支付成功。
func (p *easypayProvider) ParseCallback(r *http.Request) (*service.CallbackResult, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("easypay callback: parse form: %w", err)
	}

	get := func(key string) string {
		return r.FormValue(key)
	}

	receivedSign := get("sign")
	tradeStatus := get("trade_status")
	orderNo := get("out_trade_no")
	providerOrderNo := get("trade_no")
	moneyStr := get("money")

	if tradeStatus != "TRADE_SUCCESS" {
		return nil, fmt.Errorf("easypay callback: trade_status is %q, not TRADE_SUCCESS", tradeStatus)
	}

	// 收集所有参数用于验签
	params := make(map[string]string)
	for k, vs := range r.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	// 验签
	expectedSign := p.sign(params)
	if !strings.EqualFold(receivedSign, expectedSign) {
		return nil, fmt.Errorf("easypay callback: signature mismatch")
	}

	// 解析金额
	amount, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return nil, fmt.Errorf("easypay callback: invalid money %q: %w", moneyStr, err)
	}

	// 构造原始回调数据（用于审计）
	raw := r.Form.Encode()

	return &service.CallbackResult{
		OrderNo:         orderNo,
		ProviderOrderNo: providerOrderNo,
		Amount:          amount,
		Raw:             raw,
	}, nil
}
