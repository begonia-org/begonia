package data

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func GetRDBClient(rdb *tiga.RedisDao) *redis.Client {
	return rdb.GetClient()
}

var onceRDB sync.Once
var onceMySQL sync.Once
var onceEtcd sync.Once
var onceLayered sync.Once
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
	})
	mysql.RegisterTimeSerializer()

	return mysql

}
func NewEtcd(config *tiga.Configuration) *tiga.EtcdDao {
	onceEtcd.Do(func() {
		etcd = tiga.NewEtcdDao(config)
	})
	return etcd
}

var ProviderSet = wire.NewSet(NewMySQL,
	NewRDB,
	NewEtcd,
	GetRDBClient,
	NewData,
	NewCurdImpl,
	NewLayeredCache,

	NewDataLock,
	NewAuthzRepoImpl,
	NewUserRepoImpl,
	NewEndpointRepoImpl,
	NewAppRepoImpl,
	NewDataOperatorRepo)

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
}

func NewData(mysql *tiga.MySQLDao, rdb *tiga.RedisDao, etcd *tiga.EtcdDao) *Data {
	return &Data{db: mysql, rdb: rdb, etcd: etcd}
}

// func (d *Data) CreateInBatches(models []SourceType) error {

// 	db, err := d.CreateInBatchesByTx(models)
// 	if err != nil {
// 		return err
// 	}
// 	return db.Commit().Error
// }
// func (d *Data) newResource(model SourceType) (*common.Resource, error) {
// 	name, err := d.db.TableName(model)
// 	if err != nil {
// 		return nil, fmt.Errorf("get table name failed: %w", err)
// 	}
// 	_, resourceKey, err := getPrimaryColumnValue(model, "primary")
// 	if err != nil || resourceKey == nil {
// 		return nil, fmt.Errorf("get primary column value failed: %w", err)

// 	}
// 	_, owner, err := getPrimaryColumnValue(model, "owner")
// 	if err != nil || owner == nil {
// 		return nil, fmt.Errorf("get owner column value failed: %w", err)

// 	}
// 	return &common.Resource{
// 		ResourceKey:  resourceKey.(string),
// 		ResourceName: name,
// 		Uid:          owner.(string),
// 		CreatedAt:    timestamppb.New(time.Now()),
// 		UpdatedAt:    timestamppb.New(time.Now()),
// 	}, nil
// }
// func (d *Data) CreateInBatchesByTx(models interface{}) (*gorm.DB, error) {
// 	db := d.db.Begin()
// 	size, _ := tiga.GetElementCount(models)
// 	if size == 0 {
// 		return nil, fmt.Errorf("数据为空")
// 	}
// 	err := db.CreateInBatches(models, size).Error

// 	if err != nil {
// 		db.Rollback()
// 		return nil, fmt.Errorf("批量插入数据失败: %w", err)
// 	}
// 	resources := make([]*common.Resource, 0)
// 	sources := NewSourceTypeArray(models)
// 	for _, item := range sources {
// 		_, err := d.db.TableName(item)
// 		if err != nil {
// 			db.Rollback()
// 			return nil, fmt.Errorf("获取表名失败: %w", err)
// 		}
// 		resource, err := d.newResource(item)
// 		if err != nil {
// 			return nil, fmt.Errorf("new resource failed: %w", err)
// 		}
// 		resources = append(resources, resource)
// 	}
// 	// logger.Logger.Infoln("resources", resources)
// 	err = db.CreateInBatches(resources, len(resources)).Error
// 	if err != nil {
// 		db.Rollback()
// 		return nil, fmt.Errorf("批量插入资源失败: %w", err)
// 	}
// 	return db, nil
// }
// func (d *Data) BatchUpdates(ctx context.Context, models []SourceType) error {
// 	size, err := tiga.GetElementCount(models)
// 	if err != nil {
// 		return errors.Wrap(err, "获取元素数量失败")
// 	}
// 	if size == 0 {
// 		return nil
// 	}
// 	model, err := tiga.GetFirstElement(models)

//		if size == 1 {
//			if err != nil {
//				return fmt.Errorf("获取第一个元素失败")
//			}
//			return d.db.Update(ctx, model, model)
//		}
//		db := d.db.Begin()
//		for _, item := range models {
//			// 更新任务信息
//			paths := make([]string, 0)
//			updateMask := item.GetUpdateMask()
//			if updateMask != nil {
//				paths = updateMask.Paths
//				if !tiga.ArrayContainsString(paths, "updated_at") {
//					paths = append(paths, "updated_at")
//				}
//			}
//			key, value, err := getPrimaryColumnValue(item, "primary")
//			if err != nil {
//				db.Rollback()
//				return errors.Wrap(err, "get primary column value failed")
//			}
//			query := db.Model(item).Where(fmt.Sprintf("%s=?", key), value)
//			if len(paths) > 0 {
//				query = query.Select(paths)
//			}
//			err = query.Updates(item).Error
//			if err != nil {
//				db.Rollback()
//				return errors.Wrap(err, "批量更新失败")
//			}
//		}
//		return db.Commit().Error
//	}
func getPrimaryColumnValue(model interface{}, tagName string) (string, interface{}, error) {
	// 获取结构体类型
	modelType := reflect.TypeOf(model)
	modelVal := reflect.ValueOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
		modelVal = modelVal.Elem()

	}
	if modelType.Kind() != reflect.Struct {
		return "", "", fmt.Errorf("%s not a struct type", modelType.Kind().String())
	}

	// 遍历结构体的字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		// 检查字段名称是否匹配
		if field.Tag.Get(tagName) != "" {
			// 获取字段值
			tag := field.Tag.Get("gorm")
			// 	// 解析GORM标签
			tagParts := strings.Split(tag, ";")
			for _, part := range tagParts {
				kv := strings.Split(part, ":")
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "column" {
					value := modelVal.Field(i).Interface()
					return strings.TrimSpace(kv[1]), value, nil
				}
			}
		}

	}
	return "", nil, fmt.Errorf("not found primary column")
}

// func (d *Data) Update(ctx context.Context, model SourceType) error {
// 	paths := make([]string, 0)
// 	updateMask := model.GetUpdateMask()
// 	if updateMask != nil {
// 		paths = updateMask.Paths
// 	}
// 	key, val, err := getPrimaryColumnValue(model, "primary")
// 	if err != nil {
// 		return errors.Wrap(err, "get column value failed")

// 	}
// 	err = d.db.UpdateSelectColumns(ctx, fmt.Sprintf("%s=%s", key, val), model, paths...)

//		if err != nil {
//			return errors.Wrap(err, "更新失败")
//		}
//		return nil
//	}
//
//	func (d *Data) Cache(ctx context.Context, key string, value string, exp time.Duration) error {
//		return d.rdb.Set(ctx, key, value, exp)
//	}
//
//	func (d *Data) GetCache(ctx context.Context, key string) string {
//		return d.rdb.Get(ctx, key)
//	}
func (d *Data) BatchCacheByTx(ctx context.Context, exp time.Duration, kv ...interface{}) redis.Pipeliner {
	pipe := d.rdb.GetClient().TxPipeline()
	for i := 0; i < len(kv); i += 2 {
		pipe.Set(ctx, kv[i].(string), kv[i+1], exp)
	}
	return pipe
}

// func (d *Data) DelCacheByTx(ctx context.Context, keys ...string) redis.Pipeliner {
// 	pipe := d.rdb.GetClient().TxPipeline()
// 	for _, key := range keys {
// 		pipe.Del(ctx, key)
// 	}
// 	return pipe
// }

//	func (d *Data) BatchEtcdDelete(models []SourceType) error {
//		if len(models) == 0 {
//			return nil
//		}
//		if len(models) == 1 {
//			key, val, err := getPrimaryColumnValue(models[0], "primary")
//			if err != nil {
//				return fmt.Errorf("get primary column value failed: %w", err)
//			}
//			return d.db.Delete(models[0], fmt.Sprintf("%s=?", key), val)
//		}
//		keys := make([]string, 0)
//		dataModel := models[0]
//		_in := "uid in ?"
//		for _, item := range models {
//			key, val, err := getPrimaryColumnValue(item, "primary")
//			_in = fmt.Sprintf("%s in ?", key)
//			if err != nil {
//				return fmt.Errorf("get primary column value failed: %w", err)
//			}
//			keys = append(keys, val.(string))
//		}
//		return d.db.Delete(dataModel, _in, keys)
//	}
//
//	func (d *Data) EtcdPut(ctx context.Context, key string, value string, opts ...clientv3.OpOption) error {
//		return d.etcd.Put(ctx, key, value, opts...)
//	}
func (d *Data) PutEtcdWithTxn(ctx context.Context, ops []clientv3.Op) (bool, error) {
	return d.etcd.BatchOps(ctx, ops)
}

// func NewSourceTypeArray(models interface{}) []SourceType {
// 	items := tiga.GetArrayOrSlice(models)
// 	sources := make([]SourceType, 0)
// 	for _, item := range items {
// 		sources = append(sources, item.(SourceType))
// 	}
// 	return sources
// }
// func (d *Data) EtcdGet(ctx context.Context, key string) (string, error) {
// 	return d.etcd.GetString(ctx, key)
// }
