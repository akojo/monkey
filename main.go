package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akojo/monkey/repl"
	"golang.org/x/term"
)

var useEvaluator = flag.Bool("interp", false, "Use direct evaluation instead of bytecode compiler\n")

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprint(out, "Usage: monkey [option]... [file | -]\n")
		fmt.Fprint(out, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) > 2 {
		flag.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) == 2 {
		runFile(flag.Args()[1])
		return
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		repl.Start(os.Stdin, *useEvaluator)
	} else {
		runFile("-")
	}
}

func runFile(filename string) {
	if filename == "-" {
		repl.Run(os.Stdin, "<stdin>", *useEvaluator)
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	repl.Run(file, filename, *useEvaluator)
}
