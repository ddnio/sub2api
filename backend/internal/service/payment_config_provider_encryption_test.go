package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCreateProviderInstanceEncryptsStoredConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentConfigService{
		entClient:     client,
		encryptionKey: []byte("0123456789abcdef0123456789abcdef"),
	}

	instance, err := svc.CreateProviderInstance(ctx, CreateProviderInstanceRequest{
		ProviderKey: payment.TypeEasyPay,
		Name:        "encrypted-config",
		Config: map[string]string{
			"pid":       "1001",
			"pkey":      "secret-pkey",
			"apiBase":   "https://pay.example.com",
			"notifyUrl": "https://merchant.example.com/notify",
			"returnUrl": "https://merchant.example.com/return",
		},
		SupportedTypes: []string{payment.TypeAlipay},
		Enabled:        false,
	})
	require.NoError(t, err)

	saved, err := client.PaymentProviderInstance.Get(ctx, instance.ID)
	require.NoError(t, err)
	require.NotContains(t, saved.Config, "secret-pkey")

	cfg, err := svc.decryptConfig(saved.Config)
	require.NoError(t, err)
	require.Equal(t, "secret-pkey", cfg["pkey"])
}
