// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bindtype.proto

package bindtype

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

type BindType int32

const (
	BindType_RESOURCE_BINDING_NONE BindType = 0
	// bind or replace existing
	BindType_RESOURCE_BINDING_CONTAINER_BIND            BindType = 1001
	BindType_RESOURCE_BINDING_CONTAINER_ADD             BindType = 1002
	BindType_RESOURCE_BINDING_CONTAINER_CLEAR           BindType = 1010
	BindType_RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION BindType = 1011
	// bind or replace existing
	BindType_RESOURCE_BINDING_SLOT_BIND  BindType = 1101
	BindType_RESOURCE_BINDING_SLOT_CLEAR BindType = 1110
	// bind or replace existing
	BindType_RESOURCE_BINDING_COLLECTION_BIND  BindType = 1201
	BindType_RESOURCE_BINDING_COLLECTION_ADD   BindType = 1202
	BindType_RESOURCE_BINDING_COLLECTION_CLEAR BindType = 1210
	// bind or replace an element of the queue
	BindType_RESOURCE_BINDING_QUEUE_BIND BindType = 2101
	// clear the resources bound to an element of the queue
	BindType_RESOURCE_BINDING_QUEUE_CLEAR BindType = 2110
	// add new member to the tail of the queue
	BindType_RESOURCE_BINDING_QUEUE_PUSH BindType = 2121
	// add new member to the tail while removing the head member
	BindType_RESOURCE_BINDING_QUEUE_PUSHPOP BindType = 2122
	// remove the head member of the queue
	BindType_RESOURCE_BINDING_QUEUE_POP BindType = 2123
	// remove a member of the queue
	//
	// an empty slot will be left after the object is removed
	BindType_RESOURCE_BINDING_QUEUE_REMOVE BindType = 2124
	// bind or replace an element of the queue
	BindType_RESOURCE_BINDING_COLQUEUE_BIND BindType = 2201
	// add to an element of the queue
	BindType_RESOURCE_BINDING_COLQUEUE_ADD BindType = 2202
	// clear the resources bound to an element of the queue
	BindType_RESOURCE_BINDING_COLQUEUE_CLEAR BindType = 2210
	// add new member to the tail of the queue
	BindType_RESOURCE_BINDING_COLQUEUE_PUSH BindType = 2221
	// add new member to the tail while removing the head member
	BindType_RESOURCE_BINDING_COLQUEUE_PUSHPOP BindType = 2222
	// remove the head member of the queue
	BindType_RESOURCE_BINDING_COLQUEUE_POP BindType = 2223
	// remove a member of the queue
	//
	// an empty collection will be left after the object is removed
	BindType_RESOURCE_BINDING_COLQUEUE_REMOVE BindType = 2224
)

var BindType_name = map[int32]string{
	0:    "RESOURCE_BINDING_NONE",
	1001: "RESOURCE_BINDING_CONTAINER_BIND",
	1002: "RESOURCE_BINDING_CONTAINER_ADD",
	1010: "RESOURCE_BINDING_CONTAINER_CLEAR",
	1011: "RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION",
	1101: "RESOURCE_BINDING_SLOT_BIND",
	1110: "RESOURCE_BINDING_SLOT_CLEAR",
	1201: "RESOURCE_BINDING_COLLECTION_BIND",
	1202: "RESOURCE_BINDING_COLLECTION_ADD",
	1210: "RESOURCE_BINDING_COLLECTION_CLEAR",
	2101: "RESOURCE_BINDING_QUEUE_BIND",
	2110: "RESOURCE_BINDING_QUEUE_CLEAR",
	2121: "RESOURCE_BINDING_QUEUE_PUSH",
	2122: "RESOURCE_BINDING_QUEUE_PUSHPOP",
	2123: "RESOURCE_BINDING_QUEUE_POP",
	2124: "RESOURCE_BINDING_QUEUE_REMOVE",
	2201: "RESOURCE_BINDING_COLQUEUE_BIND",
	2202: "RESOURCE_BINDING_COLQUEUE_ADD",
	2210: "RESOURCE_BINDING_COLQUEUE_CLEAR",
	2221: "RESOURCE_BINDING_COLQUEUE_PUSH",
	2222: "RESOURCE_BINDING_COLQUEUE_PUSHPOP",
	2223: "RESOURCE_BINDING_COLQUEUE_POP",
	2224: "RESOURCE_BINDING_COLQUEUE_REMOVE",
}

var BindType_value = map[string]int32{
	"RESOURCE_BINDING_NONE":                      0,
	"RESOURCE_BINDING_CONTAINER_BIND":            1001,
	"RESOURCE_BINDING_CONTAINER_ADD":             1002,
	"RESOURCE_BINDING_CONTAINER_CLEAR":           1010,
	"RESOURCE_BINDING_CONTAINER_CLEAN_DEVIATION": 1011,
	"RESOURCE_BINDING_SLOT_BIND":                 1101,
	"RESOURCE_BINDING_SLOT_CLEAR":                1110,
	"RESOURCE_BINDING_COLLECTION_BIND":           1201,
	"RESOURCE_BINDING_COLLECTION_ADD":            1202,
	"RESOURCE_BINDING_COLLECTION_CLEAR":          1210,
	"RESOURCE_BINDING_QUEUE_BIND":                2101,
	"RESOURCE_BINDING_QUEUE_CLEAR":               2110,
	"RESOURCE_BINDING_QUEUE_PUSH":                2121,
	"RESOURCE_BINDING_QUEUE_PUSHPOP":             2122,
	"RESOURCE_BINDING_QUEUE_POP":                 2123,
	"RESOURCE_BINDING_QUEUE_REMOVE":              2124,
	"RESOURCE_BINDING_COLQUEUE_BIND":             2201,
	"RESOURCE_BINDING_COLQUEUE_ADD":              2202,
	"RESOURCE_BINDING_COLQUEUE_CLEAR":            2210,
	"RESOURCE_BINDING_COLQUEUE_PUSH":             2221,
	"RESOURCE_BINDING_COLQUEUE_PUSHPOP":          2222,
	"RESOURCE_BINDING_COLQUEUE_POP":              2223,
	"RESOURCE_BINDING_COLQUEUE_REMOVE":           2224,
}

func (x BindType) String() string {
	return proto.EnumName(BindType_name, int32(x))
}

func (BindType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_909dbdd80987cb9a, []int{0}
}

func init() {
	proto.RegisterEnum("bindtype.BindType", BindType_name, BindType_value)
}

func init() { proto.RegisterFile("bindtype.proto", fileDescriptor_909dbdd80987cb9a) }

var fileDescriptor_909dbdd80987cb9a = []byte{
	// 332 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xca, 0xcc, 0x4b,
	0x29, 0xa9, 0x2c, 0x48, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0xb5, 0x6e,
	0xb0, 0x71, 0x71, 0x38, 0x65, 0xe6, 0xa5, 0x84, 0x54, 0x16, 0xa4, 0x0a, 0x49, 0x72, 0x89, 0x06,
	0xb9, 0x06, 0xfb, 0x87, 0x06, 0x39, 0xbb, 0xc6, 0x3b, 0x79, 0xfa, 0xb9, 0x78, 0xfa, 0xb9, 0xc7,
	0xfb, 0xf9, 0xfb, 0xb9, 0x0a, 0x30, 0x08, 0xa9, 0x70, 0xc9, 0x63, 0x48, 0x39, 0xfb, 0xfb, 0x85,
	0x38, 0x7a, 0xfa, 0xb9, 0x06, 0x81, 0x45, 0x04, 0x5e, 0xb2, 0x0b, 0x29, 0x73, 0xc9, 0xe1, 0x51,
	0xe5, 0xe8, 0xe2, 0x22, 0xf0, 0x8a, 0x5d, 0x48, 0x95, 0x4b, 0x01, 0x8f, 0x22, 0x67, 0x1f, 0x57,
	0xc7, 0x20, 0x81, 0x4f, 0xec, 0x42, 0xfa, 0x5c, 0x5a, 0x04, 0x94, 0xf9, 0xc5, 0xbb, 0xb8, 0x86,
	0x79, 0x3a, 0x86, 0x78, 0xfa, 0xfb, 0x09, 0x7c, 0x66, 0x17, 0x92, 0xe7, 0x92, 0xc2, 0xd0, 0x10,
	0xec, 0xe3, 0x1f, 0x02, 0x71, 0xdd, 0x59, 0x0e, 0x21, 0x05, 0x2e, 0x69, 0xec, 0x0a, 0x20, 0x76,
	0x5e, 0xe3, 0xc0, 0xe1, 0x34, 0x1f, 0x1f, 0x57, 0x67, 0x90, 0x25, 0x10, 0x83, 0x36, 0x72, 0xe2,
	0x08, 0x0c, 0xb8, 0x32, 0x90, 0x3f, 0x37, 0x71, 0x0a, 0xa9, 0x71, 0x29, 0xe2, 0x53, 0x05, 0xb1,
	0x74, 0x17, 0x27, 0x56, 0x67, 0x05, 0x86, 0xba, 0x86, 0x42, 0x78, 0x02, 0x5b, 0x05, 0x84, 0x14,
	0xb9, 0x64, 0x70, 0xa8, 0x80, 0x18, 0xb2, 0x4f, 0x00, 0x8f, 0x21, 0x01, 0xa1, 0xc1, 0x1e, 0x02,
	0x27, 0x05, 0xb0, 0xc6, 0x0d, 0x42, 0x45, 0x80, 0x7f, 0x80, 0xc0, 0x29, 0x01, 0xac, 0x61, 0x08,
	0x55, 0xe4, 0x1f, 0x20, 0x70, 0x5a, 0x40, 0x48, 0x89, 0x4b, 0x16, 0x87, 0x82, 0x20, 0x57, 0x5f,
	0xff, 0x30, 0x57, 0x81, 0x33, 0x02, 0x38, 0x52, 0x81, 0x0f, 0x92, 0x9f, 0x66, 0x0a, 0x62, 0x35,
	0x08, 0xae, 0x08, 0x14, 0x82, 0xb3, 0x04, 0x71, 0x85, 0x33, 0xb2, 0xd7, 0x17, 0x09, 0xe2, 0xb7,
	0x0e, 0xec, 0xfb, 0xb5, 0x82, 0xb8, 0x22, 0x03, 0x35, 0x00, 0xd6, 0x11, 0x70, 0x16, 0x48, 0xcd,
	0x7a, 0x41, 0x5c, 0xa9, 0x04, 0x25, 0x18, 0x36, 0x08, 0x26, 0xb1, 0x81, 0xf3, 0x9a, 0x31, 0x20,
	0x00, 0x00, 0xff, 0xff, 0x5c, 0xd7, 0x2e, 0x4f, 0x7d, 0x03, 0x00, 0x00,
}
