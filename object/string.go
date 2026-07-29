package object

import (
	"hash/fnv"
)

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING
}

func (s *String) Inspect() string {
	return s.Value
}

func (s *String) Equals(other Object) bool {
	if s == other {
		return true
	}

	switch other := other.(type) {
	case *String:
		return s.Value == other.Value
	default:
		return false
	}
}

func (s *String) Hash() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: STRING, Value: h.Sum64()}
}
