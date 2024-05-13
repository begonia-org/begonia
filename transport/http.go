package transport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

const GatewayXParams = "x-gateway-params"

type Template struct {
	// Version is the version number of the format.
	Version int
	// OpCodes is a sequence of operations.
	OpCodes []int
	// Pool is a constant pool
	Pool []string
	// Verb is a VERB part in the template.
	Verb string
	// Fields is a list of field paths bound in this template.
	Fields []string
	// Original template (example: /v1/a_bit_of_everything)
	Template string
}
type HttpEndpoint interface {
	RegisterHandlerClient(ctx context.Context, pd ProtobufDescription, mux *runtime.ServeMux) error
	DeleteEndpoint(ctx context.Context, pd ProtobufDescription, mux *runtime.ServeMux) error
}
type HttpEndpointItem struct {
	Pattern  runtime.Pattern `json:"-"`
	Template *Template

	HttpMethod     string
	FullMethodName string
	HttpUri        string
	PathParams     []string
	In             protoreflect.MessageDescriptor
	Out            protoreflect.MessageDescriptor
	IsClientStream bool
	IsServerStream bool
	InName         string
	OutName        string
	Pkg            string
	InPkg          string
	OutPkg         string
	HttpResponse   string `json:"http_response"`
}
type HttpEndpointImpl struct {
	// items  []*HttpEndpointItem
	// pd     ProtobufDescription
	client HttpForwardGrpcEndpoint
	mux    *sync.Mutex
}

func loadHttpEndpointItem(pd ProtobufDescription, descFile string) ([]*HttpEndpointItem, error) {
	items := make(map[string][]*HttpEndpointItem)
	endpointItems := make([]*HttpEndpointItem, 0)

	file, err := os.Open(descFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %w", err)
	}
	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal file: %w", err)
	}
	for _, binds := range items {
		item := binds[0]
		//  设置入参和出参
		item.In = pd.GetMessageTypeByName(item.InPkg, item.InName)
		item.Out = pd.GetMessageTypeByName(item.OutPkg, item.OutName)
		if item.HttpResponse != "" {
			item.Out = pd.GetMessageTypeByFullName(item.HttpResponse)
		}

		item.Pattern = runtime.MustPattern(runtime.NewPattern(item.Template.Version, item.Template.OpCodes, item.Template.Pool, item.Template.Verb))
		endpointItems = append(endpointItems, item)

	}

	return endpointItems, nil

}
func loadGlobalMessages(pd ProtobufDescription) error {
	fds := pd.GetFileDescriptorSet()
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return err
	}

	// // 注册消息类型
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Messages().Len(); i++ {
			msgType := fd.Messages().Get(i)
			if _, err := protoregistry.GlobalTypes.FindMessageByName(msgType.FullName()); err == nil {
				// 如果找到了，说明已经注册过，跳过
				continue
			}
			msg := dynamicpb.NewMessageType(msgType)
			err := protoregistry.GlobalTypes.RegisterMessage(msg)
			if err != nil {
				return false
			}
		}
		for i := 0; i < fd.Enums().Len(); i++ {
			enumType := fd.Enums().Get(i)
			if _, err := protoregistry.GlobalTypes.FindEnumByName(enumType.FullName()); err == nil {
				// 如果找到了，说明已经注册过，跳过
				continue
			}
			enum := dynamicpb.NewEnumType(enumType)
			err := protoregistry.GlobalTypes.RegisterEnum(enum)
			if err != nil {
				return false
			}
		}

		return true
	})
	return nil
}
func NewHttpEndpoint(client HttpForwardGrpcEndpoint) (HttpEndpoint, error) {

	return &HttpEndpointImpl{
		client: client,
		mux:    &sync.Mutex{},
	}, nil
}
func (h *HttpEndpointImpl) stream(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, _ map[string]string) (StreamClient, runtime.ServerMetadata, error) {
	var metadata runtime.ServerMetadata
	grpcReq := &GrpcRequestImpl{
		Ctx:            ctx,
		InType:         item.In,
		OutType:        item.Out,
		FullMethodName: item.FullMethodName,
	}
	stream, err := h.client.Stream(grpcReq)
	if err != nil {
		log.Errorf("Failed to start streaming: %v", err)
		return nil, metadata, err
	}
	if req.Body == nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "body is empty")
	}
	dec := marshaler.NewDecoder(req.Body)
	handleSend := func() error {
		var protoReq = dynamicpb.NewMessage(item.In)
		err := dec.Decode(protoReq)
		if err == io.EOF {
			return err
		}
		if err != nil {
			log.Errorf("Failed to decode request: %v", err)
			return err
		}
		if err := stream.Send(protoReq); err != nil {
			log.Errorf("Failed to send request: %v", err)
			return err
		}
		return nil
	}
	go func() {
		for {
			if err := handleSend(); err != nil {
				break
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Errorf("Failed to terminate client stream: %v", err)
		}
	}()
	header, err := stream.Header()
	if err != nil {
		log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil
}
func (h *HttpEndpointImpl) serverStreamRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (ServerSideStream, runtime.ServerMetadata, error) {
	grpcReq := &GrpcRequestImpl{
		Ctx:            ctx,
		InType:         item.In,
		OutType:        item.Out,
		FullMethodName: item.FullMethodName,
		Marshaler:      marshaler,
		Req:            req,
		PathParams:     pathParams,
	}
	var metadata runtime.ServerMetadata
	var protoReq = dynamicpb.NewMessage(item.In)

	if req.Body != nil {
		dec := marshaler.NewDecoder(req.Body)
		err := dec.Decode(protoReq)

		if err != nil && err != io.EOF {
			log.Errorf("Failed to decode request: %v", err)
			return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
		}
	}

	err := h.inParamsHandle(pathParams, req, protoReq)
	if err != nil {
		return nil, metadata, err
	}
	grpcReq.In = protoReq
	stream, err := h.client.ServerSideStream(grpcReq)
	if err != nil {
		log.Errorf("Failed to start streaming: %v", err)
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil
}
func (h *HttpEndpointImpl) clientStreamRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (protoreflect.ProtoMessage, runtime.ServerMetadata, error) {
	grpcReq := &GrpcRequestImpl{
		Ctx:            ctx,
		InType:         item.In,
		OutType:        item.Out,
		FullMethodName: item.FullMethodName,
		Marshaler:      marshaler,
		Req:            req,
		PathParams:     pathParams,
	}
	var metadata runtime.ServerMetadata
	stream, err := h.client.ClientSideStream(grpcReq)
	if err != nil {
		log.Errorf("Failed to start streaming: %v", err)
		return nil, metadata, err
	}
	if req.Body == nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "body is empty")
	}
	dec := marshaler.NewDecoder(req.Body)
	for {
		var protoReq = dynamicpb.NewMessage(item.In)
		err = dec.Decode(protoReq)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("client stream failed to decode request: %v", err)
			return nil, metadata, status.Errorf(codes.InvalidArgument, "Failed to decode request:%v", err)
		}
		err := h.inParamsHandle(pathParams, req, protoReq)
		if err != nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "Failed to add request params:%v", err)
		}
		if err = stream.Send(protoReq); err != nil {
			if err == io.EOF {
				break
			}
			return nil, metadata, status.Errorf(codes.Internal, "Failed to send request:%v", err)
		}
	}

	if err := stream.CloseSend(); err != nil {
		log.Errorf("Failed to terminate client stream: %v", err)
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header

	msg, err := stream.CloseAndRecv()
	metadata.TrailerMD = stream.Trailer()
	return msg, metadata, err
}
func (h *HttpEndpointImpl) inParamsHandle(params map[string]string, req *http.Request, in *dynamicpb.Message) error {
	for val, param := range params {

		msg, err := runtime.String(val)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", param, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(param))
		if field == nil {
			return status.Errorf(codes.InvalidArgument, "no parameter %s", param)
		}
		in.Set(field, protoreflect.ValueOfString(msg))

	}
	if err := req.ParseForm(); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(in, req.Form, &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return nil
}
func (h *HttpEndpointImpl) addHexEncodeSHA256Hash(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("Failed to read body: %w", err)

	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	hashStruct := sha256.New()
	if len(body) == 0 {
		hashStruct.Write([]byte("{}"))
	} else {
		_, _ = hashStruct.Write(body)

	}
	hexStr := fmt.Sprintf("%x", hashStruct.Sum(nil))
	req.Header.Set("X-Content-Sha256", hexStr)
	return nil
}
func (h *HttpEndpointImpl) newRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (GrpcRequest, runtime.ServerMetadata, error) {
	// var protoReq HelloRequest
	var metadata runtime.ServerMetadata
	if item.In == nil {
		return nil, metadata, status.Errorf(codes.Internal, "no request type for %q", item.FullMethodName)
	}
	in := dynamicpb.NewMessage(item.In)
	if req.Body != nil {
		decoder := marshaler.NewDecoder(req.Body)
		if formdata, ok := decoder.(FormatDataDecoder); ok {
			formdata.SetBoundary(req.Header.Get("Content-Type"))
		}
		if err := decoder.Decode(in); err != nil && err != io.EOF {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "decode request body err %v", err)
		}
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	for _, param := range item.PathParams {
		val, ok = pathParams[param]
		if !ok {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", param)
		}
		msg, err := runtime.String(val)
		if err != nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", param, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(param))
		if field == nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "no parameter %s", param)
		}
		in.Set(field, protoreflect.ValueOfString(msg))

	}
	query := req.URL.Query()
	for k, v := range query {
		msg, err := runtime.String(v[0])
		if err != nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", k, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(k))
		if field == nil {
			return nil, metadata, status.Errorf(codes.InvalidArgument, "no parameter %s", k)
		}
		in.Set(field, protoreflect.ValueOfString(msg))
	}

	return &GrpcRequestImpl{
		Ctx:            ctx,
		In:             in,
		Out:            dynamicpb.NewMessage(item.Out),
		FullMethodName: item.FullMethodName,
	}, metadata, nil
}
func (h *HttpEndpointImpl) NotFound(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	w.WriteHeader(http.StatusNotFound)
}
func (h *HttpEndpointImpl) DeleteEndpoint(ctx context.Context, pd ProtobufDescription, mux *runtime.ServeMux) error {
	h.mux.Lock()
	defer h.mux.Unlock()
	items, err := loadHttpEndpointItem(pd, pd.GetGatewayJsonSchema())
	if err != nil {
		return err
	}
	for _, item := range items {
		mux.Handle(strings.ToUpper(item.HttpMethod), item.Pattern, h.NotFound)
	}
	return nil
}
func (h *HttpEndpointImpl) RegisterHandlerClient(ctx context.Context, pd ProtobufDescription, mux *runtime.ServeMux) error {
	h.mux.Lock()
	defer h.mux.Unlock()
	items, err := loadHttpEndpointItem(pd, pd.GetGatewayJsonSchema())
	if err != nil {
		return err
	}
	err = loadGlobalMessages(pd)
	if err != nil {
		return err
	}
	for _, item := range items {
		item := item
		mux.Handle(strings.ToUpper(item.HttpMethod), item.Pattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			if req.Header.Get("accept") == "" {
				req.Header.Set("accept", "application/json")
			}
			ctx, cancel := context.WithCancel(req.Context())
			defer cancel()
			inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
			var err error
			var annotatedContext context.Context
			// 添加sha256 hash
			err = h.addHexEncodeSHA256Hash(req)
			if err != nil {
				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
				return
			}
			if len(pathParams) > 0 {
				params := make([]string, 0)
				for key := range pathParams {
					params = append(params, key)
				}
				// fmt.Printf("params: %v\n", params)
				req.Header.Set(GatewayXParams, strings.Join(params, ","))

			}
			annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, item.FullMethodName, runtime.WithHTTPPathPattern(item.HttpUri))
			if err != nil {
				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
				return
			}

			// 普通请求
			if !item.IsServerStream && !item.IsClientStream {
				reqInstance, md, err := h.newRequest(annotatedContext, item, inboundMarshaler, req, pathParams)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				resp, md, err := h.client.Request(reqInstance)

				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				// log.Printf("request: %s, response: %v", req.URL.String())
				runtime.ForwardResponseMessage(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
			} else if item.IsServerStream && !item.IsClientStream {
				// 服务端推流,升级为sse服务
				resp, md, err := h.serverStreamRequest(annotatedContext, item, inboundMarshaler, req, pathParams)

				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					// req.Header.Set("Content-Type", "text/event-stream")
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				sse, err := NewServerSendEventForwarder(w, req, 3, item.OutName)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				recv := func() (proto.Message, error) {
					return resp.Recv()
				}
				runtime.ForwardResponseStream(annotatedContext, mux, outboundMarshaler, sse, req, recv, mux.GetForwardResponseOptions()...)
			} else if !item.IsServerStream && item.IsClientStream {
				// 客户端推流
				resp, md, err := h.clientStreamRequest(annotatedContext, item, inboundMarshaler, req, pathParams)
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				runtime.ForwardResponseMessage(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
			} else {
				// 双向流，升级为websocket
				resp, md, err := h.stream(annotatedContext, item, inboundMarshaler, req, pathParams)
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				// websocket
				ws, err := NewWebsocketForwarder(w, req, resp, websocket.BinaryMessage)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				recv := func() (proto.Message, error) {
					out := item.Out
					ws := ws
					msg, err := ws.Read()
					if err != nil {
						return nil, err
					}
					pb := dynamicpb.NewMessage(out)
					err = proto.Unmarshal(msg, pb)
					if err != nil {
						return nil, err
					}
					return pb, nil

				}
				runtime.ForwardResponseStream(annotatedContext, mux, outboundMarshaler, ws, req, recv, mux.GetForwardResponseOptions()...)
			}
		})
	}
	return nil

}
