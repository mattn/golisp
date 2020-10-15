package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mattn/golisp"
)

func repl() {
	env := golisp.NewEnv(nil)
	err := golisp.LoadLib(env)
	if err != nil {

	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parser := golisp.NewParser(strings.NewReader(scanner.Text()))
		node, err := parser.Parse()
		if err != nil {
			log.Fatal(err)
		}

		ret, err := env.Eval(node)
		if err != nil {
			log.Fatal(err)
		}
		s := ret.String()
		if s != "" {
			fmt.Println(s)
		}
	}
}

func main() {
	flag.Parse()

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}

	var f *os.File
	var err error

	if flag.NArg() == 0 {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			repl()
			return
		}
		f = os.Stdin
	}

	if flag.NArg() == 1 {
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	parser := golisp.NewParser(f)
	node, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	env := golisp.NewEnv(nil)
	err = golisp.LoadLib(env)
	if err != nil {
		log.Fatal(err)
	}
	_, err = env.Eval(node)
	if err != nil {
		log.Fatal(err)
	}
}
