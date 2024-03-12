package data

import (
	"context"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/begonia-org/begonia/internal/pkg/config"
	lbf "github.com/begonia-org/go-layered-bloom"
	glc "github.com/begonia-org/go-layered-cache"
	"github.com/begonia-org/go-layered-bloom/gocuckoo"
	"github.com/sirupsen/logrus"
)

type LocalCache struct {
	cache       glc.LayeredKeyValueCache
	data        *Data
	config      *config.Config
	log         *logrus.Logger
	mux         sync.Mutex
	filters     glc.LayeredFilter
	cuckoo      glc.LayeredCuckooFilter
	onceOnStart sync.Once
	kv 		glc.LayeredKeyValueCache
}

func NewLocalCache(data *Data, config *config.Config, log *logrus.Logger) *LocalCache {
	// cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	// if err != nil {
	// 	panic(err)
	// }
	// local := &LocalCache{cache: cache, data: data, config: config, log: log, filters: bf, mux: sync.Mutex{}, cuckoo: cuckoo}
	// local.LoadRemoteCache(context.Background(), config.GetAPPAccessKeyPrefix())
	
	return nil
}
func (l *LocalCache) OnStart() {
	l.onceOnStart.Do(func() {
		l.Sync()
		errChan := l.filters.OnStart(context.Background())
		go func() {
			for err := range errChan {
				l.log.Errorf("bloom filter error %v", err)
			}
		}()
	})
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

func (l *LocalCache) LoadRemoteCache(ctx context.Context, key string) {
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
			val := l.data.GetCache(ctx, key)
			err = l.cache.Set(key, []byte(val))
			if err != nil {
				l.log.Errorf("set uid blacklist cache error %v", err)
				continue
			}
		}
	}

}
func (l *LocalCache) BloomFilterTest(ctx context.Context, key string, value []byte) bool {
	return l.filters.Test(ctx, key, value)
}
func (l *LocalCache) FilterTest(ctx context.Context, key string, value []byte) bool {
	return l.cuckoo.Check(ctx, key, value)
}
func (l *LocalCache) BloomFilterAdd(ctx context.Context, key string, value []byte) error {
	m := l.config.GetBlacklistBloomM()
	errRate := l.config.GetBlacklistBloomErrRate()
	return l.filters.Add(ctx, key, value, uint(m), errRate)
}
func (l *LocalCache) FilterAdd(ctx context.Context, key string, value []byte) error {
	return l.cuckoo.Insert(ctx, key, value)
}
func (l *LocalCache) BloomFilterDel(ctx context.Context, key string, value []byte) error {
	return l.cuckoo.Delete(ctx, key,value)
}

func (l *LocalCache) Sync() {
	ticker := time.NewTicker(5 * time.Minute) // 5分钟同步一次
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			l.LoadRemoteCache(context.Background(), l.config.GetAPPAccessKeyPrefix())

		}
	}()

}
