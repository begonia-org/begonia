package serialization

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type FormDataMarshaler struct {
	runtime.JSONPb
}
type FormUrlEncodedMarshaler struct {
	runtime.JSONPb
}

type FormUrlEncodedDecoder struct {
	r io.Reader
}

type FormDataDecoder struct {
	r        io.Reader
	boundary string
}

func (f *FormDataMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &FormDataDecoder{r: r}
}

func (f *FormDataDecoder) Decode(v interface{}) error {
	if response, ok := v.(map[string]interface{}); ok {
		if _, ok := response["result"]; ok {
			v = response["result"]
		}

	}

	// 使用multipart.Reader来解析formData
	reader := multipart.NewReader(f.r, f.boundary)
	formData, err := reader.ReadForm(32 << 20) // 32MB是formData的最大内存使用
	if err != nil && err != io.EOF {
		return err
	}

	for key, files := range formData.File {
		file := files[0]
		fd, err := file.Open()
		if err != nil {
			return errors.New(fmt.Errorf("read file from form data error,%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
		}
		fileBytes, err := io.ReadAll(fd)
		fd.Close()
		if err != nil {
			return errors.New(fmt.Errorf("read file from form data error,%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")

		}
		if _, ok := formData.Value[key]; !ok {
			formData.Value[key] = make([]string, 0)
		}
		formData.Value[key] = append(formData.Value[key], string(fileBytes))

	}
	if pb, ok := v.(protoreflect.ProtoMessage); ok {
		err := parseFormToProto(formData.Value, pb)
		if err != nil {
			return errors.New(fmt.Errorf("parse form data to proto error,%w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "parse_form_data")
		}
		return nil
	}
	return nil

}

func (f *FormDataMarshaler) ContentType(v interface{}) string {
	return "multipart/form-data"
}
func (f *FormDataDecoder) SetBoundary(boundary string) {
	if strings.Contains(boundary, ";") {
		boundary = strings.Split(boundary, "=")[1]
	}
	f.boundary = boundary

}

func getProtoreflectValue(value string, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	// 处理repeated字段和单一字段
	empty := protoreflect.Value{}
	kind := field.Kind()
	if kind == protoreflect.StringKind {
		return protoreflect.ValueOfString(value), nil
	} else if kind == protoreflect.BoolKind {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfBool(boolValue), nil
	} else if kind == protoreflect.DoubleKind {
		doubleValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfFloat64(doubleValue), nil
	} else if kind == protoreflect.EnumKind {
		enumValue := field.Enum().Values().ByName(protoreflect.Name(value))
		fieldName := string(field.Name())
		if enumValue == nil {
			return empty, fmt.Errorf("invalid enum value %s for field %s", value, fieldName)
		}
		return protoreflect.ValueOfEnum(enumValue.Number()), nil
	} else if kind == protoreflect.MessageKind || kind == protoreflect.GroupKind {
		nestedMessage := field.Message()
		dpb := dynamicpb.NewMessage(nestedMessage).New()
		err := protojson.Unmarshal([]byte(value), dpb.(proto.Message))
		if err != nil {
			return empty, fmt.Errorf("Failed to unmarshal nested message %s for field %s: %w", value, string(field.Name()), err)
		}
		return protoreflect.ValueOfMessage(dpb), nil
	} else if kind == protoreflect.BytesKind {
		return protoreflect.ValueOfBytes([]byte(value)), nil
	} else if kind == protoreflect.Uint64Kind || kind == protoreflect.Fixed64Kind {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfUint64(uint64(intValue)), nil
	} else if kind == protoreflect.Uint32Kind || kind == protoreflect.Fixed32Kind {
		intValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfUint32(uint32(intValue)), nil
	} else if kind == protoreflect.Int32Kind || kind == protoreflect.Sint32Kind || kind == protoreflect.Sfixed32Kind {
		intValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfInt32(int32(intValue)), nil
	} else if kind == protoreflect.Int64Kind || kind == protoreflect.Sint64Kind || kind == protoreflect.Sfixed64Kind {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfInt64(intValue), nil
	} else if kind == protoreflect.FloatKind {
		floatValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return empty, err
		}
		return protoreflect.ValueOfFloat32(float32(floatValue)), nil

	} else {
		return empty, fmt.Errorf("Unsupported field type %s for field %s", kind, string(field.Name()))
	}
	// 根据需要处理其他类型的字段
}
func parseFormToProto(values url.Values, pb proto.Message) error {
	pbReflect := pb.ProtoReflect()
	fields := pbReflect.Descriptor().Fields()
	mask := make([]string, 0)
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := field.JSONName()
		if value, ok := values[fieldName]; ok {
			if field.IsList() {
				list := pbReflect.Mutable(field).List()
				for _, v := range value {
					elem, err := getProtoreflectValue(v, field)
					if err != nil {
						return err
					}
					list.Append(elem)
				}
				pbReflect.Set(field, protoreflect.ValueOfList(list))
				continue
			}
			elem, err := getProtoreflectValue(value[0], field)
			if err != nil {
				return err
			}
			pbReflect.Set(field, elem)
			mask = append(mask, fieldName)

		}
	}
	return SetUpdateMaskFields(pb, mask)

	// return nil
}
func NewFormDataMarshaler() *FormDataMarshaler {
	return &FormDataMarshaler{JSONPb: runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
			UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
			UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

		},
	}}
}

func (f *FormUrlEncodedDecoder) Decode(v interface{}) error {
	buf, err := io.ReadAll(f.r)
	if err != nil && err != io.EOF {
		return errors.New(fmt.Errorf("read form data error,%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_format")
	}
	data, err := url.ParseQuery(string(buf))
	if err != nil {
		return errors.New(fmt.Errorf("parse form data error,%w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "parse_form_data")
	}
	if pb, ok := v.(protoreflect.ProtoMessage); ok {
		err := parseFormToProto(data, pb)
		if err != nil {
			return errors.New(fmt.Errorf("parse form data to proto error,%w", err), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "parse_form_data")
		}
		return nil
	}
	if mapData, ok := v.(map[string]interface{}); ok {
		for k, v := range data {
			mapData[k] = v
		}
		return nil
	}
	return nil
}

func (f *FormUrlEncodedMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &FormUrlEncodedDecoder{r: r}
}

func (f *FormUrlEncodedMarshaler) ContentType(v interface{}) string {
	return "application/x-www-form-urlencoded"
}

func NewFormUrlEncodedMarshaler() *FormUrlEncodedMarshaler {
	return &FormUrlEncodedMarshaler{JSONPb: runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
			UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
			UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

		},
	}}
}
