// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: user.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Role int32

const (
	Role_ADMIN Role = 0
)

// Enum value maps for Role.
var (
	Role_name = map[int32]string{
		0: "ADMIN",
	}
	Role_value = map[string]int32{
		"ADMIN": 0,
	}
)

func (x Role) Enum() *Role {
	p := new(Role)
	*p = x
	return p
}

func (x Role) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Role) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[0].Descriptor()
}

func (Role) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[0]
}

func (x Role) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Role.Descriptor instead.
func (Role) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0}
}

type USER_STATUS int32

const (
	USER_STATUS_ACTIVTE USER_STATUS = 0
)

// Enum value maps for USER_STATUS.
var (
	USER_STATUS_name = map[int32]string{
		0: "ACTIVTE",
	}
	USER_STATUS_value = map[string]int32{
		"ACTIVTE": 0,
	}
)

func (x USER_STATUS) Enum() *USER_STATUS {
	p := new(USER_STATUS)
	*p = x
	return p
}

func (x USER_STATUS) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (USER_STATUS) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[1].Descriptor()
}

func (USER_STATUS) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[1]
}

func (x USER_STATUS) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use USER_STATUS.Descriptor instead.
func (USER_STATUS) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{1}
}

type UserSvrCode int32

const (
	UserSvrCode_USER_UNKONW           UserSvrCode = 0
	UserSvrCode_USER_LOGIN_ERR        UserSvrCode = 4107
	UserSvrCode_USER_TOKEN_EXPIRE_ERR UserSvrCode = 4108
	// 非法的token
	UserSvrCode_USER_TOKEN_INVILDAT_ERR    UserSvrCode = 4109
	UserSvrCode_USER_TOKEN_NOT_ACTIVTE_ERR UserSvrCode = 4114
	UserSvrCode_USER_AUTH_DECRYPT_ERR      UserSvrCode = 4110
	UserSvrCode_USER_ACCOUNT_ERR           UserSvrCode = 4111
	UserSvrCode_USER_PASSWORD_ERR          UserSvrCode = 4112
	UserSvrCode_USER_NOT_FOUND_ERR         UserSvrCode = 4113
)

// Enum value maps for UserSvrCode.
var (
	UserSvrCode_name = map[int32]string{
		0:    "USER_UNKONW",
		4107: "USER_LOGIN_ERR",
		4108: "USER_TOKEN_EXPIRE_ERR",
		4109: "USER_TOKEN_INVILDAT_ERR",
		4114: "USER_TOKEN_NOT_ACTIVTE_ERR",
		4110: "USER_AUTH_DECRYPT_ERR",
		4111: "USER_ACCOUNT_ERR",
		4112: "USER_PASSWORD_ERR",
		4113: "USER_NOT_FOUND_ERR",
	}
	UserSvrCode_value = map[string]int32{
		"USER_UNKONW":                0,
		"USER_LOGIN_ERR":             4107,
		"USER_TOKEN_EXPIRE_ERR":      4108,
		"USER_TOKEN_INVILDAT_ERR":    4109,
		"USER_TOKEN_NOT_ACTIVTE_ERR": 4114,
		"USER_AUTH_DECRYPT_ERR":      4110,
		"USER_ACCOUNT_ERR":           4111,
		"USER_PASSWORD_ERR":          4112,
		"USER_NOT_FOUND_ERR":         4113,
	}
)

func (x UserSvrCode) Enum() *UserSvrCode {
	p := new(UserSvrCode)
	*p = x
	return p
}

func (x UserSvrCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UserSvrCode) Descriptor() protoreflect.EnumDescriptor {
	return file_user_proto_enumTypes[2].Descriptor()
}

func (UserSvrCode) Type() protoreflect.EnumType {
	return &file_user_proto_enumTypes[2]
}

func (x UserSvrCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UserSvrCode.Descriptor instead.
func (UserSvrCode) EnumDescriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{2}
}

type Users struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"
	ID int64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty" doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"`
	// @gotags: doc:"任务ID" json:"uid" gorm:"column:uid;type:varchar(36);unique;comment:uid"
	Uid string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid" doc:"任务ID" gorm:"column:uid;type:varchar(36);unique;comment:uid"`
	// @gotags: doc:"任务ID" json:"name" gorm:"column:name;type:varchar(128);unique;comment:username" aes:"true"
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name" doc:"任务ID" gorm:"column:name;type:varchar(128);unique;comment:username" aes:"true"`
	// @gotags: doc:"任务ID" json:"email" gorm:"column:email;type:varchar(128);unique;comment:Users Email" aes:"true"
	Email string `protobuf:"bytes,4,opt,name=email,proto3" json:"email" doc:"任务ID" gorm:"column:email;type:varchar(128);unique;comment:Users Email" aes:"true"`
	// @gotags: doc:"phone number" gorm:"type:varchar(128);unique;comment:Users Phone Number" aes:"true"
	Phone string `protobuf:"bytes,5,opt,name=phone,proto3" json:"phone,omitempty" doc:"phone number" gorm:"type:varchar(128);unique;comment:Users Phone Number" aes:"true"`
	// @gotags: doc:"account password" gorm:"type:varchar(128);comment:Users account password" aes:"true"
	Password string `protobuf:"bytes,6,opt,name=password,proto3" json:"password,omitempty" doc:"account password" gorm:"type:varchar(128);comment:Users account password" aes:"true"`
	// @gotags: doc:"account avatar" gorm:"type:varchar(512);comment:Users account avatar"
	Avatar string `protobuf:"bytes,7,opt,name=avatar,proto3" json:"avatar,omitempty" doc:"account avatar" gorm:"type:varchar(512);comment:Users account avatar"`
	// @gotags: doc:"account role" gorm:"comment:Users account Role"
	Role Role `protobuf:"varint,8,opt,name=role,proto3,enum=begonia-org.begonia.Role" json:"role,omitempty" doc:"account role" gorm:"comment:Users account Role"`
	// @gotags: doc:"account status" gorm:"comment:Users account status"
	Status USER_STATUS `protobuf:"varint,9,opt,name=status,proto3,enum=begonia-org.begonia.USER_STATUS" json:"status,omitempty" doc:"account status" gorm:"comment:Users account status"`
	// @gotags: doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty" doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"`
	// @gotags: doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,11,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty" doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"`
}

func (x *Users) Reset() {
	*x = Users{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Users) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Users) ProtoMessage() {}

func (x *Users) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Users.ProtoReflect.Descriptor instead.
func (*Users) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{0}
}

func (x *Users) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *Users) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *Users) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Users) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Users) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

func (x *Users) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *Users) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Users) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_ADMIN
}

func (x *Users) GetStatus() USER_STATUS {
	if x != nil {
		return x.Status
	}
	return USER_STATUS_ACTIVTE
}

func (x *Users) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Users) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

type BasicAuth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid         string `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Role        Role   `protobuf:"varint,3,opt,name=role,proto3,enum=begonia-org.begonia.Role" json:"role,omitempty"`
	Audience    string `protobuf:"bytes,4,opt,name=audience,proto3" json:"audience,omitempty"`
	Issuer      string `protobuf:"bytes,5,opt,name=issuer,proto3" json:"issuer,omitempty"`
	NotBefore   int64  `protobuf:"varint,6,opt,name=not_before,json=nbf,proto3" json:"not_before,omitempty"`
	Expiration  int64  `protobuf:"varint,7,opt,name=expiration,json=exp,proto3" json:"expiration,omitempty"`
	IssuedAt    int64  `protobuf:"varint,8,opt,name=issued_at,json=iat,proto3" json:"issued_at,omitempty"`
	IsKeepLogin bool   `protobuf:"varint,9,opt,name=is_keep_login,proto3" json:"is_keep_login,omitempty"`
	Token       string `protobuf:"bytes,10,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *BasicAuth) Reset() {
	*x = BasicAuth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BasicAuth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BasicAuth) ProtoMessage() {}

func (x *BasicAuth) ProtoReflect() protoreflect.Message {
	mi := &file_user_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BasicAuth.ProtoReflect.Descriptor instead.
func (*BasicAuth) Descriptor() ([]byte, []int) {
	return file_user_proto_rawDescGZIP(), []int{1}
}

func (x *BasicAuth) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *BasicAuth) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BasicAuth) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_ADMIN
}

func (x *BasicAuth) GetAudience() string {
	if x != nil {
		return x.Audience
	}
	return ""
}

func (x *BasicAuth) GetIssuer() string {
	if x != nil {
		return x.Issuer
	}
	return ""
}

func (x *BasicAuth) GetNotBefore() int64 {
	if x != nil {
		return x.NotBefore
	}
	return 0
}

func (x *BasicAuth) GetExpiration() int64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

func (x *BasicAuth) GetIssuedAt() int64 {
	if x != nil {
		return x.IssuedAt
	}
	return 0
}

func (x *BasicAuth) GetIsKeepLogin() bool {
	if x != nil {
		return x.IsKeepLogin
	}
	return false
}

func (x *BasicAuth) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

var File_user_proto protoreflect.FileDescriptor

var file_user_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x77, 0x65,
	0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xf8, 0x02, 0x0a, 0x05, 0x55, 0x73, 0x65, 0x72, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x49, 0x44, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74,
	0x61, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72,
	0x12, 0x2b, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17,
	0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e,
	0x69, 0x61, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x36, 0x0a,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e,
	0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69,
	0x61, 0x2e, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x98, 0x02, 0x0a, 0x09,
	0x42, 0x61, 0x73, 0x69, 0x63, 0x41, 0x75, 0x74, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x2b, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e,
	0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69,
	0x61, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x61, 0x75, 0x64, 0x69, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x61, 0x75, 0x64, 0x69, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x73, 0x73, 0x75,
	0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72,
	0x12, 0x17, 0x0a, 0x0a, 0x6e, 0x6f, 0x74, 0x5f, 0x62, 0x65, 0x66, 0x6f, 0x72, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6e, 0x62, 0x66, 0x12, 0x17, 0x0a, 0x0a, 0x65, 0x78, 0x70,
	0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x65,
	0x78, 0x70, 0x12, 0x16, 0x0a, 0x09, 0x69, 0x73, 0x73, 0x75, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x69, 0x61, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x69, 0x73,
	0x5f, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0d, 0x69, 0x73, 0x5f, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2a, 0x11, 0x0a, 0x04, 0x52, 0x6f, 0x6c, 0x65, 0x12, 0x09,
	0x0a, 0x05, 0x41, 0x44, 0x4d, 0x49, 0x4e, 0x10, 0x00, 0x2a, 0x1a, 0x0a, 0x0b, 0x55, 0x53, 0x45,
	0x52, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x43, 0x54, 0x49,
	0x56, 0x54, 0x45, 0x10, 0x00, 0x2a, 0xf2, 0x01, 0x0a, 0x0b, 0x55, 0x73, 0x65, 0x72, 0x53, 0x76,
	0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x55, 0x4e,
	0x4b, 0x4f, 0x4e, 0x57, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0e, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4c,
	0x4f, 0x47, 0x49, 0x4e, 0x5f, 0x45, 0x52, 0x52, 0x10, 0x8b, 0x20, 0x12, 0x1a, 0x0a, 0x15, 0x55,
	0x53, 0x45, 0x52, 0x5f, 0x54, 0x4f, 0x4b, 0x45, 0x4e, 0x5f, 0x45, 0x58, 0x50, 0x49, 0x52, 0x45,
	0x5f, 0x45, 0x52, 0x52, 0x10, 0x8c, 0x20, 0x12, 0x1c, 0x0a, 0x17, 0x55, 0x53, 0x45, 0x52, 0x5f,
	0x54, 0x4f, 0x4b, 0x45, 0x4e, 0x5f, 0x49, 0x4e, 0x56, 0x49, 0x4c, 0x44, 0x41, 0x54, 0x5f, 0x45,
	0x52, 0x52, 0x10, 0x8d, 0x20, 0x12, 0x1f, 0x0a, 0x1a, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x54, 0x4f,
	0x4b, 0x45, 0x4e, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x41, 0x43, 0x54, 0x49, 0x56, 0x54, 0x45, 0x5f,
	0x45, 0x52, 0x52, 0x10, 0x92, 0x20, 0x12, 0x1a, 0x0a, 0x15, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x41,
	0x55, 0x54, 0x48, 0x5f, 0x44, 0x45, 0x43, 0x52, 0x59, 0x50, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x10,
	0x8e, 0x20, 0x12, 0x15, 0x0a, 0x10, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x41, 0x43, 0x43, 0x4f, 0x55,
	0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x10, 0x8f, 0x20, 0x12, 0x16, 0x0a, 0x11, 0x55, 0x53, 0x45,
	0x52, 0x5f, 0x50, 0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x10, 0x90,
	0x20, 0x12, 0x17, 0x0a, 0x12, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f,
	0x55, 0x4e, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x10, 0x91, 0x20, 0x42, 0x25, 0x5a, 0x23, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f,
	0x64, 0x65, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_user_proto_rawDescOnce sync.Once
	file_user_proto_rawDescData = file_user_proto_rawDesc
)

func file_user_proto_rawDescGZIP() []byte {
	file_user_proto_rawDescOnce.Do(func() {
		file_user_proto_rawDescData = protoimpl.X.CompressGZIP(file_user_proto_rawDescData)
	})
	return file_user_proto_rawDescData
}

var file_user_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_user_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_user_proto_goTypes = []interface{}{
	(Role)(0),                     // 0: begonia-org.begonia.Role
	(USER_STATUS)(0),              // 1: begonia-org.begonia.USER_STATUS
	(UserSvrCode)(0),              // 2: begonia-org.begonia.UserSvrCode
	(*Users)(nil),                 // 3: begonia-org.begonia.Users
	(*BasicAuth)(nil),             // 4: begonia-org.begonia.BasicAuth
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_user_proto_depIdxs = []int32{
	0, // 0: begonia-org.begonia.Users.role:type_name -> begonia-org.begonia.Role
	1, // 1: begonia-org.begonia.Users.status:type_name -> begonia-org.begonia.USER_STATUS
	5, // 2: begonia-org.begonia.Users.created_at:type_name -> google.protobuf.Timestamp
	5, // 3: begonia-org.begonia.Users.updated_at:type_name -> google.protobuf.Timestamp
	0, // 4: begonia-org.begonia.BasicAuth.role:type_name -> begonia-org.begonia.Role
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_user_proto_init() }
func file_user_proto_init() {
	if File_user_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_user_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Users); i {
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
		file_user_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BasicAuth); i {
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
			RawDescriptor: file_user_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_user_proto_goTypes,
		DependencyIndexes: file_user_proto_depIdxs,
		EnumInfos:         file_user_proto_enumTypes,
		MessageInfos:      file_user_proto_msgTypes,
	}.Build()
	File_user_proto = out.File
	file_user_proto_rawDesc = nil
	file_user_proto_goTypes = nil
	file_user_proto_depIdxs = nil
}
