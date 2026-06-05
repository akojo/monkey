package vm

import (
	"errors"
	"fmt"
	"slices"

	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/compiler"
	"github.com/akojo/monkey/lib"
	"github.com/akojo/monkey/object"
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, 1),
		sp:    0,

		globals: make([]object.Object, bytecode.GlobalsSize),
	}
}

func NewWithGlobals(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) StackAboveTop() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			idx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.push(vm.constants[idx])
		case code.OpPop:
			vm.sp--
		case code.OpGetGlobal:
			index := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.push(vm.globals[index])
		case code.OpSetGlobal:
			index := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[index] = vm.pop()
		case code.OpNull:
			vm.push(lib.NULL)
		case code.OpFalse:
			vm.push(lib.FALSE)
		case code.OpTrue:
			vm.push(lib.TRUE)
		case code.OpBranch:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpBranchNotEqual:
			if !lib.IsTruthy(vm.pop()) {
				pos := int(code.ReadUint16(vm.instructions[ip+1:]))
				ip = pos - 1
			} else {
				ip += 2
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv,
			code.OpEqual, code.OpNotEqual, code.OpLessThan:

			left, right := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			result := executeBinaryOp(op, left, right)
			if result.Type() == object.ERROR {
				return errors.New(result.(*object.Error).Message)
			}

			vm.sp--
			vm.stack[vm.sp-1] = result
		case code.OpMinus:
			if vm.stack[vm.sp-1].Type() != object.INTEGER {
				return fmt.Errorf("unknown operator: -%s", vm.stack[vm.sp-1].Type())
			}
			top := vm.stack[vm.sp-1].(*object.Integer)
			top.Value = -top.Value
		case code.OpBang:
			top := vm.stack[vm.sp-1]
			vm.stack[vm.sp-1] = lib.Boolean(top == lib.FALSE || top == lib.NULL)
		default:
			return fmt.Errorf("unknown op: %s", fmtOp(op))
		}
	}
	return nil
}

func (vm *VM) push(obj object.Object) {
	if vm.sp == len(vm.stack) {
		vm.stack = slices.Grow(vm.stack, vm.sp)
		vm.stack = vm.stack[:cap(vm.stack)]
	}

	vm.stack[vm.sp] = obj
	vm.sp++
}

func (vm *VM) pop() object.Object {
	vm.sp--
	return vm.stack[vm.sp]
}

func executeBinaryOp(op code.Opcode, left, right object.Object) object.Object {
	switch {
	case op == code.OpEqual:
		return lib.Boolean(lib.Equals(left, right))
	case op == code.OpNotEqual:
		return lib.Boolean(!lib.Equals(left, right))
	case op == code.OpAdd:
		return lib.Add(left, right)
	case op == code.OpMul:
		return lib.Multiply(left, right)
	case left.Type() != right.Type():
		return lib.Error("%s: type mismatch: %s %s", fmtOp(op), left.Type(), right.Type())
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		leftInt := left.(*object.Integer)
		rightInt := right.(*object.Integer)
		return executeIntegerOp(op, leftInt, rightInt)
	}
	return lib.Error("unknown op: %s %s %s", left.Type(), fmtOp(op), right.Type())
}

func executeIntegerOp(op code.Opcode, left, right *object.Integer) object.Object {
	switch op {
	case code.OpSub:
		return &object.Integer{Value: left.Value - right.Value}
	case code.OpDiv:
		return &object.Integer{Value: left.Value / right.Value}
	case code.OpLessThan:
		return lib.Boolean(left.Value < right.Value)
	}
	return lib.Error("unknown operator: %s %s %s", left.Type(), fmtOp(op), right.Type())
}

func fmtOp(op code.Opcode) string {
	def, err := code.Lookup(byte(op))
	if err != nil {
		return "<unknown>"
	}
	return def.Name
}
