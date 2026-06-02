package vm

import (
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
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, 1),
		sp:    0,
	}
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
		case code.OpAdd:
			left, right := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			vm.sp--
			vm.stack[vm.sp-1] = lib.Add(left, right)
		case code.OpPop:
			vm.sp--
		}
	}
	return nil
}

func (vm *VM) push(obj object.Object) {
	if vm.sp == cap(vm.stack) {
		newLength := vm.sp * 2
		newStack := make([]object.Object, newLength)
		copy(newStack, vm.stack)
		vm.stack = newStack
	}

	vm.stack[vm.sp] = obj
	vm.sp++
}
