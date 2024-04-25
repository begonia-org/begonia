package transport

import (
	"google.golang.org/grpc"
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
}
type clientSideStreamClient struct {
	grpc.ClientStream
	out protoreflect.MessageDescriptor
}

type streamClient struct {
	grpc.ClientStream
	out protoreflect.MessageDescriptor
}

func (x *serverSideStreamClient) Recv() (protoreflect.ProtoMessage, error) {
	out := dynamicpb.NewMessage(x.out)
	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (x *clientSideStreamClient) Send(m protoreflect.ProtoMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *clientSideStreamClient) CloseAndRecv() (protoreflect.ProtoMessage, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	out := dynamicpb.NewMessage(x.out)

	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (x *streamClient) Send(m protoreflect.ProtoMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamClient) Recv() (protoreflect.ProtoMessage, error) {
	out := dynamicpb.NewMessage(x.out)

	if err := x.ClientStream.RecvMsg(out); err != nil {
		return nil, err
	}
	return out, nil
}
