package repository

import (
	"context"
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
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type wxpayProvider struct {
	client  *core.Client
	handler *notify.Handler
	cfg     config.PaymentConfig
}

// NewWxpayProvider 初始化微信支付 Native Pay v3 Provider。
//
// 支持两种验签模式：
//   - 公钥模式（推荐，新商户默认）：配置 wxpay_public_key + wxpay_public_key_id
//   - 平台证书模式（旧商户）：无需额外配置，SDK 自动下载证书
func NewWxpayProvider(cfg config.PaymentConfig) (service.PaymentProvider, error) {
	if cfg.WxpayMchID == "" || cfg.WxpayAppID == "" || cfg.WxpayApiV3Key == "" ||
		cfg.WxpayPrivateKey == "" || cfg.WxpaySerialNo == "" {
		return nil, fmt.Errorf("wxpay: missing required config (mch_id/app_id/api_v3_key/private_key/serial_no)")
	}

	privateKey, err := utils.LoadPrivateKey(cfg.WxpayPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("wxpay: load private key: %w", err)
	}

	ctx := context.Background()

	var client *core.Client
	var handler *notify.Handler

	if cfg.WxpayPublicKey != "" && cfg.WxpayPublicKeyID != "" {
		// ── 公钥模式（新版，2024 年后新开通商户默认启用）──
		publicKey, err := utils.LoadPublicKey(cfg.WxpayPublicKey)
		if err != nil {
			return nil, fmt.Errorf("wxpay: load public key: %w", err)
		}

		client, err = core.NewClient(ctx,
			option.WithWechatPayPublicKeyAuthCipher(
				cfg.WxpayMchID,
				cfg.WxpaySerialNo,
				privateKey,
				cfg.WxpayPublicKeyID,
				publicKey,
			),
		)
		if err != nil {
			return nil, fmt.Errorf("wxpay: create client (pubkey mode): %w", err)
		}

		// 公钥模式的回调验签：使用微信支付公钥
		pubkeyVerifier := verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.WxpayPublicKeyID, *publicKey)
		handler, err = notify.NewRSANotifyHandler(cfg.WxpayApiV3Key, pubkeyVerifier)
		if err != nil {
			return nil, fmt.Errorf("wxpay: create notify handler (pubkey mode): %w", err)
		}
	} else {
		// ── 平台证书模式（旧商户）── SDK 自动下载并缓存平台证书
		client, err = core.NewClient(ctx,
			option.WithWechatPayAutoAuthCipher(
				cfg.WxpayMchID,
				cfg.WxpaySerialNo,
				privateKey,
				cfg.WxpayApiV3Key,
			),
		)
		if err != nil {
			return nil, fmt.Errorf("wxpay: create client (cert mode): %w", err)
		}

		certVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.WxpayMchID)
		handler, err = notify.NewRSANotifyHandler(cfg.WxpayApiV3Key, verifiers.NewSHA256WithRSAVerifier(certVisitor))
		if err != nil {
			return nil, fmt.Errorf("wxpay: create notify handler (cert mode): %w", err)
		}
	}

	return &wxpayProvider{
		client:  client,
		handler: handler,
		cfg:     cfg,
	}, nil
}

// CreatePayment 调用微信支付 Native Pay v3 下单接口，返回 code_url。
// code_url 是 weixin:// 协议 URL，前端用 qrcode 库渲染成二维码供用户扫码支付。
func (p *wxpayProvider) CreatePayment(ctx context.Context, req service.PaymentRequest) (*service.PaymentResult, error) {
	notifyURL := fmt.Sprintf("%s/api/v1/payment/callback/wxpay", p.cfg.CallbackBaseURL)
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

// Refund 调用微信支付退款 API（国内退款）。
// 微信退款为异步处理，Create 成功（返回 SUCCESS 或 PROCESSING）表示退款已受理。
func (p *wxpayProvider) Refund(ctx context.Context, req service.RefundRequest) (*service.RefundResult, error) {
	svc := refunddomestic.RefundsApiService{Client: p.client}

	amountFen := yuanToFen(req.Amount)

	createReq := refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.OrderNo),
		OutRefundNo: core.String(req.RefundNo),
		Reason:      core.String(req.Reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(amountFen),
			Total:    core.Int64(amountFen), // 全额退款：refund == total
			Currency: core.String("CNY"),
		},
	}

	// 优先使用微信支付订单号（更精确）
	if req.ProviderOrderNo != "" {
		createReq.TransactionId = core.String(req.ProviderOrderNo)
	}

	resp, _, err := svc.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("wxpay: refund: %w", err)
	}

	result := &service.RefundResult{}
	if resp.RefundId != nil {
		result.ProviderRefundNo = *resp.RefundId
	}
	if resp.Status != nil {
		result.Status = string(*resp.Status) // Status 是 refunddomestic.Status 枚举类型，转为 string
	}

	return result, nil
}

// yuanToFen 将元转换为分（微信支付金额单位为分）。
func yuanToFen(yuan float64) int64 {
	return int64(yuan*100 + 0.5)
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
