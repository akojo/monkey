package lib

import (
	"fmt"

	"github.com/akojo/monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Equals(left, right object.Object) bool {
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
		if !Equals(left.Elements[i], right.Elements[i]) {
			return false
		}
	}
	return true
}

func Add(left, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return Error("type mismatch: %s + %s", left.Type(), right.Type())
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

		result := Realloc(l, len(l)+len(r))
		copy(result[len(l):], r)
		return &object.Array{Elements: result}
	}
	return Error("invalid types: %s + %s", left.Type(), right.Type())
}

func Multiply(left, right object.Object) object.Object {
	if left.Type() != right.Type() {
		return Error("type mismatch: %s * %s", left.Type(), right.Type())
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
	return Error("invalid types: %s * %s", left.Type(), right.Type())
}

func Index(obj object.Object, index object.Object) object.Object {
	switch {
	case obj.Type() == object.ARRAY && index.Type() == object.INTEGER:
		array := obj.(*object.Array)
		i := index.(*object.Integer).Value
		if i < 0 || i > int64(len(array.Elements)-1) {
			return NULL
		}
		return array.Elements[i]
	case obj.Type() == object.ARRAY && index.Type() == object.SLICE:
		array := obj.(*object.Array)
		slice := index.(*object.Slice)
		return sliceIndex(array, slice)
	case obj.Type() == object.HASH:
		return hashIndex(obj, index)
	}
	return Error("index operator not supported: %s", obj.Type())
}

func sliceIndex(array *object.Array, sliceObj *object.Slice) object.Object {
	var end int64
	if sliceObj.End == nil {
		end = int64(len(array.Elements))
	} else {
		end = *sliceObj.End
	}

	return SliceArray(array, sliceObj.Start, end)
}

func hashIndex(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return Error("cannot use as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.Hash()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func SliceArray(array *object.Array, start int64, end int64) object.Object {
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

func IsTruthy(obj object.Object) bool {
	if obj == FALSE || obj == NULL {
		return false
	}
	return true
}

func Boolean(value bool) object.Object {
	if value {
		return TRUE
	}
	return FALSE
}

func Error(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func Realloc[T any](s []T, size int) []T {
	r := make([]T, size)
	copy(r, s)
	return r
}
