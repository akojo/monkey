package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/akojo/monkey/repl"
	"golang.org/x/term"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprint(out, "Usage: monkey [file | -]\n")
	}

	flag.Parse()

	if len(flag.Args()) > 1 {
		flag.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) == 1 {
		runFile(flag.Args()[0])
		return
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		repl.Start(os.Stdin)
	} else {
		runFile("-")
	}
}

func runFile(filename string) {
	if filename == "-" {
		repl.Run(os.Stdin, "<stdin>")
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	repl.Run(file, filename)
}
