// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/handler/resulthdl/protos/jobresult_request.proto

package pb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetJobResultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (x *GetJobResultRequest) Reset() {
	*x = GetJobResultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetJobResultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetJobResultRequest) ProtoMessage() {}

func (x *GetJobResultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetJobResultRequest.ProtoReflect.Descriptor instead.
func (*GetJobResultRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescGZIP(), []int{0}
}

func (x *GetJobResultRequest) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

type DeleteJobResultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (x *DeleteJobResultRequest) Reset() {
	*x = DeleteJobResultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteJobResultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteJobResultRequest) ProtoMessage() {}

func (x *DeleteJobResultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteJobResultRequest.ProtoReflect.Descriptor instead.
func (*DeleteJobResultRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescGZIP(), []int{1}
}

func (x *DeleteJobResultRequest) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

var File_internal_handler_resulthdl_protos_jobresult_request_proto protoreflect.FileDescriptor

var file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDesc = []byte{
	0x0a, 0x39, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x2f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x2f, 0x6a, 0x6f, 0x62, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x5f, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6a, 0x6f, 0x62,
	0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x2c, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x4a, 0x6f, 0x62,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a,
	0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a,
	0x6f, 0x62, 0x49, 0x64, 0x22, 0x2f, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4a, 0x6f,
	0x62, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15,
	0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6a, 0x6f, 0x62, 0x49, 0x64, 0x42, 0x2e, 0x5a, 0x2c, 0x76, 0x61, 0x6c, 0x65, 0x74, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f,
	0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescOnce sync.Once
	file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescData = file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDesc
)

func file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescGZIP() []byte {
	file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescOnce.Do(func() {
		file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescData)
	})
	return file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDescData
}

var file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_internal_handler_resulthdl_protos_jobresult_request_proto_goTypes = []interface{}{
	(*GetJobResultRequest)(nil),    // 0: jobresult.GetJobResultRequest
	(*DeleteJobResultRequest)(nil), // 1: jobresult.DeleteJobResultRequest
}
var file_internal_handler_resulthdl_protos_jobresult_request_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_internal_handler_resulthdl_protos_jobresult_request_proto_init() }
func file_internal_handler_resulthdl_protos_jobresult_request_proto_init() {
	if File_internal_handler_resulthdl_protos_jobresult_request_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetJobResultRequest); i {
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
		file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteJobResultRequest); i {
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
			RawDescriptor: file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_handler_resulthdl_protos_jobresult_request_proto_goTypes,
		DependencyIndexes: file_internal_handler_resulthdl_protos_jobresult_request_proto_depIdxs,
		MessageInfos:      file_internal_handler_resulthdl_protos_jobresult_request_proto_msgTypes,
	}.Build()
	File_internal_handler_resulthdl_protos_jobresult_request_proto = out.File
	file_internal_handler_resulthdl_protos_jobresult_request_proto_rawDesc = nil
	file_internal_handler_resulthdl_protos_jobresult_request_proto_goTypes = nil
	file_internal_handler_resulthdl_protos_jobresult_request_proto_depIdxs = nil
}
