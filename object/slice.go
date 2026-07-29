package object

import (
	"fmt"
)

type Slice struct {
	Start int64
	End   *int64
}

func (s *Slice) Type() ObjectType {
	return SLICE
}

func (s *Slice) Inspect() string {
	return fmt.Sprintf("%d:%d", s.Start, *s.End)
}

func (s *Slice) Equals(other Object) bool {
	if s == other {
		return true
	}

	switch other := other.(type) {
	case *Slice:
		if s.Start != other.Start {
			return false
		}
		if s.End == nil && other.End == nil {
			return true
		}
		if s.End == nil || other.End == nil {
			return false
		}
		return *s.End == *other.End
	default:
		return false
	}
}
