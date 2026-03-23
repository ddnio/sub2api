package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

type paymentCache struct {
	rdb *redis.Client
}

func NewPaymentCache(rdb *redis.Client) service.PaymentCache {
	return &paymentCache{rdb: rdb}
}

func (c *paymentCache) AcquireCallbackLock(ctx context.Context, orderNo string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("payment:lock:%s", orderNo)
	return c.rdb.SetNX(ctx, key, "1", ttl).Result()
}

func (c *paymentCache) ReleaseCallbackLock(ctx context.Context, orderNo string) error {
	key := fmt.Sprintf("payment:lock:%s", orderNo)
	return c.rdb.Del(ctx, key).Err()
}

func (c *paymentCache) GetOrderCreateCount(ctx context.Context, userID int64) (int, error) {
	key := fmt.Sprintf("payment:rate:%d", userID)
	val, err := c.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *paymentCache) IncrementOrderCreateCount(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("payment:rate:%d", userID)
	pipe := c.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}
