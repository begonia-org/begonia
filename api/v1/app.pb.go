// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: app.proto

package v1

import (
	v1 "github.com/begonia-org/begonia/common/api/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
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

type APPStatus int32

const (
	APPStatus_APP_ENABLED  APPStatus = 0
	APPStatus_APP_DISABLED APPStatus = 1
)

// Enum value maps for APPStatus.
var (
	APPStatus_name = map[int32]string{
		0: "APP_ENABLED",
		1: "APP_DISABLED",
	}
	APPStatus_value = map[string]int32{
		"APP_ENABLED":  0,
		"APP_DISABLED": 1,
	}
)

func (x APPStatus) Enum() *APPStatus {
	p := new(APPStatus)
	*p = x
	return p
}

func (x APPStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (APPStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_app_proto_enumTypes[0].Descriptor()
}

func (APPStatus) Type() protoreflect.EnumType {
	return &file_app_proto_enumTypes[0]
}

func (x APPStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use APPStatus.Descriptor instead.
func (APPStatus) EnumDescriptor() ([]byte, []int) {
	return file_app_proto_rawDescGZIP(), []int{0}
}

type Apps struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"`
	// @gotags: json:"name" gorm:"column:name;type:varchar(128);unique;comment:app服务名称"
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name" gorm:"column:name;type:varchar(128);unique;comment:app服务名称"`
	// @gotags: json:"description" gorm:"column:description;type:varchar(256);comment:app描述服务"
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description" gorm:"column:description;type:varchar(256);comment:app描述服务"`
	// @gotags: json:"status" gorm:"column:status;type:tinyint;comment:app服务状态"
	Status APPStatus `protobuf:"varint,4,opt,name=status,proto3,enum=begonia.org.begonia.APPStatus" json:"status" gorm:"column:status;type:tinyint;comment:app服务状态"`
	// @gotags: json:"appid" gorm:"column:appid;type:varchar(64);unique;comment:app服务id"
	Appid string `protobuf:"bytes,5,opt,name=appid,proto3" json:"appid" gorm:"column:appid;type:varchar(64);unique;comment:app服务id"`
	// @gotags: json:"access_key" gorm:"column:access_key;type:varchar(64);unique;comment:app服务access_key"
	AccessKey string `protobuf:"bytes,6,opt,name=access_key,proto3" json:"access_key" gorm:"column:access_key;type:varchar(64);unique;comment:app服务access_key"`
	// @gotags json:"secret" gorm:"column:secret;type:varchar(64);unique;comment:app服务secret"
	Secret string `protobuf:"bytes,7,opt,name=secret,proto3" json:"secret,omitempty"`
	// @gotags: json:"owner" gorm:"column:owner;type:varchar(64);comment:app服务拥有者"
	Owner string `protobuf:"bytes,8,opt,name=owner,proto3" json:"owner" gorm:"column:owner;type:varchar(64);comment:app服务拥有者"`
	// @gotags: json:"is_deleted" gorm:"column:is_deleted;type:tinyint;comment:proto服务是否删除"
	IsDeleted bool `protobuf:"varint,9,opt,name=is_deleted,proto3" json:"is_deleted" gorm:"column:is_deleted;type:tinyint;comment:proto服务是否删除"`
	// @gotags: json:"tags" gorm:"column:tags;type:json;comment:app服务标签"
	Tags []string `protobuf:"bytes,10,rep,name=tags,proto3" json:"tags" gorm:"column:tags;type:json;comment:app服务标签"`
	// @gotags: doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,12,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty" doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"`
	// @gotags: doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,13,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty" doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"`
	// @gotags: gorm:"-" json:"update_mask"
	UpdateMask *fieldmaskpb.FieldMask `protobuf:"bytes,18,opt,name=update_mask,json=updateMask,proto3" json:"update_mask" gorm:"-"`
}

func (x *Apps) Reset() {
	*x = Apps{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Apps) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Apps) ProtoMessage() {}

func (x *Apps) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Apps.ProtoReflect.Descriptor instead.
func (*Apps) Descriptor() ([]byte, []int) {
	return file_app_proto_rawDescGZIP(), []int{0}
}

func (x *Apps) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Apps) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Apps) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Apps) GetStatus() APPStatus {
	if x != nil {
		return x.Status
	}
	return APPStatus_APP_ENABLED
}

func (x *Apps) GetAppid() string {
	if x != nil {
		return x.Appid
	}
	return ""
}

func (x *Apps) GetAccessKey() string {
	if x != nil {
		return x.AccessKey
	}
	return ""
}

func (x *Apps) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *Apps) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *Apps) GetIsDeleted() bool {
	if x != nil {
		return x.IsDeleted
	}
	return false
}

func (x *Apps) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Apps) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Apps) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *Apps) GetUpdateMask() *fieldmaskpb.FieldMask {
	if x != nil {
		return x.UpdateMask
	}
	return nil
}

type AddAppsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: json:"apps"
	Apps []*Apps `protobuf:"bytes,1,rep,name=apps,proto3" json:"apps"`
}

func (x *AddAppsRequest) Reset() {
	*x = AddAppsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddAppsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddAppsRequest) ProtoMessage() {}

func (x *AddAppsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddAppsRequest.ProtoReflect.Descriptor instead.
func (*AddAppsRequest) Descriptor() ([]byte, []int) {
	return file_app_proto_rawDescGZIP(), []int{1}
}

func (x *AddAppsRequest) GetApps() []*Apps {
	if x != nil {
		return x.Apps
	}
	return nil
}

type AppsListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: json:"page" gorm:"column:page;type:int;comment:页码"
	Page int32 `protobuf:"varint,1,opt,name=page,proto3" json:"page" gorm:"column:page;type:int;comment:页码"`
	// @gotags: json:"page_size" gorm:"column:page_size;type:int;comment:每页数量"
	PageSize int32    `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size" gorm:"column:page_size;type:int;comment:每页数量"`
	Tags     []string `protobuf:"bytes,3,rep,name=tags,proto3" json:"tags,omitempty"`
	Name     []string `protobuf:"bytes,4,rep,name=name,proto3" json:"name,omitempty"`
	Appid    []string `protobuf:"bytes,5,rep,name=appid,proto3" json:"appid,omitempty"`
	Owner    []string `protobuf:"bytes,6,rep,name=owner,proto3" json:"owner,omitempty"`
	// @gotags: json:"access_key"
	AccessKey []string `protobuf:"bytes,7,rep,name=access_key,json=accessKey,proto3" json:"access_key"`
	// @gotags: json:"status" gorm:"column:status;type:tinyint;comment:app服务状态"
	Status []APPStatus `protobuf:"varint,8,rep,packed,name=status,proto3,enum=begonia.org.begonia.APPStatus" json:"status" gorm:"column:status;type:tinyint;comment:app服务状态"`
}

func (x *AppsListRequest) Reset() {
	*x = AppsListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppsListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppsListRequest) ProtoMessage() {}

func (x *AppsListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppsListRequest.ProtoReflect.Descriptor instead.
func (*AppsListRequest) Descriptor() ([]byte, []int) {
	return file_app_proto_rawDescGZIP(), []int{2}
}

func (x *AppsListRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *AppsListRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *AppsListRequest) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *AppsListRequest) GetName() []string {
	if x != nil {
		return x.Name
	}
	return nil
}

func (x *AppsListRequest) GetAppid() []string {
	if x != nil {
		return x.Appid
	}
	return nil
}

func (x *AppsListRequest) GetOwner() []string {
	if x != nil {
		return x.Owner
	}
	return nil
}

func (x *AppsListRequest) GetAccessKey() []string {
	if x != nil {
		return x.AccessKey
	}
	return nil
}

func (x *AppsListRequest) GetStatus() []APPStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

type AppsListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: json:"apps"
	Apps []*Apps `protobuf:"bytes,1,rep,name=apps,proto3" json:"apps"`
	// @gotags: json:"total" gorm:"column:total;type:int;comment:总数"
	Total int32 `protobuf:"varint,2,opt,name=total,proto3" json:"total" gorm:"column:total;type:int;comment:总数"`
}

func (x *AppsListResponse) Reset() {
	*x = AppsListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_app_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppsListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppsListResponse) ProtoMessage() {}

func (x *AppsListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_app_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppsListResponse.ProtoReflect.Descriptor instead.
func (*AppsListResponse) Descriptor() ([]byte, []int) {
	return file_app_proto_rawDescGZIP(), []int{3}
}

func (x *AppsListResponse) GetApps() []*Apps {
	if x != nil {
		return x.Apps
	}
	return nil
}

func (x *AppsListResponse) GetTotal() int32 {
	if x != nil {
		return x.Total
	}
	return 0
}

var File_app_proto protoreflect.FileDescriptor

var file_app_proto_rawDesc = []byte{
	0x0a, 0x09, 0x61, 0x70, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x62, 0x65, 0x67,
	0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x77, 0x65, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xcf, 0x03, 0x0a, 0x04,
	0x41, 0x70, 0x70, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x62, 0x65, 0x67,
	0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61,
	0x2e, 0x41, 0x50, 0x50, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x70, 0x70, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x61, 0x70, 0x70, 0x69, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x73, 0x5f, 0x64, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x69, 0x73, 0x5f, 0x64,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x0a,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x3b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x18,
	0x12, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4d, 0x61, 0x73,
	0x6b, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x61, 0x73, 0x6b, 0x22, 0x3f, 0x0a,
	0x0e, 0x41, 0x64, 0x64, 0x41, 0x70, 0x70, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x2d, 0x0a, 0x04, 0x61, 0x70, 0x70, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f,
	0x6e, 0x69, 0x61, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x52, 0x04, 0x61, 0x70, 0x70, 0x73, 0x22, 0xed,
	0x01, 0x0a, 0x0f, 0x41, 0x70, 0x70, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73,
	0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53,
	0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61,
	0x70, 0x70, 0x69, 0x64, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x61, 0x70, 0x70, 0x69,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x12, 0x36, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x08, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61,
	0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41, 0x50, 0x50,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x57,
	0x0a, 0x10, 0x41, 0x70, 0x70, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x2d, 0x0a, 0x04, 0x61, 0x70, 0x70, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62,
	0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x52, 0x04, 0x61, 0x70, 0x70,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x2a, 0x2e, 0x0a, 0x09, 0x41, 0x50, 0x50, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x0f, 0x0a, 0x0b, 0x41, 0x50, 0x50, 0x5f, 0x45, 0x4e, 0x41, 0x42,
	0x4c, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x41, 0x50, 0x50, 0x5f, 0x44, 0x49, 0x53,
	0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x32, 0x8d, 0x02, 0x0a, 0x0b, 0x41, 0x70, 0x70, 0x73,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7d, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x41, 0x70,
	0x70, 0x73, 0x12, 0x23, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67,
	0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41, 0x64, 0x64, 0x41, 0x70, 0x70, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69,
	0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x50, 0x49, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x3a,
	0x01, 0x2a, 0x22, 0x12, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x70, 0x2f,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x79, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x41, 0x70, 0x70,
	0x73, 0x12, 0x24, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e,
	0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x41, 0x70, 0x70, 0x73, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69,
	0x61, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x50, 0x49, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12,
	0x10, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x6c, 0x69, 0x73,
	0x74, 0x1a, 0x04, 0x88, 0xb7, 0x18, 0x01, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2d, 0x6f, 0x72,
	0x67, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_app_proto_rawDescOnce sync.Once
	file_app_proto_rawDescData = file_app_proto_rawDesc
)

func file_app_proto_rawDescGZIP() []byte {
	file_app_proto_rawDescOnce.Do(func() {
		file_app_proto_rawDescData = protoimpl.X.CompressGZIP(file_app_proto_rawDescData)
	})
	return file_app_proto_rawDescData
}

var file_app_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_app_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_app_proto_goTypes = []interface{}{
	(APPStatus)(0),                // 0: begonia.org.begonia.APPStatus
	(*Apps)(nil),                  // 1: begonia.org.begonia.Apps
	(*AddAppsRequest)(nil),        // 2: begonia.org.begonia.AddAppsRequest
	(*AppsListRequest)(nil),       // 3: begonia.org.begonia.AppsListRequest
	(*AppsListResponse)(nil),      // 4: begonia.org.begonia.AppsListResponse
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
	(*fieldmaskpb.FieldMask)(nil), // 6: google.protobuf.FieldMask
	(*v1.APIResponse)(nil),        // 7: begonia.org.begonia.common.api.v1.APIResponse
}
var file_app_proto_depIdxs = []int32{
	0, // 0: begonia.org.begonia.Apps.status:type_name -> begonia.org.begonia.APPStatus
	5, // 1: begonia.org.begonia.Apps.created_at:type_name -> google.protobuf.Timestamp
	5, // 2: begonia.org.begonia.Apps.updated_at:type_name -> google.protobuf.Timestamp
	6, // 3: begonia.org.begonia.Apps.update_mask:type_name -> google.protobuf.FieldMask
	1, // 4: begonia.org.begonia.AddAppsRequest.apps:type_name -> begonia.org.begonia.Apps
	0, // 5: begonia.org.begonia.AppsListRequest.status:type_name -> begonia.org.begonia.APPStatus
	1, // 6: begonia.org.begonia.AppsListResponse.apps:type_name -> begonia.org.begonia.Apps
	2, // 7: begonia.org.begonia.AppsService.AddApps:input_type -> begonia.org.begonia.AddAppsRequest
	3, // 8: begonia.org.begonia.AppsService.GetApps:input_type -> begonia.org.begonia.AppsListRequest
	7, // 9: begonia.org.begonia.AppsService.AddApps:output_type -> begonia.org.begonia.common.api.v1.APIResponse
	7, // 10: begonia.org.begonia.AppsService.GetApps:output_type -> begonia.org.begonia.common.api.v1.APIResponse
	9, // [9:11] is the sub-list for method output_type
	7, // [7:9] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_app_proto_init() }
func file_app_proto_init() {
	if File_app_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_app_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Apps); i {
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
		file_app_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddAppsRequest); i {
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
		file_app_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppsListRequest); i {
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
		file_app_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppsListResponse); i {
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
			RawDescriptor: file_app_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_app_proto_goTypes,
		DependencyIndexes: file_app_proto_depIdxs,
		EnumInfos:         file_app_proto_enumTypes,
		MessageInfos:      file_app_proto_msgTypes,
	}.Build()
	File_app_proto = out.File
	file_app_proto_rawDesc = nil
	file_app_proto_goTypes = nil
	file_app_proto_depIdxs = nil
}
