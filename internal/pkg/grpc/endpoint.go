package grpc

import (
	_ "github.com/begonia-org/begonia/api/v1"
)

type Endpoint interface {
	// RegisterService registers a service and its implementation to the gRPC

}
type PkgName string
type EndpointImpl struct {
	pkg PkgName
}

func NewEndpointImpl(pkg PkgName) *EndpointImpl {
	return &EndpointImpl{
		pkg: pkg,
	}
}

func (e *EndpointImpl) RegisterService(serviceName string) {
	// const realSvrName= fmt.Sprintf("%s.%s", e.pkg, service)
	// protoregistry.GlobalTypes.F(protoreflect.FullName(service))
	// RegisterManagerServiceHandlerFromEndpoint(ctx context.Context, mux *ServeMux, endpoint string, opts []grpc.DialOption) error
	// registerFunc := fmt.Sprintf("Register%sHandlerFromEndpoint", serviceName)
	// pkgVal := reflect.ValueOf(api{})
	// protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(e.pkg), func(fd protoreflect.FileDescriptor) bool {
	// 	fmt.Println()
	// 	return true
	// })

	// protoregistry.GlobalFiles.RangeFilesByPackage(protoreflect.FullName(e.pkg), func(fd protoreflect.FileDescriptor) bool {
	// 	fmt.Println()
	// 	return true
	// })

	// c:=&loader.Config{}
	// c.Import("github.com/begonia-org/begonia/api/v1")
	// c.Load()
	// for _,pkg:=range c.(){

	// }

}
