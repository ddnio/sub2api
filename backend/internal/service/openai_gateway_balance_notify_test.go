//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayServiceBillingDepsIncludesBalanceNotifyService(t *testing.T) {
	notifyService := &BalanceNotifyService{}
	svc := &OpenAIGatewayService{balanceNotifyService: notifyService}

	deps := svc.billingDeps()

	require.Same(t, notifyService, deps.balanceNotifyService)
}
