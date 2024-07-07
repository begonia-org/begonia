package biz

import (
	"context"

	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type CURD interface {
	Add(ctx context.Context, model Model, needEncrypt bool, tx *gorm.DB) error

	Get(ctx context.Context, model interface{}, needDecrypt bool, query string, args ...interface{}) error

	Update(ctx context.Context, model Model, needEncrypt bool, tx *gorm.DB) error
	Del(ctx context.Context, model interface{}, needEncrypt bool, tx *gorm.DB) error
	List(ctx context.Context, models interface{}, pagination *tiga.Pagination) error
	BeginTx(ctx context.Context) *gorm.DB
}
type SourceType interface {
	// 获取数据源类型
	GetUpdateMask() *fieldmaskpb.FieldMask
}
type Model interface {
	SourceType
	ProtoReflect() protoreflect.Message
	GetCreatedAt() *timestamppb.Timestamp
	GetUpdatedAt() *timestamppb.Timestamp
}

type DeleteModel interface {
	Model
	GetIsDeleted() bool
}
