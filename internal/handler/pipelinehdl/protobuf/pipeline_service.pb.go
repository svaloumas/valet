// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/handler/pipelinehdl/protos/pipeline_service.proto

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

var File_internal_handler_pipelinehdl_protos_pipeline_service_proto protoreflect.FileDescriptor

var file_internal_handler_pipelinehdl_protos_pipeline_service_proto_rawDesc = []byte{
	0x0a, 0x3a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x68, 0x64, 0x6c, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x69,
	0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x1a, 0x3a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x70, 0x69, 0x70, 0x65,
	0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x3b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e,
	0x64, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x68, 0x64, 0x6c,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32,
	0xe8, 0x03, 0x0a, 0x08, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x4d, 0x0a, 0x06,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x1f, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69,
	0x6e, 0x65, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x44, 0x0a, 0x03, 0x47,
	0x65, 0x74, 0x12, 0x1c, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x47, 0x65,
	0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1d, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x4f, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x73, 0x12, 0x1d, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x47, 0x65, 0x74,
	0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1e, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x58, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x4a, 0x6f, 0x62, 0x73, 0x12, 0x20, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x2e, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x4a, 0x6f, 0x62, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69,
	0x6e, 0x65, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x4a, 0x6f,
	0x62, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4d, 0x0a, 0x06,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1f, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69,
	0x6e, 0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4d, 0x0a, 0x06, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x1f, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x45, 0x5a, 0x43, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x76, 0x61, 0x6c, 0x6f, 0x75, 0x6d,
	0x61, 0x73, 0x2f, 0x76, 0x61, 0x6c, 0x65, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69,
	0x6e, 0x65, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x3b, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_internal_handler_pipelinehdl_protos_pipeline_service_proto_goTypes = []interface{}{
	(*CreatePipelineRequest)(nil),   // 0: pipeline.CreatePipelineRequest
	(*GetPipelineRequest)(nil),      // 1: pipeline.GetPipelineRequest
	(*GetPipelinesRequest)(nil),     // 2: pipeline.GetPipelinesRequest
	(*GetPipelineJobsRequest)(nil),  // 3: pipeline.GetPipelineJobsRequest
	(*UpdatePipelineRequest)(nil),   // 4: pipeline.UpdatePipelineRequest
	(*DeletePipelineRequest)(nil),   // 5: pipeline.DeletePipelineRequest
	(*CreatePipelineResponse)(nil),  // 6: pipeline.CreatePipelineResponse
	(*GetPipelineResponse)(nil),     // 7: pipeline.GetPipelineResponse
	(*GetPipelinesResponse)(nil),    // 8: pipeline.GetPipelinesResponse
	(*GetPipelineJobsResponse)(nil), // 9: pipeline.GetPipelineJobsResponse
	(*UpdatePipelineResponse)(nil),  // 10: pipeline.UpdatePipelineResponse
	(*DeletePipelineResponse)(nil),  // 11: pipeline.DeletePipelineResponse
}
var file_internal_handler_pipelinehdl_protos_pipeline_service_proto_depIdxs = []int32{
	0,  // 0: pipeline.Pipeline.Create:input_type -> pipeline.CreatePipelineRequest
	1,  // 1: pipeline.Pipeline.Get:input_type -> pipeline.GetPipelineRequest
	2,  // 2: pipeline.Pipeline.GetPipelines:input_type -> pipeline.GetPipelinesRequest
	3,  // 3: pipeline.Pipeline.GetPipelineJobs:input_type -> pipeline.GetPipelineJobsRequest
	4,  // 4: pipeline.Pipeline.Update:input_type -> pipeline.UpdatePipelineRequest
	5,  // 5: pipeline.Pipeline.Delete:input_type -> pipeline.DeletePipelineRequest
	6,  // 6: pipeline.Pipeline.Create:output_type -> pipeline.CreatePipelineResponse
	7,  // 7: pipeline.Pipeline.Get:output_type -> pipeline.GetPipelineResponse
	8,  // 8: pipeline.Pipeline.GetPipelines:output_type -> pipeline.GetPipelinesResponse
	9,  // 9: pipeline.Pipeline.GetPipelineJobs:output_type -> pipeline.GetPipelineJobsResponse
	10, // 10: pipeline.Pipeline.Update:output_type -> pipeline.UpdatePipelineResponse
	11, // 11: pipeline.Pipeline.Delete:output_type -> pipeline.DeletePipelineResponse
	6,  // [6:12] is the sub-list for method output_type
	0,  // [0:6] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_internal_handler_pipelinehdl_protos_pipeline_service_proto_init() }
func file_internal_handler_pipelinehdl_protos_pipeline_service_proto_init() {
	if File_internal_handler_pipelinehdl_protos_pipeline_service_proto != nil {
		return
	}
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_init()
	file_internal_handler_pipelinehdl_protos_pipeline_response_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_handler_pipelinehdl_protos_pipeline_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_handler_pipelinehdl_protos_pipeline_service_proto_goTypes,
		DependencyIndexes: file_internal_handler_pipelinehdl_protos_pipeline_service_proto_depIdxs,
	}.Build()
	File_internal_handler_pipelinehdl_protos_pipeline_service_proto = out.File
	file_internal_handler_pipelinehdl_protos_pipeline_service_proto_rawDesc = nil
	file_internal_handler_pipelinehdl_protos_pipeline_service_proto_goTypes = nil
	file_internal_handler_pipelinehdl_protos_pipeline_service_proto_depIdxs = nil
}
