package amino

import "errors"

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
