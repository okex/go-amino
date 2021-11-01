package amino

// ParseProtoPosAndTypeMustOneByte Parse feild number and type from one byte,
// if original feild number and type encode to multiple bytes, you should not use this function.
func ParseProtoPosAndTypeMustOneByte(data byte) (pos int, aminoType Typ3) {
	if data&0x80 == 0x80 {
		panic("varint more than one byte")
	}
	aminoType = Typ3(data & 0x07)
	pos = int(data) >> 3
	return
}
