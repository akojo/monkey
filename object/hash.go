package object

import (
	"bytes"
	"fmt"
	"strings"
)

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	Hash() HashKey
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH }

func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := make([]string, 0)
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (h *Hash) Equals(other Object) bool {
	if h == other {
		return true
	}

	switch other := other.(type) {
	case *Hash:
		if len(h.Pairs) != len(other.Pairs) {
			return false
		}

		for key, pair := range h.Pairs {
			otherpair, found := other.Pairs[key]
			if !found {
				return false
			}
			if !pair.Value.Equals(otherpair.Value) {
				return false
			}
		}

		return true
	default:
		return false
	}
}
