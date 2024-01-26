// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.22.2
// source: authentication.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	AuthService_Login_FullMethodName    = "/wetrycode.begonia.AuthService/login"
	AuthService_Logout_FullMethodName   = "/wetrycode.begonia.AuthService/logout"
	AuthService_Account_FullMethodName  = "/wetrycode.begonia.AuthService/account"
	AuthService_AuthSeed_FullMethodName = "/wetrycode.begonia.AuthService/authSeed"
	AuthService_Regsiter_FullMethodName = "/wetrycode.begonia.AuthService/regsiter"
)

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Login(ctx context.Context, in *LoginAPIRequest, opts ...grpc.CallOption) (*APIResponse, error)
	Logout(ctx context.Context, in *LogoutAPIRequest, opts ...grpc.CallOption) (*APIResponse, error)
	Account(ctx context.Context, in *AccountAPIRequest, opts ...grpc.CallOption) (*APIResponse, error)
	AuthSeed(ctx context.Context, in *AuthLogAPIRequest, opts ...grpc.CallOption) (*APIResponse, error)
	Regsiter(ctx context.Context, in *RegsiterAPIRequest, opts ...grpc.CallOption) (*APIResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Login(ctx context.Context, in *LoginAPIRequest, opts ...grpc.CallOption) (*APIResponse, error) {
	out := new(APIResponse)
	err := c.cc.Invoke(ctx, AuthService_Login_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Logout(ctx context.Context, in *LogoutAPIRequest, opts ...grpc.CallOption) (*APIResponse, error) {
	out := new(APIResponse)
	err := c.cc.Invoke(ctx, AuthService_Logout_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Account(ctx context.Context, in *AccountAPIRequest, opts ...grpc.CallOption) (*APIResponse, error) {
	out := new(APIResponse)
	err := c.cc.Invoke(ctx, AuthService_Account_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) AuthSeed(ctx context.Context, in *AuthLogAPIRequest, opts ...grpc.CallOption) (*APIResponse, error) {
	out := new(APIResponse)
	err := c.cc.Invoke(ctx, AuthService_AuthSeed_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Regsiter(ctx context.Context, in *RegsiterAPIRequest, opts ...grpc.CallOption) (*APIResponse, error) {
	out := new(APIResponse)
	err := c.cc.Invoke(ctx, AuthService_Regsiter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility
type AuthServiceServer interface {
	Login(context.Context, *LoginAPIRequest) (*APIResponse, error)
	Logout(context.Context, *LogoutAPIRequest) (*APIResponse, error)
	Account(context.Context, *AccountAPIRequest) (*APIResponse, error)
	AuthSeed(context.Context, *AuthLogAPIRequest) (*APIResponse, error)
	Regsiter(context.Context, *RegsiterAPIRequest) (*APIResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (UnimplementedAuthServiceServer) Login(context.Context, *LoginAPIRequest) (*APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAuthServiceServer) Logout(context.Context, *LogoutAPIRequest) (*APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedAuthServiceServer) Account(context.Context, *AccountAPIRequest) (*APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Account not implemented")
}
func (UnimplementedAuthServiceServer) AuthSeed(context.Context, *AuthLogAPIRequest) (*APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthSeed not implemented")
}
func (UnimplementedAuthServiceServer) Regsiter(context.Context, *RegsiterAPIRequest) (*APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Regsiter not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginAPIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Login(ctx, req.(*LoginAPIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutAPIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Logout_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Logout(ctx, req.(*LogoutAPIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Account_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccountAPIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Account(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Account_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Account(ctx, req.(*AccountAPIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_AuthSeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthLogAPIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).AuthSeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_AuthSeed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).AuthSeed(ctx, req.(*AuthLogAPIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Regsiter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegsiterAPIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Regsiter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Regsiter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Regsiter(ctx, req.(*RegsiterAPIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wetrycode.begonia.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "login",
			Handler:    _AuthService_Login_Handler,
		},
		{
			MethodName: "logout",
			Handler:    _AuthService_Logout_Handler,
		},
		{
			MethodName: "account",
			Handler:    _AuthService_Account_Handler,
		},
		{
			MethodName: "authSeed",
			Handler:    _AuthService_AuthSeed_Handler,
		},
		{
			MethodName: "regsiter",
			Handler:    _AuthService_Regsiter_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "authentication.proto",
}
