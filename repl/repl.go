package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/akojo/monkey/compiler"
	"github.com/akojo/monkey/evaluator"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
	"github.com/akojo/monkey/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, useEvaluator bool) {
	fmt.Print("monkey 1.7\n")

	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	constants := []object.Object{}
	globals := []object.Object{}
	symbolTable := compiler.NewSymbolTable()

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		p := parser.New(lexer.New(strings.NewReader(line), "<stdin>"))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(os.Stderr, p.Errors())
			continue
		}

		if useEvaluator {
			err := evaluator.Eval(program, env)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
			continue
		}

		c := compiler.NewWithState(symbolTable, constants)
		err := c.Compile(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "compiler error: %s\n", err)
			continue
		}

		code := c.Bytecode()
		constants = code.Constants

		if code.GlobalsSize > len(globals) {
			globals = slices.Grow(globals, code.GlobalsSize-len(globals))
			globals = globals[:cap(globals)]
		}

		machine := vm.NewWithGlobals(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			continue
		}

		result := machine.StackAboveTop()
		if result != nil {
			fmt.Fprintf(os.Stdout, "%s\n", result.Inspect())
		}
	}
}

func Run(in io.Reader, filename string, useEvaluator bool) object.Object {
	p := parser.New(lexer.New(in, filename))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(os.Stderr, p.Errors())
		return nil
	}

	if useEvaluator {
		return evaluator.Eval(program, object.NewEnvironment())
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compiler error: %s\n", err)
		return nil
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return nil
	}

	result := machine.StackAboveTop()
	if result != nil {
		fmt.Fprintf(os.Stdout, "%s\n", result.Inspect())
	}

	return nil
}

func printParseErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		fmt.Fprintf(out, "\t%s\n", err)
	}
}
