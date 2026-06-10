package testutil

import (
	"fmt"

	"github.com/akojo/monkey/lib"
	"github.com/akojo/monkey/object"
)

func ExpectObject(actual object.Object, expected any) error {
	switch e := expected.(type) {
	case int:
		return expectInteger(actual, int64(e))
	case int64:
		return expectInteger(actual, e)
	case bool:
		return expectBoolean(actual, e)
	case string:
		return expectString(actual, e)
	case []int:
		return expectIntegerArray(actual, e)
	case map[object.HashKey]int64:
		return expectIntegerHash(actual, e)
	case error:
		return expectError(actual, e.Error())
	case nil:
		return expectNull(actual)
	default:
		panic(fmt.Sprintf("invalid type %T", e))
	}
}

func expectInteger(actual object.Object, expected int64) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("expected Integer, got %T: %s", actual, actual.Inspect())
	}
	if result.Value != expected {
		return fmt.Errorf("integer: expected %d, got %d", expected, result.Value)
	}
	return nil
}

func expectBoolean(actual object.Object, expected bool) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("expected Boolean, got %q", actual.Inspect())
	}
	if result.Value != expected {
		return fmt.Errorf("boolean: expected %t, got %t", expected, result.Value)
	}
	return nil
}

func expectString(actual object.Object, expected string) error {
	str, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("expected String, got %q", actual.Type())
	}
	if str.Value != expected {
		return fmt.Errorf("string: expected %q, got %q", expected, str.Value)
	}
	return nil
}

func expectError(actual object.Object, expected string) error {
	err, ok := actual.(*object.Error)
	if !ok {
		return fmt.Errorf("expected Error: got %q", actual.Inspect())
	}
	if err.Message != expected {
		return fmt.Errorf("error: expected %q, got %q", expected, err.Message)
	}
	return nil
}

func expectIntegerArray(actual object.Object, expected []int) error {
	array, ok := actual.(*object.Array)
	if !ok {
		return fmt.Errorf("want Array, got %T (%+v)", actual, actual)
	}

	if len(array.Elements) != len(expected) {
		return fmt.Errorf("len(array): want %d, got %d", len(expected), len(array.Elements))
	}

	for i, expectedElement := range expected {
		err := expectInteger(array.Elements[i], int64(expectedElement))
		if err != nil {
			return fmt.Errorf("array[%d]: %w", i, err)
		}
	}
	return nil
}

func expectIntegerHash(actual object.Object, expected map[object.HashKey]int64) error {
	hash, err := assertType[*object.Hash](actual)
	if err != nil {
		return err
	}

	if len(hash.Pairs) != len(expected) {
		return fmt.Errorf("len(hash): want %d, got %d", len(expected), len(hash.Pairs))
	}

	for expectKey, expectValue := range expected {
		pair, ok := hash.Pairs[expectKey]
		if !ok {
			return fmt.Errorf("missing key in Pairs")
		}

		err := expectInteger(pair.Value, expectValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func expectNull(actual object.Object) error {
	if actual != lib.NULL {
		return fmt.Errorf("expected NULL, got %q", actual.Inspect())
	}
	return nil
}

func assertType[T object.Object](obj object.Object) (T, error) {
	result, ok := obj.(T)
	if !ok {
		return *new(T), fmt.Errorf("want %T, got %T (%+v)", *new(T), obj, obj)
	}
	return result, nil
}
