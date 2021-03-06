// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: internal/handler/pipelinehdl/protos/pipeline_request.proto

package pb

import (
	protobuf "github.com/svaloumas/valet/internal/handler/jobhdl/protobuf"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreatePipelineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string                       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description string                       `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	RunAt       string                       `protobuf:"bytes,3,opt,name=run_at,json=runAt,proto3" json:"run_at,omitempty"`
	Jobs        []*protobuf.CreateJobRequest `protobuf:"bytes,4,rep,name=jobs,proto3" json:"jobs,omitempty"`
}

func (x *CreatePipelineRequest) Reset() {
	*x = CreatePipelineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreatePipelineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreatePipelineRequest) ProtoMessage() {}

func (x *CreatePipelineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreatePipelineRequest.ProtoReflect.Descriptor instead.
func (*CreatePipelineRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{0}
}

func (x *CreatePipelineRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreatePipelineRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *CreatePipelineRequest) GetRunAt() string {
	if x != nil {
		return x.RunAt
	}
	return ""
}

func (x *CreatePipelineRequest) GetJobs() []*protobuf.CreateJobRequest {
	if x != nil {
		return x.Jobs
	}
	return nil
}

type GetPipelineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetPipelineRequest) Reset() {
	*x = GetPipelineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPipelineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPipelineRequest) ProtoMessage() {}

func (x *GetPipelineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPipelineRequest.ProtoReflect.Descriptor instead.
func (*GetPipelineRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{1}
}

func (x *GetPipelineRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetPipelinesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *GetPipelinesRequest) Reset() {
	*x = GetPipelinesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPipelinesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPipelinesRequest) ProtoMessage() {}

func (x *GetPipelinesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPipelinesRequest.ProtoReflect.Descriptor instead.
func (*GetPipelinesRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{2}
}

func (x *GetPipelinesRequest) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type GetPipelineJobsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetPipelineJobsRequest) Reset() {
	*x = GetPipelineJobsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPipelineJobsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPipelineJobsRequest) ProtoMessage() {}

func (x *GetPipelineJobsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPipelineJobsRequest.ProtoReflect.Descriptor instead.
func (*GetPipelineJobsRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{3}
}

func (x *GetPipelineJobsRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type UpdatePipelineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *UpdatePipelineRequest) Reset() {
	*x = UpdatePipelineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdatePipelineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdatePipelineRequest) ProtoMessage() {}

func (x *UpdatePipelineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdatePipelineRequest.ProtoReflect.Descriptor instead.
func (*UpdatePipelineRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{4}
}

func (x *UpdatePipelineRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UpdatePipelineRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *UpdatePipelineRequest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

type DeletePipelineRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeletePipelineRequest) Reset() {
	*x = DeletePipelineRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeletePipelineRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeletePipelineRequest) ProtoMessage() {}

func (x *DeletePipelineRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeletePipelineRequest.ProtoReflect.Descriptor instead.
func (*DeletePipelineRequest) Descriptor() ([]byte, []int) {
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP(), []int{5}
}

func (x *DeletePipelineRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_internal_handler_pipelinehdl_protos_pipeline_request_proto protoreflect.FileDescriptor

var file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDesc = []byte{
	0x0a, 0x3a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x72, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x68, 0x64, 0x6c, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x69,
	0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x1a, 0x30, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x68, 0x64, 0x6c, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6a, 0x6f, 0x62, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8f, 0x01, 0x0a, 0x15, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f,
	0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x75, 0x6e, 0x41, 0x74, 0x12,
	0x29, 0x0a, 0x04, 0x6a, 0x6f, 0x62, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e,
	0x6a, 0x6f, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x52, 0x04, 0x6a, 0x6f, 0x62, 0x73, 0x22, 0x24, 0x0a, 0x12, 0x47, 0x65,
	0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x22, 0x2d, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22,
	0x28, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x4a, 0x6f,
	0x62, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x5d, 0x0a, 0x15, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x27, 0x0a, 0x15, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x42, 0x45, 0x5a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x76, 0x61, 0x6c, 0x6f, 0x75, 0x6d, 0x61, 0x73, 0x2f, 0x76, 0x61, 0x6c, 0x65, 0x74, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72,
	0x2f, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x68, 0x64, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescOnce sync.Once
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescData = file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDesc
)

func file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescGZIP() []byte {
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescOnce.Do(func() {
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescData)
	})
	return file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDescData
}

var file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_internal_handler_pipelinehdl_protos_pipeline_request_proto_goTypes = []interface{}{
	(*CreatePipelineRequest)(nil),     // 0: pipeline.CreatePipelineRequest
	(*GetPipelineRequest)(nil),        // 1: pipeline.GetPipelineRequest
	(*GetPipelinesRequest)(nil),       // 2: pipeline.GetPipelinesRequest
	(*GetPipelineJobsRequest)(nil),    // 3: pipeline.GetPipelineJobsRequest
	(*UpdatePipelineRequest)(nil),     // 4: pipeline.UpdatePipelineRequest
	(*DeletePipelineRequest)(nil),     // 5: pipeline.DeletePipelineRequest
	(*protobuf.CreateJobRequest)(nil), // 6: job.CreateJobRequest
}
var file_internal_handler_pipelinehdl_protos_pipeline_request_proto_depIdxs = []int32{
	6, // 0: pipeline.CreatePipelineRequest.jobs:type_name -> job.CreateJobRequest
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_handler_pipelinehdl_protos_pipeline_request_proto_init() }
func file_internal_handler_pipelinehdl_protos_pipeline_request_proto_init() {
	if File_internal_handler_pipelinehdl_protos_pipeline_request_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreatePipelineRequest); i {
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
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPipelineRequest); i {
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
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPipelinesRequest); i {
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
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPipelineJobsRequest); i {
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
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdatePipelineRequest); i {
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
		file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeletePipelineRequest); i {
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
			RawDescriptor: file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_handler_pipelinehdl_protos_pipeline_request_proto_goTypes,
		DependencyIndexes: file_internal_handler_pipelinehdl_protos_pipeline_request_proto_depIdxs,
		MessageInfos:      file_internal_handler_pipelinehdl_protos_pipeline_request_proto_msgTypes,
	}.Build()
	File_internal_handler_pipelinehdl_protos_pipeline_request_proto = out.File
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_rawDesc = nil
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_goTypes = nil
	file_internal_handler_pipelinehdl_protos_pipeline_request_proto_depIdxs = nil
}
