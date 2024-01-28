// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: endpoint.proto

package v1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
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

type EndpointStatus int32

const (
	EndpointStatus_ENABLED  EndpointStatus = 0
	EndpointStatus_DISABLED EndpointStatus = 1
)

// Enum value maps for EndpointStatus.
var (
	EndpointStatus_name = map[int32]string{
		0: "ENABLED",
		1: "DISABLED",
	}
	EndpointStatus_value = map[string]int32{
		"ENABLED":  0,
		"DISABLED": 1,
	}
)

func (x EndpointStatus) Enum() *EndpointStatus {
	p := new(EndpointStatus)
	*p = x
	return p
}

func (x EndpointStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EndpointStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_endpoint_proto_enumTypes[0].Descriptor()
}

func (EndpointStatus) Type() protoreflect.EnumType {
	return &file_endpoint_proto_enumTypes[0]
}

func (x EndpointStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EndpointStatus.Descriptor instead.
func (EndpointStatus) EnumDescriptor() ([]byte, []int) {
	return file_endpoint_proto_rawDescGZIP(), []int{0}
}

type Endpoints struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty" doc:"历史ID" gorm:"primaryKey;autoIncrement;comment:自增id"`
	// @gotags: json:"name" gorm:"column:name;type:varchar(128);unique;comment:服务名称"
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name" gorm:"column:name;type:varchar(128);unique;comment:服务名称"`
	// @gotags: json:"service_name" gorm:"column:service_name;type:varchar(128);unique;comment:proto服务名称"
	ServiceName string `protobuf:"bytes,3,opt,name=service_name,proto3" json:"service_name" gorm:"column:service_name;type:varchar(128);unique;comment:proto服务名称"`
	// @gotags: json:"service_name" gorm:"column:service_name;type:text;comment:proto服务描述"
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"service_name" gorm:"column:service_name;type:text;comment:proto服务描述"`
	// @gotags: json:"go_package" gorm:"column:go_package;type:varchar(512);comment:proto服务对应的go mod 包名"
	GoPackage string `protobuf:"bytes,5,opt,name=go_package,proto3" json:"go_package" gorm:"column:go_package;type:varchar(512);comment:proto服务对应的go mod 包名"`
	// @gotags: json:"proto_path" gorm:"column:proto_path;type:varchar(512);comment:proto文件路径"
	ProtoPath string `protobuf:"bytes,6,opt,name=proto_path,proto3" json:"proto_path" gorm:"column:proto_path;type:varchar(512);comment:proto文件路径"`
	// @gotags: json:"endpoint" gorm:"column:endpoint;type:varchar(512);comment:proto服务对应的endpoint 服务地址"
	Endpoint string `protobuf:"bytes,7,opt,name=endpoint,proto3" json:"endpoint" gorm:"column:endpoint;type:varchar(512);comment:proto服务对应的endpoint 服务地址"`
	// @gotags: json:"plugin_id" gorm:"column:plugin_id;type:bigint;comment:proto服务对应的插件id"
	PluginId string `protobuf:"bytes,9,opt,name=plugin_id,proto3" json:"plugin_id" gorm:"column:plugin_id;type:bigint;comment:proto服务对应的插件id"`
	// @gotags: json:"status" gorm:"column:status;type:tinyint;comment:proto服务状态"
	Status EndpointStatus `protobuf:"varint,10,opt,name=status,proto3,enum=begonia-org.begonia.EndpointStatus" json:"status" gorm:"column:status;type:tinyint;comment:proto服务状态"`
	// @gotags: json:"is_deleted" gorm:"column:is_deleted;type:tinyint;comment:proto服务是否删除"
	IsDeleted bool `protobuf:"varint,11,opt,name=is_deleted,proto3" json:"is_deleted" gorm:"column:is_deleted;type:tinyint;comment:proto服务是否删除"`
	// @gotags: doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,12,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty" doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"`
	// @gotags: doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,13,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty" doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"`
}

func (x *Endpoints) Reset() {
	*x = Endpoints{}
	if protoimpl.UnsafeEnabled {
		mi := &file_endpoint_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Endpoints) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Endpoints) ProtoMessage() {}

func (x *Endpoints) ProtoReflect() protoreflect.Message {
	mi := &file_endpoint_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Endpoints.ProtoReflect.Descriptor instead.
func (*Endpoints) Descriptor() ([]byte, []int) {
	return file_endpoint_proto_rawDescGZIP(), []int{0}
}

func (x *Endpoints) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Endpoints) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Endpoints) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *Endpoints) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Endpoints) GetGoPackage() string {
	if x != nil {
		return x.GoPackage
	}
	return ""
}

func (x *Endpoints) GetProtoPath() string {
	if x != nil {
		return x.ProtoPath
	}
	return ""
}

func (x *Endpoints) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *Endpoints) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}

func (x *Endpoints) GetStatus() EndpointStatus {
	if x != nil {
		return x.Status
	}
	return EndpointStatus_ENABLED
}

func (x *Endpoints) GetIsDeleted() bool {
	if x != nil {
		return x.IsDeleted
	}
	return false
}

func (x *Endpoints) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Endpoints) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

type EndpointRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	GoPackage   string `protobuf:"bytes,3,opt,name=go_package,proto3" json:"go_package,omitempty"`
	ProtoPath   string `protobuf:"bytes,4,opt,name=proto_path,proto3" json:"proto_path,omitempty"`
}

func (x *EndpointRequest) Reset() {
	*x = EndpointRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_endpoint_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EndpointRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndpointRequest) ProtoMessage() {}

func (x *EndpointRequest) ProtoReflect() protoreflect.Message {
	mi := &file_endpoint_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndpointRequest.ProtoReflect.Descriptor instead.
func (*EndpointRequest) Descriptor() ([]byte, []int) {
	return file_endpoint_proto_rawDescGZIP(), []int{1}
}

func (x *EndpointRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EndpointRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *EndpointRequest) GetGoPackage() string {
	if x != nil {
		return x.GoPackage
	}
	return ""
}

func (x *EndpointRequest) GetProtoPath() string {
	if x != nil {
		return x.ProtoPath
	}
	return ""
}

var File_endpoint_proto protoreflect.FileDescriptor

var file_endpoint_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x11, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f,
	0x6e, 0x69, 0x61, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc0, 0x03, 0x0a, 0x09, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x0a, 0x0a,
	0x67, 0x6f, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x67, 0x6f, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1a, 0x0a, 0x08,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f,
	0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x73, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x69, 0x73, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x64, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x87, 0x01, 0x0a, 0x0f, 0x45, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x1e, 0x0a, 0x0a, 0x67, 0x6f, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x67, 0x6f, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x74,
	0x68, 0x2a, 0x2b, 0x0a, 0x0e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x0c, 0x0a, 0x08, 0x44, 0x49, 0x53, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x42, 0x25,
	0x5a, 0x23, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x74,
	0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_endpoint_proto_rawDescOnce sync.Once
	file_endpoint_proto_rawDescData = file_endpoint_proto_rawDesc
)

func file_endpoint_proto_rawDescGZIP() []byte {
	file_endpoint_proto_rawDescOnce.Do(func() {
		file_endpoint_proto_rawDescData = protoimpl.X.CompressGZIP(file_endpoint_proto_rawDescData)
	})
	return file_endpoint_proto_rawDescData
}

var file_endpoint_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_endpoint_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_endpoint_proto_goTypes = []interface{}{
	(EndpointStatus)(0),           // 0: begonia-org.begonia.EndpointStatus
	(*Endpoints)(nil),             // 1: begonia-org.begonia.Endpoints
	(*EndpointRequest)(nil),       // 2: begonia-org.begonia.EndpointRequest
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_endpoint_proto_depIdxs = []int32{
	0, // 0: begonia-org.begonia.Endpoints.status:type_name -> begonia-org.begonia.EndpointStatus
	3, // 1: begonia-org.begonia.Endpoints.created_at:type_name -> google.protobuf.Timestamp
	3, // 2: begonia-org.begonia.Endpoints.updated_at:type_name -> google.protobuf.Timestamp
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_endpoint_proto_init() }
func file_endpoint_proto_init() {
	if File_endpoint_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_endpoint_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Endpoints); i {
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
		file_endpoint_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EndpointRequest); i {
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
			RawDescriptor: file_endpoint_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_endpoint_proto_goTypes,
		DependencyIndexes: file_endpoint_proto_depIdxs,
		EnumInfos:         file_endpoint_proto_enumTypes,
		MessageInfos:      file_endpoint_proto_msgTypes,
	}.Build()
	File_endpoint_proto = out.File
	file_endpoint_proto_rawDesc = nil
	file_endpoint_proto_goTypes = nil
	file_endpoint_proto_depIdxs = nil
}
