// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.22.2
// source: options.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_options_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         50033,
		Name:          "begonia-org.begonia.common.api.v1.auth_reqiured",
		Tag:           "varint,50033,opt,name=auth_reqiured",
		Filename:      "options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         50034,
		Name:          "begonia-org.begonia.common.api.v1.method_auth_reqiured",
		Tag:           "varint,50034,opt,name=method_auth_reqiured",
		Filename:      "options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         50035,
		Name:          "begonia-org.begonia.common.api.v1.jsontag",
		Tag:           "varint,50035,opt,name=jsontag",
		Filename:      "options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         50036,
		Name:          "begonia-org.begonia.common.api.v1.msg",
		Tag:           "bytes,50036,opt,name=msg",
		Filename:      "options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         50037,
		Name:          "begonia-org.begonia.common.api.v1.go_mod_pkg",
		Tag:           "bytes,50037,opt,name=go_mod_pkg",
		Filename:      "options.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional bool auth_reqiured = 50033;
	E_AuthReqiured = &file_options_proto_extTypes[0]
	// optional bool method_auth_reqiured = 50034;
	E_MethodAuthReqiured = &file_options_proto_extTypes[1]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional bool jsontag = 50035;
	E_Jsontag = &file_options_proto_extTypes[2]
)

// Extension fields to descriptorpb.EnumValueOptions.
var (
	// optional string msg = 50036;
	E_Msg = &file_options_proto_extTypes[3]
)

// Extension fields to descriptorpb.FileOptions.
var (
	// optional string go_mod_pkg = 50037;
	E_GoModPkg = &file_options_proto_extTypes[4]
)

var File_options_proto protoreflect.FileDescriptor

var file_options_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x1f, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x62, 0x65, 0x67, 0x6f, 0x6e,
	0x69, 0x61, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x3a, 0x49, 0x0a, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x72, 0x65, 0x71, 0x69, 0x75,
	0x72, 0x65, 0x64, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf1, 0x86, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x61, 0x75,
	0x74, 0x68, 0x52, 0x65, 0x71, 0x69, 0x75, 0x72, 0x65, 0x64, 0x88, 0x01, 0x01, 0x3a, 0x56, 0x0a,
	0x14, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x72, 0x65, 0x71,
	0x69, 0x75, 0x72, 0x65, 0x64, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf2, 0x86, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12,
	0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x41, 0x75, 0x74, 0x68, 0x52, 0x65, 0x71, 0x69, 0x75, 0x72,
	0x65, 0x64, 0x88, 0x01, 0x01, 0x3a, 0x3c, 0x0a, 0x07, 0x6a, 0x73, 0x6f, 0x6e, 0x74, 0x61, 0x67,
	0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0xf3, 0x86, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x6a, 0x73, 0x6f, 0x6e, 0x74, 0x61, 0x67,
	0x88, 0x01, 0x01, 0x3a, 0x38, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x21, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6e, 0x75,
	0x6d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf4, 0x86,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x88, 0x01, 0x01, 0x3a, 0x3c, 0x0a,
	0x0a, 0x67, 0x6f, 0x5f, 0x6d, 0x6f, 0x64, 0x5f, 0x70, 0x6b, 0x67, 0x12, 0x1c, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69,
	0x6c, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf5, 0x86, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x67, 0x6f, 0x4d, 0x6f, 0x64, 0x50, 0x6b, 0x67, 0x42, 0x2c, 0x5a, 0x2a, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x65, 0x74, 0x72, 0x79, 0x63,
	0x6f, 0x64, 0x65, 0x2f, 0x62, 0x65, 0x67, 0x6f, 0x6e, 0x69, 0x61, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var file_options_proto_goTypes = []interface{}{
	(*descriptorpb.ServiceOptions)(nil),   // 0: google.protobuf.ServiceOptions
	(*descriptorpb.FieldOptions)(nil),     // 1: google.protobuf.FieldOptions
	(*descriptorpb.EnumValueOptions)(nil), // 2: google.protobuf.EnumValueOptions
	(*descriptorpb.FileOptions)(nil),      // 3: google.protobuf.FileOptions
}
var file_options_proto_depIdxs = []int32{
	0, // 0: begonia-org.begonia.common.api.v1.auth_reqiured:extendee -> google.protobuf.ServiceOptions
	0, // 1: begonia-org.begonia.common.api.v1.method_auth_reqiured:extendee -> google.protobuf.ServiceOptions
	1, // 2: begonia-org.begonia.common.api.v1.jsontag:extendee -> google.protobuf.FieldOptions
	2, // 3: begonia-org.begonia.common.api.v1.msg:extendee -> google.protobuf.EnumValueOptions
	3, // 4: begonia-org.begonia.common.api.v1.go_mod_pkg:extendee -> google.protobuf.FileOptions
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	0, // [0:5] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_options_proto_init() }
func file_options_proto_init() {
	if File_options_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_options_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 5,
			NumServices:   0,
		},
		GoTypes:           file_options_proto_goTypes,
		DependencyIndexes: file_options_proto_depIdxs,
		ExtensionInfos:    file_options_proto_extTypes,
	}.Build()
	File_options_proto = out.File
	file_options_proto_rawDesc = nil
	file_options_proto_goTypes = nil
	file_options_proto_depIdxs = nil
}
