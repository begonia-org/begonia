package data

import (
	"context"
	"sync"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/config"
	glc "github.com/begonia-org/go-layered-cache"
	"github.com/begonia-org/go-layered-cache/gocuckoo"
	"github.com/begonia-org/go-layered-cache/source"
	"github.com/sirupsen/logrus"
)

type LayeredCache struct {
	kv          glc.LayeredKeyValueCache
	data        *Data
	config      *config.Config
	log         *logrus.Logger
	mux         sync.Mutex
	filters     glc.LayeredCuckooFilter
	onceOnStart sync.Once
}

func NewLayeredCache(ctx context.Context, data *Data, config *config.Config, log *logrus.Logger) *LayeredCache {

	kvWatcher := source.NewWatchOptions([]interface{}{config.GetKeyValuePubsubKey()})
	strategy := glc.CacheReadStrategy(config.GetMultiCacheReadStrategy())
	KvOptions := glc.LayeredBuildOptions{
		RDB:       data.rdb.GetClient(),
		Strategy:  glc.CacheReadStrategy(strategy),
		Watcher:   kvWatcher,
		Channel:   config.GetKeyValuePubsubKey(),
		Log:       log,
		KeyPrefix: config.GetKeyValuePrefix(),
	}
	kv, err := glc.NewKeyValueCache(ctx, KvOptions, 5*100*100)
	if err != nil {
		panic(err)

	}
	filterWatcher := source.NewWatchOptions([]interface{}{config.GetFilterPubsubKey()})
	filterOptions := glc.LayeredBuildOptions{
		RDB:       data.rdb.GetClient(),
		Strategy:  glc.LocalOnly,
		Watcher:   filterWatcher,
		Channel:   config.GetFilterPubsubKey(),
		Log:       log,
		KeyPrefix: config.GetFilterPrefix(),
	}
	filter := glc.NewLayeredCuckoo(&filterOptions, gocuckoo.CuckooBuildOptions{
		Entries:       100000,
		BucketSize:    4,
		MaxIterations: 20,
		Expansion:     2,
	})
	local := &LayeredCache{kv: kv, data: data, config: config, log: log, mux: sync.Mutex{}, onceOnStart: sync.Once{}, filters: filter}
	return local
}
func (l *LayeredCache) OnStart() {
	l.onceOnStart.Do(func() {
		l.Sync()
		errChan := l.filters.Watch(context.Background())
		go func() {
			for err := range errChan {
				l.log.Errorf("bloom filter error %v", err)
			}
		}()
		errKvChan := l.kv.Watch(context.Background())
		go func() {
			for err := range errKvChan {
				l.log.Errorf("kv error %v", err)
			}
		}()
	})
}
func (l *LayeredCache) Set(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return l.kv.Set(ctx, key, value, exp)
}
func (l *LayeredCache) Get(ctx context.Context, key string) ([]byte, error) {
	return l.kv.Get(ctx, key)
}
func (l *LayeredCache) Del(ctx context.Context, key string) error {
	return l.kv.Del(ctx, key)
}

func (l *LayeredCache) LoadRemoteCache(ctx context.Context, key string) {

	err := l.kv.LoadDump(ctx)
	if err != nil {
		l.log.Errorf("load remote cache error %v", err)
	}
	err = l.filters.LoadDump(ctx)
	if err != nil {
		l.log.Errorf("load remote cache error %v", err)

	}

}

func (l *LayeredCache) CheckInFilter(ctx context.Context, key string, value []byte) (bool, error) {
	return l.filters.Check(ctx, key, value)
}

func (l *LayeredCache) AddToFilter(ctx context.Context, key string, value []byte) error {
	return l.filters.Add(ctx, key, value)
}
func (l *LayeredCache) DelInFilter(ctx context.Context, key string, value []byte) error {
	return l.filters.Del(ctx, key, value)
}

func (l *LayeredCache) Sync() {
	ticker := time.NewTicker(5 * time.Minute) // 5分钟同步一次
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			l.LoadRemoteCache(context.Background(), l.config.GetAPPAccessKeyPrefix())

		}
	}()

}
