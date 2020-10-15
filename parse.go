package golisp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	EOF = errors.New("unexpected end of file")
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
	NodeGoValue
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
			p.unreadRune()
			return
		}
	}
}

func (p *Parser) ParseParen(bq bool) (*Node, error) {
	first := true
	head := &Node{
		t: NodeCell,
	}
	curr := head
	for {
		p.SkipWhite()
		b, err := p.buf.Peek(1)
		if err == io.EOF || (len(b) > 0 && b[0] == ')') {
			break
		}
		quote := err == nil && b[0] == ','
		if quote {
			p.buf.ReadByte()
		}

		child, err := p.ParseAny(false)
		if err != nil {
			return nil, err
		}
		if child == nil {
			break
		}
		if bq && !quote {
			child = &Node{
				t:   NodeQuote,
				car: child,
			}
		}

		if child.t == NodeIdent && child.v.(string) == "." && !first {
			child, err = p.ParseAny(false)
			if err != nil {
				return nil, err
			}
			curr.cdr = child
			break
		} else if head.car != nil {
			x := &Node{
				t: NodeCell,
				car: &Node{
					t: NodeNil,
				},
			}
			curr.cdr = x
			curr = x
		}
		first = false
		curr.car = child
	}
	if head.car == nil && head.cdr == nil {
		head.t = NodeNil
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

		if r == '\\' {
			r, err = p.readRune()
			if err != nil {
				return nil, err
			}
			switch r {
			case '\\':
				r = '\\'
			case 'n':
				r = '\n'
			case 'r':
				r = '\r'
			case 't':
				r = '\t'
			case 'b':
				r = '\b'
			case 'f':
				r = '\f'
			case '"':
				buf.WriteRune(r)
				continue
			}
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

func (p *Parser) ParseQuote() (*Node, error) {
	node, err := p.ParseAny(false)
	if err != nil {
		return nil, err
	}
	return &Node{
		t:   NodeQuote,
		car: node,
	}, nil
}

func (p *Parser) ParseBquote() (*Node, error) {
	node, err := p.ParseAny(true)
	if err != nil {
		return nil, err
	}
	return &Node{
		t:   NodeBquote,
		car: node,
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

func (p *Parser) unreadRune() error {
	err := p.buf.UnreadRune()
	p.pos -= 1
	return err
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
			p.unreadRune()
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
	return &Node{
		t: NodeIdent,
		v: s,
	}, nil
}

func (p *Parser) ParseAny(bq bool) (*Node, error) {
	p.SkipWhite()
	r, err := p.readRune()
	if err != nil {
		return nil, err
	}

	if r == '(' {
		node, err := p.ParseParen(bq)
		if err != nil {
			return nil, err
		}
		r, err := p.readRune()
		if err != nil || r != ')' {
			return nil, EOF
		}
		return node, nil
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || isSymbolLetter(r) {
		p.unreadRune()
		return p.ParsePrimitive()
	}
	if r == '\'' {
		return p.ParseQuote()
	}
	if r == '`' {
		return p.ParseBquote()
	}
	if r == '"' {
		return p.ParseString()
	}
	return nil, fmt.Errorf("invalid token: '%c' (%d)", r, p.Pos())
}

func (n *Node) String() string {
	if n == nil {
		return "nil"
	}
	var buf bytes.Buffer
	switch n.t {
	case NodeCell:
		curr := n
		fmt.Fprint(&buf, "(")
		for curr != nil {
			if curr.car != nil {
				fmt.Fprint(&buf, curr.car)
			} else {
				fmt.Fprint(&buf, "nil")
			}
			if curr.cdr == nil || curr.cdr.t == NodeNil {
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
		fmt.Fprint(&buf, ")")
	case NodeNil:
		fmt.Fprint(&buf, "nil")
	case NodeT:
		fmt.Fprint(&buf, "t")
	case NodeQuote:
		fmt.Fprintf(&buf, "'%v", n.car)
	case NodeBquote:
		fmt.Fprintf(&buf, "`%v", n.car)
	case NodeString:
		fmt.Fprintf(&buf, "%q", n.v)
	case NodeLambda:
		fmt.Fprintf(&buf, "(lambda %v %v)", n.car, n.cdr.car)
	case NodeEnv:
		if n.car != nil {
			fmt.Fprintf(&buf, "(defun %v %v %v)", n.v, n.car, n.cdr.car)
		} else {
			fmt.Fprintf(&buf, "(defun %v %v)", n.v, n.cdr.car)
		}
	case NodeGoValue:
		rv, ok := n.v.(reflect.Value)
		if ok {
			switch rv.Kind() {
			case reflect.String:
				fmt.Fprintf(&buf, "%q", rv.Interface())
			default:
				fmt.Fprint(&buf, rv.Interface())
			}
		}
	default:
		fmt.Fprint(&buf, n.v)
	}
	return buf.String()
}

func (p *Parser) Parse() (*Node, error) {
	return p.ParseParen(false)
}
