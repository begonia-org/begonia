package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/bsm/redislock"
	"github.com/cockroachdb/errors"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func GetRDBClient(rdb *tiga.RedisDao) *redis.Client {
	return rdb.GetClient()
}

var onceRDB sync.Once
var onceMySQL sync.Once
var onceEtcd sync.Once
var rdb *tiga.RedisDao
var mysql *tiga.MySQLDao
var etcd *tiga.EtcdDao

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
func NewEtcd(config *tiga.Configuration) *tiga.EtcdDao {
	onceEtcd.Do(func() {
		etcd = tiga.NewEtcdDao(config)
	})
	return etcd
}

var ProviderSet = wire.NewSet(NewMySQL, NewRDB, NewEtcd, GetRDBClient, NewData, NewLayeredCache, NewUserRepo, NewFileRepoImpl, NewEndpointRepoImpl, NewAppRepoImpl, NewDataOperatorRepo)

type Data struct {
	// mysql
	db *tiga.MySQLDao
	// redis
	rdb  *tiga.RedisDao
	etcd *tiga.EtcdDao
}
type SourceType interface {
	// 获取数据源类型
	GetUpdateMask() *fieldmaskpb.FieldMask
	GetKey() string
	GetOwner() string
}

func NewData(mysql *tiga.MySQLDao, rdb *tiga.RedisDao, etcd *tiga.EtcdDao) *Data {
	return &Data{db: mysql, rdb: rdb, etcd: etcd}
}
func (d *Data) Get(model interface{}, data interface{}, conds ...interface{}) error {
	return d.db.GetModel(model).First(data, conds...).Error
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

//	func (d *Data) Get(model interface{}, data interface{}, conds ...interface{}) error {
//		return d.db.GetModel(model).First(data, conds...).Error
//	}
func (d *Data) Create(model interface{}) error {
	return d.db.Create(model)
}
func (d *Data) CreateInBatches(models []SourceType) error {

	db, err := d.CreateInBatchesByTx(models)
	if err != nil {
		return err
	}
	return db.Commit().Error
}
func (d *Data) CreateInBatchesByTx(models interface{}) (*gorm.DB, error) {
	db := d.db.Begin()
	size, _ := tiga.GetElementCount(models)
	if size == 0 {
		return nil, fmt.Errorf("数据为空")
	}
	err := db.CreateInBatches(models, size).Error

	if err != nil {
		db.Rollback()
		return nil, fmt.Errorf("批量插入数据失败: %w", err)
	}
	resources := make([]*common.Resource, 0)
	sources := NewSourceTypeArray(models)
	for _, item := range sources {
		name, err := d.db.TableName(item)
		if err != nil {
			db.Rollback()
			return nil, fmt.Errorf("获取表名失败: %w", err)
		}
		resources = append(resources, &common.Resource{
			ResourceKey:  item.GetKey(),
			ResourceName: name,
			Uid:          item.GetOwner(),
			CreatedAt:    timestamppb.New(time.Now()),
			UpdatedAt:    timestamppb.New(time.Now()),
		})
	}
	// logger.Logger.Infoln("resources", resources)
	err = db.CreateInBatches(resources, len(resources)).Error
	if err != nil {
		db.Rollback()
		return nil, fmt.Errorf("批量插入资源失败: %w", err)
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
func (d *Data) Update(ctx context.Context, model SourceType, dataModel interface{}) error {
	paths := make([]string, 0)
	updateMask := model.GetUpdateMask()
	if updateMask != nil {
		paths = updateMask.Paths
		if !tiga.ArrayContainsString(paths, "updated_at") {
			paths = append(paths, "updated_at")
		}
	}
	query := d.db.GetModel(dataModel).Where("uid=?", model.GetKey())
	if len(paths) > 0 {
		query = query.Select(paths)
	}
	err := query.Updates(model).Error
	if err != nil {
		return errors.Wrap(err, "更新失败")
	}
	return nil
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
func (d *Data) EtcdPut(ctx context.Context, key string, value string, opts ...clientv3.OpOption) error {
	return d.etcd.Put(ctx, key, value, opts...)
}
func (d *Data) PutEtcdWithTxn(ctx context.Context, ops []clientv3.Op) (bool, error) {
	return d.etcd.BatchOps(ctx, ops)
}
func NewSourceTypeArray(models interface{}) []SourceType {
	items := tiga.GetArrayOrSlice(models)
	sources := make([]SourceType, 0)
	for _, item := range items {
		sources = append(sources, item.(SourceType))
	}
	return sources
}
func (d *Data) EtcdGet(ctx context.Context, key string) (string, error) {
	return d.etcd.GetString(ctx, key)
}
