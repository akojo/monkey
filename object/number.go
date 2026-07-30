package object

import (
	"math"
	"strconv"
)

type Number struct {
	Value float64
}

func (i *Number) Type() ObjectType {
	return NUMBER
}

func (i *Number) Inspect() string {
	return strconv.FormatFloat(i.Value, 'g', -1, 64)
}

func (i *Number) Equals(other Object) bool {
	switch other := other.(type) {
	case *Number:
		return i.Value == other.Value
	default:
		return false
	}
}

func (i *Number) Hash() HashKey {
	return HashKey{Type: NUMBER, Value: math.Float64bits(i.Value)}
}
