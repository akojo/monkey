package compiler

import (
	"fmt"
	"slices"
	"strings"

	"github.com/akojo/monkey/ast"
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions

	currentInstruction  EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes       []CompilationScope
	currentScope int
}

const INVALID_OFFSET = 65535

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		currentInstruction:  EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),

		scopes:       []CompilationScope{mainScope},
		currentScope: 0,
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
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
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
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b ast.Expression) int {
			return strings.Compare(a.String(), b.String())
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs))
	case *ast.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(code.OpIndex)
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

		c.replaceOperands(jumpToElse, len(c.instructions()))

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.compileBlockExpression(node.Alternative)
			if err != nil {
				return err
			}
		}

		c.replaceOperands(jumpToEnd, len(c.instructions()))
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable '%s'", node.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
		}
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.currentInstruction().Opcode == code.OpPop {
			c.replaceLastPopWithReturn()
		}

		if c.currentInstruction().Opcode != code.OpReturnValue {
			c.emit(code.OpReturnValue)
		}

		locals := len(c.symbolTable.store)
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{
			Instructions: instructions,
			Locals:       locals,
			Parameters:   len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFn))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
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
	if c.currentInstruction().Opcode == code.OpPop {
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
		Instructions: c.instructions(),
		Constants:    c.constants,
		GlobalsSize:  len(c.symbolTable.store),
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.add(ins)

	scope := c.scopes[c.currentScope]
	c.scopes[c.currentScope].previousInstruction = scope.currentInstruction
	c.scopes[c.currentScope].currentInstruction = EmittedInstruction{Opcode: op, Position: pos}

	return pos
}

func (c *Compiler) addConstant(obj object.Object) int {
	pos := len(c.constants)
	c.constants = append(c.constants, obj)
	return pos
}

func (c *Compiler) add(ins []byte) int {
	instructions := c.instructions()
	pos := len(instructions)

	c.scopes[c.currentScope].instructions = append(instructions, ins...)
	return pos
}

func (c *Compiler) instructions() code.Instructions {
	return c.scopes[c.currentScope].instructions
}

func (c *Compiler) currentInstruction() EmittedInstruction {
	return c.scopes[c.currentScope].currentInstruction
}

func (c *Compiler) removeCurrent() {
	scope := c.scopes[c.currentScope]
	c.scopes[c.currentScope].instructions = scope.instructions[:scope.currentInstruction.Position]
	c.scopes[c.currentScope].currentInstruction = scope.previousInstruction
}

func (c *Compiler) replaceOperands(pos int, operand ...int) {
	// Get the current op to calculate correct operand lengths
	instructions := c.instructions()
	op := code.Opcode(instructions[pos])
	instruction := code.Make(op, operand...)

	copy(instructions[pos+1:], instruction[1:])
}

func (c *Compiler) replaceLastPopWithReturn() {
	pos := c.scopes[c.currentScope].currentInstruction.Position
	returnValue := code.Make(code.OpReturnValue)
	copy(c.scopes[c.currentScope].instructions[pos:], returnValue)

	c.scopes[c.currentScope].currentInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		currentInstruction:  EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.currentScope++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.instructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.currentScope--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	GlobalsSize  int
}
