package object

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/akojo/monkey/ast"
)

type ObjectType string

const (
	ARRAY    = "ARRAY"
	BOOLEAN  = "BOOLEAN"
	BUILTIN  = "BUILTIN"
	ERROR    = "ERROR"
	FUNCTION = "FUNCTION"
	INTEGER  = "INTEGER"
	NULL     = "NULL"
	RETURN   = "RETURN"
	SLICE    = "SLICE"
	STRING   = "STRING"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY }
func (a *Array) Inspect() string {
	elements := make([]string, len(a.Elements))
	for i, elem := range a.Elements {
		elements[i] = elem.Inspect()
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN }
func (b *Boolean) Inspect() string  { return strconv.FormatBool(b.Value) }

type BuiltinFunction func(args ...Object) Object
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN }
func (b *Builtin) Inspect() string  { return "<builtin>" }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }

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
	return fmt.Sprintf("fn (%s) %s", strings.Join(params, ", "), f.Body.PrettyPrint(1))
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL }
func (n *Null) Inspect() string  { return "null" }

type Return struct {
	Value Object
}

func (r *Return) Type() ObjectType { return RETURN }
func (r *Return) Inspect() string  { return r.Value.Inspect() }

type Slice struct {
	Start int64
	End   *int64
}

func (s *Slice) Type() ObjectType { return SLICE }
func (s *Slice) Inspect() string  { return fmt.Sprintf("%d:%d", s.Start, *s.End) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }
