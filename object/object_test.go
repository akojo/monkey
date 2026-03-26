package object

import "testing"

func TestBooleanHashKey(t *testing.T) {
	t1 := &Boolean{Value: true}
	t2 := &Boolean{Value: true}
	f := &Boolean{Value: false}

	expectHashEquals(t, t1, t2)
	expectHashNotEquals(t, t1, f)
}

func TestIntegerHashKey(t *testing.T) {
	one1 := &Integer{Value: 1}
	one2 := &Integer{Value: 1}
	hundred := &Integer{Value: 100}

	expectHashEquals(t, one1, one2)
	expectHashNotEquals(t, one1, hundred)
}

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "hello world"}
	hello2 := &String{Value: "hello world"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	expectHashEquals(t, hello1, hello2)
	expectHashEquals(t, diff1, diff2)
	expectHashNotEquals(t, hello1, diff1)
}

func expectHashEquals(t *testing.T, a, b Hashable) {
	if a.Hash() != b.Hash() {
		t.Errorf("%v == %v: got false, expect true", a, b)
	}
}

func expectHashNotEquals(t *testing.T, a, b Hashable) {
	if a.Hash() == b.Hash() {
		t.Errorf("%v == %v: got false, expect true", a, b)
	}
}
