package object

import (
	"strconv"
)

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER
}

func (i *Integer) Inspect() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i *Integer) Equals(other Object) bool {
	switch other := other.(type) {
	case *Integer:
		return i.Value == other.Value
	default:
		return false
	}
}

func (i *Integer) Hash() HashKey {
	return HashKey{Type: INTEGER, Value: uint64(i.Value)}
}
