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

// DecodeIntUpdateBytes will decode a int from the input bytes,
// and update the input bytes to be the remaining bytes after the int.
func DecodeIntUpdateBytes(bz *[]byte) (i int, err error) {
	var n int
	i, n, err = DecodeInt(*bz)
	if err != nil {
		return
	}
	*bz = (*bz)[n:]
	return
}

// DecodeUvarintUpdateBytes will decode a uvarint from the input bytes,
// and update the input bytes to be the remaining bytes after the uvarint.
func DecodeUvarintUpdateBytes(bz *[]byte) (i uint64, err error) {
	var n int
	i, n, err = DecodeUvarint(*bz)
	if err != nil {
		return
	}
	*bz = (*bz)[n:]
	return
}

// UpdateByteSlice will try copy src to dst with amino rules.
func UpdateByteSlice(dst *[]byte, src []byte) {
	if len(src) == 0 {
		*dst = nil
	} else {
		var newBz []byte
		dstBz := *dst
		if cap(dstBz) >= len(src) {
			newBz = dstBz[:len(src)]
		} else {
			newBz = make([]byte, len(src))
		}
		copy(newBz, src)
		*dst = newBz
	}
}

// DecodeByteSliceWithoutCopy will decode a byte slice from the input bytes,
// and update the input bytes to be the remaining bytes after the byte slice.
// the decoded byte slice just a reference to the input bytes.
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
