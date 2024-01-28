package routers

import (
	_ "github.com/begonia-org/begonia/api/v1"
	_ "github.com/begonia-org/begonia/common/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type APIMethodDetails struct {
	// 服务名
	ServiceName string
	// 方法名
	MethodName    string
	AuthRequired  bool
	RequestMethod string
}
type HttpURIRouteToSrvMethod struct {
	routers map[string]*APIMethodDetails
}

func NewHttpURIRouteToSrvMethod() *HttpURIRouteToSrvMethod {
	return &HttpURIRouteToSrvMethod{
		routers: make(map[string]*APIMethodDetails),
	}
}

func (r *HttpURIRouteToSrvMethod) AddRoute(uri string, srvMethod *APIMethodDetails) {
	r.routers[uri] = srvMethod
}
func (r *HttpURIRouteToSrvMethod) GetRoute(uri string) *APIMethodDetails {
	return r.routers[uri]
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
func (r *HttpURIRouteToSrvMethod) getServiceOptionByExt(service protoreflect.ServiceDescriptor, ext protoreflect.ExtensionType) interface{} {
	if options := r.getServiceOptions(service); options != nil {
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
func (r *HttpURIRouteToSrvMethod) getHttpRule(method protoreflect.MethodDescriptor) *annotations.HttpRule {
	if options := r.getMethodOptions(method); options != nil {
		if ext := proto.GetExtension(options, annotations.E_Http); ext != nil {
			if httpRule, ok := ext.(*annotations.HttpRule); ok {
				return httpRule
			}
		}
	}
	return nil
}
func (r *HttpURIRouteToSrvMethod) addRouterDetails(serviceName string, authRequired bool, methodName protoreflect.MethodDescriptor) {
	// 获取并打印 google.api.http 注解
	if httpRule := r.getHttpRule(methodName); httpRule != nil {
		var path, method string
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
		}
		if path != "" {
			r.AddRoute(path, &APIMethodDetails{
				ServiceName:   serviceName,
				MethodName:    string(methodName.Name()),
				AuthRequired:  authRequired,
				RequestMethod: method,
			})
		}
	}

}
func (r *HttpURIRouteToSrvMethod) LoadAllRouters() {
	protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(config.APIPkg), func(fd protoreflect.FileDescriptor) bool {
		// 遍历文件中的所有服务
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			srvOptions := r.getServiceOptions(service)
			if srvOptions == nil {
				return true
			}
			authRequired := false
			// 获取并打印 pb.auth_reqiured 注解
			if authRequiredExt := r.getServiceOptionByExt(service, common.E_AuthReqiured); authRequiredExt != nil {
				authRequired, _ = authRequiredExt.(bool)
			}
			// 遍历服务中的所有方法
			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				methodName := methods.Get(j)
				r.addRouterDetails(string(service.FullName()), authRequired, methodName)
			}
		}
		return true
	})
	// if len(r.routers) == 0 {
	// 	panic("没有找到任何路由")
	// }
}
