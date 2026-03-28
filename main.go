package main

import (
	"os"

	"github.com/akojo/monkey/repl"
	"golang.org/x/term"
)

func main() {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		repl.Start(os.Stdin, os.Stdout)
	} else {
		repl.Run(os.Stdin, "<stdin>")
	}
}
