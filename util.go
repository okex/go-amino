package amino

import (
	"errors"
	"fmt"
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
