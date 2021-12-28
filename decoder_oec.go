package amino

import "math"

func DecodeIntFromUvarint(bz []byte) (i int, n int, err error) {
	var num uint64
	num, n, err = DecodeUvarint(bz)
	if err != nil {
		return
	}
	if num > math.MaxInt {
		err = ErrOverflowInt
		return
	}
	i = int(num)
	return
}
