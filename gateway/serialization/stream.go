package serialization

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const ClientStreamContentType = "application/begonia-client-stream"

// ProtobufWithLengthPrefix 是一个编码器，它将消息编码为长度前缀的字节切片
// 方便stream处理,不需要使用分隔符,将连续的消息流转换为字节流
type HttpProtobufStream interface {
	runtime.Marshaler
}
type HttpProtobufStreamImpl struct {
	*runtime.ProtoMarshaller
}
type LengthPrefixMarshalerOption func(cxt context.Context, message proto.Message) error
type LengthPrefixUnmarshalerOption func(cxt context.Context, data []byte) error

// Marshal 将消息编码为长度前缀的字节切片
func (*HttpProtobufStreamImpl) Marshal(value interface{}) ([]byte, error) {
	data, err := proto.Marshal(value.(proto.Message))
	if err != nil {
		return nil, err
	}
	// 创建一个足够大的缓冲区来存储长度前缀和消息本身
	buf := make([]byte, len(data)+4) // 4 字节用于存储长度

	// 将消息长度作为前缀写入
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	// 将编码后的消息追加到长度后面
	copy(buf[4:], data)

	return buf, nil

}

func NewProtobufWithLengthPrefix() HttpProtobufStream {
	return &HttpProtobufStreamImpl{
		&runtime.ProtoMarshaller{},
	}
}
func (p *HttpProtobufStreamImpl) NewDecoder(reader io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(value interface{}) error {
		// 读取长度前缀
		var length uint32
		err := binary.Read(reader, binary.BigEndian, &length)
		if err != nil {
			if err == io.EOF {
				// 如果到达流的末尾，则结束循环
				return err
			}
			return fmt.Errorf("failed to read message length: %v", err)
		}

		// 读取指定长度的消息数据
		messageData := make([]byte, length)
		_, err = reader.Read(messageData)
		if err != nil && err != io.EOF {

			return fmt.Errorf("failed to read message data: %v", err)
		}
		return p.Unmarshal(messageData, value)
	})
}
func (p *HttpProtobufStreamImpl) NewEncoder(writer io.Writer) runtime.Encoder {
	return runtime.EncoderFunc(func(value interface{}) error {
		buf, err := p.Marshal(value)
		if err != nil {
			return err
		}
		_, err = writer.Write(buf)
		return err
	})
}
func (p *HttpProtobufStreamImpl) ContentType(v interface{}) string {
	return ClientStreamContentType
}

// Unmarshal 将消息解码为长度前缀的字节切片
func (p *HttpProtobufStreamImpl) Unmarshal(data []byte, value interface{}) error {

	// 反序列化消息
	err := protojson.Unmarshal(data, value.(proto.Message))
	if err != nil {
		return fmt.Errorf("failed to unmarshal message from HttpProtobufStreamImpl: %w", err)
	}
	return nil
}
