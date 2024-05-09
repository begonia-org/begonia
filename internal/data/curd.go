package data

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/cockroachdb/errors"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type curdImpl struct {
	db   *tiga.MySQLDao
	conf *config.Config
}


func NewCurdImpl(db *tiga.MySQLDao, conf *config.Config) biz.CURD {
	return &curdImpl{db: db, conf: conf}
}
func (c *curdImpl) SetDatetimeAt(model biz.Model, jsonName string) error {
	field := model.ProtoReflect().Descriptor().Fields().ByJSONName(jsonName)
	if field == nil {
		return fmt.Errorf("field %s not found", jsonName)

	}
	val:= timestamppb.Now()
	model.ProtoReflect().Set(field, protoreflect.ValueOfMessage(val.ProtoReflect()))
	return nil
}
func (c *curdImpl) SetBoolean(model biz.DeleteModel, jsonName string) error {
	field := model.ProtoReflect().Descriptor().Fields().ByJSONName(jsonName)
	if field == nil {
		return fmt.Errorf("field %s not found", jsonName)
	}
	model.ProtoReflect().Set(field, protoreflect.ValueOfBool(true))
	return nil
}
func (c *curdImpl) Add(ctx context.Context, model biz.Model,needEncrypt bool) error {
	if err := c.SetDatetimeAt(model, "created_at"); err != nil {
		return err
	}
	if err := c.SetDatetimeAt(model, "updated_at"); err != nil {
		return err
	}
	if needEncrypt {
		ivKey := c.conf.GetAesIv()
		aseKey := c.conf.GetAesKey()

		err := tiga.EncryptStructAES([]byte(aseKey), model, ivKey)
		if err != nil {
			return fmt.Errorf("encrypt model struct failed:%w", err)

		}
	
	}
	return c.db.Create(ctx,model)
}
func (c *curdImpl) Get(ctx context.Context, model interface{},needDecrypt bool, query string, args ...interface{}) error {
	if _, ok := model.(biz.DeleteModel); ok {
		query = fmt.Sprintf("(%s) and is_deleted=0", query)

	}
	if err := c.db.First(ctx,model, query, args...); err != nil {
		return fmt.Errorf("get model failed: %w", err)
	}
	if needDecrypt {
		ivKey := c.conf.GetAesIv()
		aseKey := c.conf.GetAesKey()

		err := tiga.DecryptStructAES([]byte(aseKey), model, ivKey)
		if err != nil {
			return fmt.Errorf("decrypt model struct failed:%w", err)

		}
	}
	return nil
}
func (c *curdImpl) Update(ctx context.Context, model biz.Model, needEncrypt bool) error {
	paths := make([]string, 0)
	updateMask := model.GetUpdateMask()
	if updateMask != nil {
		paths = updateMask.Paths
	}
	key, val, err := getPrimaryColumnValue(model, "primary")
	if err != nil {
		return errors.Wrap(err, "get primary column value failed")
	}
	_=c.SetDatetimeAt(model, "updated_at")
	if needEncrypt {
		ivKey := c.conf.GetAesIv()
		aseKey := c.conf.GetAesKey()

		err := tiga.EncryptStructAES([]byte(aseKey), model, ivKey)
		if err != nil {
			return fmt.Errorf("encrypt model struct failed:%w", err)

		}
	}
	
	err = c.db.UpdateSelectColumns(ctx,fmt.Sprintf("%s=%s", key, val), model, paths...)
	if err != nil {
		return fmt.Errorf("update model for %s=%v failed: %w", key, val, err)
	}
	return nil
}
func (c *curdImpl) renameUniqueFields(model biz.Model) ([]string, error) {
	// 获取结构体类型
	modelType := reflect.TypeOf(model)
	modelVal := reflect.ValueOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
		modelVal = modelVal.Elem()

	}
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s not a struct type", modelType.Kind().String())
	}
	updated := make([]string, 0)
	// 遍历结构体的字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		// 检查字段名称是否匹配
		if field.Tag.Get("ondeleted") != "" {
			val := modelVal.Field(i).Interface()
			name := fmt.Sprintf("%s_%s", val, time.Now().Format("20060102150405"))
			modelVal.Field(i).SetString(name)
			tag := field.Tag.Get("gorm")
			// 	// 解析GORM标签
			tagParts := strings.Split(tag, ";")
			for _, part := range tagParts {
				kv := strings.Split(part, ":")
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "column" {
					updated = append(updated, strings.TrimSpace(kv[1]))
				}
			}

		}

	}
	return updated, nil
}
func (c *curdImpl) Del(ctx context.Context, model interface{}, needEncrypt bool) error {
	key, val, err := getPrimaryColumnValue(model, "primary")
	if err != nil {
		return errors.Wrap(err, "get primary column value failed")
	}
	if delModel, ok := model.(biz.DeleteModel); ok {
		if err := c.SetBoolean(delModel, "is_deleted"); err != nil {
			return err
		}
		_=c.SetDatetimeAt(delModel, "updated_at")
		updated, err := c.renameUniqueFields(delModel)
		if err != nil {
			return fmt.Errorf("rename unique fields failed: %w", err)
		}
		updated = append(updated, "is_deleted", "updated_at")
		if needEncrypt {
			if needEncrypt {
				ivKey := c.conf.GetAesIv()
				aseKey := c.conf.GetAesKey()

				err := tiga.EncryptStructAES([]byte(aseKey), model, ivKey)
				if err != nil {
					return fmt.Errorf("encrypt model struct failed:%w", err)

				}
			}
		}
		return c.db.UpdateSelectColumns(ctx,fmt.Sprintf("%s=%s", key, val), model, updated...)
	} else {
		return c.db.Delete(model, fmt.Sprintf("%s=?", key), val)
	}

}
func (c *curdImpl) List(ctx context.Context, models interface{}, pagination *tiga.Pagination) error {
	if _, ok := models.(biz.DeleteModel); ok {
		pagination.Query = fmt.Sprintf("(%s) and is_deleted=0", pagination.Query)
	}
	return c.db.Pagination(ctx,models, pagination)
}
