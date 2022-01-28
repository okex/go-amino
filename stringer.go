package amino

import "fmt"

type BytesHexStringer []byte

func (b BytesHexStringer) String() string {
	return fmt.Sprintf("%X", []byte(b))
}

type FuncStringer func() string

func (cus FuncStringer) String() string {
	if cus == nil {
		return ""
	}
	return cus()
}
