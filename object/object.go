package object

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/akojo/monkey/ast"
)

type ObjectType string

const (
	BOOLEAN  = "BOOLEAN"
	BUILTIN  = "BUILTIN"
	ERROR    = "ERROR"
	FUNCTION = "FUNCTION"
	INTEGER  = "INTEGER"
	NULL     = "NULL"
	RETURN   = "RETURN"
	STRING   = "STRING"
)

type Object interface {
	Type() ObjectType
	Inspect() string
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

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }
