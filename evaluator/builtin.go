package evaluator

import (
	"unicode/utf8"

	"github.com/akojo/monkey/object"
)

func builtin_len(args ...object.Object) object.Object {
	if err := argCheck(1, args); err != nil {
		return err
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.String:
		length := utf8.RuneCountInString(arg.Value)
		return &object.Integer{Value: int64(length)}
	}
	return newError("argument to `len` not supported, got %s", args[0].Type())
}

func builtin_append(args ...object.Object) object.Object {
	if err := argCheck(2, args); err != nil {
		return err
	}
	if args[0].Type() != object.ARRAY {
		return newError("append: got %s, want ARRAY", args[0].Type())
	}

	array := args[0].(*object.Array)
	length := len(array.Elements)

	newElements := make([]object.Object, length+1)
	copy(newElements, array.Elements)
	newElements[length] = args[1]

	return &object.Array{Elements: newElements}
}

func argCheck(want int, args []object.Object) object.Object {
	if len(args) != want {
		return newError("wrong number of arguments: got %d, want %d", len(args), want)
	}
	return nil
}
