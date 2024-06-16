package middleware

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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

func getFieldNamesFromJSONTags(input interface{}) map[string]string {
	fieldMap := make(map[string]string)

	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			if strings.Contains(jsonTag, ",") {
				jsonTag = strings.Split(jsonTag, ",")[0]
			}
			fieldMap[jsonTag] = field.Name
		}
	}

	return fieldMap
}

// FiltersFields 从FieldMask中获取过滤字段,获取待验证字段
func (p *ParamsValidatorImpl) FiltersFields(v interface{}, parent string) []string {
	fieldsMap := getFieldNamesFromJSONTags(v)
	fieldsName := make([]string, 0)
	if message, ok := v.(protoreflect.ProtoMessage); ok {
		md := message.ProtoReflect().Descriptor()

		// 遍历所有字段
		for i := 0; i < md.Fields().Len(); i++ {
			field := md.Fields().Get(i)
			// 检查字段是否是FieldMask类型
			if field.Kind() == protoreflect.MessageKind && !field.IsList() && !field.IsMap() {

				// 获取字段的值（确保它是FieldMask类型）
				fieldValue := message.ProtoReflect().Get(field).Message()
				mask, ok := fieldValue.Interface().(*fieldmaskpb.FieldMask)
				if mask == nil || !ok {
					continue
				}
				for _, path := range mask.Paths {
					if fd := message.ProtoReflect().Descriptor().Fields().ByJSONName(path); fd != nil {
						fieldName := ""
						if parent != "" {
							fieldName = parent + "." + fieldsMap[fd.JSONName()]
						} else {
							fieldName = fieldsMap[fd.JSONName()]
						}
						if fd.Kind() == protoreflect.MessageKind {
							if fd.IsList() {
								for j := 0; j < message.ProtoReflect().Get(fd).List().Len(); j++ {
									if fd.Kind() == protoreflect.MessageKind {
										fieldsName = append(fieldsName, p.FiltersFields(message.ProtoReflect().Get(fd).List().Get(j).Message().Interface(), fmt.Sprintf("%s[%d]", fieldName, j))...)
									} else {
										fieldsName = append(fieldsName, fmt.Sprintf("%s[%d]", fieldName, j))
									}
								}
							} else {
								fieldsName = append(fieldsName, p.FiltersFields(message.ProtoReflect().Get(fd).Message().Interface(), fieldName)...)
							}
							// fieldsName = append(fieldsName, p.FiltersFields(message.ProtoReflect().Get(fd).Interface(), fieldName)...)
						} else {
							fieldsName = append(fieldsName, fieldName)
						}
					}

				}
			}
		}
		return fieldsName
	}
	return nil
}
func RegisterCustomValidators(v *validator.Validate) {
	_=v.RegisterValidation("required_if", requiredIf)
}

// requiredIf 自定义验证器逻辑
func requiredIf(fl validator.FieldLevel) bool {
	param := fl.Param()
	field := fl.Field()

	// 获取参数字段值
	paramField := fl.Parent().FieldByName(param)

	// 如果参数字段为空，当前字段必须非空
	if paramField.String() == "" {
		return field.String() != ""
	}

	return true
}
func (p *ParamsValidatorImpl) ValidateParams(v interface{}) error {
	validate := validator.New()
	RegisterCustomValidators(validate)
	err := validate.Struct(v)
	filters := p.FiltersFields(v, "")
	if len(filters) > 0 {
		err = validate.StructPartial(v, filters...)
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		clientMsg := fmt.Sprintf("params %s validation failed with %v,except %s,%v", errs[0].Field(), errs[0].Value(), errs[0].ActualTag(), filters)
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
	defer validatePluginStreamPool.Put(validateStream)

	validateStream.ServerStream = ss
	validateStream.validator = p
	validateStream.ctx = ss.Context()
	err := handler(srv, validateStream)
	return err
}

func NewParamsValidator() ParamsValidator {
	return &ParamsValidatorImpl{}
}
