// backend/internal/service/payment_expiry_service.go
package service

import (
	"context"
	"log"
	"sync"
	"time"
)

type PaymentExpiryService struct {
	orderRepo PaymentOrderRepository
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewPaymentExpiryService(orderRepo PaymentOrderRepository, interval time.Duration) *PaymentExpiryService {
	return &PaymentExpiryService{
		orderRepo: orderRepo,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (s *PaymentExpiryService) Start() {
	if s == nil || s.orderRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *PaymentExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *PaymentExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expired, err := s.orderRepo.ExpirePendingOrders(ctx)
	if err != nil {
		log.Printf("[PaymentExpiry] Failed to expire pending orders: %v", err)
		return
	}
	if expired > 0 {
		log.Printf("[PaymentExpiry] Expired %d pending orders", expired)
	}
}
