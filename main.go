package main

import (
	"fmt"
	"os"

	"github.com/Glorforidor/didactic_compiler/checker"
	"github.com/Glorforidor/didactic_compiler/compiler"
	"github.com/Glorforidor/didactic_compiler/lexer"
	"github.com/Glorforidor/didactic_compiler/parser"
	"github.com/Glorforidor/didactic_compiler/repl"
	"github.com/Glorforidor/didactic_compiler/resolver"
	"github.com/Glorforidor/didactic_compiler/symbol"
)

func main() {
	if 1 < len(os.Args) {
		b, err := os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}

		l := lexer.New(string(b))
		p := parser.New(l)
		c := compiler.New()

		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				fmt.Println(msg)
			}

			return
		}

		if err := resolver.Resolve(program, symbol.NewTable()); err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		if err := checker.Check(program); err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		if err := c.Compile(program); err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		fmt.Println(c.Asm())
	} else {
		fmt.Println("Welcome to the Didactic Compiler")
		repl.Start(os.Stdin, os.Stdout)
	}
}
