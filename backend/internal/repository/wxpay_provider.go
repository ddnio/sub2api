package repository

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type wxpayProvider struct {
	client     *core.Client
	handler    *notify.Handler
	cfg        config.PaymentConfig
	privateKey *rsa.PrivateKey
}

// NewWxpayProvider 初始化微信支付 Native Pay v3 Provider。
// 在 Wire 启动时调用，会向 SDK 全局 DownloaderManager 注册平台证书下载器。
func NewWxpayProvider(cfg config.PaymentConfig) (service.PaymentProvider, error) {
	if cfg.WxpayMchID == "" || cfg.WxpayAppID == "" || cfg.WxpayApiV3Key == "" ||
		cfg.WxpayPrivateKey == "" || cfg.WxpaySerialNo == "" {
		return nil, fmt.Errorf("wxpay: missing required config fields (mch_id/app_id/api_v3_key/private_key/serial_no)")
	}

	privateKey, err := utils.LoadPrivateKey(cfg.WxpayPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("wxpay: load private key: %w", err)
	}

	ctx := context.Background()

	// WithWechatPayAutoAuthCipher 自动下载并缓存平台证书，同时启用签名/验签/加解密
	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(
			cfg.WxpayMchID,
			cfg.WxpaySerialNo,
			privateKey,
			cfg.WxpayApiV3Key,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("wxpay: create client: %w", err)
	}

	// 使用全局 DownloaderManager 的证书访问器构造回调验签 handler
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.WxpayMchID)
	handler, err := notify.NewRSANotifyHandler(
		cfg.WxpayApiV3Key,
		verifiers.NewSHA256WithRSAVerifier(certVisitor),
	)
	if err != nil {
		return nil, fmt.Errorf("wxpay: create notify handler: %w", err)
	}

	return &wxpayProvider{
		client:     client,
		handler:    handler,
		cfg:        cfg,
		privateKey: privateKey,
	}, nil
}

// CreatePayment 调用微信支付 Native Pay v3 下单接口，返回 code_url。
// code_url 是 weixin:// 协议 URL，前端用 qrcode 库渲染成二维码供用户扫码支付。
func (p *wxpayProvider) CreatePayment(ctx context.Context, req service.PaymentRequest) (*service.PaymentResult, error) {
	notifyURL := fmt.Sprintf("%s/api/v1/payment/callback/wxpay", p.cfg.CallbackBaseURL)

	// 订单有效期：与 service 层 order_expiry_sec 保持一致（默认 15 分钟）
	expireTime := time.Now().Add(15 * time.Minute)

	svc := native.NativeApiService{Client: p.client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(p.cfg.WxpayAppID),
		Mchid:       core.String(p.cfg.WxpayMchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OrderNo),
		TimeExpire:  core.Time(expireTime),
		NotifyUrl:   core.String(notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(yuanToFen(req.Amount)),
			Currency: core.String("CNY"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wxpay: prepay: %w", err)
	}
	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		return nil, fmt.Errorf("wxpay: empty code_url in prepay response")
	}

	return &service.PaymentResult{QRCodeURL: *resp.CodeUrl}, nil
}

// ParseCallback 解析并验证微信支付异步通知（POST JSON + RSA 签名）。
// SDK 自动完成：验签、AES-GCM 解密、JSON 反序列化。
func (p *wxpayProvider) ParseCallback(r *http.Request) (*service.CallbackResult, error) {
	transaction := new(payments.Transaction)
	_, err := p.handler.ParseNotifyRequest(r.Context(), r, transaction)
	if err != nil {
		return nil, fmt.Errorf("wxpay callback: parse/verify: %w", err)
	}

	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		state := "unknown"
		if transaction.TradeState != nil {
			state = *transaction.TradeState
		}
		return nil, fmt.Errorf("wxpay callback: trade_state is %q, not SUCCESS", state)
	}

	orderNo := ptrStr(transaction.OutTradeNo)
	providerOrderNo := ptrStr(transaction.TransactionId)

	// 金额：分 → 元
	amount := 0.0
	if transaction.Amount != nil && transaction.Amount.Total != nil {
		amount = fenToYuan(*transaction.Amount.Total)
	}

	raw := fmt.Sprintf("trade_state=%s&out_trade_no=%s&transaction_id=%s",
		ptrStr(transaction.TradeState),
		orderNo,
		providerOrderNo,
	)

	return &service.CallbackResult{
		OrderNo:         orderNo,
		ProviderOrderNo: providerOrderNo,
		Amount:          amount,
		Raw:             raw,
	}, nil
}

// yuanToFen 将元转换为分（微信支付金额单位为分）。
func yuanToFen(yuan float64) int64 {
	return int64(yuan*100 + 0.5) // 四舍五入
}

// fenToYuan 将分转换为元。
func fenToYuan(fen int64) float64 {
	return float64(fen) / 100.0
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
