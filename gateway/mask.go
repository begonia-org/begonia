package gateway

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// type MaskDecoder interface {
// 	runtime.Decoder
// 	GetMask() []string
// }

func SetUpdateMaskFields(message protoreflect.ProtoMessage, fields []string) {
	// 反射获取消息的描述符
	md := message.ProtoReflect().Descriptor()

	// 遍历所有字段
	for i := 0; i < md.Fields().Len(); i++ {
		field := md.Fields().Get(i)

		// 检查字段是否是FieldMask类型
		if field.Message() != nil && field.Message().FullName() == "google.protobuf.FieldMask" {
			// 获取字段的值（确保它是FieldMask类型）
			// fieldValue := message.ProtoReflect().Get(field).Message()

			// 创建一个新的FieldMask并设置fields
			fieldMask := &fieldmaskpb.FieldMask{Paths: fields}
			// 更新原始消息中的FieldMask字段
			message.ProtoReflect().Set(field, protoreflect.ValueOf(fieldMask.ProtoReflect()))

		}
	}
}

type maskDecoder struct {
	runtime.Decoder
	newDecoder func(r io.Reader) runtime.Decoder
}

func NewJsonDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderWrapper{Decoder: json.NewDecoder(r)}
}
func NewMaskDecoder(dec runtime.Decoder) *maskDecoder {
	return &maskDecoder{
		Decoder:    dec,
		newDecoder: NewJsonDecoder,
	}
}
func (d *maskDecoder) Decode(v interface{}) error {

	// 检查消息是否实现了ProtoMessage接口
	message, ok := v.(protoreflect.ProtoMessage)
	if !ok {
		return nil
	}
	mask := make(map[string]interface{})
	// 解码更新掩码字段
	err := d.Decoder.Decode(&mask)
	if err != nil {
		return err

	}

	fields := make([]string, 0)
	for k := range mask {
		fields = append(fields, k)

	}
	bData, err := json.Marshal(mask)
	if err != nil {
		return err
	}

	decoder := d.newDecoder(bytes.NewReader(bData))
	err = decoder.Decode(message)
	if err != nil {
		return err
	}
	// 设置更新掩码字段
	SetUpdateMaskFields(message, fields)
	
	return nil
}
