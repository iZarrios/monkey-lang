package main

import (
	"os"

	"github.com/iZarrios/monkey-lang/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
