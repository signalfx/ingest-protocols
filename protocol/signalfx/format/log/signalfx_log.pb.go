// Code generated by protoc-gen-go. DO NOT EDIT.
// source: signalfx_log.proto

package sfx_log_model

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type LogRecordFlags int32

const (
	LogRecordFlags_UNDEFINED       LogRecordFlags = 0
	LogRecordFlags_TRACE_FLAG_MASK LogRecordFlags = 255
)

var LogRecordFlags_name = map[int32]string{
	0:   "UNDEFINED",
	255: "TRACE_FLAG_MASK",
}

var LogRecordFlags_value = map[string]int32{
	"UNDEFINED":       0,
	"TRACE_FLAG_MASK": 255,
}

func (x LogRecordFlags) String() string {
	return proto.EnumName(LogRecordFlags_name, int32(x))
}

func (LogRecordFlags) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{0}
}

// LogRequest represents the data that is incoming which is a list of ResourceLogs
type LogRequest struct {
	ResourceLogs         []*ResourceLogs `protobuf:"bytes,1,rep,name=resourceLogs,proto3" json:"resourceLogs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *LogRequest) Reset()         { *m = LogRequest{} }
func (m *LogRequest) String() string { return proto.CompactTextString(m) }
func (*LogRequest) ProtoMessage()    {}
func (*LogRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{0}
}

func (m *LogRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LogRequest.Unmarshal(m, b)
}
func (m *LogRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LogRequest.Marshal(b, m, deterministic)
}
func (m *LogRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LogRequest.Merge(m, src)
}
func (m *LogRequest) XXX_Size() int {
	return xxx_messageInfo_LogRequest.Size(m)
}
func (m *LogRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_LogRequest.DiscardUnknown(m)
}

var xxx_messageInfo_LogRequest proto.InternalMessageInfo

func (m *LogRequest) GetResourceLogs() []*ResourceLogs {
	if m != nil {
		return m.ResourceLogs
	}
	return nil
}

// ResourceLogs handles either sharing a resource with list of LogRecord or unique for each logRecord
type ResourceLogs struct {
	Resource             *KeyValueList `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	LogRecords           []*LogRecord  `protobuf:"bytes,2,rep,name=logRecords,proto3" json:"logRecords,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ResourceLogs) Reset()         { *m = ResourceLogs{} }
func (m *ResourceLogs) String() string { return proto.CompactTextString(m) }
func (*ResourceLogs) ProtoMessage()    {}
func (*ResourceLogs) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{1}
}

func (m *ResourceLogs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResourceLogs.Unmarshal(m, b)
}
func (m *ResourceLogs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResourceLogs.Marshal(b, m, deterministic)
}
func (m *ResourceLogs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResourceLogs.Merge(m, src)
}
func (m *ResourceLogs) XXX_Size() int {
	return xxx_messageInfo_ResourceLogs.Size(m)
}
func (m *ResourceLogs) XXX_DiscardUnknown() {
	xxx_messageInfo_ResourceLogs.DiscardUnknown(m)
}

var xxx_messageInfo_ResourceLogs proto.InternalMessageInfo

func (m *ResourceLogs) GetResource() *KeyValueList {
	if m != nil {
		return m.Resource
	}
	return nil
}

func (m *ResourceLogs) GetLogRecords() []*LogRecord {
	if m != nil {
		return m.LogRecords
	}
	return nil
}

// model is based off https://github.com/open-telemetry/oteps/blob/master/text/0097-log-data-model.md#log-and-event-record-definition
type LogRecord struct {
	Timestamp            *TimeField    `protobuf:"bytes,1,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	TraceID              []byte        `protobuf:"bytes,2,opt,name=TraceID,proto3" json:"TraceID,omitempty"`
	SpanID               []byte        `protobuf:"bytes,3,opt,name=SpanID,proto3" json:"SpanID,omitempty"`
	TraceFlags           uint32        `protobuf:"fixed32,4,opt,name=TraceFlags,proto3" json:"TraceFlags,omitempty"`
	SeverityText         string        `protobuf:"bytes,5,opt,name=SeverityText,proto3" json:"SeverityText,omitempty"`
	SeverityNumber       uint32        `protobuf:"fixed32,6,opt,name=SeverityNumber,proto3" json:"SeverityNumber,omitempty"`
	Name                 string        `protobuf:"bytes,7,opt,name=Name,proto3" json:"Name,omitempty"`
	Body                 *Value        `protobuf:"bytes,8,opt,name=Body,proto3" json:"Body,omitempty"`
	Attributes           *KeyValueList `protobuf:"bytes,9,opt,name=Attributes,proto3" json:"Attributes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *LogRecord) Reset()         { *m = LogRecord{} }
func (m *LogRecord) String() string { return proto.CompactTextString(m) }
func (*LogRecord) ProtoMessage()    {}
func (*LogRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{2}
}

func (m *LogRecord) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LogRecord.Unmarshal(m, b)
}
func (m *LogRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LogRecord.Marshal(b, m, deterministic)
}
func (m *LogRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LogRecord.Merge(m, src)
}
func (m *LogRecord) XXX_Size() int {
	return xxx_messageInfo_LogRecord.Size(m)
}
func (m *LogRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_LogRecord.DiscardUnknown(m)
}

var xxx_messageInfo_LogRecord proto.InternalMessageInfo

func (m *LogRecord) GetTimestamp() *TimeField {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *LogRecord) GetTraceID() []byte {
	if m != nil {
		return m.TraceID
	}
	return nil
}

func (m *LogRecord) GetSpanID() []byte {
	if m != nil {
		return m.SpanID
	}
	return nil
}

func (m *LogRecord) GetTraceFlags() uint32 {
	if m != nil {
		return m.TraceFlags
	}
	return 0
}

func (m *LogRecord) GetSeverityText() string {
	if m != nil {
		return m.SeverityText
	}
	return ""
}

func (m *LogRecord) GetSeverityNumber() uint32 {
	if m != nil {
		return m.SeverityNumber
	}
	return 0
}

func (m *LogRecord) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *LogRecord) GetBody() *Value {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *LogRecord) GetAttributes() *KeyValueList {
	if m != nil {
		return m.Attributes
	}
	return nil
}

// AnyValue is used to store Body attribute different types of values
// @TODO: benchmark protobuf auto-gen with gogo/protobuf
type Value struct {
	// Types that are valid to be assigned to Value:
	//	*Value_StringValue
	//	*Value_BoolValue
	//	*Value_IntValue
	//	*Value_DoubleValue
	//	*Value_ArrayValue
	//	*Value_KvlistValue
	Value                isValue_Value `protobuf_oneof:"value"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Value) Reset()         { *m = Value{} }
func (m *Value) String() string { return proto.CompactTextString(m) }
func (*Value) ProtoMessage()    {}
func (*Value) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{3}
}

func (m *Value) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Value.Unmarshal(m, b)
}
func (m *Value) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Value.Marshal(b, m, deterministic)
}
func (m *Value) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Value.Merge(m, src)
}
func (m *Value) XXX_Size() int {
	return xxx_messageInfo_Value.Size(m)
}
func (m *Value) XXX_DiscardUnknown() {
	xxx_messageInfo_Value.DiscardUnknown(m)
}

var xxx_messageInfo_Value proto.InternalMessageInfo

type isValue_Value interface {
	isValue_Value()
}

type Value_StringValue struct {
	StringValue string `protobuf:"bytes,1,opt,name=string_value,json=stringValue,proto3,oneof"`
}

type Value_BoolValue struct {
	BoolValue bool `protobuf:"varint,2,opt,name=bool_value,json=boolValue,proto3,oneof"`
}

type Value_IntValue struct {
	IntValue int64 `protobuf:"varint,3,opt,name=int_value,json=intValue,proto3,oneof"`
}

type Value_DoubleValue struct {
	DoubleValue float64 `protobuf:"fixed64,4,opt,name=double_value,json=doubleValue,proto3,oneof"`
}

type Value_ArrayValue struct {
	ArrayValue *ValueList `protobuf:"bytes,5,opt,name=array_value,json=arrayValue,proto3,oneof"`
}

type Value_KvlistValue struct {
	KvlistValue *KeyValueList `protobuf:"bytes,6,opt,name=kvlist_value,json=kvlistValue,proto3,oneof"`
}

func (*Value_StringValue) isValue_Value() {}

func (*Value_BoolValue) isValue_Value() {}

func (*Value_IntValue) isValue_Value() {}

func (*Value_DoubleValue) isValue_Value() {}

func (*Value_ArrayValue) isValue_Value() {}

func (*Value_KvlistValue) isValue_Value() {}

func (m *Value) GetValue() isValue_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Value) GetStringValue() string {
	if x, ok := m.GetValue().(*Value_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (m *Value) GetBoolValue() bool {
	if x, ok := m.GetValue().(*Value_BoolValue); ok {
		return x.BoolValue
	}
	return false
}

func (m *Value) GetIntValue() int64 {
	if x, ok := m.GetValue().(*Value_IntValue); ok {
		return x.IntValue
	}
	return 0
}

func (m *Value) GetDoubleValue() float64 {
	if x, ok := m.GetValue().(*Value_DoubleValue); ok {
		return x.DoubleValue
	}
	return 0
}

func (m *Value) GetArrayValue() *ValueList {
	if x, ok := m.GetValue().(*Value_ArrayValue); ok {
		return x.ArrayValue
	}
	return nil
}

func (m *Value) GetKvlistValue() *KeyValueList {
	if x, ok := m.GetValue().(*Value_KvlistValue); ok {
		return x.KvlistValue
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Value) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Value_StringValue)(nil),
		(*Value_BoolValue)(nil),
		(*Value_IntValue)(nil),
		(*Value_DoubleValue)(nil),
		(*Value_ArrayValue)(nil),
		(*Value_KvlistValue)(nil),
	}
}

// some existing schemas have time field as string and hence we should be able to accept such
// time field into the system. Parsing of such fields will happen at data store layer
type TimeField struct {
	// Types that are valid to be assigned to Value:
	//	*TimeField_StringValue
	//	*TimeField_NumericValue
	Value                isTimeField_Value `protobuf_oneof:"value"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *TimeField) Reset()         { *m = TimeField{} }
func (m *TimeField) String() string { return proto.CompactTextString(m) }
func (*TimeField) ProtoMessage()    {}
func (*TimeField) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{4}
}

func (m *TimeField) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TimeField.Unmarshal(m, b)
}
func (m *TimeField) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TimeField.Marshal(b, m, deterministic)
}
func (m *TimeField) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TimeField.Merge(m, src)
}
func (m *TimeField) XXX_Size() int {
	return xxx_messageInfo_TimeField.Size(m)
}
func (m *TimeField) XXX_DiscardUnknown() {
	xxx_messageInfo_TimeField.DiscardUnknown(m)
}

var xxx_messageInfo_TimeField proto.InternalMessageInfo

type isTimeField_Value interface {
	isTimeField_Value()
}

type TimeField_StringValue struct {
	StringValue string `protobuf:"bytes,1,opt,name=string_value,json=stringValue,proto3,oneof"`
}

type TimeField_NumericValue struct {
	NumericValue uint64 `protobuf:"fixed64,2,opt,name=numeric_value,json=numericValue,proto3,oneof"`
}

func (*TimeField_StringValue) isTimeField_Value() {}

func (*TimeField_NumericValue) isTimeField_Value() {}

func (m *TimeField) GetValue() isTimeField_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *TimeField) GetStringValue() string {
	if x, ok := m.GetValue().(*TimeField_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (m *TimeField) GetNumericValue() uint64 {
	if x, ok := m.GetValue().(*TimeField_NumericValue); ok {
		return x.NumericValue
	}
	return 0
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*TimeField) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*TimeField_StringValue)(nil),
		(*TimeField_NumericValue)(nil),
	}
}

// KeyAnyValue is a wrapper for map key's and its corresponding any value used for Attributes
type KeyValue struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                *Value   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeyValue) Reset()         { *m = KeyValue{} }
func (m *KeyValue) String() string { return proto.CompactTextString(m) }
func (*KeyValue) ProtoMessage()    {}
func (*KeyValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{5}
}

func (m *KeyValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyValue.Unmarshal(m, b)
}
func (m *KeyValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyValue.Marshal(b, m, deterministic)
}
func (m *KeyValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyValue.Merge(m, src)
}
func (m *KeyValue) XXX_Size() int {
	return xxx_messageInfo_KeyValue.Size(m)
}
func (m *KeyValue) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyValue.DiscardUnknown(m)
}

var xxx_messageInfo_KeyValue proto.InternalMessageInfo

func (m *KeyValue) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KeyValue) GetValue() *Value {
	if m != nil {
		return m.Value
	}
	return nil
}

// Value list is a list of any values
type ValueList struct {
	Values               []*Value `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValueList) Reset()         { *m = ValueList{} }
func (m *ValueList) String() string { return proto.CompactTextString(m) }
func (*ValueList) ProtoMessage()    {}
func (*ValueList) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{6}
}

func (m *ValueList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValueList.Unmarshal(m, b)
}
func (m *ValueList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValueList.Marshal(b, m, deterministic)
}
func (m *ValueList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValueList.Merge(m, src)
}
func (m *ValueList) XXX_Size() int {
	return xxx_messageInfo_ValueList.Size(m)
}
func (m *ValueList) XXX_DiscardUnknown() {
	xxx_messageInfo_ValueList.DiscardUnknown(m)
}

var xxx_messageInfo_ValueList proto.InternalMessageInfo

func (m *ValueList) GetValues() []*Value {
	if m != nil {
		return m.Values
	}
	return nil
}

// KeyValueMap is a list of {key, any value}
type KeyValueList struct {
	Values               []*KeyValue `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *KeyValueList) Reset()         { *m = KeyValueList{} }
func (m *KeyValueList) String() string { return proto.CompactTextString(m) }
func (*KeyValueList) ProtoMessage()    {}
func (*KeyValueList) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc573788dbb9537, []int{7}
}

func (m *KeyValueList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyValueList.Unmarshal(m, b)
}
func (m *KeyValueList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyValueList.Marshal(b, m, deterministic)
}
func (m *KeyValueList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyValueList.Merge(m, src)
}
func (m *KeyValueList) XXX_Size() int {
	return xxx_messageInfo_KeyValueList.Size(m)
}
func (m *KeyValueList) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyValueList.DiscardUnknown(m)
}

var xxx_messageInfo_KeyValueList proto.InternalMessageInfo

func (m *KeyValueList) GetValues() []*KeyValue {
	if m != nil {
		return m.Values
	}
	return nil
}

func init() {
	proto.RegisterEnum("sfx_log_format.LogRecordFlags", LogRecordFlags_name, LogRecordFlags_value)
	proto.RegisterType((*LogRequest)(nil), "sfx_log_format.LogRequest")
	proto.RegisterType((*ResourceLogs)(nil), "sfx_log_format.ResourceLogs")
	proto.RegisterType((*LogRecord)(nil), "sfx_log_format.LogRecord")
	proto.RegisterType((*Value)(nil), "sfx_log_format.Value")
	proto.RegisterType((*TimeField)(nil), "sfx_log_format.TimeField")
	proto.RegisterType((*KeyValue)(nil), "sfx_log_format.KeyValue")
	proto.RegisterType((*ValueList)(nil), "sfx_log_format.ValueList")
	proto.RegisterType((*KeyValueList)(nil), "sfx_log_format.KeyValueList")
}

func init() { proto.RegisterFile("signalfx_log.proto", fileDescriptor_2cc573788dbb9537) }

var fileDescriptor_2cc573788dbb9537 = []byte{
	// 598 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0x5d, 0x8f, 0xd2, 0x4c,
	0x14, 0xc7, 0x5b, 0x58, 0x0a, 0x3d, 0xdb, 0x7d, 0x79, 0x26, 0x8f, 0x66, 0x4c, 0x7c, 0x21, 0x35,
	0x1a, 0xd4, 0x48, 0x8c, 0xc6, 0xf8, 0xb6, 0x17, 0x0b, 0x02, 0x42, 0x16, 0xb9, 0x18, 0xd0, 0x0b,
	0x6f, 0xb0, 0xc0, 0x6c, 0xd3, 0x6c, 0xdb, 0xc1, 0x99, 0x29, 0x59, 0xae, 0xfd, 0x18, 0x7e, 0x38,
	0x3f, 0x8a, 0xa6, 0xd3, 0x17, 0x0b, 0x2e, 0x71, 0xef, 0x7a, 0xce, 0xf9, 0x9d, 0xff, 0x9c, 0xf9,
	0xcf, 0x4c, 0x01, 0x09, 0xcf, 0x0d, 0x1d, 0xff, 0xfc, 0x72, 0xea, 0x33, 0xb7, 0xb9, 0xe4, 0x4c,
	0x32, 0x74, 0x28, 0x92, 0x70, 0x7a, 0xce, 0x78, 0xe0, 0x48, 0x7b, 0x04, 0x30, 0x64, 0x2e, 0xa1,
	0xdf, 0x22, 0x2a, 0x24, 0x3a, 0x05, 0x8b, 0x53, 0xc1, 0x22, 0x3e, 0xa7, 0x43, 0xe6, 0x0a, 0xac,
	0xd7, 0xcb, 0x8d, 0xfd, 0xe7, 0xb7, 0x9b, 0x9b, 0x4d, 0x4d, 0x52, 0x60, 0xc8, 0x46, 0x87, 0xfd,
	0x5d, 0x07, 0xab, 0x58, 0x46, 0xaf, 0xa1, 0x96, 0x01, 0x58, 0xaf, 0xeb, 0x57, 0xc9, 0x9d, 0xd1,
	0xf5, 0x67, 0xc7, 0x8f, 0xe8, 0xd0, 0x13, 0x92, 0xe4, 0x34, 0x7a, 0x03, 0xe0, 0xc7, 0xa3, 0xcd,
	0x19, 0x5f, 0x08, 0x5c, 0x52, 0xa3, 0xdc, 0xda, 0xee, 0x1d, 0x66, 0x04, 0x29, 0xc0, 0xf6, 0xcf,
	0x12, 0x98, 0x79, 0x05, 0xbd, 0x02, 0x73, 0xe2, 0x05, 0x54, 0x48, 0x27, 0x58, 0xa6, 0x33, 0xfc,
	0xa5, 0x13, 0x03, 0x3d, 0x8f, 0xfa, 0x0b, 0xf2, 0x87, 0x45, 0x18, 0xaa, 0x13, 0xee, 0xcc, 0xe9,
	0xa0, 0x83, 0x4b, 0x75, 0xbd, 0x61, 0x91, 0x2c, 0x44, 0x37, 0xc1, 0x18, 0x2f, 0x9d, 0x70, 0xd0,
	0xc1, 0x65, 0x55, 0x48, 0x23, 0x74, 0x17, 0x40, 0x21, 0x3d, 0xdf, 0x71, 0x05, 0xde, 0xab, 0xeb,
	0x8d, 0x2a, 0x29, 0x64, 0x90, 0x0d, 0xd6, 0x98, 0xae, 0x28, 0xf7, 0xe4, 0x7a, 0x42, 0x2f, 0x25,
	0xae, 0xd4, 0xf5, 0x86, 0x49, 0x36, 0x72, 0xe8, 0x21, 0x1c, 0x66, 0xf1, 0x28, 0x0a, 0x66, 0x94,
	0x63, 0x43, 0xe9, 0x6c, 0x65, 0x11, 0x82, 0xbd, 0x91, 0x13, 0x50, 0x5c, 0x55, 0x1a, 0xea, 0x1b,
	0x3d, 0x82, 0xbd, 0x36, 0x5b, 0xac, 0x71, 0x4d, 0xed, 0xf2, 0xc6, 0xf6, 0x2e, 0x95, 0xcd, 0x44,
	0x21, 0xe8, 0x04, 0xa0, 0x25, 0x25, 0xf7, 0x66, 0x91, 0xa4, 0x02, 0x9b, 0xd7, 0x38, 0x9a, 0x02,
	0x6f, 0xff, 0x28, 0x41, 0x45, 0x55, 0xd0, 0x7d, 0xb0, 0x84, 0xe4, 0x5e, 0xe8, 0x4e, 0x57, 0x71,
	0xac, 0x0c, 0x36, 0xfb, 0x1a, 0xd9, 0x4f, 0xb2, 0x09, 0x74, 0x0f, 0x60, 0xc6, 0x98, 0x9f, 0x22,
	0xb1, 0x99, 0xb5, 0xbe, 0x46, 0xcc, 0x38, 0x97, 0x00, 0x77, 0xc0, 0xf4, 0x42, 0x99, 0xd6, 0x63,
	0x4f, 0xcb, 0x7d, 0x8d, 0xd4, 0xbc, 0x50, 0xe6, 0x8b, 0x2c, 0x58, 0x34, 0xf3, 0x69, 0x4a, 0xc4,
	0xce, 0xea, 0xf1, 0x22, 0x49, 0x36, 0x81, 0x4e, 0x60, 0xdf, 0xe1, 0xdc, 0x59, 0xa7, 0x4c, 0xe5,
	0xea, 0x93, 0xce, 0xf7, 0xd3, 0xd7, 0x08, 0x28, 0x3e, 0xe9, 0x6e, 0x81, 0x75, 0xb1, 0xf2, 0x3d,
	0x91, 0x0d, 0x61, 0xfc, 0xdb, 0x91, 0x78, 0x80, 0xa4, 0x47, 0xa5, 0xda, 0x55, 0xa8, 0xa8, 0x5e,
	0xfb, 0x6b, 0x72, 0xe3, 0xd4, 0x85, 0xba, 0x9e, 0x41, 0x0f, 0xe0, 0x20, 0x8c, 0x02, 0xca, 0xbd,
	0x79, 0xc1, 0x23, 0xa3, 0xaf, 0x11, 0x2b, 0x4d, 0x6f, 0xad, 0x30, 0x80, 0x5a, 0x36, 0x09, 0x3a,
	0x86, 0xf2, 0x05, 0x5d, 0x27, 0xba, 0x24, 0xfe, 0x44, 0x4f, 0x52, 0x4c, 0xa9, 0xec, 0xbc, 0x07,
	0xa9, 0xd4, 0x5b, 0x30, 0xf3, 0x1d, 0xa1, 0xa7, 0x60, 0xa8, 0x6c, 0xf6, 0xf6, 0x77, 0xb4, 0xa6,
	0x90, 0x7d, 0x0a, 0x56, 0xd1, 0x10, 0xf4, 0x6c, 0xab, 0x1d, 0xef, 0xb2, 0x2f, 0x53, 0x78, 0xfc,
	0x12, 0x0e, 0xf3, 0x97, 0x9a, 0xbc, 0x91, 0x03, 0x30, 0x3f, 0x8d, 0x3a, 0xdd, 0xde, 0x60, 0xd4,
	0xed, 0x1c, 0x6b, 0xe8, 0x7f, 0x38, 0x9a, 0x90, 0xd6, 0xfb, 0xee, 0xb4, 0x37, 0x6c, 0x7d, 0x98,
	0x7e, 0x6c, 0x8d, 0xcf, 0x8e, 0x7f, 0xe9, 0xed, 0xff, 0xbe, 0x1c, 0x35, 0xdf, 0x65, 0xda, 0x01,
	0x5b, 0x50, 0x7f, 0x66, 0xa8, 0x3f, 0xdc, 0x8b, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x0b, 0xc7,
	0x0d, 0x68, 0xf7, 0x04, 0x00, 0x00,
}
