package service

// affiliate_service_stub.go
// Minimal AffiliateService stub to satisfy payment module compilation.
// The affiliateService field in PaymentService is nil-guarded:
//   payment_fulfillment.go:372: if s.affiliateService == nil { return nil }
// Full affiliate system not implemented in this fork.
// [fork patch]

import "context"

// AffiliateService stub — only the methods called by payment code.
type AffiliateService struct{}

// AccrueInviteRebate is a no-op stub; rebates are skipped when affiliateService is nil.
func (s *AffiliateService) AccrueInviteRebate(_ context.Context, _ int64, _ float64) (float64, error) {
	return 0, nil
}
