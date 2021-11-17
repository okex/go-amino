package amino

import (
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
