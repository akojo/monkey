package object

import (
	"strconv"
)

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }
func (i *Integer) Hash() HashKey {
	return HashKey{Type: INTEGER, Value: uint64(i.Value)}
}
