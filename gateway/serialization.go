package gateway

import (
	"fmt"
	"io"
	"reflect"

	_ "github.com/begonia-org/go-sdk/api/app/v1"
	_ "github.com/begonia-org/go-sdk/api/endpoint/v1"
	_ "github.com/begonia-org/go-sdk/api/file/v1"
	_ "github.com/begonia-org/go-sdk/api/iam/v1"
	_ "github.com/begonia-org/go-sdk/api/plugin/v1"
	_ "github.com/begonia-org/go-sdk/api/sys/v1"
	_ "github.com/begonia-org/go-sdk/api/user/v1"
	_ "github.com/begonia-org/go-sdk/common/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

type RawBinaryUnmarshaler runtime.HTTPBodyMarshaler
type JSONMarshaler struct {
	runtime.JSONPb
}

type EventSourceMarshaler struct {
	JSONMarshaler
}
type BinaryDecoder struct {
	fieldName string
	r         io.Reader
	marshaler runtime.Marshaler
}

func (d *BinaryDecoder) fn() string {
	if d.fieldName == "" {
		return "Data"
	}
	return d.fieldName
}

// var typeOfBytes = reflect.TypeOf([]byte(nil))
// var typeOfHttpbody = reflect.TypeOf(&httpbody.HttpBody{})

func (d *BinaryDecoder) Decode(v interface{}) error {
	if v == nil {
		return nil

	}
	rv := reflect.ValueOf(v).Elem() // assert it must be a pointer
	if rv.Kind() != reflect.Struct {
		return d
	}
	if dpb, ok := v.(*dynamicpb.Message); ok {
		typ := dpb.Type().Descriptor().Name()
		if string(typ) == "HttpBody" {
			p, err := io.ReadAll(d.r)
			if err != nil {
				return err
			}
			if len(p) == 0 {
				return io.EOF
			}
			httpBody := &httpbody.HttpBody{
				ContentType: "application/octet-stream",
				Data:        p,
			}
			body, err := proto.Marshal(httpBody)
			if err != nil {
				return err
			}
			return proto.Unmarshal(body, v.(proto.Message))
		}
		return d.marshaler.NewDecoder(d.r).Decode(v)
	}
	return d.marshaler.NewDecoder(d.r).Decode(v)
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
	var httpBody *httpbody.HttpBody = new(httpbody.HttpBody)
	if err := proto.Unmarshal(serialized, httpBody); err != nil {
		return nil, err
	}

	return httpBody, nil
}
func (m *RawBinaryUnmarshaler) ContentType(v interface{}) string {
	if dpb, ok := v.(*dynamicpb.Message); ok {
		typ := dpb.Type().Descriptor().Name()
		if typ == "HttpBody" {
			if httpBody, err := ConvertDynamicMessageToHttpBody(dpb); err == nil && httpBody != nil {
				if t := httpBody.GetContentType(); t != "" {
					return t
				}
				return "application/octet-stream"

			}
		}
	}
	return "application/octet-stream"
}

func (m *RawBinaryUnmarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &BinaryDecoder{"Data", r, m.Marshaler}

}
func (m *RawBinaryUnmarshaler) Marshal(v interface{}) ([]byte, error) {
	if dpb, ok := v.(*dynamicpb.Message); ok {
		typ := dpb.Type().Descriptor().Name()
		if typ == "HttpBody" {
			if httpBody, err := ConvertDynamicMessageToHttpBody(dpb); err == nil && httpBody != nil {
				return httpBody.GetData(), nil
			}
		}
	}
	return m.Marshaler.Marshal(v)
}

func (m *EventSourceMarshaler) ContentType(v interface{}) string {
	return "text/event-stream"
}

func (m *EventSourceMarshaler) Marshal(v interface{}) ([]byte, error) {
	if response, ok := v.(map[string]interface{}); ok {
		// result:=response
		if _, ok := response["result"]; ok {
			v = response["result"]
		}

	}
	// 在这里定义你的自定义序列化逻辑
	if stream, ok := v.(*common.EventStream); ok {
		line := fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n", stream.Id, stream.Event, stream.Retry, stream.Data)
		return []byte(line), nil

	}
	return m.JSONPb.Marshal(v)
}
func NewJSONMarshaler() *JSONMarshaler {
	return &JSONMarshaler{JSONPb: runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
			UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
			// UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

		},
	}}
}
func NewEventSourceMarshaler() *EventSourceMarshaler {
	return &EventSourceMarshaler{JSONMarshaler: JSONMarshaler{
		JSONPb: runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true, // 设置为 true 以确保默认值（例如 0 或空字符串）被序列化
				UseEnumNumbers:  true, // 设置为 true 以确保枚举值被序列化为数字而不是字符串
				UseProtoNames:   true, // 设置为 true 以确保 proto 消息的原始名称（而不是 Go 字段名称）被序列化

			},
		}}}
}

func (m *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
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
func (m *JSONMarshaler) ContentType(v interface{}) string {
	return "application/json"
}

// func (m *JSONMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
// 	// return NewMaskDecoder(m.JSONPb.NewDecoder(r))
// 	return json.NewDecoder(r)
// }
