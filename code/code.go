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
	OpAdd
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"PUSHI", []int{2}}, // load immediate
	OpAdd:      {"ADD", []int{}},
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

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.LittleEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
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
		return fmt.Sprintf("%s %d", def.Name, operands[0]), read
	}

	return fmt.Sprintf("ERROR: %s: unhandled operand count %d", def.Name, operandCount), read
}

func ReadOperands(def *Definition, instructions Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(instructions[offset:]))
		}
		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.LittleEndian.Uint16(ins)
}
