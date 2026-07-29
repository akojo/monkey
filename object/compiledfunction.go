package object

import (
	"fmt"

	"github.com/akojo/monkey/code"
)

type CompiledFunction struct {
	Instructions code.Instructions
	Locals       int
	Parameters   int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
