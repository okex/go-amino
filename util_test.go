package amino

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

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

func TestMarshalBigIntToText(t *testing.T) {
	testCases := []*big.Int{
		(*big.Int)(nil),
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(-1),
		big.NewInt(-100000),
		big.NewInt(99999999),
		new(big.Int).Add(new(big.Int).Mul(big.NewInt(-999999999999999), big.NewInt(100000000000)), big.NewInt(-99999999999)),
		big.NewInt(math.MaxInt64),
		big.NewInt(math.MinInt64),
		big.NewInt(math.MaxInt64).Add(big.NewInt(math.MaxInt64), big.NewInt(1)),
		big.NewInt(math.MaxInt64).Mul(big.NewInt(math.MaxInt64), big.NewInt(math.MinInt64)),
		big.NewInt(math.MaxInt64).Mul(big.NewInt(math.MaxInt64), big.NewInt(math.MaxInt64)),
		big.NewInt(math.MinInt64).Sub(big.NewInt(math.MinInt64), big.NewInt(1)),
	}
	for _, i := range testCases {
		bi := i
		str, err := bi.MarshalText()
		require.NoError(t, err)

		str2, err := MarshalBigIntToText(bi)
		require.NoError(t, err)

		require.Equal(t, BytesToStr(str), str2)

		require.Equal(t, len(str), CalcBigIntTextSize(bi))
	}
}

func BenchmarkMarshalBigIntToText(b *testing.B) {
	bi := big.NewInt(math.MaxInt64)
	b.Run("big", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = bi.MarshalText()
		}
	})
	b.Run("opt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = MarshalBigIntToText(bi)
		}
	})
	bi.Mul(bi, big.NewInt(2))
	bi.Add(bi, big.NewInt(1))
	bi.Add(bi, big.NewInt(1))
	b.Run("big2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = bi.MarshalText()
		}
	})
	b.Run("opt2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = MarshalBigIntToText(bi)
		}
	})
}

func BenchmarkHexEncodeToString(b *testing.B) {
	var buf = make([]byte, 512)
	rand.Read(buf)
	b.ResetTimer()

	b.Run("hex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = hex.EncodeToString(buf)
		}
	})

	b.Run("amino hex", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = HexEncodeToString(buf)
		}
	})

	b.Run("string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = string(buf)
		}
	})
}

func TestEncodedTimeSize(t *testing.T) {
	testCases := []time.Time{
		time.Now(),
		time.Unix(0, 0),
	}

	buf := &bytes.Buffer{}

	for _, ti := range testCases {
		err := EncodeTimeToBuffer(buf, ti)
		require.NoError(t, err)

		require.Equal(t, buf.Len(), TimeSize(ti))

		buf.Reset()
	}
}

func BenchmarkBigIntNums(b *testing.B) {
	smallBigInt := big.NewInt(math.MaxInt64)
	b.Run("small", func(b *testing.B) {
		d := smallBigInt
		b.Run("marshal", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = d.MarshalText()
			}
		})
		b.Run("calc", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				CalcBigIntTextSize(d)
			}
		})
		expect, _ := d.MarshalText()
		require.Equal(b, len(expect), CalcBigIntTextSize(d))
	})
	mediumBigInt := big.NewInt(1).Mul(new(big.Int).SetUint64(10000000000000000000), big.NewInt(50000))
	b.Run("medium", func(b *testing.B) {
		d := mediumBigInt
		b.Run("marshal", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = d.MarshalText()
			}

		})
		b.Run("calc", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				CalcBigIntTextSize(d)
			}
		})
		expect, _ := d.MarshalText()
		require.Equal(b, len(expect), CalcBigIntTextSize(d))
	})
	bigBigInt := big.NewInt(1).Mul(new(big.Int).SetUint64(math.MaxUint64), new(big.Int).SetUint64(math.MaxUint64))
	bigBigInt = bigBigInt.Mul(bigBigInt, bigBigInt)
	b.Run("big", func(b *testing.B) {
		d := bigBigInt
		b.Run("marshal", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = d.MarshalText()
			}

		})
		b.Run("calc", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				CalcBigIntTextSize(d)
			}
		})
		expect, _ := d.MarshalText()
		require.Equal(b, len(expect), CalcBigIntTextSize(d))
	})
}

func TestHexEncodeToStringUpper(t *testing.T) {
	var bz = make([]byte, 256)
	for i := 0; i < 10; i++ {
		rand.Read(bz)
		require.Equal(t, strings.ToUpper(hex.EncodeToString(bz)), HexEncodeToStringUpper(bz))
	}
}
