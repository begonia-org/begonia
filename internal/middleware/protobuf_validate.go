package middleware

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type protobufValidator struct {
	validate *validator.Validate
}

// NewProtobufValidate Create a new protobuf validator
//
// The validate parameter is a validator instance that can be used to validate the structure of the protobuf message
func NewProtobufValidate(validate *validator.Validate) *protobufValidator {
	return &protobufValidator{validate: validate}
}

// getValue Get the value of the field
func (p *protobufValidator) getValue(v protoreflect.Value, k protoreflect.Kind, f protoreflect.FieldDescriptor) interface{} {
	switch k {
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return v.Int()
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return v.Uint()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float()
	case protoreflect.MessageKind:
		return v.Message().Interface()
	case protoreflect.EnumKind:
		return f.Enum().Values().ByNumber(v.Enum()).Name()
	case protoreflect.BytesKind:
		return v.Bytes()

	}
	return nil
}

// getFieldTag Get the tag of the field
//
// The validateTag parameter is the validate tag of the field
// For Enum fields, the oneof tag is added to the validate tag
// Json tag is added to the field tag
func (p *protobufValidator) getFieldTag(field protoreflect.FieldDescriptor, validateTag interface{}) string {
	tag := fmt.Sprintf(`json:"%s"`, field.JSONName())
	if validateTag != nil {
		validate := validateTag.(string)
		tag = fmt.Sprintf(`json:"%s" validate:"%s"`, field.JSONName(), validate)
	}
	if field.Enum() != nil && !strings.Contains(tag, "oneof") {
		oneOfEnum := make([]string, 0)
		for i := 0; i < field.Enum().Values().Len(); i++ {
			oneOfEnum = append(oneOfEnum, string(field.Enum().Values().Get(i).Name()))
		}
		if !field.IsList() && !field.IsMap() {
			tag = fmt.Sprintf(`json:"%s" validate:"oneof=%s"`, field.JSONName(), strings.Join(oneOfEnum, " "))
		}
	}
	return tag
}

// handleFieldValue Handle the field value, convert the protobuf message field to a struct field by recursion
//
// see: `reflect.StructField.Type`
func (p *protobufValidator) handleFieldValue(field protoreflect.FieldDescriptor, fieldValue protoreflect.Value, ext protoreflect.ExtensionType) interface{} {
	if field.IsMap() {
		// convert map field to struct map field
		mapTyp := make(map[string]interface{})
		mapValue := fieldValue.Map()
		mapValue.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
			if field.MapValue().Kind() == protoreflect.MessageKind {
				mapTyp[key.String()] = p.protobufToStructType(value.Message().Interface(), ext).Interface()
			} else {
				mapTyp[key.String()] = p.getValue(value, field.MapValue().Kind(), field.MapValue())
			}
			return true
		})
		return mapTyp
	}

	if field.Kind() == protoreflect.MessageKind {
		if field.IsList() {

			list := make([]interface{}, 0)
			for j := 0; j < fieldValue.List().Len(); j++ {
				list = append(list, p.protobufToStructType(fieldValue.List().Get(j).Message().Interface(), ext).Interface())
			}
			return list
		}
		return p.protobufToStructType(fieldValue.Message().Interface(), ext).Interface()
	}

	if field.IsList() {

		list := make([]interface{}, 0)
		for j := 0; j < fieldValue.List().Len(); j++ {
			list = append(list, p.getValue(fieldValue.List().Get(j), field.Kind(), field))
		}
		return list
	}

	return p.getValue(fieldValue, field.Kind(), field)
}

// isValid Determine whether the field value is valid,
//
// If the field is a list or map, the function will return true if the field is valid
func (p *protobufValidator) isValid(value protoreflect.Value, f protoreflect.FieldDescriptor) bool {
	if f.IsList() {
		return value.List().IsValid()
	}
	if f.IsMap() {
		return value.Map().IsValid()
	}
	switch f.Kind() {
	case protoreflect.MessageKind:
		return value.Message().IsValid()
	case protoreflect.StringKind:
		return value.String() != ""
	default:
		return value.IsValid()
	}
}

// protobufToStructType Convert the protobuf message to a struct type
//
// see: `reflect.StructField`
func (p *protobufValidator) protobufToStructType(message proto.Message, ext protoreflect.ExtensionType) reflect.Value {
	md := message.ProtoReflect().Descriptor()
	fieldsValues := make(map[string]reflect.Value)
	structFields := make([]reflect.StructField, 0)

	for i := 0; i < md.Fields().Len(); i++ {
		field := md.Fields().Get(i)
		fieldName := strcase.ToCamel(string(field.Name()))
		fieldValue := message.ProtoReflect().Get(field)
		value := p.handleFieldValue(field, fieldValue, ext)
		validateTag := proto.GetExtension(field.Options(), ext)
		tag := p.getFieldTag(field, validateTag)

		structFields = append(structFields, reflect.StructField{
			Name: fieldName,
			Type: reflect.TypeOf(value),
			Tag:  reflect.StructTag(tag),
		})
		fieldsValues[fieldName] = reflect.ValueOf(value)
		if !p.isValid(fieldValue, field) {
			fieldsValues[fieldName] = reflect.Zero(reflect.TypeOf(value))
		}
	}

	structType := reflect.StructOf(structFields)
	newTypVal := reflect.New(structType)
	for k, v := range fieldsValues {
		newTypVal.Elem().FieldByName(k).Set(v)
	}
	return newTypVal
}

func (p *protobufValidator) Protobuf(message proto.Message, ext protoreflect.ExtensionType) error {
	v := p.protobufToStructType(message, ext).Interface()
	return p.validate.Struct(v)
}

func (p *protobufValidator) NewStructFromProtobuf(message proto.Message, ext protoreflect.ExtensionType) interface{} {
	return p.protobufToStructType(message, ext).Interface()
}

func (p *protobufValidator) ProtobufPartial(message proto.Message, ext protoreflect.ExtensionType, fields ...string) error {
	v := p.protobufToStructType(message, ext).Interface()
	return p.validate.StructPartial(v, fields...)
}
func (p *protobufValidator) ProtobufPartialCtx(ctx context.Context, message proto.Message, ext protoreflect.ExtensionType, fields ...string) error {
	v := p.protobufToStructType(message, ext).Interface()
	return p.validate.StructPartialCtx(ctx, v, fields...)
}
