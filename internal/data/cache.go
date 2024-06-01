package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/config"
	glc "github.com/begonia-org/go-layered-cache"
	"github.com/begonia-org/go-layered-cache/gocuckoo"
	"github.com/begonia-org/go-layered-cache/source"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/spark-lence/tiga"
)

type LayeredCache struct {
	kv          glc.LayeredKeyValueCache
	config      *config.Config
	log         logger.Logger
	mux         sync.Mutex
	filters     glc.LayeredCuckooFilter
	onceOnStart sync.Once
}

var layered *LayeredCache

func newCache(rdb *tiga.RedisDao, config *config.Config, log logger.Logger) *LayeredCache {
	kvWatcher := source.NewWatchOptions([]interface{}{config.GetKeyValuePubsubKey()})
	strategy := glc.CacheReadStrategy(config.GetMultiCacheReadStrategy())
	KvOptions := glc.LayeredBuildOptions{
		RDB:       rdb.GetClient(),
		Strategy:  glc.CacheReadStrategy(strategy),
		Watcher:   kvWatcher,
		Channel:   config.GetKeyValuePubsubKey(),
		Log:       log.Logurs(),
		KeyPrefix: config.GetKeyValuePrefix(),
	}
	kv, err := glc.NewKeyValueCache(context.Background(), KvOptions, 5*100*100)
	if err != nil {
		panic(err)

	}
	filterWatcher := source.NewWatchOptions([]interface{}{config.GetFilterPubsubKey()})
	filterOptions := glc.LayeredBuildOptions{
		RDB:       rdb.GetClient(),
		Strategy:  glc.LocalOnly,
		Watcher:   filterWatcher,
		Channel:   config.GetFilterPubsubKey(),
		Log:       log.Logurs(),
		KeyPrefix: config.GetFilterPrefix(),
	}
	filter := glc.NewLayeredCuckoo(&filterOptions, gocuckoo.CuckooBuildOptions{
		Entries:       100000,
		BucketSize:    4,
		MaxIterations: 20,
		Expansion:     2,
	})
	return &LayeredCache{kv: kv, config: config, log: log, mux: sync.Mutex{}, onceOnStart: sync.Once{}, filters: filter}
}

// NewLayeredCache creates a new layered cache only once
func NewLayeredCache(rdb *tiga.RedisDao, config *config.Config, log logger.Logger) *LayeredCache {
	onceLayered.Do(func() {
		layered = newCache(rdb, config, log)
	})
	return layered

}

func (l *LayeredCache) Set(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return l.kv.Set(ctx, key, value, exp)
}
func (l *LayeredCache) Get(ctx context.Context, key string) ([]byte, error) {
	return l.kv.Get(ctx, key)
}
func (l *LayeredCache) GetFromLocal(ctx context.Context, key string) ([]byte, error) {
	values, err := l.kv.GetFromLocal(ctx, key)
	if err != nil {
		return nil, err
	}

	for _, val := range values {
		if val, ok := val.([]byte); ok {
			return val, nil
		}
	}
	return nil, fmt.Errorf("local cache value is not found")
}
func (l *LayeredCache) Del(ctx context.Context, key string) error {
	return l.kv.Del(ctx, key)
}
func (l *LayeredCache) SetToLocal(ctx context.Context, key string, value []byte, exp time.Duration) error {
	return l.kv.SetToLocal(ctx, key, value, exp)
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
func (l *LayeredCache) Watch(ctx context.Context) {
	errChan := l.kv.Watch(ctx)
	for err := range errChan {
		l.log.Errorf(ctx, "Watch layered-cache error:%v", err)
	}

}
