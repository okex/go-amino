package amino

import (
	"fmt"
	"math"
)

func DecodeInt(bz []byte) (i int, n int, err error) {
	var num uint64
	num, n, err = DecodeUvarint(bz)
	if err != nil {
		return
	}
	if int64(num) > math.MaxInt || int64(num) < math.MinInt {
		err = ErrOverflowInt
		return
	}
	i = int(num)
	return
}

func DecodeIntUpdateBytes(bz *[]byte) (i int, err error) {
	var n int
	i, n, err = DecodeInt(*bz)
	if err != nil {
		return
	}
	*bz = (*bz)[n:]
	return
}

// Deprecated: use DecodeInt
func DecodeIntFromUvarint(bz []byte) (i int, n int, err error) {
	return DecodeInt(bz)
}

func DecodeByteSliceWithoutCopy(source *[]byte) ([]byte, error) {
	bz := *source
	count, _n, err := DecodeUvarint(bz)
	if err != nil {
		return nil, err
	}
	if int(count) < 0 {
		err = fmt.Errorf("invalid negative length %v decoding []byte", count)
		return nil, err
	}
	bz = bz[_n:]
	if len(bz) < int(count) {
		err = fmt.Errorf("insufficient bytes decoding []byte of length %v", count)
		return nil, err
	}
	ret := bz[:int(count)]
	*source = bz[int(count):]
	return ret, nil
}
