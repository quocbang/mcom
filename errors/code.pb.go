// Code generated by protoc-gen-go. DO NOT EDIT.
// source: code.proto

package errors

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

// Must advise MUI developer of any change.
type Code int32

const (
	Code_NONE                              Code = 0
	Code_ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD Code = 10000
	Code_USER_NO_PERMISSION                Code = 10100
	Code_USER_UNKNOWN_TOKEN                Code = 10300
	Code_USER_ALREADY_EXISTS               Code = 10400
	Code_ACCOUNT_ROLES_NOT_PERMIT          Code = 10500
	Code_ACCOUNT_SAME_AS_OLD_PASSWORD      Code = 10600
	Code_ACCOUNT_BAD_OLD_PASSWORD          Code = 10700
	Code_USER_NOT_FOUND                    Code = 10800
	Code_ACCOUNT_ALREADY_EXISTS            Code = 11000
	Code_ACCOUNT_NOT_FOUND                 Code = 12000
	// 13xxx for user sign in/out errors
	Code_PREVIOUS_USER_NOT_SIGNED_OUT          Code = 13000
	Code_USER_HAS_NOT_SIGNED_IN                Code = 13100
	Code_STATION_OPERATOR_NOT_MATCH            Code = 13200
	Code_STATION_NOT_FOUND                     Code = 20000
	Code_STATION_ALREADY_EXISTS                Code = 20200
	Code_STATION_PRINTER_NOT_DEFINED           Code = 21000
	Code_STATION_GROUP_ALREADY_EXISTS          Code = 25100
	Code_STATION_GROUP_ID_NOT_FOUND            Code = 25200
	Code_STATION_SITE_NOT_FOUND                Code = 26000
	Code_STATION_SITE_BIND_RECORD_NOT_FOUND    Code = 26010
	Code_STATION_SITE_REMAINING_OBJECTS        Code = 27000
	Code_STATION_SITE_ALREADY_EXISTS           Code = 28000
	Code_STATION_SITE_SUB_TYPE_MISMATCH        Code = 29000
	Code_RESOURCE_NOT_FOUND                    Code = 30000
	Code_RESOURCE_MATERIAL_SHORTAGE            Code = 30010
	Code_RESOURCE_UNAVAILABLE                  Code = 30020
	Code_RESOURCE_EXPIRED                      Code = 30100
	Code_RESOURCE_CONTROL_ABOVE_EXTENDED_COUNT Code = 30401
	Code_RESOURCE_EXISTED                      Code = 30500
	Code_RESOURCES_COUNT_MISMATCH              Code = 30700
	// 31xxx for resource site errors
	Code_RESOURCE_SITE_NOT_SHARED Code = 31100
	Code_WORKORDER_NOT_FOUND      Code = 40000
	// WORKORDER_BAD_BATCH the batch is not allowed to be operated. e.g. CLOSED batch can
	// not be closed.
	Code_WORKORDER_BAD_BATCH Code = 40100
	// WORKORDER_BAD_STATUS work order unexpected status, e.g: load feed work order only
	// allowed ACTIVE status, load collect work order only allowed
	// ACTIVE or CLOSING status.
	Code_WORKORDER_BAD_STATUS Code = 40300
	Code_CARRIER_NOT_FOUND    Code = 50000
	// CARRIER_IN_USE carrier in use and cannot be shared.
	Code_CARRIER_IN_USE         Code = 50100
	Code_CARRIER_QUANTITY_LIMIT Code = 50400
	Code_BATCH_NOT_FOUND        Code = 60000
	Code_BATCH_ALREADY_EXISTS   Code = 60200
	// BATCH_NOT_READY e.g. not enough fed batch to collect.
	Code_BATCH_NOT_READY             Code = 60100
	Code_DEPARTMENT_NOT_FOUND        Code = 70000
	Code_DEPARTMENT_ALREADY_EXISTS   Code = 70200
	Code_PRODUCTION_PLAN_NOT_FOUND   Code = 80000
	Code_PRODUCTION_PLAN_EXISTED     Code = 80100
	Code_RECORD_NOT_FOUND            Code = 81000
	Code_RECORD_ALREADY_EXISTS       Code = 81100
	Code_RECIPE_NOT_FOUND            Code = 90000
	Code_RECIPE_ALREADY_EXISTS       Code = 90100
	Code_PRODUCT_ID_NOT_FOUND        Code = 91000
	Code_SUBSTITUTION_ALREADY_EXISTS Code = 91010
	// duplicated product type in limitary_hour table.
	Code_LIMITARY_HOUR_ALREADY_EXISTS Code = 91110
	Code_LIMITARY_HOUR_NOT_FOUND      Code = 91111
	Code_PROCESS_NOT_FOUND            Code = 92000
	Code_PROCESS_ALREADY_EXISTS       Code = 92100
	Code_INSUFFICIENT_REQUEST         Code = 100000
	Code_INVALID_NUMBER               Code = 100100
	Code_BAD_REQUEST                  Code = 100200
	Code_PRODUCT_ID_MISMATCH          Code = 100300
	Code_BAD_WORK_DATE                Code = 100400
	Code_FAILED_TO_PRINT_RESOURCE     Code = 100500
	Code_WAREHOUSE_NOT_FOUND          Code = 101000
	// USER_STATION_MISMATCH the user is not the station/site operator.
	Code_USER_STATION_MISMATCH Code = 120100
	// STATION_WORKORDER_MISMATCH the work order is not being executed the station
	Code_STATION_WORKORDER_MISMATCH Code = 240100
	// RESOURCE_WORKORDER_QUANTITY_BELOW_MIN the used quantity is less than the minimum
	// quantity in the recipe.
	Code_RESOURCE_WORKORDER_QUANTITY_BELOW_MIN Code = 340201
	// RESOURCE_WORKORDER_QUANTITY_ABOVE_MAX the used quantity is greater than the
	// maximum quantity in the recipe.
	Code_RESOURCE_WORKORDER_QUANTITY_ABOVE_MAX Code = 340202
	// RESOURCE_WORKORDER_BAD_GRADE required specified grade but got others.
	Code_RESOURCE_WORKORDER_BAD_GRADE           Code = 340301
	Code_RESOURCE_WORKORDER_RESOURCE_UNEXPECTED Code = 340400
	// RESOURCE_WORKORDER_RESOURCE_MISSING the resource is required but missing.
	Code_RESOURCE_WORKORDER_RESOURCE_MISSING Code = 340500
	Code_BLOB_ALREADY_EXIST                  Code = 341000
)

var Code_name = map[int32]string{
	0:      "NONE",
	10000:  "ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD",
	10100:  "USER_NO_PERMISSION",
	10300:  "USER_UNKNOWN_TOKEN",
	10400:  "USER_ALREADY_EXISTS",
	10500:  "ACCOUNT_ROLES_NOT_PERMIT",
	10600:  "ACCOUNT_SAME_AS_OLD_PASSWORD",
	10700:  "ACCOUNT_BAD_OLD_PASSWORD",
	10800:  "USER_NOT_FOUND",
	11000:  "ACCOUNT_ALREADY_EXISTS",
	12000:  "ACCOUNT_NOT_FOUND",
	13000:  "PREVIOUS_USER_NOT_SIGNED_OUT",
	13100:  "USER_HAS_NOT_SIGNED_IN",
	13200:  "STATION_OPERATOR_NOT_MATCH",
	20000:  "STATION_NOT_FOUND",
	20200:  "STATION_ALREADY_EXISTS",
	21000:  "STATION_PRINTER_NOT_DEFINED",
	25100:  "STATION_GROUP_ALREADY_EXISTS",
	25200:  "STATION_GROUP_ID_NOT_FOUND",
	26000:  "STATION_SITE_NOT_FOUND",
	26010:  "STATION_SITE_BIND_RECORD_NOT_FOUND",
	27000:  "STATION_SITE_REMAINING_OBJECTS",
	28000:  "STATION_SITE_ALREADY_EXISTS",
	29000:  "STATION_SITE_SUB_TYPE_MISMATCH",
	30000:  "RESOURCE_NOT_FOUND",
	30010:  "RESOURCE_MATERIAL_SHORTAGE",
	30020:  "RESOURCE_UNAVAILABLE",
	30100:  "RESOURCE_EXPIRED",
	30401:  "RESOURCE_CONTROL_ABOVE_EXTENDED_COUNT",
	30500:  "RESOURCE_EXISTED",
	30700:  "RESOURCES_COUNT_MISMATCH",
	31100:  "RESOURCE_SITE_NOT_SHARED",
	40000:  "WORKORDER_NOT_FOUND",
	40100:  "WORKORDER_BAD_BATCH",
	40300:  "WORKORDER_BAD_STATUS",
	50000:  "CARRIER_NOT_FOUND",
	50100:  "CARRIER_IN_USE",
	50400:  "CARRIER_QUANTITY_LIMIT",
	60000:  "BATCH_NOT_FOUND",
	60200:  "BATCH_ALREADY_EXISTS",
	60100:  "BATCH_NOT_READY",
	70000:  "DEPARTMENT_NOT_FOUND",
	70200:  "DEPARTMENT_ALREADY_EXISTS",
	80000:  "PRODUCTION_PLAN_NOT_FOUND",
	80100:  "PRODUCTION_PLAN_EXISTED",
	81000:  "RECORD_NOT_FOUND",
	81100:  "RECORD_ALREADY_EXISTS",
	90000:  "RECIPE_NOT_FOUND",
	90100:  "RECIPE_ALREADY_EXISTS",
	91000:  "PRODUCT_ID_NOT_FOUND",
	91010:  "SUBSTITUTION_ALREADY_EXISTS",
	91110:  "LIMITARY_HOUR_ALREADY_EXISTS",
	91111:  "LIMITARY_HOUR_NOT_FOUND",
	92000:  "PROCESS_NOT_FOUND",
	92100:  "PROCESS_ALREADY_EXISTS",
	100000: "INSUFFICIENT_REQUEST",
	100100: "INVALID_NUMBER",
	100200: "BAD_REQUEST",
	100300: "PRODUCT_ID_MISMATCH",
	100400: "BAD_WORK_DATE",
	100500: "FAILED_TO_PRINT_RESOURCE",
	101000: "WAREHOUSE_NOT_FOUND",
	120100: "USER_STATION_MISMATCH",
	240100: "STATION_WORKORDER_MISMATCH",
	340201: "RESOURCE_WORKORDER_QUANTITY_BELOW_MIN",
	340202: "RESOURCE_WORKORDER_QUANTITY_ABOVE_MAX",
	340301: "RESOURCE_WORKORDER_BAD_GRADE",
	340400: "RESOURCE_WORKORDER_RESOURCE_UNEXPECTED",
	340500: "RESOURCE_WORKORDER_RESOURCE_MISSING",
	341000: "BLOB_ALREADY_EXIST",
}

var Code_value = map[string]int32{
	"NONE":                                   0,
	"ACCOUNT_NOT_FOUND_OR_BAD_PASSWORD":      10000,
	"USER_NO_PERMISSION":                     10100,
	"USER_UNKNOWN_TOKEN":                     10300,
	"USER_ALREADY_EXISTS":                    10400,
	"ACCOUNT_ROLES_NOT_PERMIT":               10500,
	"ACCOUNT_SAME_AS_OLD_PASSWORD":           10600,
	"ACCOUNT_BAD_OLD_PASSWORD":               10700,
	"USER_NOT_FOUND":                         10800,
	"ACCOUNT_ALREADY_EXISTS":                 11000,
	"ACCOUNT_NOT_FOUND":                      12000,
	"PREVIOUS_USER_NOT_SIGNED_OUT":           13000,
	"USER_HAS_NOT_SIGNED_IN":                 13100,
	"STATION_OPERATOR_NOT_MATCH":             13200,
	"STATION_NOT_FOUND":                      20000,
	"STATION_ALREADY_EXISTS":                 20200,
	"STATION_PRINTER_NOT_DEFINED":            21000,
	"STATION_GROUP_ALREADY_EXISTS":           25100,
	"STATION_GROUP_ID_NOT_FOUND":             25200,
	"STATION_SITE_NOT_FOUND":                 26000,
	"STATION_SITE_BIND_RECORD_NOT_FOUND":     26010,
	"STATION_SITE_REMAINING_OBJECTS":         27000,
	"STATION_SITE_ALREADY_EXISTS":            28000,
	"STATION_SITE_SUB_TYPE_MISMATCH":         29000,
	"RESOURCE_NOT_FOUND":                     30000,
	"RESOURCE_MATERIAL_SHORTAGE":             30010,
	"RESOURCE_UNAVAILABLE":                   30020,
	"RESOURCE_EXPIRED":                       30100,
	"RESOURCE_CONTROL_ABOVE_EXTENDED_COUNT":  30401,
	"RESOURCE_EXISTED":                       30500,
	"RESOURCES_COUNT_MISMATCH":               30700,
	"RESOURCE_SITE_NOT_SHARED":               31100,
	"WORKORDER_NOT_FOUND":                    40000,
	"WORKORDER_BAD_BATCH":                    40100,
	"WORKORDER_BAD_STATUS":                   40300,
	"CARRIER_NOT_FOUND":                      50000,
	"CARRIER_IN_USE":                         50100,
	"CARRIER_QUANTITY_LIMIT":                 50400,
	"BATCH_NOT_FOUND":                        60000,
	"BATCH_ALREADY_EXISTS":                   60200,
	"BATCH_NOT_READY":                        60100,
	"DEPARTMENT_NOT_FOUND":                   70000,
	"DEPARTMENT_ALREADY_EXISTS":              70200,
	"PRODUCTION_PLAN_NOT_FOUND":              80000,
	"PRODUCTION_PLAN_EXISTED":                80100,
	"RECORD_NOT_FOUND":                       81000,
	"RECORD_ALREADY_EXISTS":                  81100,
	"RECIPE_NOT_FOUND":                       90000,
	"RECIPE_ALREADY_EXISTS":                  90100,
	"PRODUCT_ID_NOT_FOUND":                   91000,
	"SUBSTITUTION_ALREADY_EXISTS":            91010,
	"LIMITARY_HOUR_ALREADY_EXISTS":           91110,
	"LIMITARY_HOUR_NOT_FOUND":                91111,
	"PROCESS_NOT_FOUND":                      92000,
	"PROCESS_ALREADY_EXISTS":                 92100,
	"INSUFFICIENT_REQUEST":                   100000,
	"INVALID_NUMBER":                         100100,
	"BAD_REQUEST":                            100200,
	"PRODUCT_ID_MISMATCH":                    100300,
	"BAD_WORK_DATE":                          100400,
	"FAILED_TO_PRINT_RESOURCE":               100500,
	"WAREHOUSE_NOT_FOUND":                    101000,
	"USER_STATION_MISMATCH":                  120100,
	"STATION_WORKORDER_MISMATCH":             240100,
	"RESOURCE_WORKORDER_QUANTITY_BELOW_MIN":  340201,
	"RESOURCE_WORKORDER_QUANTITY_ABOVE_MAX":  340202,
	"RESOURCE_WORKORDER_BAD_GRADE":           340301,
	"RESOURCE_WORKORDER_RESOURCE_UNEXPECTED": 340400,
	"RESOURCE_WORKORDER_RESOURCE_MISSING":    340500,
	"BLOB_ALREADY_EXIST":                     341000,
}

func (x Code) String() string {
	return proto.EnumName(Code_name, int32(x))
}

func (Code) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_6e9b0151640170c3, []int{0}
}

func init() {
	proto.RegisterEnum("errors.Code", Code_name, Code_value)
}

func init() { proto.RegisterFile("code.proto", fileDescriptor_6e9b0151640170c3) }

var fileDescriptor_6e9b0151640170c3 = []byte{
	// 1188 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x56, 0xdb, 0x8b, 0x1c, 0xc5,
	0x17, 0xfe, 0xf5, 0xce, 0x25, 0x4d, 0xfd, 0x70, 0xad, 0xad, 0xed, 0x6c, 0x36, 0x77, 0x33, 0x9a,
	0x78, 0x65, 0xf3, 0xe0, 0x5f, 0x50, 0xdd, 0x5d, 0x3b, 0x53, 0xa6, 0xa7, 0xaa, 0x53, 0x55, 0xbd,
	0x17, 0x41, 0x8a, 0x5c, 0xd6, 0x20, 0x1a, 0x47, 0xd6, 0x80, 0xaf, 0x22, 0x1b, 0x5d, 0x41, 0xe3,
	0x0a, 0x11, 0x82, 0x18, 0x08, 0xb2, 0x0f, 0x62, 0x7c, 0x58, 0xc1, 0x87, 0x98, 0x08, 0x51, 0x58,
	0x4c, 0x20, 0x51, 0x83, 0x46, 0x09, 0x92, 0x87, 0x4c, 0x0c, 0xba, 0xee, 0x8e, 0x9a, 0x48, 0x1e,
	0x56, 0xf0, 0x41, 0xaa, 0x67, 0x7a, 0xa6, 0xbb, 0x13, 0xf4, 0xad, 0xb7, 0xbe, 0xaf, 0xce, 0xf9,
	0x4e, 0x9d, 0xf3, 0x9d, 0x1d, 0x00, 0xf6, 0x35, 0xf6, 0x4f, 0x8d, 0xbc, 0x30, 0xdd, 0x38, 0xd4,
	0x40, 0xe5, 0xa9, 0xe9, 0xe9, 0xc6, 0xf4, 0x8b, 0x8f, 0x7c, 0x3c, 0x00, 0x8a, 0x5e, 0x63, 0xff,
	0x14, 0xb2, 0x41, 0x91, 0x71, 0x46, 0xe0, 0xff, 0xd0, 0x0e, 0xb0, 0x0d, 0x7b, 0x1e, 0x8f, 0x98,
	0xd2, 0x8c, 0x2b, 0x3d, 0xca, 0x23, 0xe6, 0x6b, 0x2e, 0xb4, 0x8b, 0x7d, 0x1d, 0x62, 0x29, 0xc7,
	0xb9, 0xf0, 0xe1, 0x1c, 0x43, 0xeb, 0x00, 0x8a, 0x24, 0x11, 0x9a, 0x71, 0x1d, 0x12, 0x51, 0xa7,
	0x52, 0x52, 0xce, 0xe0, 0xed, 0x1e, 0x10, 0xb1, 0x5d, 0x8c, 0x8f, 0x33, 0xad, 0xf8, 0x2e, 0xc2,
	0xe0, 0x67, 0x21, 0x1a, 0x06, 0x83, 0x31, 0x80, 0x03, 0x41, 0xb0, 0x3f, 0xa9, 0xc9, 0x04, 0x95,
	0x4a, 0xc2, 0x13, 0xbb, 0xd1, 0x66, 0x30, 0x9c, 0xe4, 0x14, 0x3c, 0x20, 0x32, 0xce, 0x1c, 0x47,
	0x55, 0x70, 0x46, 0xa0, 0x6d, 0x60, 0x53, 0x02, 0x4b, 0x5c, 0x27, 0x1a, 0x4b, 0xcd, 0x83, 0x94,
	0x9a, 0x25, 0x91, 0x8e, 0x60, 0x84, 0x66, 0xe0, 0x8b, 0x12, 0x0d, 0x82, 0xfe, 0x8e, 0xd8, 0x4e,
	0x45, 0x70, 0x41, 0xa1, 0x8d, 0x60, 0x28, 0xb9, 0x93, 0x93, 0xb4, 0x1a, 0xa1, 0x21, 0x30, 0x70,
	0xc7, 0x33, 0xc0, 0x6b, 0x4f, 0x19, 0x2d, 0xa1, 0x20, 0x63, 0x94, 0x47, 0x52, 0x77, 0x43, 0x4a,
	0x5a, 0x65, 0xc4, 0xd7, 0x3c, 0x52, 0xf0, 0xfc, 0x94, 0x89, 0x1b, 0x23, 0x35, 0x2c, 0xd3, 0x28,
	0x65, 0xf0, 0xa3, 0xa7, 0xd1, 0x56, 0xb0, 0x41, 0x2a, 0xac, 0x28, 0x67, 0x9a, 0x87, 0x44, 0x60,
	0xc5, 0xdb, 0x21, 0xea, 0x58, 0x79, 0x35, 0x38, 0x77, 0x00, 0xad, 0x03, 0x03, 0x09, 0xa1, 0x97,
	0xf8, 0xc4, 0x7b, 0x16, 0xda, 0x04, 0x86, 0x12, 0x20, 0x27, 0x77, 0xe9, 0xb8, 0x85, 0xb6, 0x81,
	0x8d, 0x09, 0x1a, 0x0a, 0xca, 0x54, 0x47, 0x99, 0x4f, 0x46, 0x29, 0x23, 0x3e, 0x9c, 0x9d, 0xb7,
	0x50, 0x05, 0x6c, 0x4a, 0x28, 0x55, 0xc1, 0xa3, 0x30, 0x1f, 0xe6, 0x8d, 0x45, 0x0b, 0xdd, 0xd7,
	0x93, 0xd7, 0xe6, 0x50, 0x3f, 0x25, 0xe3, 0xe6, 0x62, 0x46, 0x86, 0xa4, 0x8a, 0xa4, 0xd0, 0xb9,
	0x0b, 0x16, 0x7a, 0x08, 0x54, 0x32, 0xa8, 0x4b, 0x99, 0xaf, 0x05, 0xf1, 0xb8, 0x48, 0xc7, 0x79,
	0xf7, 0x82, 0x85, 0x1e, 0x00, 0x5b, 0x32, 0x4c, 0x41, 0xea, 0x98, 0x32, 0xca, 0xaa, 0x9a, 0xbb,
	0x4f, 0x10, 0xcf, 0x74, 0xe1, 0xdb, 0x4c, 0x59, 0x31, 0x2b, 0x27, 0xf9, 0xda, 0x8f, 0x77, 0x06,
	0x92, 0x91, 0xab, 0xd5, 0x64, 0x48, 0x74, 0x9d, 0xca, 0xf6, 0xab, 0x9e, 0xbf, 0x6e, 0xa1, 0x61,
	0x80, 0x04, 0x91, 0x3c, 0x12, 0x5e, 0x5a, 0xf2, 0xc2, 0x72, 0x5c, 0x72, 0x17, 0xa9, 0x63, 0x45,
	0x04, 0xc5, 0x81, 0x96, 0x35, 0x2e, 0x14, 0xae, 0x12, 0x78, 0x7a, 0xd9, 0x42, 0x1b, 0x80, 0xd3,
	0x65, 0x44, 0x0c, 0x8f, 0x61, 0x1a, 0x60, 0x37, 0x20, 0x70, 0x71, 0xd9, 0x42, 0x43, 0x00, 0x76,
	0x31, 0x32, 0x11, 0x52, 0x41, 0x7c, 0x78, 0x74, 0xc5, 0x42, 0x8f, 0x82, 0xed, 0xdd, 0x73, 0x8f,
	0x33, 0x25, 0x78, 0xa0, 0xb1, 0xcb, 0xc7, 0x0c, 0x4b, 0x11, 0xe6, 0x13, 0x5f, 0xc7, 0xd3, 0x05,
	0xbf, 0xf8, 0x2d, 0x1f, 0x84, 0x4a, 0x45, 0x7c, 0x38, 0xff, 0xbb, 0x85, 0xb6, 0x80, 0xe1, 0xe4,
	0x5c, 0xb6, 0xe9, 0xbd, 0xa2, 0x5a, 0x7f, 0x64, 0xf0, 0x5e, 0x33, 0x64, 0x0d, 0x1b, 0x11, 0x7f,
	0xff, 0x69, 0xa1, 0xf5, 0x60, 0x70, 0x9c, 0x8b, 0x5d, 0x5c, 0xf8, 0x99, 0xd9, 0xff, 0xfc, 0x54,
	0x5f, 0x16, 0x32, 0x96, 0x71, 0xe3, 0xa8, 0xf3, 0x9f, 0xf6, 0x99, 0x72, 0xb3, 0x90, 0x79, 0xde,
	0x48, 0xc2, 0xd6, 0xe9, 0x3e, 0x33, 0x9d, 0x1e, 0x16, 0x82, 0x66, 0xe2, 0x5d, 0x7a, 0xb5, 0x80,
	0x1c, 0xd0, 0x9f, 0x00, 0x94, 0x19, 0x67, 0xc0, 0x4f, 0x5e, 0x2b, 0x98, 0x61, 0x49, 0x4e, 0x77,
	0x47, 0x98, 0x29, 0xaa, 0x26, 0x75, 0x40, 0x8d, 0xad, 0xaf, 0xbd, 0x5e, 0x40, 0x6b, 0xc1, 0xbd,
	0x71, 0xd6, 0xb4, 0xc3, 0x2e, 0x17, 0x4c, 0xfe, 0xf6, 0x71, 0xae, 0xd9, 0x1f, 0xfc, 0x90, 0xbb,
	0x12, 0xa3, 0x70, 0xf1, 0xfb, 0xf8, 0x8a, 0x4f, 0x42, 0x2c, 0x54, 0x9d, 0x64, 0x0c, 0x7b, 0xf3,
	0xfd, 0x22, 0xda, 0x0a, 0xd6, 0xa7, 0xb0, 0x5c, 0xcc, 0x53, 0xf3, 0x31, 0x21, 0x14, 0xdc, 0x8f,
	0xbc, 0xb6, 0x7b, 0x02, 0x9c, 0x76, 0xde, 0xcb, 0xb7, 0x8a, 0x68, 0x33, 0x58, 0x97, 0x27, 0x24,
	0x5d, 0xba, 0x71, 0xab, 0xd8, 0xee, 0x5e, 0x6e, 0xc2, 0x97, 0x56, 0x8b, 0x68, 0x23, 0x58, 0xdb,
	0x39, 0xcf, 0x25, 0xbd, 0xf8, 0x57, 0x72, 0x89, 0x86, 0x19, 0x03, 0x9d, 0x2d, 0x75, 0x2e, 0x99,
	0xf3, 0xdc, 0xa5, 0xdb, 0x67, 0x4b, 0xa6, 0xcc, 0x8e, 0x90, 0xac, 0x2f, 0x57, 0xbf, 0x2c, 0xc5,
	0x4e, 0x89, 0x5c, 0xa9, 0xa8, 0x8a, 0xee, 0xb6, 0x23, 0x5e, 0x39, 0x57, 0x32, 0x0b, 0x20, 0x7e,
	0x7c, 0x2c, 0x26, 0x75, 0x8d, 0x47, 0x77, 0x6c, 0xe2, 0x9f, 0xcf, 0x95, 0x4c, 0xad, 0x59, 0x4e,
	0x2f, 0xcb, 0x2f, 0xe7, 0x4a, 0xa6, 0xff, 0xa1, 0xe0, 0x1e, 0x91, 0x32, 0xdd, 0xb4, 0xaf, 0x4b,
	0xa6, 0xd3, 0x09, 0x90, 0x8b, 0xba, 0xf8, 0x4d, 0x2c, 0x9c, 0x32, 0x19, 0x8d, 0x8e, 0x52, 0x8f,
	0x9a, 0x2e, 0x08, 0xb2, 0x3b, 0x22, 0x52, 0xc1, 0x13, 0x6f, 0x96, 0xcd, 0xe4, 0x50, 0x36, 0x86,
	0x03, 0x53, 0x51, 0x54, 0x77, 0x89, 0x80, 0x33, 0x47, 0xca, 0x68, 0x00, 0xfc, 0xdf, 0x8c, 0x5e,
	0x42, 0x5c, 0x3a, 0x52, 0x36, 0x23, 0x9b, 0xaa, 0xbe, 0x6b, 0x84, 0x8b, 0x6f, 0x95, 0xd1, 0x20,
	0xb8, 0xc7, 0xb0, 0xcd, 0xd8, 0x6a, 0x1f, 0x2b, 0x02, 0x17, 0xe6, 0xca, 0xc6, 0x1d, 0xa3, 0x98,
	0x06, 0xc4, 0xd7, 0x8a, 0xb7, 0x97, 0xa2, 0x4e, 0xdc, 0x02, 0x8f, 0xbe, 0x1d, 0xc7, 0x1b, 0xc7,
	0x82, 0xd4, 0x78, 0x24, 0xd3, 0x5d, 0x98, 0x7d, 0xa7, 0x6c, 0xba, 0x10, 0xaf, 0xf0, 0x64, 0xb1,
	0x74, 0x93, 0xcd, 0x7f, 0xb8, 0x26, 0xbd, 0x23, 0x7b, 0x3e, 0xe9, 0x32, 0x6e, 0x7c, 0xd7, 0x9f,
	0x31, 0x7f, 0x8f, 0xd2, 0x75, 0x80, 0x4b, 0x02, 0x3e, 0xae, 0xeb, 0x94, 0xc1, 0x5f, 0x9b, 0xce,
	0x7f, 0x91, 0xdb, 0x4b, 0xa3, 0x8e, 0x27, 0xe0, 0x72, 0xd3, 0x31, 0x2d, 0xbc, 0x0b, 0xd9, 0xd4,
	0x5e, 0x15, 0xd8, 0x27, 0xf0, 0xab, 0xeb, 0x0e, 0x7a, 0x0c, 0xec, 0xb8, 0x0b, 0x27, 0xb5, 0xc1,
	0xc8, 0x44, 0x48, 0x3c, 0x33, 0xbd, 0x0b, 0x3f, 0x39, 0xe8, 0x61, 0x70, 0xff, 0xbf, 0xb1, 0xe3,
	0x7f, 0xec, 0xac, 0x0a, 0x8f, 0xde, 0x70, 0xcc, 0x0e, 0x75, 0x03, 0xee, 0x66, 0x1b, 0x0c, 0x67,
	0x97, 0x9c, 0x4a, 0xd9, 0xbe, 0xc2, 0xe1, 0x15, 0x5e, 0xb1, 0xed, 0x99, 0xe3, 0x16, 0x9c, 0x39,
	0x6e, 0x55, 0x6c, 0x7b, 0x75, 0xc5, 0x82, 0xab, 0x2b, 0xe6, 0xeb, 0x6a, 0xcb, 0x82, 0x57, 0x5b,
	0xe6, 0xeb, 0xd2, 0x99, 0x3e, 0x78, 0xe9, 0x4c, 0x5f, 0xc5, 0xb6, 0x8f, 0xcd, 0x16, 0xe0, 0xb1,
	0xd9, 0x42, 0xc5, 0xb6, 0x5b, 0x27, 0xd7, 0xc0, 0xd6, 0xc9, 0x35, 0x15, 0xdb, 0xbe, 0x7c, 0xb8,
	0x1f, 0x5e, 0x3e, 0xdc, 0x6f, 0xa2, 0x34, 0x1d, 0x38, 0xd3, 0x74, 0x2a, 0xb6, 0xbd, 0xd2, 0x74,
	0xe0, 0x4a, 0xfc, 0xd5, 0x6a, 0x3a, 0xb0, 0xd5, 0x74, 0xdc, 0x07, 0x9f, 0xdc, 0x7e, 0xe0, 0x99,
	0x43, 0xcf, 0xed, 0xd9, 0x3b, 0xf2, 0xec, 0xd4, 0xf3, 0xfb, 0xf7, 0x8c, 0xec, 0x6b, 0x1c, 0x1c,
	0x39, 0xf4, 0xd2, 0xce, 0xf8, 0x8f, 0x9d, 0x07, 0xf7, 0x35, 0x0e, 0xee, 0x6c, 0xff, 0xb8, 0xd9,
	0x5b, 0x8e, 0x7f, 0xeb, 0x3c, 0xfe, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x79, 0xc0, 0xda, 0xda,
	0xf9, 0x08, 0x00, 0x00,
}
