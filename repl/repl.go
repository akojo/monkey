package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/akojo/monkey/evaluator"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/object"
	"github.com/akojo/monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	fmt.Print("monkey 1.7\n")

	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(strings.NewReader(line), "<stdin>")
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		result := evaluator.Eval(program, env)
		if result != nil {
			fmt.Fprintf(out, "%s\n", result.Inspect())
		}
	}
}

func Run(in io.Reader, filename string) {
	p := parser.New(lexer.New(in, filename))

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParseErrors(os.Stderr, p.Errors())
		return
	}

	evaluator.Eval(program, object.NewEnvironment())
}

func printParseErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		fmt.Fprintf(out, "\t%s\n", err)
	}
}
