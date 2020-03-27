package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	parser := parser{
		buf: bufio.NewReader(os.Stdin),
	}
	node, err := parser.ParseParen()
	if err != nil {
		log.Fatal(err)
	}

	env := NewEnv()
	ret, err := eval(env, node)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ret)
}
