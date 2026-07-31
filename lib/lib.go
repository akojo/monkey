package lib

import (
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/akojo/monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

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
	case *object.Number:
		return &object.Number{Value: left.Value + right.(*object.Number).Value}
	case object.String:
		return left.Add(right.(object.String))
	case *object.Array:
		result := slices.Concat(left.Elements, right.(*object.Array).Elements)
		return &object.Array{Elements: result}
	case *object.Hash:
		pairs := maps.Clone(left.Pairs)
		maps.Copy(pairs, right.(*object.Hash).Pairs)
		return &object.Hash{Pairs: pairs}
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
	case *object.Number:
		return &object.Number{Value: left.Value * right.(*object.Number).Value}
	}
	return Error("invalid types: %s * %s", left.Type(), right.Type())
}

func Index(obj object.Object, index object.Object) object.Object {
	switch {
	case obj.Type() == object.ARRAY && index.Type() == object.NUMBER:
		array := obj.(*object.Array)

		f := index.(*object.Number).Value
		if math.Round(f) != f {
			return Error("invalid index")
		}

		i := int(f)
		if i < 0 || i > (len(array.Elements)-1) {
			return NULL
		}
		return array.Elements[i]
	case obj.Type() == object.ARRAY && index.Type() == object.SLICE:
		array := obj.(*object.Array)
		slice := index.(*object.Slice)
		return sliceIndex(array, slice)
	case obj.Type() == object.ARRAY:
		return Error("invalid index type: %s", index.Type())
	case obj.Type() == object.HASH:
		return hashIndex(obj, index)
	}
	return Error("index operator not supported: %s[%s]", obj.Type(), index.Type())
}

func SetIndex(obj object.Object, index object.Object, value object.Object) error {
	switch obj := obj.(type) {
	case *object.Array:
		return arraySet(obj, index, value)
	case *object.Hash:
		key, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("cannot use as hash key: %s", index.Type())
		}

		obj.Pairs[key.Hash()] = object.HashPair{Key: index, Value: value}
		return nil
	default:
		return fmt.Errorf("index operator not supported: %s[%s]", obj.Type(), index.Type())
	}
}

func SliceArray(array *object.Array, start int, end int) object.Object {
	if start < 0 {
		start = 0
	}
	if start > len(array.Elements) {
		start = len(array.Elements)
	}
	if end < 0 {
		end = 0
	}
	if end > len(array.Elements) {
		end = len(array.Elements)
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

func sliceIndex(array *object.Array, sliceObj *object.Slice) object.Object {
	var end int
	if sliceObj.End == nil {
		end = len(array.Elements)
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

func arraySet(array *object.Array, index object.Object, value object.Object) error {
	switch index := index.(type) {
	case *object.Number:
		return arraySetIndex(array, index, value)
	case *object.Slice:
		return arraySetSlice(array, index, value)
	default:
		return fmt.Errorf("index operator not supported: ARRAY[%s]", index.Type())
	}
}

func arraySetIndex(target *object.Array, index *object.Number, value object.Object) error {
	f := index.Value
	if math.Round(f) != f {
		return fmt.Errorf("invalid index")
	}

	i := int(f)
	if i < 0 || i > (len(target.Elements)-1) {
		return fmt.Errorf("index out of range")
	}

	target.Elements[i] = value
	return nil
}

func arraySetSlice(target *object.Array, index *object.Slice, value object.Object) error {
	array, ok := value.(*object.Array)
	if !ok {
		return fmt.Errorf("invalid value %s, must be ARRAY", value.Type())
	}

	start := index.Start
	var end int
	if index.End != nil {
		end = *index.End
	} else {
		end = start + len(array.Elements)
	}

	if len(array.Elements) != (end - start) {
		return fmt.Errorf("invalid value length %d, must be %d", len(array.Elements), end-start)
	}

	if end > len(target.Elements)-1 {
		target.Elements = slices.Concat(target.Elements[:start], array.Elements)
	} else {
		copy(target.Elements[start:end], array.Elements)
	}
	return nil
}
