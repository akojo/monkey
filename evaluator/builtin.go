package evaluator

import (
	"unicode/utf8"

	"github.com/akojo/monkey/object"
)

func builtin_len(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments: got %d, want 1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		length := utf8.RuneCountInString(arg.Value)
		return &object.Integer{Value: int64(length)}
	}
	return newError("argument to `len` not supported, got %s", args[0].Type())
}
