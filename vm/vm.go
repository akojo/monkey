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
	constants []object.Object

	stack []object.Object
	sp    int

	globals []object.Object

	frames []*Frame
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, 1),
		sp:    0,

		globals: make([]object.Object, bytecode.GlobalsSize),

		frames: []*Frame{NewFrame(mainFn, 0)},
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
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.frame().ip < len(vm.frame().Instructions())-1 {
		vm.frame().ip++

		ip = vm.frame().ip
		ins = vm.frame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			idx := code.ReadUint16(ins[ip+1:])
			vm.frame().ip += 2

			vm.push(vm.constants[idx])
		case code.OpPop:
			vm.sp--
		case code.OpGetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.frame().ip += 2

			vm.push(vm.globals[index])
		case code.OpSetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.frame().ip += 2

			vm.globals[index] = vm.pop()
		case code.OpGetLocal:
			index := int(code.ReadUint8(ins[ip+1:]))
			vm.frame().ip += 1

			vm.push(vm.stack[vm.frame().fp+index])
		case code.OpSetLocal:
			index := int(code.ReadUint8(ins[ip+1:]))
			vm.frame().ip += 1

			vm.stack[vm.frame().fp+index] = vm.pop()
		case code.OpNull:
			vm.push(lib.NULL)
		case code.OpFalse:
			vm.push(lib.FALSE)
		case code.OpTrue:
			vm.push(lib.TRUE)
		case code.OpArray:
			n := int(code.ReadUint16(ins[ip+1:]))
			vm.frame().ip += 2

			start, end := vm.sp-n, vm.sp
			elements := make([]object.Object, end-start)
			copy(elements, vm.stack[start:end])
			vm.sp -= n

			vm.stack[vm.sp] = &object.Array{Elements: elements}
			vm.sp++
		case code.OpHash:
			npairs := int(code.ReadUint16(ins[ip+1:]))
			vm.frame().ip += 2

			start, end := vm.sp-npairs*2, vm.sp
			pairs := make(map[object.HashKey]object.HashPair)

			for i := start; i < end; i += 2 {
				key := vm.stack[i]
				value := vm.stack[i+1]

				pair := object.HashPair{Key: key, Value: value}

				hashKey, ok := key.(object.Hashable)
				if !ok {
					return fmt.Errorf("cannot use as hash key: %s", key.Type())
				}

				pairs[hashKey.Hash()] = pair
			}
			vm.sp -= npairs * 2

			vm.stack[vm.sp] = &object.Hash{Pairs: pairs}
			vm.sp++
		case code.OpIndex:
			obj, index := vm.stack[vm.sp-2], vm.stack[vm.sp-1]

			result := lib.Index(obj, index)
			if result.Type() == object.ERROR {
				return errors.New(result.(*object.Error).Message)
			}

			vm.sp--
			vm.stack[vm.sp-1] = result
		case code.OpBranch:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.frame().ip = pos - 1
		case code.OpBranchNotEqual:
			if !lib.IsTruthy(vm.pop()) {
				pos := int(code.ReadUint16(ins[ip+1:]))
				vm.frame().ip = pos - 1
			} else {
				vm.frame().ip += 2
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
		case code.OpCall:
			nargs := int(code.ReadUint8(ins[ip+1:]))
			vm.frame().ip += 1

			err := vm.call(nargs)
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.fp

			vm.stack[vm.sp-1] = returnValue
		default:
			return fmt.Errorf("unknown op: %s", fmtOp(op))
		}
	}
	return nil
}

func (vm *VM) call(nargs int) error {
	vm.sp -= nargs

	fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("calling non-function: %T", vm.stack[vm.sp-1].Type())
	}

	if nargs != fn.Parameters {
		return fmt.Errorf("wrong number of arguments: want %d, got %d", fn.Parameters, nargs)
	}

	frame := NewFrame(fn, vm.sp)
	vm.pushFrame(frame)

	vm.sp = frame.fp
	vm.alloc(fn.Locals)
	return nil
}

// allocate n slots from the top of the stack
func (vm *VM) alloc(n int) {
	for len(vm.stack) <= vm.sp+n {
		vm.stack = slices.Grow(vm.stack, len(vm.stack))
		vm.stack = vm.stack[:cap(vm.stack)]
	}
	vm.sp += n
}

func (vm *VM) push(obj object.Object) {
	if vm.sp == len(vm.stack) {
		vm.stack = slices.Grow(vm.stack, len(vm.stack))
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

func (vm *VM) frame() *Frame {
	return vm.frames[len(vm.frames)-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames = append(vm.frames, f)
}

func (vm *VM) popFrame() *Frame {
	f := vm.frame()
	vm.frames = vm.frames[:len(vm.frames)-1]
	return f
}

func fmtOp(op code.Opcode) string {
	def, err := code.Lookup(byte(op))
	if err != nil {
		return "<unknown>"
	}
	return def.Name
}
