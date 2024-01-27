// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: authentication.proto

package v1

import (
	v1 "github.com/wetrycode/begonia/common/api/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LoginAPIRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Auth        string `protobuf:"bytes,1,opt,name=auth,proto3" json:"auth,omitempty"`
	Seed        int64  `protobuf:"varint,2,opt,name=seed,proto3" json:"seed,omitempty"`
	IsKeepLogin bool   `protobuf:"varint,3,opt,name=is_keep_login,proto3" json:"is_keep_login,omitempty"`
}

func (x *LoginAPIRequest) Reset() {
	*x = LoginAPIRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LoginAPIRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoginAPIRequest) ProtoMessage() {}

func (x *LoginAPIRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LoginAPIRequest.ProtoReflect.Descriptor instead.
func (*LoginAPIRequest) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{0}
}

func (x *LoginAPIRequest) GetAuth() string {
	if x != nil {
		return x.Auth
	}
	return ""
}

func (x *LoginAPIRequest) GetSeed() int64 {
	if x != nil {
		return x.Seed
	}
	return 0
}

func (x *LoginAPIRequest) GetIsKeepLogin() bool {
	if x != nil {
		return x.IsKeepLogin
	}
	return false
}

type LogoutAPIRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LogoutAPIRequest) Reset() {
	*x = LogoutAPIRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogoutAPIRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogoutAPIRequest) ProtoMessage() {}

func (x *LogoutAPIRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogoutAPIRequest.ProtoReflect.Descriptor instead.
func (*LogoutAPIRequest) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{1}
}

type LogoutAPIResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LogoutAPIResponse) Reset() {
	*x = LogoutAPIResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogoutAPIResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogoutAPIResponse) ProtoMessage() {}

func (x *LogoutAPIResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogoutAPIResponse.ProtoReflect.Descriptor instead.
func (*LogoutAPIResponse) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{2}
}

type LoginAPIResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	User  *Users `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	Token string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *LoginAPIResponse) Reset() {
	*x = LoginAPIResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LoginAPIResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoginAPIResponse) ProtoMessage() {}

func (x *LoginAPIResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LoginAPIResponse.ProtoReflect.Descriptor instead.
func (*LoginAPIResponse) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{3}
}

func (x *LoginAPIResponse) GetUser() *Users {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *LoginAPIResponse) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type AccountAPIRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uids []string `protobuf:"bytes,1,rep,name=uids,proto3" json:"uids,omitempty"`
}

func (x *AccountAPIRequest) Reset() {
	*x = AccountAPIRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountAPIRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountAPIRequest) ProtoMessage() {}

func (x *AccountAPIRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountAPIRequest.ProtoReflect.Descriptor instead.
func (*AccountAPIRequest) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{4}
}

func (x *AccountAPIRequest) GetUids() []string {
	if x != nil {
		return x.Uids
	}
	return nil
}

type AccountAPIResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users []*Users `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
}

func (x *AccountAPIResponse) Reset() {
	*x = AccountAPIResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountAPIResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountAPIResponse) ProtoMessage() {}

func (x *AccountAPIResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountAPIResponse.ProtoReflect.Descriptor instead.
func (*AccountAPIResponse) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{5}
}

func (x *AccountAPIResponse) GetUsers() []*Users {
	if x != nil {
		return x.Users
	}
	return nil
}

type UserAuth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account  string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
}

func (x *UserAuth) Reset() {
	*x = UserAuth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserAuth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserAuth) ProtoMessage() {}

func (x *UserAuth) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserAuth.ProtoReflect.Descriptor instead.
func (*UserAuth) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{6}
}

func (x *UserAuth) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *UserAuth) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type RegsiterAPIRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// string username=1;
	// string password=2;
	Auth string            `protobuf:"bytes,1,opt,name=auth,proto3" json:"auth,omitempty"`
	Ext  map[string]string `protobuf:"bytes,3,rep,name=ext,proto3" json:"ext,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *RegsiterAPIRequest) Reset() {
	*x = RegsiterAPIRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegsiterAPIRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegsiterAPIRequest) ProtoMessage() {}

func (x *RegsiterAPIRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegsiterAPIRequest.ProtoReflect.Descriptor instead.
func (*RegsiterAPIRequest) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{7}
}

func (x *RegsiterAPIRequest) GetAuth() string {
	if x != nil {
		return x.Auth
	}
	return ""
}

func (x *RegsiterAPIRequest) GetExt() map[string]string {
	if x != nil {
		return x.Ext
	}
	return nil
}

type AuthLogAPIRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp string `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *AuthLogAPIRequest) Reset() {
	*x = AuthLogAPIRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthLogAPIRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthLogAPIRequest) ProtoMessage() {}

func (x *AuthLogAPIRequest) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthLogAPIRequest.ProtoReflect.Descriptor instead.
func (*AuthLogAPIRequest) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{8}
}

func (x *AuthLogAPIRequest) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

type AuthSeed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Seed int64  `protobuf:"varint,1,opt,name=seed,proto3" json:"seed,omitempty"`
	Key  string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *AuthSeed) Reset() {
	*x = AuthSeed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthSeed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthSeed) ProtoMessage() {}

func (x *AuthSeed) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthSeed.ProtoReflect.Descriptor instead.
func (*AuthSeed) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{9}
}

func (x *AuthSeed) GetSeed() int64 {
	if x != nil {
		return x.Seed
	}
	return 0
}

func (x *AuthSeed) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

// 加密种子
type AuthLogAPIResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg       string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	Timestamp string `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *AuthLogAPIResponse) Reset() {
	*x = AuthLogAPIResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_authentication_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthLogAPIResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthLogAPIResponse) ProtoMessage() {}

func (x *AuthLogAPIResponse) ProtoReflect() protoreflect.Message {
	mi := &file_authentication_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthLogAPIResponse.ProtoReflect.Descriptor instead.
func (*AuthLogAPIResponse) Descriptor() ([]byte, []int) {
	return file_authentication_proto_rawDescGZIP(), []int{10}
}

func (x *AuthLogAPIResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *AuthLogAPIResponse) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

var File_authentication_proto protoreflect.FileDescriptor

var file_authentication_proto_rawDesc = []byte{
	0x0a, 0x14, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64,
	0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x77, 0x65, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x75, 0x73, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5f, 0x0a, 0x0f, 0x4c, 0x6f, 0x67, 0x69, 0x6e,
	0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x75,
	0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x65, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x65,
	0x65, 0x64, 0x12, 0x24, 0x0a, 0x0d, 0x69, 0x73, 0x5f, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x6c, 0x6f,
	0x67, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x69, 0x73, 0x5f, 0x6b, 0x65,
	0x65, 0x70, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x22, 0x12, 0x0a, 0x10, 0x4c, 0x6f, 0x67, 0x6f,
	0x75, 0x74, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x13, 0x0a, 0x11,
	0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x56, 0x0a, 0x10, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e,
	0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x73, 0x52, 0x04, 0x75,
	0x73, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x27, 0x0a, 0x11, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x75, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x75, 0x69,
	0x64, 0x73, 0x22, 0x44, 0x0a, 0x12, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x50, 0x49,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x05, 0x75, 0x73, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63,
	0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x55, 0x73, 0x65, 0x72,
	0x73, 0x52, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x22, 0x40, 0x0a, 0x08, 0x55, 0x73, 0x65, 0x72,
	0x41, 0x75, 0x74, 0x68, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1a,
	0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0xa2, 0x01, 0x0a, 0x12, 0x52,
	0x65, 0x67, 0x73, 0x69, 0x74, 0x65, 0x72, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x40, 0x0a, 0x03, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62,
	0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x52, 0x65, 0x67, 0x73, 0x69, 0x74, 0x65, 0x72, 0x41,
	0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x45, 0x78, 0x74, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x03, 0x65, 0x78, 0x74, 0x1a, 0x36, 0x0a, 0x08, 0x45, 0x78, 0x74, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x31, 0x0a, 0x11, 0x41, 0x75, 0x74, 0x68, 0x4c, 0x6f, 0x67, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x22, 0x30, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x53, 0x65, 0x65, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x65, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x65,
	0x65, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x22, 0x44, 0x0a, 0x12, 0x41, 0x75, 0x74, 0x68, 0x4c, 0x6f, 0x67, 0x41,
	0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x32, 0x82, 0x05, 0x0a, 0x0b, 0x41,
	0x75, 0x74, 0x68, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x78, 0x0a, 0x05, 0x6c, 0x6f,
	0x67, 0x69, 0x6e, 0x12, 0x22, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e,
	0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x41, 0x50, 0x49,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63,
	0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x3a, 0x01, 0x2a,
	0x22, 0x12, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x6c,
	0x6f, 0x67, 0x69, 0x6e, 0x12, 0x78, 0x0a, 0x06, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x12, 0x23,
	0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e,
	0x69, 0x61, 0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e,
	0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x1b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x12, 0x13, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x12, 0x7e,
	0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x2e, 0x77, 0x65, 0x74, 0x72,
	0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2c, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f,
	0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1f, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x19, 0x3a, 0x01, 0x2a, 0x22, 0x14, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x7b,
	0x0a, 0x08, 0x61, 0x75, 0x74, 0x68, 0x53, 0x65, 0x65, 0x64, 0x12, 0x24, 0x2e, 0x77, 0x65, 0x74,
	0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x4c, 0x6f, 0x67, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2c, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67,
	0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1b,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x3a, 0x01, 0x2a, 0x22, 0x10, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x6c, 0x6f, 0x67, 0x12, 0x81, 0x01, 0x0a, 0x08,
	0x72, 0x65, 0x67, 0x73, 0x69, 0x74, 0x65, 0x72, 0x12, 0x25, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79,
	0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x52, 0x65, 0x67,
	0x73, 0x69, 0x74, 0x65, 0x72, 0x41, 0x50, 0x49, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2c, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f,
	0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x50, 0x49, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x20, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x1a, 0x3a, 0x01, 0x2a, 0x22, 0x15, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x72, 0x65, 0x67, 0x73, 0x69, 0x74, 0x65, 0x72, 0x42,
	0x25, 0x5a, 0x23, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65,
	0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_authentication_proto_rawDescOnce sync.Once
	file_authentication_proto_rawDescData = file_authentication_proto_rawDesc
)

func file_authentication_proto_rawDescGZIP() []byte {
	file_authentication_proto_rawDescOnce.Do(func() {
		file_authentication_proto_rawDescData = protoimpl.X.CompressGZIP(file_authentication_proto_rawDescData)
	})
	return file_authentication_proto_rawDescData
}

var file_authentication_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_authentication_proto_goTypes = []interface{}{
	(*LoginAPIRequest)(nil),    // 0: wetrycode.begonia.LoginAPIRequest
	(*LogoutAPIRequest)(nil),   // 1: wetrycode.begonia.LogoutAPIRequest
	(*LogoutAPIResponse)(nil),  // 2: wetrycode.begonia.LogoutAPIResponse
	(*LoginAPIResponse)(nil),   // 3: wetrycode.begonia.LoginAPIResponse
	(*AccountAPIRequest)(nil),  // 4: wetrycode.begonia.AccountAPIRequest
	(*AccountAPIResponse)(nil), // 5: wetrycode.begonia.AccountAPIResponse
	(*UserAuth)(nil),           // 6: wetrycode.begonia.UserAuth
	(*RegsiterAPIRequest)(nil), // 7: wetrycode.begonia.RegsiterAPIRequest
	(*AuthLogAPIRequest)(nil),  // 8: wetrycode.begonia.AuthLogAPIRequest
	(*AuthSeed)(nil),           // 9: wetrycode.begonia.AuthSeed
	(*AuthLogAPIResponse)(nil), // 10: wetrycode.begonia.AuthLogAPIResponse
	nil,                        // 11: wetrycode.begonia.RegsiterAPIRequest.ExtEntry
	(*Users)(nil),              // 12: wetrycode.begonia.Users
	(*v1.APIResponse)(nil),     // 13: wetrycode.begonia.common.api.v1.APIResponse
}
var file_authentication_proto_depIdxs = []int32{
	12, // 0: wetrycode.begonia.LoginAPIResponse.user:type_name -> wetrycode.begonia.Users
	12, // 1: wetrycode.begonia.AccountAPIResponse.users:type_name -> wetrycode.begonia.Users
	11, // 2: wetrycode.begonia.RegsiterAPIRequest.ext:type_name -> wetrycode.begonia.RegsiterAPIRequest.ExtEntry
	0,  // 3: wetrycode.begonia.AuthService.login:input_type -> wetrycode.begonia.LoginAPIRequest
	1,  // 4: wetrycode.begonia.AuthService.logout:input_type -> wetrycode.begonia.LogoutAPIRequest
	4,  // 5: wetrycode.begonia.AuthService.account:input_type -> wetrycode.begonia.AccountAPIRequest
	8,  // 6: wetrycode.begonia.AuthService.authSeed:input_type -> wetrycode.begonia.AuthLogAPIRequest
	7,  // 7: wetrycode.begonia.AuthService.regsiter:input_type -> wetrycode.begonia.RegsiterAPIRequest
	13, // 8: wetrycode.begonia.AuthService.login:output_type -> wetrycode.begonia.common.api.v1.APIResponse
	13, // 9: wetrycode.begonia.AuthService.logout:output_type -> wetrycode.begonia.common.api.v1.APIResponse
	13, // 10: wetrycode.begonia.AuthService.account:output_type -> wetrycode.begonia.common.api.v1.APIResponse
	13, // 11: wetrycode.begonia.AuthService.authSeed:output_type -> wetrycode.begonia.common.api.v1.APIResponse
	13, // 12: wetrycode.begonia.AuthService.regsiter:output_type -> wetrycode.begonia.common.api.v1.APIResponse
	8,  // [8:13] is the sub-list for method output_type
	3,  // [3:8] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_authentication_proto_init() }
func file_authentication_proto_init() {
	if File_authentication_proto != nil {
		return
	}
	file_user_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_authentication_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LoginAPIRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogoutAPIRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogoutAPIResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LoginAPIResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountAPIRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountAPIResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserAuth); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegsiterAPIRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthLogAPIRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthSeed); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_authentication_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthLogAPIResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_authentication_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_authentication_proto_goTypes,
		DependencyIndexes: file_authentication_proto_depIdxs,
		MessageInfos:      file_authentication_proto_msgTypes,
	}.Build()
	File_authentication_proto = out.File
	file_authentication_proto_rawDesc = nil
	file_authentication_proto_goTypes = nil
	file_authentication_proto_depIdxs = nil
}
