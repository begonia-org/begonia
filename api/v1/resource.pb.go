// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: resource.proto

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

// 权限资源表 (Resources)
type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @gotags: gorm:"primaryKey;autoIncrement;comment:自增id"
	ID int64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty" gorm:"primaryKey;autoIncrement;comment:自增id"`
	// @gotags: json:"resource_key" gorm:"column:resource_key;type:varchar(36);comment:资源id"
	ResourceKey string `protobuf:"bytes,2,opt,name=ResourceKey,proto3" json:"resource_key" gorm:"column:resource_key;type:varchar(36);comment:资源id"`
	// @gotags: json:"resource_name" gorm:"column:resource_name;type:varchar(128);comment:资源名称"
	ResourceName string `protobuf:"bytes,3,opt,name=ResourceName,proto3" json:"resource_name" gorm:"column:resource_name;type:varchar(128);comment:资源名称"`
	// @gotags: json:"resource_table" gorm:"column:resource_table;type:varchar(128);comment:资源表名称"
	ResourceTable string `protobuf:"bytes,4,opt,name=ResourceTable,proto3" json:"resource_table" gorm:"column:resource_table;type:varchar(128);comment:资源表名称"`
	// @gotags: doc:"任务ID" json:"uid" gorm:"column:uid;type:varchar(36);comment:uid"
	Uid string `protobuf:"bytes,5,opt,name=uid,proto3" json:"uid" doc:"任务ID" gorm:"column:uid;type:varchar(36);comment:uid"`
	// @gotags: doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty" doc:"完成时间" gorm:"column:completed_at;type:datetime;serializer:timepb;comment:定时任务的最近一次完成时间"`
	// @gotags: doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"
	UpdatedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty" doc:"更新时间" gorm:"autoUpdateTime;column:updated_at;type:datetime;serializer:timepb;comment:定时任务的更新时间"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_resource_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_resource_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_resource_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetID() int64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *Resource) GetResourceKey() string {
	if x != nil {
		return x.ResourceKey
	}
	return ""
}

func (x *Resource) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *Resource) GetResourceTable() string {
	if x != nil {
		return x.ResourceTable
	}
	return ""
}

func (x *Resource) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *Resource) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Resource) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

var File_resource_proto protoreflect.FileDescriptor

var file_resource_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x11, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f,
	0x6e, 0x69, 0x61, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x02, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x49,
	0x44, 0x12, 0x20, 0x0a, 0x0b, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4b, 0x65, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x4b, 0x65, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12,
	0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x64, 0x41, 0x74, 0x42, 0x25, 0x5a, 0x23, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2f, 0x62, 0x65,
	0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_resource_proto_rawDescOnce sync.Once
	file_resource_proto_rawDescData = file_resource_proto_rawDesc
)

func file_resource_proto_rawDescGZIP() []byte {
	file_resource_proto_rawDescOnce.Do(func() {
		file_resource_proto_rawDescData = protoimpl.X.CompressGZIP(file_resource_proto_rawDescData)
	})
	return file_resource_proto_rawDescData
}

var file_resource_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_resource_proto_goTypes = []interface{}{
	(*Resource)(nil),              // 0: wetrycode.begonia.Resource
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_resource_proto_depIdxs = []int32{
	1, // 0: wetrycode.begonia.Resource.created_at:type_name -> google.protobuf.Timestamp
	1, // 1: wetrycode.begonia.Resource.updated_at:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_resource_proto_init() }
func file_resource_proto_init() {
	if File_resource_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_resource_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Resource); i {
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
			RawDescriptor: file_resource_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_resource_proto_goTypes,
		DependencyIndexes: file_resource_proto_depIdxs,
		MessageInfos:      file_resource_proto_msgTypes,
	}.Build()
	File_resource_proto = out.File
	file_resource_proto_rawDesc = nil
	file_resource_proto_goTypes = nil
	file_resource_proto_depIdxs = nil
}
