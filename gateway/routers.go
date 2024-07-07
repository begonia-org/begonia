package gateway

import (
	"fmt"
	"strings"
	"sync"

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
	HttpMethodName  string
	AuthRequired    bool
	UseJsonResponse bool
	RequestMethod   string
	GrpcFullRouter  string
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
func GetRouter() *HttpURIRouteToSrvMethod {
	return NewHttpURIRouteToSrvMethod()
}

func (r *HttpURIRouteToSrvMethod) AddRoute(uri string, srvMethod *APIMethodDetails) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.routers[uri] = srvMethod
	// log.Printf("add srv method grpc router:%s,pointer:%p", srvMethod.GrpcFullRouter, r)
	r.grpcRouter[strings.ToUpper(srvMethod.GrpcFullRouter)] = srvMethod
}
func (r *HttpURIRouteToSrvMethod) deleteRoute(uri string, grpcFullMethod string) {
	delete(r.routers, uri)
	delete(r.grpcRouter, grpcFullMethod)
}

func (r *HttpURIRouteToSrvMethod) GetRoute(uri string) *APIMethodDetails {
	return r.routers[uri]
}
func (r *HttpURIRouteToSrvMethod) GetRouteByGrpcMethod(method string) *APIMethodDetails {
	// log.Printf("get grpc method,%v:%s,pointer:%p",r.grpcRouter,strings.ToUpper(method),r)
	return r.grpcRouter[strings.ToUpper(method)]
}
func (r *HttpURIRouteToSrvMethod) GetAllRoutes() map[string]*APIMethodDetails {
	return r.routers
}

func (r *HttpURIRouteToSrvMethod) getServiceOptionByExt(service *descriptorpb.ServiceDescriptorProto, ext protoreflect.ExtensionType) interface{} {
	if options := service.GetOptions(); options != nil {
		if ext := proto.GetExtension(options, ext); ext != nil {
			return ext
		}
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) getMethodOptionByExt(method *descriptorpb.MethodDescriptorProto, ext protoreflect.ExtensionType) interface{} {
	if options := method.GetOptions(); options != nil {
		if ext := proto.GetExtension(options, ext); ext != nil {
			return ext
		}
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
	// log.Printf("add local srv:%s", fullMethod)
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
func (r *HttpURIRouteToSrvMethod) addRouterDetails(serviceName string, useJsonResponse, authRequired bool, methodName *descriptorpb.MethodDescriptorProto) {
	// 获取并打印 google.api.http 注解
	if path, method := r.getUri(methodName); path != "" {
		r.AddRoute(path, &APIMethodDetails{
			ServiceName:     serviceName,
			HttpMethodName:  string(methodName.GetName()),
			AuthRequired:    authRequired,
			RequestMethod:   method,
			GrpcFullRouter:  serviceName,
			UseJsonResponse: useJsonResponse,
		})

	}

}

// LoadAllRouters Load all routers from protobuf description
// for service methods, if the method has a google.api.http annotation, then add the router
// to the router list, and set the authRequired flag to true if the method has a pb.auth_required annotation,
// if the method has a pb.http_response annotation, then set the useJsonResponse flag to true,
// if the method has a pb.dont_use_http_response annotation, then set the useJsonResponse flag to false.
func (r *HttpURIRouteToSrvMethod) LoadAllRouters(pd ProtobufDescription) {
	fds := pd.GetFileDescriptorSet()
	for _, fd := range fds.File {
		for _, service := range fd.Service {

			authRequired := false
			httpResponse := false
			// 获取并打印 pb.auth_reqiured 注解
			if authRequiredExt := r.getServiceOptionByExt(service, common.E_AuthReqiured); authRequiredExt != nil {
				authRequired, _ = authRequiredExt.(bool)
			}
			if httpResponseExt := r.getServiceOptionByExt(service, common.E_HttpResponse); httpResponseExt != nil && httpResponseExt.(string) != "" {
				httpResponse = true
			}
			// 遍历服务中的所有方法
			for _, method := range service.GetMethod() {
				key := fmt.Sprintf("/%s.%s/%s", fd.GetPackage(), service.GetName(), method.GetName())
				// log.Printf("add router:%s,%v", key, httpResponse)
				// do not use HttpResponse for this method if it is set
				dontUseHttpResponse := r.getMethodOptionByExt(method, common.E_DontUseHttpResponse)
				useHttpResponse := httpResponse
				if dontUseHttpResponse != nil && dontUseHttpResponse.(bool) {
					useHttpResponse = false
				}
				r.addRouterDetails(strings.ToUpper(key), useHttpResponse, authRequired, method)
			}

		}
	}

}

func (h *HttpURIRouteToSrvMethod) DeleteRouters(pd ProtobufDescription) {
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
