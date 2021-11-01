package amino

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProtoPosAndTypeMustOneByte(t *testing.T) {
	pos, pbType := ParseProtoPosAndTypeMustOneByte(0x0a)
	if pos != 1 {
		t.Fatal("parse pos error")
	}
	if pbType != 2 {
		t.Fatal("parse type error")
	}

	require.Panics(t, func() {
		ParseProtoPosAndTypeMustOneByte(0x82)
	})
}
