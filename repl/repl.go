package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/Glorforidor/didactic_compiler/checker"
	"github.com/Glorforidor/didactic_compiler/compiler"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
	"github.com/Glorforidor/didactic_compiler/resolver"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

const prompt = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	t := symbol.NewTable()

	for {
		fmt.Fprintf(out, prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		c := compiler.New()

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrros(out, p.Errors())
			continue
		}

		if err := resolver.Resolve(program, t); err != nil {
			fmt.Fprintf(out, "%s\n", err)
			continue
		}

		if err := checker.Check(program); err != nil {
			fmt.Fprintf(out, "%s\n", err)
			continue
		}

		if err := c.Compile(program); err != nil {
			fmt.Fprintf(out, "%s\n", err)
			continue
		}

		fmt.Fprintln(out, c.Asm())
	}
}

func printParserErrros(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintln(out, msg)
	}
}
