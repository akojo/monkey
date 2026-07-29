package object

import (
	"hash/fnv"
	"unicode/utf8"
	"unique"
)

type String unique.Handle[string]

func NewString(value string) String {
	return String(unique.Make(value))
}

func (s String) Type() ObjectType {
	return STRING
}

func (s String) Inspect() string {
	return s.value()
}

func (s String) Equals(other Object) bool {
	return s == other
}

func (s String) Len() *Integer {
	return &Integer{Value: int64(utf8.RuneCountInString(s.value()))}
}

func (s String) Add(other String) String {
	return NewString(s.value() + other.value())
}

func (s String) Hash() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.value()))
	return HashKey{Type: STRING, Value: h.Sum64()}
}

func (s String) value() string {
	return unique.Handle[string](s).Value()
}
