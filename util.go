package amino

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// ParseProtoPosAndTypeMustOneByte Parse field number and type from one byte,
// if original field number and type encode to multiple bytes, you should not use this function.
func ParseProtoPosAndTypeMustOneByte(data byte) (pos int, pb3Type Typ3, err error) {
	if data&0x80 == 0x80 {
		err = errors.New("func ParseProtoPosAndTypeMustOneBytevarint can't parse more than one byte")
		return
	}
	pb3Type = Typ3(data & 0x07)
	pos = int(data) >> 3
	return
}

func EncodeProtoPosAndTypeMustOneByte(pos int, typ Typ3) (byte, error) {
	// 1 1111 111
	if pos > 15 {
		return 0, fmt.Errorf("pos must be less than 16")
	}
	if typ > 7 {
		return 0, fmt.Errorf("typ must be less than 8")
	}
	data := byte(pos)
	data <<= 3
	data |= byte(typ)
	return data, nil
}

// StrToBytes is meant to make a zero allocation conversion
// from string -> []byte to speed up operations, it is not meant
// to be used generally
func StrToBytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Cap = stringHeader.Len
	hdr.Len = stringHeader.Len
	hdr.Data = stringHeader.Data
	return b
}

// BytesToStr is meant to make a zero allocation conversion
// from []byte -> string to speed up operations, it is not meant
// to be used generally, but for a specific pattern to delete keys
// from a map.
func BytesToStr(b []byte) string {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&b))
	return *(*string)(unsafe.Pointer(hdr))
}
