// backend/internal/repository/easypay_provider.go
package repository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type easypayProvider struct {
	cfg config.PaymentConfig
}

func NewEasyPayProvider(cfg config.PaymentConfig) service.PaymentProvider {
	return &easypayProvider{cfg: cfg}
}

func (p *easypayProvider) CreatePayment(ctx context.Context, req service.PaymentRequest) (*service.PaymentResult, error) {
	// TODO: 对接实际支付 API
	// 1. 构造签名参数
	// 2. 发送 HTTP 请求到 EasyPay API
	// 3. 解析返回的支付二维码 URL
	return nil, fmt.Errorf("easypay provider not implemented yet")
}

func (p *easypayProvider) ParseCallback(r *http.Request) (*service.CallbackResult, error) {
	// TODO: 对接实际支付回调
	// 1. 读取回调参数
	// 2. 验证签名
	// 3. 解析订单号和金额
	return nil, fmt.Errorf("easypay callback parser not implemented yet")
}
