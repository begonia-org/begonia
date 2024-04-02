package serialization

import (
	"fmt"
	"io"
	"reflect"

	"github.com/begonia-org/begonia/internal/pkg/config"
	_ "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/spark-lence/tiga"
)

type RawBinaryUnmarshaler runtime.HTTPBodyMarshaler
type ResponseJSONMarshaler struct {
	runtime.JSONPb
}

type EventSourceMarshaler struct {
	ResponseJSONMarshaler
}
type BinaryDecoder struct {
	fieldName string
	r         io.Reader
}

func (d *BinaryDecoder) fn() string {
	if d.fieldName == "" {
		return "Data"
	}
	return d.fieldName
}

var typeOfBytes = reflect.TypeOf([]byte(nil))
var typeOfHttpbody = reflect.TypeOf(&httpbody.HttpBody{})

func (d *BinaryDecoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v).Elem() // assert it must be a pointer
	if rv.Kind() != reflect.Struct {
		return d
	}

	data := rv.FieldByName(d.fn())
	if !data.CanSet() || (data.Type() != typeOfBytes && data.Type() != typeOfHttpbody) {
		return d
	}
	p, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}
	if len(p) == 0 {
		return io.EOF
	}

	if _, ok := data.Interface().(*httpbody.HttpBody); ok {
		httpBody := &httpbody.HttpBody{
			ContentType: "application/octet-stream",
			Data:        p,
		}
		data.Set(reflect.ValueOf(httpBody))
		return nil
	}
	data.SetBytes(p)

	return err
}

func (d *BinaryDecoder) Error() string {
	d.r = nil
	return "cannot set: " + d.fn()
}
func NewRawBinaryUnmarshaler() *RawBinaryUnmarshaler {
	return &RawBinaryUnmarshaler{

		Marshaler: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
}

//	func (m *RawBinaryUnmarshaler) NewDecoder(r io.Reader) runtime.Decoder {
//		return &BinaryDecoder{"Data", r}
//	}
func ConvertDynamicMessageToHttpBody(dynamicMessage *dynamicpb.Message) (*httpbody.HttpBody, error) {
	// 序列化dynamicpb.Message为字节流
	serialized, err := proto.Marshal(dynamicMessage)
	if err != nil {
		return nil, err
	}

	// 反序列化字节流回原始的HttpBody
	var httpBody httpbody.HttpBody
	if err := proto.Unmarshal(serialized, &httpBody); err != nil {
		return nil, err
	}

	return &httpBody, nil
}
func (m *RawBinaryUnmarshaler) ContentType(v interface{}) string {
	if dpb, ok := v.(*dynamicpb.Message); ok {
		if httpBody, err := ConvertDynamicMessageToHttpBody(dpb); err == nil && httpBody != nil {
			return httpBody.GetContentType()
		}
	}
	return "application/octet-stream"
}
func (m *RawBinaryUnmarshaler) Marshal(v interface{}) ([]byte, error) {
	if dpb, ok := v.(*dynamicpb.Message); ok {
		if httpBody, err := ConvertDynamicMessageToHttpBody(dpb); err == nil && httpBody != nil {
			return httpBody.GetData(), nil
		}
	}
	return m.Marshaler.Marshal(v)
}
func (m *EventSourceMarshaler) ContentType(v interface{}) string {
	return "text/event-stream"
}
func (m *EventSourceMarshaler) ToEventStreamResponse(dynMsg *dynamicpb.Message) (*common.EventStreamResponse, error) {
	esr := &common.EventStreamResponse{}

	// 遍历所有字段
	dynMsg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch fd.Name() {
		case "event":
			esr.Event = v.String()
		case "data":
			esr.Data = v.String()
		case "id":
			esr.Id = v.Int()
		case "retry":
			esr.Retry = int32(v.Int())
		}
		return true
	})

	// 返回转换后的消息
	return esr, nil
}
func (m *EventSourceMarshaler) Marshal(v interface{}) ([]byte, error) {
	if response, ok := v.(map[string]interface{}); ok {
		// result:=response
		if _, ok := response["result"]; ok {
			v = response["result"]
		}

	}
	// 在这里定义你的自定义序列化逻辑
	if response, ok := v.(*common.APIResponse); ok {
		data, err := tiga.ProtoMsgUnserializer(fmt.Sprintf("%s.%s", config.APIPkg, response.ResponseType), response.Data)
		if err != nil {
			return nil, fmt.Errorf("marshal response error: %w", err)
		}
		stream, err := m.ToEventStreamResponse(data.(*dynamicpb.Message))
		if err != nil {
			return nil, fmt.Errorf("marshal response error: %w", err)
		}
		line := fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n\n", stream.Id, stream.Event, stream.Retry, stream.Data)
		return []byte(line), nil

	}
	return m.JSONPb.Marshal(v)
}
func NewResponseJSONMarshaler() *ResponseJSONMarshaler {
	return &ResponseJSONMarshaler{JSONPb: runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
			UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
			UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

		},
	}}
}
func NewEventSourceMarshaler() *EventSourceMarshaler {
	return &EventSourceMarshaler{ResponseJSONMarshaler: ResponseJSONMarshaler{
		JSONPb: runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
				UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
				UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

			},
		}}}
}

func (m *ResponseJSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	if response, ok := v.(map[string]interface{}); ok {
		if _, ok := response["result"]; ok {
			v = response["result"]
		}

	}

	if response, ok := v.(*dynamicpb.Message); ok {
		// log.Println("实际类型,", response.Type().Descriptor().Name())
		byteData, err := m.JSONPb.Marshal(response)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "marshal_response: %v", err)
		}
		return byteData, nil

	}
	return m.JSONPb.Marshal(v)
}
