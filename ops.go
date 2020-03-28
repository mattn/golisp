package golisp

import (
	"bytes"
	"errors"
	"fmt"
)

type Fn func(*Env, *Node) (*Node, error)

var ops map[string]Fn

func init() {
	ops = make(map[string]Fn)
	ops["dotimes"] = doDotimes
	ops["prin1"] = doPrin1
	ops["print"] = doPrint
	ops["let"] = doLet
	ops["setq"] = doSetq
	ops["+"] = doPlus
	ops["-"] = doMinus
	ops["*"] = doMul
	ops["/"] = doDiv
	ops["<"] = doLess
	ops["="] = doEqual
	ops["if"] = doIf
	ops["mod"] = doMod
	ops["and"] = doAnd
	ops["or"] = doOr
	ops["cond"] = doCond
	ops["cons"] = doCons
	ops["car"] = doCar
	ops["cdr"] = doCdr
	ops["apply"] = doApply
	ops["concatenate"] = doConcatenate
}

func eval(env *Env, node *Node) (*Node, error) {
	var ret *Node
	switch node.t {
	case NodeIdent:
		name := node.v.(string)
		_, ok := ops[name]
		if ok {
			return node, nil
		}
		e := env
		for e != nil {
			v, ok := e.vars[name]
			if ok {
				return v, nil
			}
			e = e.env
		}
		return nil, fmt.Errorf("undefined symbol: %v", node.v)
	case NodeCell:
		if node.car != nil && node.car.t == NodeCell && node.cdr == nil {
			node = node.car
		}
		lhs, err := eval(env, node.car)
		if err != nil {
			return nil, err
		}
		if lhs != nil && lhs.t == NodeIdent {
			name := lhs.v.(string)
			fn, ok := ops[name]
			if !ok {
				return nil, fmt.Errorf("invalid op: %v", name)
			}

			ret, err = fn(env, node.cdr)
			if err != nil {
				return nil, err
			}
		} else {
			ret = lhs
		}
	case NodeQuote:
		ret = node.car
	default:
		ret = node
	}
	return ret, nil
}

func doPrin1(env *Env, node *Node) (*Node, error) {
	ret, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}
	fmt.Print(ret.v)
	return ret, nil
}

func doPrint(env *Env, node *Node) (*Node, error) {
	ret, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}
	fmt.Println(ret.v)
	return ret, nil
}

func doDotimes(env *Env, node *Node) (*Node, error) {
	var ret *Node
	var err error

	if node.car == nil || node.car.car == nil {
		return nil, errors.New("invalid arguments")
	}
	if node.car == nil || node.car.cdr == nil || node.car.cdr.car == nil {
		return nil, errors.New("invalid arguments")
	}
	v := node.car.car.v.(string)
	c := node.car.cdr.car.v.(int64)

	scope := NewEnv()
	vv := &Node{
		t: NodeInt,
		v: int64(0),
		e: scope,
	}
	scope.vars[v] = vv

	node = node.cdr
	for i := int64(0); i < c; i++ {
		vv.v = i
		curr := node
		for curr != nil {
			ret, err = eval(scope, curr)
			if err != nil {
				return nil, err
			}
			curr = curr.cdr
		}
	}
	return ret, nil
}

func doLet(env *Env, node *Node) (*Node, error) {
	var ret *Node
	var err error
	v := node.car.car.v.(string)
	vv, err := eval(env, node.cdr)
	if err != nil {
		return nil, err
	}
	scope := NewEnv()
	scope.vars[v] = vv
	curr := node.cdr.cdr
	for curr != nil {
		ret, err = eval(scope, curr)
		if err != nil {
			return nil, err
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doSetq(env *Env, node *Node) (*Node, error) {
	v := node.car.v.(string)
	vv, err := eval(env, node.cdr)
	if err != nil {
		return nil, err
	}
	scope := NewEnv()
	scope.vars[v] = vv
	return vv, nil
}

func doPlus(env *Env, node *Node) (*Node, error) {
	var ret *Node

	ret = &Node{
		t: NodeInt,
		v: int64(0),
	}
	curr := node
	for curr != nil {
		v, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		switch ret.t {
		case NodeInt:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(int64) + v.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) + v.v.(float64)
				ret.t = NodeDouble
			}
		case NodeDouble:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(float64) + float64(v.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) + v.v.(float64)
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doMinus(env *Env, node *Node) (*Node, error) {
	var ret *Node

	ret = &Node{
		t: NodeInt,
		v: int64(0),
	}
	curr := node
	for curr != nil {
		v, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		switch ret.t {
		case NodeInt:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(int64) - v.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) - v.v.(float64)
				ret.t = NodeDouble
			}
		case NodeDouble:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(float64) - float64(v.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) - v.v.(float64)
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doMul(env *Env, node *Node) (*Node, error) {
	var ret *Node

	ret = &Node{
		t: NodeInt,
		v: int64(1),
	}
	curr := node
	for curr != nil {
		v, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		switch ret.t {
		case NodeInt:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(int64) * v.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) * v.v.(float64)
				ret.t = NodeDouble
			}
		case NodeDouble:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(float64) * float64(v.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) * v.v.(float64)
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doDiv(env *Env, node *Node) (*Node, error) {
	var ret *Node

	ret = &Node{
		t: NodeInt,
		v: int64(1),
	}
	curr := node
	for curr != nil {
		v, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		switch ret.t {
		case NodeInt:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(int64) / v.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) / v.v.(float64)
				ret.t = NodeDouble
			}
		case NodeDouble:
			switch v.t {
			case NodeInt:
				ret.v = ret.v.(float64) / float64(v.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) / v.v.(float64)
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doEqual(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	var f1, f2 float64
	switch lhs.t {
	case NodeInt:
		f1 = float64(lhs.v.(int64))
	case NodeDouble:
		f1 = lhs.v.(float64)
	}
	switch rhs.t {
	case NodeInt:
		f2 = float64(rhs.v.(int64))
	case NodeDouble:
		f2 = rhs.v.(float64)
	}

	if f1 == f2 {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}

	return &Node{
		t: NodeNil,
		v: nil,
	}, nil
}

func doLess(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	var f1, f2 float64
	switch lhs.t {
	case NodeInt:
		f1 = float64(lhs.v.(int64))
	case NodeDouble:
		f1 = lhs.v.(float64)
	}
	switch rhs.t {
	case NodeInt:
		f2 = float64(rhs.v.(int64))
	case NodeDouble:
		f2 = rhs.v.(float64)
	}

	if f1 < f2 {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}

	return &Node{
		t: NodeNil,
		v: nil,
	}, nil
}

func doIf(env *Env, node *Node) (*Node, error) {
	v, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.car.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	var b bool
	switch v.t {
	case NodeInt:
		b = v.v.(int64) != 0
	case NodeDouble:
		b = v.v.(float64) != 0
	case NodeT:
		b = true
	}

	if b {
		v, err = eval(env, node.car.cdr.car)
		if err != nil {
			return nil, err
		}
	} else if node.car.cdr.car != nil {
		v, err = eval(env, node.car.cdr.car.car)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func doMod(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	var i1, i2 int64
	switch lhs.t {
	case NodeInt:
		i1 = lhs.v.(int64)
	case NodeDouble:
		i1 = int64(lhs.v.(float64))
	}
	switch rhs.t {
	case NodeInt:
		i2 = rhs.v.(int64)
	case NodeDouble:
		i2 = int64(rhs.v.(float64))
	}

	return &Node{
		t: NodeInt,
		v: i1 % i2,
	}, nil
}

func doAnd(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	var b1, b2 bool
	switch lhs.t {
	case NodeInt:
		b1 = lhs.v.(int64) != 0
	case NodeDouble:
		b1 = lhs.v.(float64) != 0
	case NodeT:
		b1 = true
	}
	switch rhs.t {
	case NodeInt:
		b2 = rhs.v.(int64) != 0
	case NodeDouble:
		b2 = rhs.v.(float64) != 0
	case NodeT:
		b1 = true
	}

	if b1 && b2 {
		return &Node{
			t: NodeNil,
			v: nil,
		}, nil
	}

	return &Node{
		t: NodeNil,
		v: nil,
	}, nil
}

func doOr(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	var b1, b2 bool
	switch lhs.t {
	case NodeInt:
		b1 = lhs.v.(int64) != 0
	case NodeDouble:
		b1 = lhs.v.(float64) != 0
	case NodeT:
		b1 = true
	}
	switch rhs.t {
	case NodeInt:
		b2 = rhs.v.(int64) != 0
	case NodeDouble:
		b2 = rhs.v.(float64) != 0
	case NodeT:
		b1 = true
	}

	if b1 || b2 {
		return &Node{
			t: NodeNil,
			v: nil,
		}, nil
	}

	return &Node{
		t: NodeNil,
		v: nil,
	}, nil
}

func doCond(env *Env, node *Node) (*Node, error) {
	var ret *Node
	var err error

	curr := node
	for curr != nil {
		if curr.car == nil || curr.car.cdr == nil {
			return nil, errors.New("invalid arguments")
		}
		ret, err = eval(env, curr.car.car)
		if err != nil {
			return nil, err
		}
		var b bool
		switch ret.t {
		case NodeInt:
			b = ret.v.(int64) != 0
		case NodeDouble:
			b = ret.v.(float64) != 0
		case NodeT:
			b = true
		}
		if b {
			ret, err = eval(env, curr.car.cdr.car)
			if err != nil {
				return nil, err
			}
			break
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doCons(env *Env, node *Node) (*Node, error) {
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
	}

	if node.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	rhs, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}

	return &Node{
		t:   NodeCell,
		car: lhs,
		cdr: rhs,
	}, nil
}

func doCar(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	return node.car.cdr.car, nil
}

func doCdr(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.cdr == nil {
		return nil, errors.New("invalid arguments")
	}
	return node.car.cdr.cdr, nil
}

func doApply(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.cdr == nil || node.cdr.car == nil {
		return nil, errors.New("invalid arguments")
	}
	v := &Node{
		t:   NodeCell,
		car: node.car.car,
		cdr: node.cdr.car.car,
	}
	return eval(env, v)
}

func doAref(env *Env, node *Node) (*Node, error) {
	v := &Node{
		t:   NodeAref,
		car: node.car,
	}
	return v, nil
}

func doConcatenate(env *Env, node *Node) (*Node, error) {
	var buf bytes.Buffer
	curr := node
	for curr != nil {
		v, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		switch v.t {
		case NodeString:
			buf.WriteString(v.v.(string))
		default:
			return nil, errors.New("invalid arguments")
		}
		curr = curr.cdr
	}

	return &Node{
		t: NodeString,
		v: buf.String(),
	}, nil
}
