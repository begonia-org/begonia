package routers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/begonia-org/begonia/transport"
	_ "github.com/begonia-org/go-sdk/api/app/v1"
	_ "github.com/begonia-org/go-sdk/api/endpoint/v1"
	_ "github.com/begonia-org/go-sdk/api/example/v1"
	_ "github.com/begonia-org/go-sdk/api/iam/v1"
	_ "github.com/begonia-org/go-sdk/api/plugin/v1"
	_ "github.com/begonia-org/go-sdk/api/sys/v1"
	_ "github.com/begonia-org/go-sdk/api/user/v1"
	_ "github.com/begonia-org/go-sdk/common/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var onceRouter sync.Once
var httpURIRouteToSrvMethod *HttpURIRouteToSrvMethod

type APIMethodDetails struct {
	// 服务名
	ServiceName string
	// 方法名
	MethodName     string
	AuthRequired   bool
	RequestMethod  string
	GrpcFullRouter string
}
type HttpURIRouteToSrvMethod struct {
	routers    map[string]*APIMethodDetails
	grpcRouter map[string]*APIMethodDetails
	localSrv   map[string]bool
	mux        sync.Mutex
}

func NewHttpURIRouteToSrvMethod() *HttpURIRouteToSrvMethod {
	onceRouter.Do(func() {
		httpURIRouteToSrvMethod = &HttpURIRouteToSrvMethod{
			routers:    make(map[string]*APIMethodDetails),
			grpcRouter: make(map[string]*APIMethodDetails),
			localSrv:   make(map[string]bool),
			mux:        sync.Mutex{},
		}
	})
	return httpURIRouteToSrvMethod
}
func Get() *HttpURIRouteToSrvMethod {
	return NewHttpURIRouteToSrvMethod()
}

func (r *HttpURIRouteToSrvMethod) AddRoute(uri string, srvMethod *APIMethodDetails) {
	r.routers[uri] = srvMethod
	r.grpcRouter[srvMethod.GrpcFullRouter] = srvMethod
}
func (r *HttpURIRouteToSrvMethod) deleteRoute(uri string, grpcFullMethod string) {
	delete(r.routers, uri)
	delete(r.grpcRouter, grpcFullMethod)
}

func (r *HttpURIRouteToSrvMethod) GetRoute(uri string) *APIMethodDetails {
	return r.routers[uri]
}
func (r *HttpURIRouteToSrvMethod) GetRouteByGrpcMethod(method string) *APIMethodDetails {
	return r.grpcRouter[strings.ToUpper(method)]
}
func (r *HttpURIRouteToSrvMethod) GetAllRoutes() map[string]*APIMethodDetails {
	return r.routers
}
func (r *HttpURIRouteToSrvMethod) GetRouteByMethod(method string) *APIMethodDetails {
	for _, v := range r.routers {
		if v.MethodName == method {
			return v
		}
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) getServiceOptions(service protoreflect.ServiceDescriptor) *descriptorpb.ServiceOptions {
	if options, ok := service.Options().(*descriptorpb.ServiceOptions); ok {
		return options
	}
	return nil

}
func (r *HttpURIRouteToSrvMethod) getServiceOptionByExt(service *descriptorpb.ServiceDescriptorProto, ext protoreflect.ExtensionType) interface{} {
	if options := service.GetOptions(); options != nil {
		if ext := proto.GetExtension(options, ext); ext != nil {
			return ext
		}
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) getMethodOptions(method protoreflect.MethodDescriptor) *descriptorpb.MethodOptions {
	if options, ok := method.Options().(*descriptorpb.MethodOptions); ok {
		return options
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) getHttpRule(method *descriptorpb.MethodDescriptorProto) *annotations.HttpRule {
	if options := method.GetOptions(); options != nil {
		if ext := proto.GetExtension(options, annotations.E_Http); ext != nil {
			if httpRule, ok := ext.(*annotations.HttpRule); ok {
				return httpRule
			}
		}
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) AddLocalSrv(fullMethod string) {
	r.localSrv[strings.ToUpper(fullMethod)] = true
}
func (r *HttpURIRouteToSrvMethod) IsLocalSrv(fullMethod string) bool {
	ret := r.localSrv[strings.ToUpper(fullMethod)]
	return ret
}
func (h *HttpURIRouteToSrvMethod) getUri(methodName *descriptorpb.MethodDescriptorProto) (string, string) {
	if httpRule := h.getHttpRule(methodName); httpRule != nil {
		var path string
		var method string
		switch pattern := httpRule.Pattern.(type) {
		case *annotations.HttpRule_Get:
			path = pattern.Get
			method = "GET"
		case *annotations.HttpRule_Post:
			path = pattern.Post
			method = "POST"
		case *annotations.HttpRule_Put:
			path = pattern.Put
			method = "PUT"
		case *annotations.HttpRule_Delete:
			path = pattern.Delete
			method = "DELETE"
		case *annotations.HttpRule_Patch:
			path = pattern.Patch
			method = "PATCH"
		// 可以继续处理其他情况，如 Custom 方法
		case *annotations.HttpRule_Custom:
			path = pattern.Custom.Path
			method = pattern.Custom.Kind
		}
		return path, method
	}
	return "", ""

}
func (h *HttpURIRouteToSrvMethod) DeleteRouterDetails(fullMethod string, method *descriptorpb.MethodDescriptorProto) {
	h.mux.Lock()
	defer h.mux.Unlock()
	uri, _ := h.getUri(method)
	h.deleteRoute(uri, fullMethod)
}
func (r *HttpURIRouteToSrvMethod) addRouterDetails(serviceName string, authRequired bool, methodName *descriptorpb.MethodDescriptorProto) {
	// 获取并打印 google.api.http 注解
	if path, method := r.getUri(methodName); path != "" {
		r.AddRoute(path, &APIMethodDetails{
			ServiceName:    serviceName,
			MethodName:     string(methodName.GetName()),
			AuthRequired:   authRequired,
			RequestMethod:  method,
			GrpcFullRouter: serviceName,
		})

	}

}
func (r *HttpURIRouteToSrvMethod) LoadAllRouters(pd transport.ProtobufDescription) {
	fds := pd.GetFileDescriptorSet()
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, fd := range fds.File {
		for _, service := range fd.Service {
			// srvOptions := service.GetOptions()
			// if srvOptions == nil {
			// 	continue
			// }
			authRequired := false
			// 获取并打印 pb.auth_reqiured 注解
			if authRequiredExt := r.getServiceOptionByExt(service, common.E_AuthReqiured); authRequiredExt != nil {
				authRequired, _ = authRequiredExt.(bool)
			}
			// 遍历服务中的所有方法
			for _, method := range service.GetMethod() {
				key := fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName())
				r.addRouterDetails(strings.ToUpper(key), authRequired, method)
			}

		}
	}

}

func (h *HttpURIRouteToSrvMethod) DeleteRouters(pd transport.ProtobufDescription) {
	// h.mux.Lock()
	// defer h.mux.Unlock()
	fds := pd.GetFileDescriptorSet()
	for _, fd := range fds.File {
		for _, service := range fd.Service {
			for _, method := range service.GetMethod() {
				key := fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName())
				h.DeleteRouterDetails(strings.ToUpper(key), method)
			}
		}
	}
}
