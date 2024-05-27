package middleware

import (
	"context"
	"fmt"
	"log"
	"sync"

	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type validatePluginStream struct {
	grpc.ServerStream
	// fullName string
	// plugin   gosdk.RemotePlugin
	ctx       context.Context
	validator ParamsValidator
}

var validatePluginStreamPool = &sync.Pool{
	New: func() interface{} {
		return &validatePluginStream{
			// validate: validator,
		}
	},
}

type ParamsValidator interface {
	gosdk.LocalPlugin
	ValidateParams(v interface{}) error
}

type ParamsValidatorImpl struct {
	priority int
}

func (p *validatePluginStream) Context() context.Context {
	return p.ctx
}
func (p *validatePluginStream) RecvMsg(m interface{}) error {
	err := p.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}
	err = p.validator.ValidateParams(m)
	return err

}
// FiltersFields 从FieldMask中获取过滤字段,获取待验证字段
func (p *ParamsValidatorImpl) FiltersFields(v interface{}) []string {
	fields := make([]string, 0)
	if message, ok := v.(protoreflect.ProtoMessage); ok {
		md := message.ProtoReflect().Descriptor()

		// 遍历所有字段
		for i := 0; i < md.Fields().Len(); i++ {
			field := md.Fields().Get(i)

			// 检查字段是否是FieldMask类型
			if field.Message() != nil && field.Message().FullName() == "google.protobuf.FieldMask" {

				// 获取字段的值（确保它是FieldMask类型）
				fieldValue := message.ProtoReflect().Get(field).Message()
				mask := fieldValue.Interface().(*fieldmaskpb.FieldMask)
				if mask == nil {
					continue
				}
				for _, path := range mask.Paths {
					if fd := message.ProtoReflect().Descriptor().Fields().ByJSONName(path); fd != nil {
						fields = append(fields, string(fd.Name()))
					}

				}
			}
		}
		return fields
	}
	return nil
}
func (p *ParamsValidatorImpl) ValidateParams(v interface{}) error {
	validate := validator.New()
	err := validate.Struct(v)
	filters := p.FiltersFields(v)
	if len(filters) > 0 {
		err = validate.StructPartial(v, filters...)
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		clientMsg := fmt.Sprintf("params %s validation failed with %v,except %s,%v", errs[0].Field(), errs[0].Value(), errs[0].ActualTag(),filters)
		log.Print(clientMsg)
		return gosdk.NewError(fmt.Errorf("params %s validation failed: %v", errs[0].Field(), errs[0].Value()), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "params_validation", gosdk.WithClientMessage(clientMsg))
	}
	return nil
}

func (p *ParamsValidatorImpl) SetPriority(priority int) {
	p.priority = priority
}

func (p *ParamsValidatorImpl) Priority() int {
	return p.priority
}
func (p *ParamsValidatorImpl) Name() string {
	return "ParamsValidator"
}

func (p *ParamsValidatorImpl) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	err = p.ValidateParams(req)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (p *ParamsValidatorImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	validateStream := validatePluginStreamPool.Get().(*validatePluginStream)
	validateStream.ServerStream = ss
	validateStream.validator = p
	validateStream.ctx = ss.Context()
	err := handler(srv, validateStream)
	validatePluginStreamPool.Put(validateStream)
	return err
}

func NewParamsValidator() ParamsValidator {
	return &ParamsValidatorImpl{}
}
