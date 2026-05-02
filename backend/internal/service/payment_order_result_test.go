package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestBuildCreateOrderResponseDefaultsToOrderCreated(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	resp := buildCreateOrderResponse(
		&dbent.PaymentOrder{
			ID:         42,
			Amount:     12.34,
			FeeRate:    0.03,
			ExpiresAt:  expiresAt,
			OutTradeNo: "sub2_42",
		},
		CreateOrderRequest{PaymentType: payment.TypeWxpay},
		12.71,
		&payment.InstanceSelection{PaymentMode: "qrcode"},
		&payment.CreatePaymentResponse{
			TradeNo: "sub2_42",
			QRCode:  "weixin://wxpay/bizpayurl?pr=test",
		},
		payment.CreatePaymentResultOrderCreated,
	)

	if resp.ResultType != payment.CreatePaymentResultOrderCreated {
		t.Fatalf("result type = %q, want %q", resp.ResultType, payment.CreatePaymentResultOrderCreated)
	}
	if resp.OutTradeNo != "sub2_42" {
		t.Fatalf("out_trade_no = %q, want %q", resp.OutTradeNo, "sub2_42")
	}
	if resp.QRCode != "weixin://wxpay/bizpayurl?pr=test" {
		t.Fatalf("qr_code = %q, want %q", resp.QRCode, "weixin://wxpay/bizpayurl?pr=test")
	}
	if resp.JSAPI != nil || resp.JSAPIPayload != nil {
		t.Fatal("order_created response should not include jsapi payload")
	}
	if !resp.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expires_at = %v, want %v", resp.ExpiresAt, expiresAt)
	}
}

func TestBuildCreateOrderResponseCopiesJSAPIPayload(t *testing.T) {
	t.Parallel()

	jsapiPayload := &payment.WechatJSAPIPayload{
		AppID:     "wx123",
		TimeStamp: "1712345678",
		NonceStr:  "nonce-123",
		Package:   "prepay_id=wx123",
		SignType:  "RSA",
		PaySign:   "signed-payload",
	}
	resp := buildCreateOrderResponse(
		&dbent.PaymentOrder{
			ID:         88,
			Amount:     66.88,
			FeeRate:    0.01,
			ExpiresAt:  time.Date(2026, 4, 16, 13, 0, 0, 0, time.UTC),
			OutTradeNo: "sub2_88",
		},
		CreateOrderRequest{PaymentType: payment.TypeWxpay},
		67.55,
		&payment.InstanceSelection{PaymentMode: "popup"},
		&payment.CreatePaymentResponse{
			TradeNo:    "sub2_88",
			ResultType: payment.CreatePaymentResultJSAPIReady,
			JSAPI:      jsapiPayload,
		},
		payment.CreatePaymentResultJSAPIReady,
	)

	if resp.ResultType != payment.CreatePaymentResultJSAPIReady {
		t.Fatalf("result type = %q, want %q", resp.ResultType, payment.CreatePaymentResultJSAPIReady)
	}
	if resp.JSAPI == nil || resp.JSAPIPayload == nil {
		t.Fatal("expected jsapi payload aliases to be populated")
	}
	if resp.JSAPI != jsapiPayload || resp.JSAPIPayload != jsapiPayload {
		t.Fatal("expected jsapi aliases to preserve the original pointer")
	}
}

func TestMaybeBuildWeChatOAuthRequiredResponseSkipsOAuthInFork(t *testing.T) {
	t.Setenv("PAYMENT_RESUME_SIGNING_KEY", "0123456789abcdef0123456789abcdef")

	svc := newWeChatPaymentOAuthTestService(nil)

	resp, err := svc.maybeBuildWeChatOAuthRequiredResponse(context.Background(), CreateOrderRequest{
		Amount:          12.5,
		PaymentType:     payment.TypeWxpay,
		IsWeChatBrowser: true,
		SrcURL:          "https://merchant.example/payment?from=wechat",
		OrderType:       payment.OrderTypeBalance,
	}, 12.5, 12.88, 0.03)
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWeChatBrowserWithoutOpenIDDoesNotRequireJSAPICompatibility(t *testing.T) {
	t.Parallel()

	req := CreateOrderRequest{
		PaymentType:     payment.TypeWxpay,
		IsWeChatBrowser: true,
	}

	if requestNeedsWeChatJSAPICompatibility(req) {
		t.Fatal("WeChat browser without OpenID should not require JSAPI compatibility")
	}
	if requiresWeChatJSAPICompatibleSelection(req, &payment.InstanceSelection{ProviderKey: payment.TypeWxpay}) {
		t.Fatal("WeChat browser without OpenID should not require a JSAPI-compatible provider")
	}
}

func TestWeChatOpenIDRequiresJSAPICompatibility(t *testing.T) {
	t.Parallel()

	req := CreateOrderRequest{
		PaymentType: payment.TypeWxpay,
		OpenID:      "openid-123",
	}

	if !requestNeedsWeChatJSAPICompatibility(req) {
		t.Fatal("OpenID wxpay request should require JSAPI compatibility")
	}
	if !requiresWeChatJSAPICompatibleSelection(req, &payment.InstanceSelection{ProviderKey: payment.TypeWxpay}) {
		t.Fatal("OpenID wxpay request should require a JSAPI-compatible provider")
	}
}

func TestMaybeBuildWeChatOAuthRequiredResponseForSelectionSkipsEasyPayProvider(t *testing.T) {
	svc := newWeChatPaymentOAuthTestService(nil)

	resp, err := svc.maybeBuildWeChatOAuthRequiredResponseForSelection(context.Background(), CreateOrderRequest{
		Amount:          12.5,
		PaymentType:     payment.TypeWxpay,
		IsWeChatBrowser: true,
		OrderType:       payment.OrderTypeBalance,
	}, 12.5, 12.88, 0.03, &payment.InstanceSelection{
		ProviderKey: payment.TypeEasyPay,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
}

func newWeChatPaymentOAuthTestService(values map[string]string) *PaymentService {
	return &PaymentService{
		configService: &PaymentConfigService{
			settingRepo:   &paymentConfigSettingRepoStub{values: values},
			encryptionKey: []byte("0123456789abcdef0123456789abcdef"),
		},
	}
}
