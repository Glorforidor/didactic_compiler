package main

import (
	"fmt"
	"os"

	"github.com/Glorforidor/didactic_compiler/repl"
)

func main() {
	fmt.Println("Welcome to the Didactic Compiler")
	repl.Start(os.Stdin, os.Stdout)
}
