// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/handler/jobhdl/protos/job_service.proto

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

var File_internal_handler_jobhdl_protos_job_service_proto protoreflect.FileDescriptor

var file_internal_handler_jobhdl_protos_job_service_proto_rawDesc = []byte{
	0x0a, 0x30, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2f, 0x6a, 0x6f, 0x62, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x03, 0x6a, 0x6f, 0x62, 0x1a, 0x30, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x68, 0x64, 0x6c,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6a, 0x6f, 0x62, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x31, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x68,
	0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6a, 0x6f, 0x62, 0x5f, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xe8, 0x01, 0x0a,
	0x03, 0x4a, 0x6f, 0x62, 0x12, 0x39, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x15,
	0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x30, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x12, 0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x47, 0x65, 0x74,
	0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x6a, 0x6f, 0x62,
	0x2e, 0x47, 0x65, 0x74, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x39, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x15, 0x2e, 0x6a, 0x6f,
	0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4a,
	0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x06,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x15, 0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x6a, 0x6f, 0x62, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x2b, 0x5a, 0x29, 0x76, 0x61, 0x6c, 0x65, 0x74,
	0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_internal_handler_jobhdl_protos_job_service_proto_goTypes = []interface{}{
	(*CreateJobRequest)(nil),  // 0: job.CreateJobRequest
	(*GetJobRequest)(nil),     // 1: job.GetJobRequest
	(*UpdateJobRequest)(nil),  // 2: job.UpdateJobRequest
	(*DeleteJobRequest)(nil),  // 3: job.DeleteJobRequest
	(*CreateJobResponse)(nil), // 4: job.CreateJobResponse
	(*GetJobResponse)(nil),    // 5: job.GetJobResponse
	(*UpdateJobResponse)(nil), // 6: job.UpdateJobResponse
	(*DeleteJobResponse)(nil), // 7: job.DeleteJobResponse
}
var file_internal_handler_jobhdl_protos_job_service_proto_depIdxs = []int32{
	0, // 0: job.Job.Create:input_type -> job.CreateJobRequest
	1, // 1: job.Job.Get:input_type -> job.GetJobRequest
	2, // 2: job.Job.Update:input_type -> job.UpdateJobRequest
	3, // 3: job.Job.Delete:input_type -> job.DeleteJobRequest
	4, // 4: job.Job.Create:output_type -> job.CreateJobResponse
	5, // 5: job.Job.Get:output_type -> job.GetJobResponse
	6, // 6: job.Job.Update:output_type -> job.UpdateJobResponse
	7, // 7: job.Job.Delete:output_type -> job.DeleteJobResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_internal_handler_jobhdl_protos_job_service_proto_init() }
func file_internal_handler_jobhdl_protos_job_service_proto_init() {
	if File_internal_handler_jobhdl_protos_job_service_proto != nil {
		return
	}
	file_internal_handler_jobhdl_protos_job_request_proto_init()
	file_internal_handler_jobhdl_protos_job_response_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_handler_jobhdl_protos_job_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_handler_jobhdl_protos_job_service_proto_goTypes,
		DependencyIndexes: file_internal_handler_jobhdl_protos_job_service_proto_depIdxs,
	}.Build()
	File_internal_handler_jobhdl_protos_job_service_proto = out.File
	file_internal_handler_jobhdl_protos_job_service_proto_rawDesc = nil
	file_internal_handler_jobhdl_protos_job_service_proto_goTypes = nil
	file_internal_handler_jobhdl_protos_job_service_proto_depIdxs = nil
}
