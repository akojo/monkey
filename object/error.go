package object

import (
	"fmt"
)

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR
}

func (e *Error) Inspect() string {
	return fmt.Sprintf("ERROR: %s", e.Message)
}

func (e *Error) Equals(other Object) bool {
	if e == other {
		return true
	}

	switch other := other.(type) {
	case *Error:
		return e.Message == other.Message
	default:
		return false
	}
}
