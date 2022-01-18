package amino

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
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

func GetBinaryBareFromBinaryLengthPrefixed(bz []byte) ([]byte, error) {
	if len(bz) == 0 {
		return nil, errors.New("cannot be empty bytes")
	}

	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return nil, fmt.Errorf("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return nil, fmt.Errorf("Not enough bytes to read, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return nil, fmt.Errorf("Bytes left over, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	return bz[n:], nil
}

func UnmarshalBigIntBase10(bz []byte) (*big.Int, error) {
	ret := new(big.Int)
	if len(bz) < 19 {
		i, err := strconv.ParseInt(BytesToStr(bz), 10, 0)
		if err == nil {
			ret.SetInt64(i)
			return ret, nil
		}
	}

	err := ret.UnmarshalText(bz)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func HexEncodeToString(src []byte) string {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return BytesToStr(dst)
}
