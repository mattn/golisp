package main

import "fmt"

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
	ops["+"] = doMul
	ops["/"] = doDiv
	//ops["if"] = doIf
	//ops["="] = doEqual
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
