package object

import (
	"fmt"
	"strconv"
)

type ObjectType string

const (
	BOOLEAN = "BOOLEAN"
	ERROR   = "ERROR"
	INTEGER = "INTEGER"
	NULL    = "NULL"
	RETURN  = "RETURN"
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

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }

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
