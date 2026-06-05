package builtin

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/akojo/monkey/lib"
	"github.com/akojo/monkey/object"
)

func Len(args ...object.Object) object.Object {
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
	return lib.Error("argument to `len` not supported, got %s", args[0].Type())
}

func Append(args ...object.Object) object.Object {
	if err := argCheck(2, args); err != nil {
		return err
	}
	if args[0].Type() != object.ARRAY {
		return lib.Error("append: got %s, want ARRAY", args[0].Type())
	}

	array := args[0].(*object.Array)
	length := len(array.Elements)

	newElements := lib.Realloc(array.Elements, length+1)
	newElements[length] = args[1]

	return &object.Array{Elements: newElements}
}

func Equals(args ...object.Object) object.Object {
	if err := argCheck(2, args); err != nil {
		return err
	}
	return lib.Boolean(lib.Equals(args[0], args[1]))
}

func Slice(args ...object.Object) object.Object {
	if err := argCheck(3, args); err != nil {
		return err
	}
	array, ok := args[0].(*object.Array)
	if !ok {
		lib.Error("argument 0: got %s, want ARRAY", args[0].Type())
	}
	start, ok := args[1].(*object.Integer)
	if !ok {
		lib.Error("argument 1: got %s, want INTEGER", args[0].Type())
	}
	end, ok := args[2].(*object.Integer)
	if !ok {
		lib.Error("argument 2: got %s, want INTEGER", args[0].Type())
	}
	return lib.SliceArray(array, start.Value, end.Value)
}

func Print(args ...object.Object) object.Object {
	result := make([]string, 0)

	for _, arg := range args {
		result = append(result, arg.Inspect())
	}
	fmt.Print(strings.Join(result, " "))
	fmt.Print("\n")

	return lib.NULL
}

func argCheck(want int, args []object.Object) object.Object {
	if len(args) != want {
		return lib.Error("wrong number of arguments: got %d, want %d", len(args), want)
	}
	return nil
}
