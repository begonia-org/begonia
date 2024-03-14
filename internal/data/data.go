package data

import (
	"context"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/cockroachdb/errors"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

func GetRDBClient(rdb *tiga.RedisDao) *redis.Client {
	return rdb.GetClient()
}

var onceRDB sync.Once
var onceMySQL sync.Once
var rdb *tiga.RedisDao
var mysql *tiga.MySQLDao

func NewRDB(config *tiga.Configuration) *tiga.RedisDao {
	onceRDB.Do(func() {
		rdb = tiga.NewRedisDao(config)
	})
	return rdb
}
func NewMySQL(config *tiga.Configuration) *tiga.MySQLDao {
	onceMySQL.Do(func() {
		mysql = tiga.NewMySQLDao(config)
		mysql.RegisterTimeSerializer()
	})
	return mysql

}

var ProviderSet = wire.NewSet(NewMySQL, NewRDB, GetRDBClient, NewData, NewLayeredCache, NewUserRepo, NewFileRepoImpl, NewEndpointRepoImpl)

type Data struct {
	// mysql
	db *tiga.MySQLDao
	// redis
	rdb *tiga.RedisDao
	// etcd *tiga.EtcdDao

}
type SourceType interface {
	// 获取数据源类型
	GetUpdateMask() *fieldmaskpb.FieldMask
	GetKey() string
	GetOwner() string
}

func NewData(mysql *tiga.MySQLDao, rdb *tiga.RedisDao) *Data {
	return &Data{db: mysql, rdb: rdb}
}

func (d *Data) List(model interface{}, data interface{}, conds ...interface{}) error {
	queryTag := tiga.QueryTags{}
	query := queryTag.BuildConditions(d.db.GetModel(model), conds)
	if query == nil {
		query = d.db.GetModel(model).Find(data, conds...)
	}
	if err := query.Find(data).Error; err != nil {
		return err
	} else {
		return nil
	}
}
// func (d *Data) Get(model interface{}, data interface{}, conds ...interface{}) error {
// 	return d.db.GetModel(model).First(data, conds...).Error
// }
func (d *Data) Create(model interface{}) error {
	return d.db.Create(model)
}
func (d *Data) CreateInBatches(models []SourceType) error {
	// db := d.db.Begin()

	// err := db.CreateInBatches(models, len(models)).Error
	// if err != nil {
	// 	db.Rollback()
	// 	return errors.Wrap(err, "批量插入失败")
	// }
	db, err := d.CreateInBatchesByTx(models)
	if err != nil {
		return err
	}
	return db.Commit().Error
}
func (d *Data) CreateInBatchesByTx(models []SourceType) (*gorm.DB, error) {
	db := d.db.Begin()

	err := db.CreateInBatches(models, len(models)).Error
	if err != nil {
		db.Rollback()
		return nil, errors.Wrap(err, "批量插入失败")
	}
	return db, nil

}
func (d *Data) BatchUpdates(models []SourceType, dataModel interface{}) error {
	size, err := tiga.GetElementCount(models)
	if err != nil {
		return errors.Wrap(err, "获取元素数量失败")
	}
	if size == 0 {
		return nil
	}
	model, err := tiga.GetFirstElement(models)

	if size == 1 {
		if err != nil {
			return errors.New("获取第一个元素失败")
		}
		return d.db.Update(model, model)
	}
	db := d.db.Begin()
	for _, item := range models {
		// 更新任务信息
		paths := make([]string, 0)
		updateMask := item.GetUpdateMask()
		if updateMask != nil {
			paths = updateMask.Paths
			if !tiga.ArrayContainsString(paths, "updated_at") {
				paths = append(paths, "updated_at")
			}
		}
		query := db.Model(dataModel).Where("uid=?", item.GetKey())
		if len(paths) > 0 {
			query = query.Select(paths)
		}
		err := query.Updates(item).Error
		if err != nil {
			db.Rollback()
			return errors.Wrap(err, "批量更新失败")
		}
	}
	return db.Commit().Error
}
func (d *Data) Cache(ctx context.Context, key string, value string, exp time.Duration) error {
	return d.rdb.Set(ctx, key, value, exp)
}
func (d *Data) GetCache(ctx context.Context, key string) string {
	return d.rdb.Get(ctx, key)
}
func (d *Data) BatchCacheByTx(ctx context.Context, exp time.Duration, kv ...interface{}) redis.Pipeliner {
	pipe := d.rdb.GetClient().TxPipeline()
	for i := 0; i < len(kv); i += 2 {
		pipe.Set(ctx, kv[i].(string), kv[i+1], exp)
	}
	return pipe
}
func (d *Data) DelCacheByTx(ctx context.Context, keys ...string) redis.Pipeliner {
	pipe := d.rdb.GetClient().TxPipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	return pipe
}
func (d *Data) DelCache(ctx context.Context, key string) error {
	return d.rdb.Del(ctx, key)
}
func (d *Data) ScanCache(ctx context.Context, cur uint64, prefix string, count int64) ([]string, uint64, error) {
	return d.rdb.Scan(ctx, cur, count, prefix)
}
func (d *Data) Lock(ctx context.Context, key string, expiration time.Duration) (*redislock.Lock, error) {
	return d.rdb.Lock(ctx, key, expiration)
}
func (d *Data) BatchDelete(models []SourceType, dataModel interface{}) error {
	if len(models) == 0 {
		return nil
	}
	if len(models) == 1 {
		return d.db.Delete(models[0], "uid=?", models[0].GetKey())
	}
	keys := make([]string, 0)
	for _, item := range models {
		keys = append(keys, item.GetKey())
	}
	return d.db.Delete(dataModel, "uid in ?", keys)
}

func NewSourceTypeArray(models interface{}) []SourceType {
	items := tiga.GetArrayOrSlice(models)
	sources := make([]SourceType, 0)
	for _, item := range items {
		sources = append(sources, item.(SourceType))
	}
	return sources
}
