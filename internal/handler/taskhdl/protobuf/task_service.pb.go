// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/handler/taskhdl/protos/task_service.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_internal_handler_taskhdl_protos_task_service_proto protoreflect.FileDescriptor

var file_internal_handler_taskhdl_protos_task_service_proto_rawDesc = []byte{
	0x0a, 0x32, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x1a, 0x32, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x74, 0x61, 0x73,
	0x6b, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x74, 0x61, 0x73, 0x6b,
	0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x33,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72,
	0x2f, 0x74, 0x61, 0x73, 0x6b, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x32, 0x43, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x3b, 0x0a, 0x08, 0x47,
	0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x12, 0x15, 0x2e, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x47,
	0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16,
	0x2e, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x41, 0x5a, 0x3f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x76, 0x61, 0x6c, 0x6f, 0x75, 0x6d, 0x61, 0x73,
	0x2f, 0x76, 0x61, 0x6c, 0x65, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x68, 0x64, 0x6c, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var file_internal_handler_taskhdl_protos_task_service_proto_goTypes = []interface{}{
	(*GetTasksRequest)(nil),  // 0: task.GetTasksRequest
	(*GetTasksResponse)(nil), // 1: task.GetTasksResponse
}
var file_internal_handler_taskhdl_protos_task_service_proto_depIdxs = []int32{
	0, // 0: task.Task.GetTasks:input_type -> task.GetTasksRequest
	1, // 1: task.Task.GetTasks:output_type -> task.GetTasksResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_internal_handler_taskhdl_protos_task_service_proto_init() }
func file_internal_handler_taskhdl_protos_task_service_proto_init() {
	if File_internal_handler_taskhdl_protos_task_service_proto != nil {
		return
	}
	file_internal_handler_taskhdl_protos_task_request_proto_init()
	file_internal_handler_taskhdl_protos_task_response_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_handler_taskhdl_protos_task_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_handler_taskhdl_protos_task_service_proto_goTypes,
		DependencyIndexes: file_internal_handler_taskhdl_protos_task_service_proto_depIdxs,
	}.Build()
	File_internal_handler_taskhdl_protos_task_service_proto = out.File
	file_internal_handler_taskhdl_protos_task_service_proto_rawDesc = nil
	file_internal_handler_taskhdl_protos_task_service_proto_goTypes = nil
	file_internal_handler_taskhdl_protos_task_service_proto_depIdxs = nil
}
