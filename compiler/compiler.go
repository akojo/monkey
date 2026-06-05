package compiler

import (
	"fmt"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	currentInstruction  EmittedInstruction
	previousInstruction EmittedInstruction

	symbolTable *SymbolTable
}

const INVALID_OFFSET = 65535

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		currentInstruction:  EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants

	return compiler
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.BlockStatement:
		if len(node.Statements) == 0 {
			c.emit(code.OpNull)
		}
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.InfixExpression:
		var err error
		if node.Operator == ">" {
			err = c.CompileN(node.Right, node.Left)
		} else {
			err = c.CompileN(node.Left, node.Right)
		}
		if err != nil {
			return err
		}

		switch node.Operator {
		case "<", ">":
			c.emit(code.OpLessThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.IFExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpToElse := c.emit(code.OpBranchNotEqual, INVALID_OFFSET)

		err = c.compileBlockExpression(node.Consequence)
		if err != nil {
			return err
		}

		jumpToEnd := c.emit(code.OpBranch, INVALID_OFFSET)

		c.replaceOperands(jumpToElse, len(c.instructions))

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.compileBlockExpression(node.Alternative)
			if err != nil {
				return err
			}
		}

		c.replaceOperands(jumpToEnd, len(c.instructions))
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable '%s'", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)
	default:
		return fmt.Errorf("unknown expression: %T (%s)", node, node.String())
	}
	return nil
}

func (c *Compiler) compileBlockExpression(stmt *ast.BlockStatement) error {
	err := c.Compile(stmt)
	if err != nil {
		return err
	}

	// keep last value of block expression on the stack
	if c.currentInstruction.Opcode == code.OpPop {
		c.removeCurrent()
	}

	return nil
}

func (c *Compiler) CompileN(nodes ...ast.Node) error {
	for _, node := range nodes {
		err := c.Compile(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		GlobalsSize:  len(c.symbolTable.store),
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.add(ins)

	c.previousInstruction = c.currentInstruction
	c.currentInstruction = EmittedInstruction{Opcode: op, Position: pos}

	return pos
}

func (c *Compiler) addConstant(obj object.Object) int {
	pos := len(c.constants)
	c.constants = append(c.constants, obj)
	return pos
}

func (c *Compiler) add(ins []byte) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return pos
}

func (c *Compiler) removeCurrent() {
	c.instructions = c.instructions[:c.currentInstruction.Position]
	c.currentInstruction = c.previousInstruction
}

func (c *Compiler) replaceOperands(pos int, operand ...int) {
	// Get the current op to calculate correct operand lengths
	op := code.Opcode(c.instructions[pos])
	instruction := code.Make(op, operand...)

	copy(c.instructions[pos+1:], instruction[1:])
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	GlobalsSize  int
}
