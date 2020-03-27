package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mattn/golisp"
)

func repl() {
	env := golisp.NewEnv()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parser := golisp.NewParser(strings.NewReader(scanner.Text()))
		node, err := parser.ParseParen()
		if err != nil {
			log.Fatal(err)
		}

		ret, err := env.Eval(node)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ret)
	}
}

func main() {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		repl()
		return
	}

	parser := golisp.NewParser(os.Stdin)
	node, err := parser.ParseParen()
	if err != nil {
		log.Fatal(err)
	}

	env := golisp.NewEnv()
	_, err = env.Eval(node)
	if err != nil {
		log.Fatal(err)
	}
}
