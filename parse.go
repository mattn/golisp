package golisp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type NodeType int

const (
	NodeNil NodeType = iota
	NodeT
	NodeInt
	NodeDouble
	NodeString
	NodeQuote
	NodeBquote
	NodeIdent
	NodeLambda
	NodeSpecial
	NodeBuiltinfunc
	NodeCell
	NodeAref
	NodeEnv
	NodeError
)

type Node struct {
	t   NodeType
	v   interface{}
	e   *Env
	car *Node
	cdr *Node
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		buf: bufio.NewReader(r),
	}
}

type Parser struct {
	buf *bufio.Reader
	pos int
}

func (p *Parser) NewError(err error) *Node {
	return &Node{
		t: NodeError,
		v: err,
	}
}

func (p *Parser) SkipWhite() {
	for {
		r, err := p.readRune()
		if err != nil {
			return
		}
		if r == ';' {
			for {
				r, err = p.readRune()
				if err != nil {
					return
				}
				if r == '\n' {
					break
				}
			}
			continue
		}
		if !unicode.IsSpace(r) {
			p.buf.UnreadRune()
			return
		}
	}
}

func (p *Parser) ParseParen() (*Node, error) {
	head := &Node{
		t: NodeCell,
	}
	curr := head
	for {
		child, err := p.ParseAny()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if child == nil {
			break
		}
		if child.t == NodeIdent && child.v.(string) == "." {
			child, err = p.ParseAny()
			if err != nil {
				return nil, err
			}
			curr.cdr = child
			break
		} else if head.car != nil {
			x := &Node{
				t: NodeCell,
			}
			curr.cdr = x
			curr = x
		}
		curr.car = child
	}
	return head, nil
}

func (p *Parser) ParseString() (*Node, error) {
	var buf bytes.Buffer
	for {
		r, err := p.readRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if r == '"' {
			break
		}
		buf.WriteRune(r)
	}
	return &Node{
		t: NodeString,
		v: buf.String(),
	}, nil
}

func isSymbolLetter(r rune) bool {
	return strings.ContainsRune(`+-*/<>=&%?.@_#$:*`, r)
}

func (p *Parser) Pos() int {
	return p.pos
}

func (p *Parser) readRune() (rune, error) {
	r, n, err := p.buf.ReadRune()
	p.pos += n
	return r, err
}

func (p *Parser) ParsePrimitive() (*Node, error) {
	var buf bytes.Buffer
	for {
		r, err := p.readRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !isSymbolLetter(r) {
			p.buf.UnreadRune()
			break
		}
		buf.WriteRune(r)
	}

	s := buf.String()

	if s == "nil" {
		return &Node{
			t: NodeNil,
			v: nil,
		}, nil
	}
	if s == "t" {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &Node{
			t: NodeInt,
			v: i,
		}, nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return &Node{
			t: NodeDouble,
			v: f,
		}, nil
	}
	if s[0] == '"' {
		return p.ParseString()
	}
	return &Node{
		t: NodeIdent,
		v: s,
	}, nil
}

func (p *Parser) ParseAny() (*Node, error) {
	p.SkipWhite()
	r, err := p.readRune()
	if err != nil {
		return nil, err
	}

	if r == ')' {
		return nil, nil
	}
	if r == '(' {
		return p.ParseParen()
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || isSymbolLetter(r) {
		p.buf.UnreadRune()
		return p.ParsePrimitive()
	}
	if r == '\'' {
		node, err := p.ParseAny()
		if err != nil {
			return nil, err
		}
		return &Node{
			t:   NodeQuote,
			car: node,
		}, nil
	}
	if r == '"' {
		return p.ParseString()
	}
	return nil, fmt.Errorf("invalid token: '%c' (%d)", r, p.Pos())
}

func (n *Node) String() string {
	var buf bytes.Buffer
	switch n.t {
	case NodeCell:
		curr := n
		fmt.Fprint(&buf, "(")
		if curr.car != nil || curr.cdr != nil {
			for curr != nil {
				if curr.car != nil {
					fmt.Fprint(&buf, curr.car)
				} else {
					fmt.Fprint(&buf, "nil")
				}
				if curr.cdr == nil {
					break
				}
				if curr.cdr.t != NodeCell {
					fmt.Fprint(&buf, " . ")
					fmt.Fprint(&buf, curr.cdr)
					break
				}
				fmt.Fprint(&buf, " ")
				curr = curr.cdr
			}
		}
		fmt.Fprint(&buf, ")")
	case NodeNil:
		fmt.Fprint(&buf, "nil")
	case NodeT:
		fmt.Fprint(&buf, "t")
	case NodeQuote:
		fmt.Fprintf(&buf, "'%v", n.car)
	case NodeString:
		fmt.Fprintf(&buf, "%q", n.v)
	default:
		fmt.Fprint(&buf, n.v)
	}
	return buf.String()
}
