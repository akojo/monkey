package main

import (
	"fmt"
	"os"

	"github.com/akojo/monkey/repl"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "Usage: monkey [file | -]\n")
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		runFile(os.Args[1])
		return
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		repl.Start(os.Stdin, os.Stdout)
	} else {
		repl.Run(os.Stdin, "<stdin>")
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
