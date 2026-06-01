package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		result := run(strings.NewReader(line), "<stdin>", env, useEvaluator)
		if result != nil {
			fmt.Fprintf(os.Stdout, "%s\n", result.Inspect())
		}
	}
}

func Run(in io.Reader, filename string, useEvaluator bool) object.Object {
	return run(in, filename, object.NewEnvironment(), useEvaluator)
}

func run(in io.Reader, filename string, env *object.Environment, useEvaluator bool) object.Object {
	p := parser.New(lexer.New(in, filename))

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(os.Stderr, p.Errors())
		return nil
	}

	if useEvaluator {
		return evaluator.Eval(program, env)
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

	return machine.StackTop()
}

func printParseErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		fmt.Fprintf(out, "\t%s\n", err)
	}
}
