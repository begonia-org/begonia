package data

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/sirupsen/logrus"
)

type LocalCache struct {
	cache  *bigcache.BigCache
	data   *Data
	config *config.Config
	log    *logrus.Logger
}

func NewLocalCache(data *Data, config *config.Config, log *logrus.Logger) *LocalCache {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	if err != nil {
		panic(err)
	}
	local := &LocalCache{cache: cache, data: data, config: config, log: log}
	local.LoadRemoteCache(context.Background(),config.GetUserBlackListPrefix())
	local.LoadRemoteCache(context.Background(),config.GetAPPAccessKeyPrefix())
	local.Sync()
	return local
}

func (l *LocalCache) Set(ctx context.Context, key string, value []byte) error {
	return l.cache.Set(key, value)
}
func (l *LocalCache) Get(ctx context.Context, key string) ([]byte, error) {
	return l.cache.Get(key)
}
func (l *LocalCache) Del(ctx context.Context, key string) error {
	return l.cache.Delete(key)
}
func (l *LocalCache) LoadRemoteCache(ctx context.Context,key string) {
	var cursor uint64 = 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		keys, cursor, err := l.data.ScanCache(ctx, cursor, key, 100)
		cancel()
		if cursor == 0 {
			break
		}

		if err != nil {
			l.log.Errorf("scan cache error %v", err)
			break
		}
		for _, key := range keys {
			val:=l.data.GetCache(ctx, key)
			err = l.cache.Set(key, []byte(val))
			if err != nil {
				l.log.Errorf("set uid blacklist cache error %v", err)
				continue
			}
		}
	}

}

// 同步黑名单
func (l *LocalCache) Sync() {
	ticker := time.NewTicker(5 * time.Minute) // 5分钟同步一次
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			l.LoadRemoteCache(context.Background(),l.config.GetUserBlackListPrefix())
			l.LoadRemoteCache(context.Background(),l.config.GetAPPAccessKeyPrefix())

		}
	}()

}
