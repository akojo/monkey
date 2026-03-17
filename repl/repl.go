package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/akojo/monkey/evaluator"
	"github.com/akojo/monkey/lexer"
	"github.com/akojo/monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

		result := evaluator.Eval(program)
		if result != nil {
			fmt.Fprintf(out, "%s\n", result.Inspect())
		}
	}
}

func printParseErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		fmt.Fprintf(out, "\t%s\n", err)
	}
}
