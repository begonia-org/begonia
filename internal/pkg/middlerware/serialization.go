package middlerware

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/wetrycode/begonia/api/v1"
	api "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/spark-lence/tiga"
)

type ResponseJSONMarshaler struct {
	runtime.JSONPb
}
type EventSourceMarshaler struct {
	ResponseJSONMarshaler
}

func (m *EventSourceMarshaler) ContentType(v interface{}) string {
	return "text/event-stream"
}
func (m *EventSourceMarshaler) ConvertDynamicMessageToEventStreamResponse(dynMsg *dynamicpb.Message) (*api.EventStreamResponse, error) {
	esr := &api.EventStreamResponse{}

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
	if response, ok := v.(*api.APIResponse); ok {
		data, err := tiga.ProtoMsgUnserializer(fmt.Sprintf("%s.%s", config.APIPkg, response.ResponseType), response.Data)
		if err != nil {
			return nil, fmt.Errorf("marshal response error: %w", err)
		}
		stream, err := m.ConvertDynamicMessageToEventStreamResponse(data.(*dynamicpb.Message))
		if err != nil {
			return nil, fmt.Errorf("marshal response error: %w", err)
		}
		line := fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n\n", stream.Id, stream.Event, stream.Retry, stream.Data)
		return []byte(line), nil

	}
	return m.JSONPb.Marshal(v)
}
func (m *ResponseJSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	if response, ok := v.(map[string]interface{}); ok {
		// result:=response
		if _, ok := response["result"]; ok {
			v = response["result"]
		}

	}
	// 在这里定义你的自定义序列化逻辑
	if response, ok := v.(*api.APIResponse); ok {
		rsp, err := tiga.StructToMap(response)
		if err != nil {
			return nil, fmt.Errorf("marshal response error: %w", err)
		}
		var data interface{} = nil
		newRsp := make(map[string]interface{})
		newRsp["code"] = rsp["code"]
		newRsp["message"] = rsp["message"]
		if response.ResponseType != "" {

			data, err = tiga.ProtoMsgUnserializer(fmt.Sprintf("%s.%s", config.APIPkg, response.ResponseType), response.Data)
			if err != nil {
				return nil, fmt.Errorf("marshal response error: %w", err)
			}
		}

		newRsp["data"] = data
		return m.JSONPb.Marshal(newRsp)

	}
	return m.JSONPb.Marshal(v)
}
