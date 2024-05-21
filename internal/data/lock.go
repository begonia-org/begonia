package data

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type dataLock struct {
	// lock
	lock   *redislock.Lock
	client *redislock.Client
	key    string
	ttl    time.Duration
	retry  int
}

func NewDataLock(client *redis.Client, key string, ttl time.Duration, retry int) biz.DataLock {
	return &dataLock{client: redislock.New(client), key: key, ttl: ttl, retry: retry}
}
func (d *dataLock) UnLock(ctx context.Context) error {
	return d.lock.Release(ctx)
}

func (d *dataLock) Lock(ctx context.Context) error {
	lock, err := d.client.Obtain(ctx, d.key, d.ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(time.Second*2), d.retry),
	})
	if err != nil {
		return err
	}
	// if !d.lock.() {
	// 	return redislock.ErrLockNotObtained
	// }
	d.lock = lock

	return nil
}
