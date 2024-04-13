package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/iZarrios/monkey-lang/evaluator"
	"github.com/iZarrios/monkey-lang/lexer"
	"github.com/iZarrios/monkey-lang/object"
	"github.com/iZarrios/monkey-lang/parser"
)

const (
	PROMPT = ">> "
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()

		l := lexer.NewLexer(line)
		p, _ := parser.NewParser(l)

		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// We have parsed the whole program now and we have found no errors in it
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
