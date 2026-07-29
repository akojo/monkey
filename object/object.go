package object

type ObjectType string

const (
	ARRAY             = "ARRAY"
	BOOLEAN           = "BOOLEAN"
	BUILTIN           = "BUILTIN"
	ERROR             = "ERROR"
	FUNCTION          = "FUNCTION"
	HASH              = "HASH"
	INTEGER           = "INTEGER"
	NULL              = "NULL"
	RETURN            = "RETURN"
	SLICE             = "SLICE"
	STRING            = "STRING"
	COMPILED_FUNCTION = "COMPILED_FUNCTION"
	CLOSURE           = "CLOSURE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}
