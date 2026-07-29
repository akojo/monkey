package object

import (
	"fmt"
	"strings"

	"github.com/akojo/monkey/ast"
)

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION }
func (f *Function) Inspect() string {
	var params = make([]string, 0)
	for _, param := range f.Parameters {
		params = append(params, param.String())
	}
	return fmt.Sprintf("fn (%s) %s", strings.Join(params, ", "), f.Body.String())
}
