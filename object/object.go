package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
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
	HASH     = "HASH"
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

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	Hash() HashKey
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
func (b *Boolean) Hash() HashKey {
	if b.Value {
		return HashKey{Type: BOOLEAN, Value: 1}
	}
	return HashKey{Type: BOOLEAN, Value: 0}
}

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
	return fmt.Sprintf("fn (%s) %s", strings.Join(params, ", "), f.Body.String())
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

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) Inspect() string  { return strconv.FormatInt(i.Value, 10) }
func (i *Integer) Hash() HashKey {
	return HashKey{Type: INTEGER, Value: uint64(i.Value)}
}

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
func (s *String) Hash() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: STRING, Value: h.Sum64()}
}
