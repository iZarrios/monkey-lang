package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/iZarrios/monkey-lang/lexer"
	"github.com/iZarrios/monkey-lang/token"
)

const (
	PROMPT = ">> "
)

func Start(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)
	for {
		fmt.Fprint(w, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.NewLexer(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			if tok.Type == token.ILLEGAL {
				fmt.Fprintln(w, "Error while lexing")
				break
			} else {
				fmt.Fprintf(w, "%+v\n", tok)
			}
		}
	}
}
