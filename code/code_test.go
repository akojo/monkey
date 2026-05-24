package code

import (
	"slices"
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 254, 255}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		if len(instruction) != len(test.expected) {
			t.Errorf("len(instruction): expected %d, got %d", len(instruction), len(test.expected))
		}

		for i, b := range test.expected {
			if instruction[i] != b {
				t.Errorf("instruction[i]: expected %d, got %d", b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := Instructions(slices.Concat(Make(OpConstant, 1), Make(OpConstant, 2), Make(OpAdd)))

	expected := `0000 PUSHI 1
0003 PUSHI 2
0006 ADD
`

	if expected != instructions.String() {
		t.Errorf("instructions:\nwant %q\ngot  %q", expected, instructions.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		def, err := Lookup(byte(test.op))
		if err != nil {
			t.Fatalf("definition not found: %q", err)
		}

		operands, bytesRead := ReadOperands(def, instruction[1:])
		if test.bytesRead != bytesRead {
			t.Fatalf("bytesRead: want %d, got %d", test.bytesRead, bytesRead)
		}

		for i, want := range test.operands {
			if want != operands[i] {
				t.Errorf("operand %d: want %d, got %d", i, want, operands[i])
			}
		}
	}
}
