package middleware

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
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
type ValidateError struct {
	error
	Field string
}

// type validatePluginClientStream struct {
// 	grpc.ClientStream
// 	ctx       context.Context
// 	validator ParamsValidator
// }

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
	validate *validator.Validate
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

// func getFieldNamesFromProto(input interface{}) map[string]string {}
func getFieldNamesFromJSONTags(input interface{}) map[string]string {
	fieldMap := make(map[string]string)

	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if !val.IsValid() || val.IsZero() {
		return nil

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

// isRequiredField 检查字段是否是必填字段
// 通过proto文件中的validate标签或者struct tag中的validate标签判断
func (p *ParamsValidatorImpl) isRequiredField(field interface{}) bool {
	if fd, ok := field.(protoreflect.FieldDescriptor); ok && fd != nil {

		if v, ok := proto.GetExtension(fd.Options(), common.E_Validate).(string); ok && strings.Contains(v, "required") {
			return true

		}

	}

	if fd, ok := field.(reflect.StructField); ok && fd.Tag.Get("validate") != "" {
		if v, ok := fd.Tag.Lookup("validate"); ok && strings.Contains(v, "required") {
			return true

		}

	}
	return false
}

// getValidatePath 获取待验证字段的路径
// 路径格式参考validate.StructPartial
func (p *ParamsValidatorImpl) getValidatePath(message protoreflect.ProtoMessage, field string, parent string) []string {
	fieldsName := make([]string, 0)
	md := message.ProtoReflect().Descriptor()
	// log.Printf("get validate path,field:%s,parent:%s", field, parent)
	if fd := md.Fields().ByJSONName(field); fd != nil {
		fieldName := strcase.ToCamel(string(fd.Name()))
		// fieldName := fd.JSONName()
		if parent != "" {
			fieldName = parent + "." + fieldName
		}
		fieldsName = append(fieldsName, fieldName)

		if fd.Kind() == protoreflect.MessageKind {
			if fd.IsList() {
				list := message.ProtoReflect().Get(fd).List()
				for j := 0; j < list.Len(); j++ {
					item := list.Get(j).Message().Interface()
					fieldsName = append(fieldsName, p.FiltersFields(item, fmt.Sprintf("%s[%d]", fieldName, j))...)
				}
			} else if fd.IsMap() {
				mapValue := message.ProtoReflect().Get(fd).Map()

				mapValue.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {

					if fd.MapValue().Kind() == protoreflect.MessageKind {

						item := value.Message().Interface()
						fieldsName = append(fieldsName, p.FiltersFields(item, fmt.Sprintf("%s[%v]", fieldName, key.Interface()))...)

					} else {
						// log.Printf("map key path:%v", fmt.Sprintf("%s[%v]", fieldName, key.Interface()))
						fieldsName = append(fieldsName, fmt.Sprintf("%s[%v]", fieldName, key.Interface()))
					}
					return true
				})
			} else {
				nestedMessage := message.ProtoReflect().Get(fd).Message().Interface()
				fieldsName = append(fieldsName, p.FiltersFields(nestedMessage, fieldName)...)
			}
		}
	}

	return fieldsName
}

// FiltersFields 从FieldMask中获取过滤字段,获取待验证字段
// required 字段优先级高于FieldMask
func (p *ParamsValidatorImpl) FiltersMessageFields(v interface{}) []string {
	// fieldsMap := getFieldNamesFromJSONTags(v)
	requiredFields := make([]string, 0)
	maskFields := make([]string, 0)

	if message, ok := v.(protoreflect.ProtoMessage); ok {
		md := message.ProtoReflect().Descriptor()

		// 遍历所有字段
		for i := 0; i < md.Fields().Len(); i++ {
			field := md.Fields().Get(i)
			// require 字段必须校验
			if p.isRequiredField(field) {
				// log.Printf("required field:%s", field.JSONName())
				requiredFields = append(requiredFields, p.getValidatePath(message, field.JSONName(), "")...)
			}

			// 检查字段是否是FieldMask类型
			if field.Kind() == protoreflect.MessageKind && !field.IsList() && !field.IsMap() {

				// 获取字段的值（确保它是FieldMask类型）
				fieldValue := message.ProtoReflect().Get(field).Message()
				mask, ok := fieldValue.Interface().(*fieldmaskpb.FieldMask)
				if mask == nil || !ok {
					continue
				}
				paths := make([]string, 0)
				paths = append(paths, mask.Paths...)
				for _, path := range paths {
					maskField := strcase.ToCamel(path)
					// if parent != "" {
					// 	maskField = fmt.Sprintf("%s.%s", parent, strcase.ToCamel(path))
					// }
					maskFields = append(maskFields, maskField)
					maskFields = append(maskFields, p.getValidatePath(message, path, "")...)
				}
			}
		}
		return append(requiredFields, maskFields...)
	}
	return nil
}

// FiltersFields 从FieldMask中获取过滤字段,获取待验证字段
// required 字段优先级高于FieldMask
func (p *ParamsValidatorImpl) FiltersFields(v interface{}, parent string) []string {
	fieldsMap := getFieldNamesFromJSONTags(v)
	requiredFields := make([]string, 0)
	maskFields := make([]string, 0)

	if message, ok := v.(protoreflect.ProtoMessage); ok {
		md := message.ProtoReflect().Descriptor()
		val := reflect.ValueOf(v)
		typ := reflect.TypeOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			typ = typ.Elem()
		}
		isRequired := false
		for k := range fieldsMap {
			field := md.Fields().ByJSONName(k)
			st, ok := typ.FieldByName(fieldsMap[k])
			if val.Kind() == reflect.Struct {
				isRequired = ok && p.isRequiredField(st)

			}
			// 检查字段是否是必填字段
			// 如果是必填字段，将其加入requiredFields,用于检查
			if p.isRequiredField(field) || isRequired {
				requiredFields = append(requiredFields, p.getValidatePath(message, k, parent)...)
			}
		}
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
				paths := make([]string, 0)
				paths = append(paths, mask.Paths...)
				for _, path := range paths {
					maskField := strcase.ToCamel(path)
					if parent != "" {
						maskField = fmt.Sprintf("%s.%s", parent, strcase.ToCamel(path))
					}
					maskFields = append(maskFields, maskField)
					maskFields = append(maskFields, p.getValidatePath(message, path, parent)...)
				}
			}
		}
		return append(requiredFields, maskFields...)
	}
	return nil
}

func (p *ParamsValidatorImpl) ValidateParams(v interface{}) error {
	// p.validate.Struct()
	var err error
	if message, ok := v.(proto.Message); ok {
		filters := p.FiltersMessageFields(v)
		duplicateFilters := make([]string, 0)
		fieldsSet := make(map[string]struct{})
		for _, f := range filters {
			if _, ok := fieldsSet[f]; !ok {
				fieldsSet[f] = struct{}{}
				duplicateFilters = append(duplicateFilters, f)
			}
		}

		pv := NewProtobufValidate(p.validate)
		// log.Printf("validate fields:%v", duplicateFilters)
		err = pv.ProtobufPartial(message, common.E_Validate, duplicateFilters...)
		// err = p.ValidateProtoMessage(message, common.E_Validate, fieldsSet, strcase.ToCamel(string(message.ProtoReflect().Descriptor().Name()))+".")

	} else {

		err = gosdk.NewError(fmt.Errorf("params validation failed: params is not a proto.Message"), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "params_validation", gosdk.WithClientMessage("params validation failed: unsupported type"))
	}
	fieldName := ""

	validateErr := validator.ValidationErrors{}
	if errors.As(err, &validateErr) {
		if validateErr[0].Namespace() != "" {
			fieldName = validateErr[0].Namespace()
		}
		clientMsg := fmt.Sprintf("params %s validation failed with %v,except %s", fieldName, validateErr[0].Value(), validateErr[0].ActualTag())
		return gosdk.NewError(fmt.Errorf("params %s validation failed: %v due to %v", fieldName, validateErr[0].Value(), validateErr[0].ActualTag()), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "params_validation", gosdk.WithClientMessage(clientMsg))
	}
	return err
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
	// fmt.Print("params validator unary interceptor\n")
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
func (p *ParamsValidatorImpl) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(ctx, desc, cc, method, opts...)

}
func NewParamsValidator() ParamsValidator {

	v := &ParamsValidatorImpl{
		validate: validator.New(),
	}
	// RegisterCustomValidators(v.validate)
	return v

}
