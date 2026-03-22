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

func builtin_equals(args ...object.Object) object.Object {
	if err := argCheck(2, args); err != nil {
		return err
	}
	return toBoolean(equals(args[0], args[1]))
}

func equals(left object.Object, right object.Object) bool {
	if left.Type() != right.Type() {
		return false
	}

	switch left := left.(type) {
	case *object.Boolean:
		return left == right
	case *object.Integer:
		return left.Value == right.(*object.Integer).Value
	case *object.String:
		return left.Value == right.(*object.String).Value
	case *object.Array:
		return arrayEquals(left, right.(*object.Array))
	}
	return false
}

func arrayEquals(left *object.Array, right *object.Array) bool {
	if len(left.Elements) != len(right.Elements) {
		return false
	}

	for i := range left.Elements {
		if !equals(left.Elements[i], right.Elements[i]) {
			return false
		}
	}
	return true
}

func argCheck(want int, args []object.Object) object.Object {
	if len(args) != want {
		return newError("wrong number of arguments: got %d, want %d", len(args), want)
	}
	return nil
}
