package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPushImmediate
	OpPop
	OpDup

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal

	OpNull

	OpFalse
	OpTrue

	OpArray
	OpHash
	OpIndex
	OpSlice

	OpEqual
	OpNotEqual
	OpLessThan

	OpBranch
	OpBranchNotEqual

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpMinus
	OpBang

	OpCall
	OpReturn

	OpGetBuiltin

	OpClosure
	OpGetFree
	OpCurrentClosure
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"PUSH", []int{2}}, // push a constant
	// Optimization: push an immediate integer that fits in a half-word.
	// Since OpConstant above uses a 2-byte operand it takes the same space in
	// bytecode to encode small integer constants straight into the code itself
	// (and it saves a slot in the constants store).
	OpPushImmediate: {"PUSHI", []int{2}},
	OpPop:           {"POP", []int{}},
	OpDup:           {"DUP", []int{}}, // Duplicate top of stack

	OpGetGlobal: {"GETGLOBAL", []int{2}},
	OpSetGlobal: {"SETGLOBAL", []int{2}},
	OpGetLocal:  {"GETLOCAL", []int{1}},
	OpSetLocal:  {"SETLOCAL", []int{1}},

	OpNull: {"NULL", []int{}},

	OpFalse: {"FALSE", []int{}},
	OpTrue:  {"TRUE", []int{}},

	OpArray: {"ARRAY", []int{2}},
	OpHash:  {"HASH", []int{2}},
	OpIndex: {"INDEX", []int{}},
	OpSlice: {"SLICE", []int{}},

	OpEqual:    {"EQ", []int{}},
	OpNotEqual: {"NEQ", []int{}},
	OpLessThan: {"LT", []int{}},

	OpBranch:         {"B", []int{2}},
	OpBranchNotEqual: {"BNE", []int{2}},

	OpAdd: {"ADD", []int{}},
	OpSub: {"SUB", []int{}},
	OpMul: {"MUL", []int{}},
	OpDiv: {"DIV", []int{}},

	OpMinus: {"NEG", []int{}},
	OpBang:  {"NOT", []int{}},

	OpCall:   {"CALL", []int{1}},
	OpReturn: {"RET", []int{}},

	OpGetBuiltin: {"BUILTIN", []int{1}},

	OpClosure:        {"CLOSURE", []int{2, 1}},
	OpGetFree:        {"GETFREE", []int{1}},
	OpCurrentClosure: {"CURRCLOSURE", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("invalid opcode: %d", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instruction := []byte{byte(op)}
	for i, operand := range operands {
		switch def.OperandWidths[i] {
		case 1:
			instruction = append(instruction, byte(operand))
		case 2:
			instruction = binary.NativeEndian.AppendUint16(instruction, uint16(operand))
		}
	}

	return instruction
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	for i := 0; i < len(ins); {
		result, read := ins[i:].fmt()
		fmt.Fprintf(&out, "%04d %s\n", i, result)

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmt() (string, int) {
	def, err := Lookup(ins[0])
	if err != nil {
		return fmt.Sprintf("ERROR: %s\n", err), 0
	}

	operands, read := ReadOperands(def, ins[1:])

	operandCount := len(def.OperandWidths)
	if operandCount != len(operands) {
		return fmt.Sprintf("ERROR: operand len %d, defined %d", len(operands), operandCount), read
	}

	switch operandCount {
	case 0:
		return def.Name, 0
	case 1:
		return fmt.Sprintf("%s %04x", def.Name, operands[0]), read
	case 2:
		return fmt.Sprintf("%s %04x %04x", def.Name, operands[0], operands[1]), read
	}

	return fmt.Sprintf("ERROR: %s: unhandled operand count %d", def.Name, operandCount), read
}

func ReadOperands(def *Definition, instructions Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(ReadUint8(instructions[offset:]))
		case 2:
			operands[i] = int(ReadUint16(instructions[offset:]))
		}
		offset += width
	}

	return operands, offset
}

func ReadUint8(ins Instructions) uint8 {
	return ins[0]
}

func ReadUint16(ins Instructions) uint16 {
	return binary.NativeEndian.Uint16(ins)
}
