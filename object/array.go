package object

import (
	"fmt"
	"strings"
)

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType {
	return ARRAY
}

func (a *Array) Inspect() string {
	elements := make([]string, len(a.Elements))
	for i, elem := range a.Elements {
		elements[i] = elem.Inspect()
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (a *Array) Equals(other Object) bool {
	if a == other {
		return true
	}

	switch other := other.(type) {
	case *Array:
		if len(a.Elements) != len(other.Elements) {
			return false
		}

		for i := range a.Elements {
			if !a.Elements[i].Equals(other.Elements[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}
