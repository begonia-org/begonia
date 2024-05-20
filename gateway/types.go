package gateway

import (
	"sync/atomic"

	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ServerSideStream interface {
	Recv() (protoreflect.ProtoMessage, error)
	grpc.ClientStream
}

type ClientSideStream interface {
	Send(protoreflect.ProtoMessage) error
	CloseAndRecv() (protoreflect.ProtoMessage, error)
	grpc.ClientStream
}
type StreamClient interface {
	Send(protoreflect.ProtoMessage) error
	Recv() (protoreflect.ProtoMessage, error)
	grpc.ClientStream
}

type serverSideStreamClient struct {
	grpc.ClientStream
	out protoreflect.MessageDescriptor
	ID  int64
}
type clientSideStreamClient struct {
	grpc.ClientStream
	out protoreflect.MessageDescriptor
}

type streamClient struct {
	grpc.ClientStream
	out protoreflect.MessageDescriptor
}

func (x *serverSideStreamClient) buildEventStreamResponse(dpm *dynamicpb.Message) (*common.EventStream, error) {
	data, err := protojson.Marshal(dpm.Interface())
	if err != nil {
		return nil, err

	}

	commonEvent := &common.EventStream{
		Event: string(dpm.Descriptor().Name()),
		Id:    atomic.LoadInt64(&x.ID),
		Data:  string(data),
		Retry: 0,
	}
	defer func() {
		atomic.AddInt64(&x.ID, 1)
	}()
	return commonEvent, nil

}
func (x *serverSideStreamClient) Recv() (protoreflect.ProtoMessage, error) {
	out := dynamicpb.NewMessage(x.out)
	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return x.buildEventStreamResponse(out)
}
func (x *serverSideStreamClient)SendMsg(v any)error{
	return x.ClientStream.SendMsg(v)
}
func (x *serverSideStreamClient)CloseSend()error{
	return x.ClientStream.CloseSend()

}

func (x *clientSideStreamClient) Send(m protoreflect.ProtoMessage) error {
	return x.ClientStream.SendMsg(m)
}
func (x *clientSideStreamClient) CloseSend() error {
	return x.ClientStream.CloseSend()
}
func (x *clientSideStreamClient)RecvMsg(v any)error{
	return x.ClientStream.RecvMsg(v)
}
func (x *clientSideStreamClient) CloseAndRecv() (protoreflect.ProtoMessage, error) {
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	out := dynamicpb.NewMessage(x.out)

	if err := x.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (x *streamClient) Send(m protoreflect.ProtoMessage) error {
	err := x.ClientStream.SendMsg(m)
	return err
}

func (x *streamClient) Recv() (protoreflect.ProtoMessage, error) {
	out := dynamicpb.NewMessage(x.out)

	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}
