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

func builtin_slice(args ...object.Object) object.Object {
	if err := argCheck(3, args); err != nil {
		return err
	}
	array, ok := args[0].(*object.Array)
	if !ok {
		newError("argument 0: got %s, want ARRAY", args[0].Type())
	}
	start, ok := args[1].(*object.Integer)
	if !ok {
		newError("argument 1: got %s, want INTEGER", args[0].Type())
	}
	end, ok := args[2].(*object.Integer)
	if !ok {
		newError("argument 2: got %s, want INTEGER", args[0].Type())
	}
	return slice(array, start.Value, end.Value)
}

func equals(left, right object.Object) bool {
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

func arrayEquals(left, right *object.Array) bool {
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

func add(left, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s + %s", left.Type(), right.Type())
	}
	switch left := left.(type) {
	case *object.Boolean:
		if left == TRUE || right == TRUE {
			return TRUE
		}
		return FALSE
	case *object.Integer:
		return &object.Integer{Value: left.Value + right.(*object.Integer).Value}
	case *object.String:
		return &object.String{Value: left.Value + right.(*object.String).Value}
	case *object.Array:
		l := left.Elements
		r := right.(*object.Array).Elements

		result := &object.Array{Elements: make([]object.Object, len(l)+len(r))}
		copy(result.Elements, l)
		copy(result.Elements[len(l):], r)
		return result
	}
	return newError("invalid types: %s + %s", left.Type(), right.Type())
}

func multiply(left, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s * %s", left.Type(), right.Type())
	}
	switch left := left.(type) {
	case *object.Boolean:
		if left == FALSE || right == FALSE {
			return FALSE
		}
		return TRUE
	case *object.Integer:
		return &object.Integer{Value: left.Value * right.(*object.Integer).Value}
	}
	return newError("invalid types: %s * %s", left.Type(), right.Type())
}

func slice(array *object.Array, start int64, end int64) object.Object {
	if start < 0 {
		start = 0
	}
	if start > int64(len(array.Elements)) {
		start = int64(len(array.Elements))
	}
	if end < 0 {
		end = 0
	}
	if end > int64(len(array.Elements)) {
		end = int64(len(array.Elements))
	}
	if !(start < end) {
		return &object.Array{Elements: make([]object.Object, 0)}
	}

	newSlice := &object.Array{Elements: make([]object.Object, end-start)}

	copy(newSlice.Elements, array.Elements[start:end])

	return newSlice
}

func argCheck(want int, args []object.Object) object.Object {
	if len(args) != want {
		return newError("wrong number of arguments: got %d, want %d", len(args), want)
	}
	return nil
}
