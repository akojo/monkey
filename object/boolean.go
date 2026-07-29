package object

import (
	"strconv"
)

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN }
func (b *Boolean) Inspect() string  { return strconv.FormatBool(b.Value) }
func (b *Boolean) Hash() HashKey {
	if b.Value {
		return HashKey{Type: BOOLEAN, Value: 1}
	}
	return HashKey{Type: BOOLEAN, Value: 0}
}
