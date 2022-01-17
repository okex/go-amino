package amino

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProtoPosAndTypeMustOneByte(t *testing.T) {
	pos, pbType, err := ParseProtoPosAndTypeMustOneByte(0x0a)
	if pos != 1 {
		t.Fatal("parse pos error")
	}
	if pbType != 2 {
		t.Fatal("parse type error")
	}

	data, err := EncodeProtoPosAndTypeMustOneByte(pos, pbType)
	if err != nil {
		t.Fatal("err shoul be nil")
	}
	if data != 0x0a {
		t.Fatal("encode error")
	}

	pos, pbType, err = ParseProtoPosAndTypeMustOneByte(0x82)
	require.Error(t, err)
}

func TestUnmarshalBigIntBase10(t *testing.T) {
	testCases := []int64{
		0,
		1,
		-1,
		math.MaxInt64,
		math.MinInt64,
	}
	for _, i := range testCases {
		bi := big.NewInt(i)
		str, err := bi.MarshalText()
		require.NoError(t, err)

		var bi2 big.Int
		err = bi2.UnmarshalText(str)
		require.NoError(t, err)

		bi3, err := UnmarshalBigIntBase10(str)
		require.NoError(t, err)
		require.Equal(t, &bi2, bi3)
	}

	{
		// bigger than int64
		bi := big.NewInt(math.MaxInt64).Add(big.NewInt(math.MaxInt64), big.NewInt(1))
		str, err := bi.MarshalText()
		require.NoError(t, err)

		var bi2 big.Int
		err = bi2.UnmarshalText(str)
		require.NoError(t, err)

		bi3, err := UnmarshalBigIntBase10(str)
		require.NoError(t, err)
		require.Equal(t, &bi2, bi3)
	}

	{
		// expect error
		bi := big.NewInt(12345)
		str, err := bi.MarshalText()
		require.NoError(t, err)

		str = StrToBytes(BytesToStr(str) + "a")

		var bi2 big.Int
		err = bi2.UnmarshalText(str)
		require.Error(t, err)

		_, err = UnmarshalBigIntBase10(str)
		require.Error(t, err)
	}
}
