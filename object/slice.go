package object

import (
	"fmt"
)

type Slice struct {
	Start int64
	End   *int64
}

func (s *Slice) Type() ObjectType { return SLICE }
func (s *Slice) Inspect() string  { return fmt.Sprintf("%d:%d", s.Start, *s.End) }
