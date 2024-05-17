package gateway

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
	"github.com/spark-lence/tiga"
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
func (h *HttpEndpointImpl) stream(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, ws WebsocketForwarder) (StreamClient, runtime.ServerMetadata, error) {
	var metadata runtime.ServerMetadata
	grpcReq := NewGrpcRequest(ctx, item.In, item.Out, item.FullMethodName)
	stream, err := h.client.Stream(grpcReq)
	if err != nil {
		// Log.Errorf(ctx,"Failed to start websocket request: %v", err)
		return nil, metadata, fmt.Errorf("Failed to start websocket request: %w", err)
	}
	handleSend := func() error {
		var protoReq = dynamicpb.NewMessage(item.In)
		reader, err := ws.NextReader()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				// Log.Errorf(ctx,"Failed to read websocket: %w", err)
				return fmt.Errorf("Failed to read websocket: %w", err)
			}
			return err
		}
		if reader == nil {
			return io.EOF
		}
		dec := marshaler.NewDecoder(reader)

		err = dec.Decode(protoReq)

		if err != nil && err != io.EOF {
			// Log.Errorf("Failed to decode websocket request: %v", err)
			return fmt.Errorf("Failed to decode websocket request: %w", err)
		}
		if err != nil {
			return io.EOF
		}

		if err := stream.Send(protoReq); err != nil {
			// Log.Errorf("Failed to send request: %v", err)
			return fmt.Errorf("Failed to send request: %w", err)
		}
		return nil
	}
	go func() {
		for {
			if err := handleSend(); err != nil {
				if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					Log.Warnf(ctx, "Failed to send websocket request: %v", err)
				}
				break

			}
		}
		if err := stream.CloseSend(); err != nil {
			Log.Warnf(ctx, "Failed to terminate client stream: %v", err)
		}
	}()
	header, err := stream.Header()
	if err != nil {
		// Log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, fmt.Errorf("Failed to get header from websocket client: %w", err)
	}
	metadata.HeaderMD = header
	return stream, metadata, nil
}
func (h *HttpEndpointImpl) serverStreamRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (ServerSideStream, runtime.ServerMetadata, error) {

	var metadata runtime.ServerMetadata
	var protoReq = dynamicpb.NewMessage(item.In)

	if req.Body != nil {
		dec := marshaler.NewDecoder(req.Body)
		err := dec.Decode(protoReq)

		if err != nil && err != io.EOF {
			// Log.Errorf(ctx,"Failed to decode request: %v", err)
			return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
		}
	}

	err := h.inParamsHandle(pathParams, req, protoReq)
	if err != nil {
		return nil, metadata, err
	}
	// grpcReq.In = protoReq
	grpcReq := NewGrpcRequest(ctx, item.In, item.Out, item.FullMethodName, WithIn(protoReq), WithOut(dynamicpb.NewMessage(item.Out)))
	stream, err := h.client.ServerSideStream(grpcReq)
	if err != nil {
		// Log.Errorf("Failed to start streaming: %v", err)
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		// Log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header
	return stream, metadata, nil
}
func (h *HttpEndpointImpl) clientStreamRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (protoreflect.ProtoMessage, runtime.ServerMetadata, error) {

	var metadata runtime.ServerMetadata
	grpcReq := NewGrpcRequest(ctx, item.In, item.Out, item.FullMethodName, WithGatewayReq(req), WithGatewayPathParams(pathParams))
	stream, err := h.client.ClientSideStream(grpcReq)
	if err != nil {
		// Log.Errorf("Failed to start streaming: %v", err)
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
			// Log.Errorf("client stream failed to decode request: %v", err)
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
		// Log.Errorf("Failed to terminate client stream: %v", err)
		return nil, metadata, err
	}
	header, err := stream.Header()
	if err != nil {
		// Log.Errorf("Failed to get header from client: %v", err)
		return nil, metadata, err
	}
	metadata.HeaderMD = header

	msg, err := stream.CloseAndRecv()
	metadata.TrailerMD = stream.Trailer()
	return msg, metadata, err
}
func (h *HttpEndpointImpl) inParamsHandle(params map[string]string, req *http.Request, in *dynamicpb.Message) error {
	for param, val := range params {

		msg, err := runtime.String(val)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", param, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(param))
		if field == nil {
			// Log.Errorf("inParamsHandle no parameter %s", param)
			return status.Errorf(codes.InvalidArgument, "no such parameter %s", param)
		}
		in.Set(field, protoreflect.ValueOfString(msg))

	}
	if err := req.ParseForm(); err != nil {
		// Log.Errorf("Failed to parse form: %v", err)
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(in, req.Form, &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}); err != nil {
		// Log.Errorf("Failed to populate query parameters: %v", err)
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return nil
}

func (h *HttpEndpointImpl) addHexEncodeSHA256HashV2(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	// 创建SHA256哈希对象
	hashStruct := sha256.New()
	if req.ContentLength == 0 {
		hashStruct.Write([]byte("{}"))
		hexStr := fmt.Sprintf("%x", hashStruct.Sum(nil))
		req.Header.Set("X-Content-Sha256", hexStr)
		return nil

	}
	// 创建一个Buffer用于同时读取和写入数据
	var bodyBuffer bytes.Buffer

	// 使用io.TeeReader在读取body的同时写入Buffer和计算哈希
	teeReader := io.TeeReader(req.Body, &bodyBuffer)
	if _, err := io.Copy(hashStruct, teeReader); err != nil {
		return fmt.Errorf("failed to read and hash body: %w", err)
	}

	// 设置哈希值到请求头
	hexStr := fmt.Sprintf("%x", hashStruct.Sum(nil))
	req.Header.Set("X-Content-Sha256", hexStr)

	// 重置Body为从Buffer读取
	req.Body = io.NopCloser(&bodyBuffer)

	return nil
}
func (h *HttpEndpointImpl) newRequest(ctx context.Context, item *HttpEndpointItem, marshaler runtime.Marshaler, req *http.Request, pathParams map[string]string) (GrpcRequest, error) {
	// var metadata runtime.ServerMetadata
	if item.In == nil {
		return nil, status.Errorf(codes.Internal, "no request type for %q", item.FullMethodName)
	}
	in := dynamicpb.NewMessage(item.In)
	if req.Body != nil {
		decoder := marshaler.NewDecoder(req.Body)
		if formdata, ok := decoder.(FormatDataDecoder); ok {
			formdata.SetBoundary(req.Header.Get("Content-Type"))
		}
		if err := decoder.Decode(in); err != nil && err != io.EOF {
			return nil, status.Errorf(codes.InvalidArgument, "decode request body err %v", err)
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
			continue
		}
		msg, err := runtime.String(val)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", param, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(param))
		if field == nil {
			continue
		}
		if err := tiga.SetFieldValueFromString(in, field, msg); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "set field value error: %v", err)

		}

	}
	query := req.URL.Query()
	for k, v := range query {
		msg, err := runtime.String(v[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", k, err)
		}
		fields := in.Descriptor().Fields()
		field := fields.ByName(protoreflect.Name(k))

		if field == nil {
			continue
		}

		if err := tiga.SetFieldValueFromString(in, field, msg); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "set field value error: %v", err)
		}
	}
	grpcReq := NewGrpcRequest(ctx, item.In, item.Out, item.FullMethodName, WithIn(in), WithOut(dynamicpb.NewMessage(item.Out)))
	return grpcReq, nil
}
func (h *HttpEndpointImpl) NotFound(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	w.WriteHeader(http.StatusNotFound)
	// runtime.HTTPError(r.Context(), nil, nil, w, r, status.Errorf(codes.NotFound, "not found"))
}
func (h *HttpEndpointImpl) DeleteEndpoint(ctx context.Context, pd ProtobufDescription, mux *runtime.ServeMux) error {
	h.mux.Lock()
	defer h.mux.Unlock()
	items, err := loadHttpEndpointItem(pd, pd.GetGatewayJsonSchema())
	if err != nil {
		return err
	}
	for _, item := range items {
		Log.Debugf(ctx, "delete endpoint %s: %s %v", strings.ToUpper(item.HttpMethod), item.HttpUri, item.Pattern)
		mux.Handle(strings.ToUpper(item.HttpMethod), item.Pattern, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
			annotatedContext, _ := runtime.AnnotateIncomingContext(ctx, mux, req, item.FullMethodName, runtime.WithHTTPPathPattern(item.HttpUri))

			annotatedContext, _ = runtime.AnnotateContext(annotatedContext, mux, req, item.FullMethodName, runtime.WithHTTPPathPattern(item.HttpUri))
			Log.Warn(annotatedContext, "not found router")

			h.NotFound(w, req, pathParams)
		})
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
		Log.Debugf(ctx, "register endpoint %s: %s %v", strings.ToUpper(item.HttpMethod), item.HttpUri, item.Pattern)
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
			err = h.addHexEncodeSHA256HashV2(req)
			if err != nil {
				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
				return
			}
			if len(pathParams) > 0 {
				params := make([]string, 0)
				for key := range pathParams {
					params = append(params, key)
				}
				req.Header.Set(GatewayXParams, strings.Join(params, ","))

			}

			annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, item.FullMethodName, runtime.WithHTTPPathPattern(item.HttpUri))
			if err != nil {
				runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
				return
			}

			// 普通请求
			if !item.IsServerStream && !item.IsClientStream {
				// log.Printf("inboundMarshaler: %v, outboundMarshaler: %v", inboundMarshaler.ContentType(nil), outboundMarshaler.ContentType(nil))
				reqInstance, err := h.newRequest(annotatedContext, item, inboundMarshaler, req, pathParams)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				resp, md, err := h.client.Request(reqInstance)

				// annotatedContext = metadata.NewIncomingContext(annotatedContext, md.HeaderMD)

				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				runtime.ForwardResponseMessage(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
			} else if item.IsServerStream && !item.IsClientStream {
				// log.Println("request: ", req.URL.String())
				// 服务端推流,升级为sse服务
				resp, md, err := h.serverStreamRequest(annotatedContext, item, inboundMarshaler, req, pathParams)

				// annotatedContext = metadata.NewIncomingContext(annotatedContext, md.HeaderMD)
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}

				recv := func() (proto.Message, error) {
					return resp.Recv()
				}
				runtime.ForwardResponseStream(annotatedContext, mux, outboundMarshaler, w, req, recv, mux.GetForwardResponseOptions()...)
			} else if !item.IsServerStream && item.IsClientStream {
				// 客户端推流
				resp, md, err := h.clientStreamRequest(annotatedContext, item, inboundMarshaler, req, pathParams)
				// annotatedContext = metadata.NewIncomingContext(annotatedContext, md.HeaderMD)
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)

				if err != nil {
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				runtime.ForwardResponseMessage(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
			} else {
				// 双向流，升级为websocket
				Log.Infof(annotatedContext, "upgrade to websocket: %s %s", item.HttpMethod, item.HttpUri)
				ws, err := NewWebsocketForwarder(w, req, websocket.BinaryMessage)
				if err != nil {
					Log.Errorf(annotatedContext, "Failed to upgrade to websocket: %v", err)
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				stream, md, err := h.stream(annotatedContext, item, inboundMarshaler, ws)
				// annotatedContext = metadata.NewIncomingContext(annotatedContext, md.HeaderMD)
				annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)

				if err != nil {
					Log.Warnf(annotatedContext, "Failed to start websocket request: %v", err)
					runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
					return
				}
				runtime.ForwardResponseStream(annotatedContext, mux, outboundMarshaler, ws, req, stream.Recv, mux.GetForwardResponseOptions()...)
			}
		})
	}
	return nil

}
