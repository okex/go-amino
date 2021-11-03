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

	pos, pbType, err = ParseProtoPosAndTypeMustOneByte(0x82)
	require.Error(t, err)
}
