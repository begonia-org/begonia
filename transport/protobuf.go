package transport

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway/register"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ProtobufDescription interface {
	GetFileDescriptorSet() *descriptorpb.FileDescriptorSet
	GetMessageTypeByName(pkg string, name string) protoreflect.MessageDescriptor
	GetGatewayJsonSchema() string
	SetHttpResponse(option protoreflect.ExtensionType) error
	GetMessageTypeByFullName(fullName string) protoreflect.MessageDescriptor
	GetDescription() []byte
}

type protobufDescription struct {
	fileDescriptorSet *descriptorpb.FileDescriptorSet
	messages          map[string]protoreflect.MessageDescriptor
	gatewayJsonSchema string
	fs                *protoregistry.Files
	descriptions      []byte
}

func (p *protobufDescription) GetMessages() map[string]protoreflect.MessageDescriptor {
	return p.messages
}

// 初始化描述文件
func (p *protobufDescription) initDescriptorSet() error {
	p.messages = make(map[string]protoreflect.MessageDescriptor)
	fs, err := protodesc.NewFiles(p.fileDescriptorSet)
	if err != nil {
		return fmt.Errorf("Error creating file descriptor:%w", err)

	}
	p.fs = fs
	return nil
}

// SetHttpResponse 设置http_response
func (p *protobufDescription) SetHttpResponse(option protoreflect.ExtensionType) error {
	file, err := os.Open(p.gatewayJsonSchema)
	if err != nil {
		return err
	}
	defer file.Close()
	var gateway = make(map[string]interface{})
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&gateway)
	if err != nil {
		return err
	}
	for key, value := range gateway {
		key = strings.TrimPrefix(key, "/")
		svrAndMethod := strings.Split(key, "/")
		if len(svrAndMethod) != 2 {
			return fmt.Errorf("invalid gateway.json")
		}
		svr := svrAndMethod[0]

		desc, err := p.fs.FindDescriptorByName(protoreflect.FullName(svr))
		if err != nil {
			return err

		}
		if serviceDesc, ok := desc.(protoreflect.ServiceDescriptor); ok {
			if options, ok := serviceDesc.Options().(*descriptorpb.ServiceOptions); ok && options != nil {
				if ext := proto.GetExtension(options, option); ext != nil {
					binds := value.([]interface{})
					for _, bind := range binds {
						if bind.(map[string]interface{})["OutName"] == "HttpBody" {
							continue
						}
						bind.(map[string]interface{})["http_response"] = ext
					}
				}
			}
		}
	}
	wfile, err := os.Create(p.gatewayJsonSchema) // 使用 os.Create 覆盖原文件
	if err != nil {
		return err
	}
	defer wfile.Close()

	// 创建一个 encoder 并写入修改后的数据
	encoder := json.NewEncoder(wfile)
	encoder.SetIndent("", "  ") // 可选：设置缩进，美化输出
	err = encoder.Encode(&gateway)
	if err != nil {
		return err
	}
	return nil
}
func NewDescription(dir string) (ProtobufDescription, error) {
	descPb := filepath.Join(dir, "desc.pb")
	// 使用Glob找到所有.proto文件
	protoFiles, err := filepath.Glob(filepath.Join(dir, "*.proto"))
	if err != nil {
		return nil, fmt.Errorf("Error reading proto files: %w", err)
	}
	args := []string{
		"--descriptor_set_out=" + descPb,
		"--include_imports",
		"--proto_path=" + dir,
		"--grpc-gateway_out=" + dir,
		"--grpc-gateway_opt=only_descriptors=true",
	}
	args = append(args, protoFiles...) // 将文件路径添加到参数中

	cmd := exec.Command("protoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Error running protoc %w,Output: %s", err, output)

	}
	data, err := os.ReadFile(descPb)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %w", err)
	}
	desc := &protobufDescription{
		fileDescriptorSet: &descriptorpb.FileDescriptorSet{},
		descriptions:      data,
	}
	// 解析描述文件
	if err := proto.Unmarshal(data, desc.fileDescriptorSet); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %w", err)
	}
	err = desc.initDescriptorSet()
	if err != nil {
		return nil, err
	}
	desc.gatewayJsonSchema = filepath.Join(dir, "gateway.json")
	return desc, nil
}
func (p *protobufDescription) GetDescription() []byte {
	return p.descriptions

}
func NewDescriptionFromBinary(data []byte, outDir string) (ProtobufDescription, error) {
	desc := &protobufDescription{
		fileDescriptorSet: &descriptorpb.FileDescriptorSet{},
		descriptions:      data,
	}
	// 解析描述文件
	if err := proto.Unmarshal(data, desc.fileDescriptorSet); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal: %w", err)
	}
	err := desc.initDescriptorSet()
	if err != nil {
		return nil, err
	}
	// desc.gatewayJsonSchema = filepath.Join(outDir, "gateway.json")
	contents, err := register.Register(desc.GetFileDescriptorSet(), false, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to register: %w", err)

	}
	
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("Failed to create directory: %w", err)
	}
	err = os.WriteFile(filepath.Join(outDir, "gateway.json"), []byte(contents[0]), 0666)
	if err != nil {
		return nil, fmt.Errorf("Failed to write gateway.json: %w", err)
	}
	desc.gatewayJsonSchema = filepath.Join(outDir, "gateway.json")
	return desc, nil
}
func (p *protobufDescription) GetFileDescriptorSet() *descriptorpb.FileDescriptorSet {
	return p.fileDescriptorSet
}
func (p *protobufDescription) GetMessageTypeByName(pkg string, name string) protoreflect.MessageDescriptor {
	key := fmt.Sprintf("%s.%s", pkg, name)
	if v, ok := p.messages[key]; ok {
		return v
	}
	if desc, err := p.fs.FindDescriptorByName(protoreflect.FullName(key)); err == nil {
		v := desc.(protoreflect.MessageDescriptor)
		p.messages[key] = v
		return v
	}
	return nil
}
func (p *protobufDescription) GetMessageTypeByFullName(fullName string) protoreflect.MessageDescriptor {
	if len(strings.TrimSpace(fullName)) == 0 {
		return nil

	}
	if desc, err := p.fs.FindDescriptorByName(protoreflect.FullName(fullName)); err == nil {
		v := desc.(protoreflect.MessageDescriptor)
		return v
	}
	return nil
}
func (p *protobufDescription) GetGatewayJsonSchema() string {
	return p.gatewayJsonSchema
}

// ProtobufWithLengthPrefix 是一个编码器，它将消息编码为长度前缀的字节切片
// 方便stream处理,不需要使用分隔符,将连续的消息流转换为字节流
type HttpProtobufStream interface{}
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

// Unmarshal 将消息解码为长度前缀的字节切片
func (p *HttpProtobufStreamImpl) Unmarshal(data []byte, value interface{}) error {

	// 反序列化消息
	// var msg = dynamicpb.NewMessage(protoType)
	err := proto.Unmarshal(data, value.(proto.Message))
	if err != nil {
		return fmt.Errorf("failed to unmarshal message from HttpProtobufStreamImpl: %w", err)
	}
	return nil
}
