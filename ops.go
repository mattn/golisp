package golisp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

type Fn func(*Env, *Node) (*Node, error)

type FnType struct {
	special bool
	fn      Fn
}

var ops map[string]FnType

func makeFn(special bool, fn Fn) FnType {
	return FnType{special: special, fn: fn}
}

func init() {
	ops = make(map[string]FnType)
	ops["dotimes"] = makeFn(false, doDotimes)
	ops["prin1"] = makeFn(false, doPrin1)
	ops["print"] = makeFn(false, doPrint)
	ops["let"] = makeFn(false, doLet)
	ops["setq"] = makeFn(true, doSetq)
	ops["1+"] = makeFn(false, doPlusOne)
	ops["1-"] = makeFn(false, doMinusOne)
	ops["+"] = makeFn(false, doPlus)
	ops["-"] = makeFn(false, doMinus)
	ops["*"] = makeFn(false, doMul)
	ops["/"] = makeFn(false, doDiv)
	ops["<"] = makeFn(false, doLt)
	ops["<="] = makeFn(false, doLe)
	ops[">"] = makeFn(false, doGt)
	ops[">="] = makeFn(false, doGe)
	ops["="] = makeFn(false, doEqual)
	ops["if"] = makeFn(false, doIf)
	ops["not"] = makeFn(false, doNot)
	ops["mod"] = makeFn(false, doMod)
	ops["%"] = makeFn(false, doMod)
	ops["and"] = makeFn(false, doAnd)
	ops["or"] = makeFn(false, doOr)
	ops["cond"] = makeFn(true, doCond)
	ops["cons"] = makeFn(false, doCons)
	ops["car"] = makeFn(false, doCar)
	ops["cdr"] = makeFn(false, doCdr)
	ops["first"] = makeFn(false, doCar)
	ops["rest"] = makeFn(false, doCdr)
	ops["apply"] = makeFn(false, doApply)
	ops["concatenate"] = makeFn(false, doConcatenate)
	ops["defun"] = makeFn(true, doDefun)
	ops["quote"] = makeFn(false, doQuote)
	ops["getenv"] = makeFn(false, doGetenv)
	ops["length"] = makeFn(false, doLength)
	ops["null"] = makeFn(false, doNull)
}

type Env struct {
	vars map[string]*Node
	fncs map[string]*Node
	env  *Env
	out  io.Writer
}

func NewEnv(env *Env) *Env {
	var out io.Writer = os.Stdout
	if env != nil {
		out = env.out
	}
	return &Env{
		vars: make(map[string]*Node),
		fncs: make(map[string]*Node),
		env:  env,
		out:  out,
	}
}

func (e *Env) Eval(node *Node) (*Node, error) {
	var ret *Node
	var err error
	for node != nil && node.car != nil {
		ret, err = eval(e, node.car)
		if err != nil {
			return nil, err
		}
		node = node.cdr
	}
	return ret, nil
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

		e = env
		for e.env != nil {
			e = e.env
		}
		v, ok := e.fncs[name]
		if ok {
			return v, nil
		}

		return nil, fmt.Errorf("undefined symbol: %v", node.v)
	case NodeCell:
		if node.car == nil {
			return &Node{
				t: NodeNil,
				v: nil,
			}, nil
		}
		var err error
		if node.car != nil && node.car.t == NodeIdent {
			name := node.car.v.(string)
			ft, ok := ops[name]
			if !ok {
				e := env
				var fn *Node
				var ok bool
				for e != nil {
					fn, ok = e.fncs[name]
					if ok {
						break
					}
					e = e.env
				}
				if fn == nil {
					return nil, fmt.Errorf("invalid op: %v", name)
				}
				ev := &Node{
					t:   NodeCell,
					car: fn,
					cdr: node.cdr,
				}
				return eval(env, ev)
			}

			if ft.special {
				ret, err = ft.fn(env, node.cdr)
				if err != nil {
					return nil, err
				}
			} else {
				head := &Node{
					t: NodeCell,
					car: &Node{
						t: NodeNil,
					},
					cdr: &Node{
						t: NodeNil,
					},
				}
				arg := head
				if node.cdr != nil {
					curr := node.cdr
					for curr != nil && curr.car != nil {
						vv, err := eval(env, curr.car)
						if err != nil {
							return nil, err
						}
						arg.cdr = &Node{
							t:   NodeCell,
							car: vv,
						}
						arg = arg.cdr
						curr = curr.cdr
					}
				}
				newenv := NewEnv(env)
				ret, err = ft.fn(newenv, head.cdr)
				if err != nil {
					return nil, err
				}
			}
		} else if node.car != nil && node.car.t == NodeEnv {
			scope := NewEnv(node.car.e)
			var code *Node
			if node.car.cdr.car != nil {
				arg := node.car.cdr.car
				val := node.cdr
				for arg != nil && arg.car != nil {
					vv, err := eval(env, val.car)
					if err != nil {
						return nil, err
					}
					scope.vars[arg.car.v.(string)] = vv
					arg = arg.cdr
					val = val.cdr
				}
				if node.car.cdr.cdr != nil && node.car.cdr.cdr.car != nil {
					code = node.car.cdr.cdr
				} else {
					code = node.car.cdr
				}
			} else {
				code = node.car.cdr
			}

			for code != nil && code.car != nil {
				ret, err = eval(scope, code.car)
				if err != nil {
					return nil, err
				}
				code = code.cdr
			}
		}
		return ret, nil
	case NodeQuote:
		ret = node.car
	default:
		ret = node
	}
	return ret, nil
}

func doPrin1(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for prin1")
	}
	if node.car.t == NodeNil {
		fmt.Fprint(env.out, "nil")
	} else {
		fmt.Fprint(env.out, node.car.v)
	}
	return node.car, nil
}

func doPrint(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for print")
	}
	if node.car.t == NodeNil {
		fmt.Fprintln(env.out, "nil")
	} else if node.car.t == NodeT {
		fmt.Fprintln(env.out, "t")
	} else if node.car.t == NodeQuote {
		fmt.Fprintln(env.out, node.car)
	} else if node.car.t == NodeCell {
		fmt.Fprintln(env.out, node.car)
	} else {
		fmt.Fprintln(env.out, node.car.v)
	}
	return node.car, nil
}

func doDotimes(env *Env, node *Node) (*Node, error) {
	var err error

	if node.car == nil || node.car.car == nil {
		return nil, errors.New("invalid arguments for dotimes")
	}
	if node.car == nil || node.car.cdr == nil || node.car.cdr.car == nil {
		return nil, errors.New("invalid arguments for dotimes")
	}
	v := node.car.car.v.(string)
	c := node.car.cdr.car.v.(int64)

	scope := NewEnv(env)
	vv := &Node{
		t: NodeInt,
		v: int64(0),
		e: scope,
	}
	scope.vars[v] = vv

	node = node.cdr
	for i := int64(0); i < c; i++ {
		vv.v = i
		if node != nil {
			curr := node
			for curr != nil {
				_, err = eval(scope, curr.car)
				if err != nil {
					return nil, err
				}
				curr = curr.cdr
			}
		}
	}
	return &Node{
		t: NodeNil,
	}, nil
}

func doLet(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for if")
	}
	var ret *Node
	var err error
	v, err := eval(env, node.car.car)
	if err != nil {
		return nil, err
	}
	vv, err := eval(env, node.cdr)
	if err != nil {
		return nil, err
	}
	scope := NewEnv(env)
	scope.vars[v.v.(string)] = vv
	curr := node.cdr
	for curr != nil {
		ret, err = eval(scope, curr.car)
		if err != nil {
			return nil, err
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doSetq(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeIdent {
		return nil, errors.New("invalid arguments for setq")
	}
	env.vars[node.car.v.(string)] = node.cdr.car
	return node.cdr.car, nil
}

func doPlusOne(env *Env, node *Node) (*Node, error) {
	if node.car == nil || (node.car.t != NodeInt && node.car.t != NodeDouble) {
		return nil, errors.New("invalid arguments for 1+")
	}

	ret := &Node{
		t: node.car.t,
		v: node.car.v,
	}
	switch ret.t {
	case NodeInt:
		ret.v = ret.v.(int64) + 1
	case NodeDouble:
		ret.v = ret.v.(float64) + 1
	default:
		return nil, errors.New("invalid arguments for 1+")
	}
	return ret, nil
}

func doPlus(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return &Node{
			t: NodeInt,
			v: int64(0),
		}, nil
	}
	if node.car.t != NodeInt && node.car.t != NodeDouble {
		return nil, errors.New("invalid arguments for +")
	}

	ret := &Node{
		t: node.car.t,
		v: node.car.v,
	}
	curr := node.cdr
	for curr != nil {
		switch ret.t {
		case NodeInt:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(int64) + curr.car.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) + curr.car.v.(float64)
				ret.t = NodeDouble
			default:
				return nil, errors.New("invalid arguments for +")
			}
		case NodeDouble:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(float64) + float64(curr.car.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) + curr.car.v.(float64)
			default:
				return nil, errors.New("invalid arguments for +")
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doMinusOne(env *Env, node *Node) (*Node, error) {
	if node.car == nil || (node.car.t != NodeInt && node.car.t != NodeDouble) {
		return nil, errors.New("invalid arguments for 1-")
	}

	ret := &Node{
		t: node.car.t,
		v: node.car.v,
	}
	switch ret.t {
	case NodeInt:
		ret.v = ret.v.(int64) - 1
	case NodeDouble:
		ret.v = ret.v.(float64) - 1
	default:
		return nil, errors.New("invalid arguments for 1-")
	}
	return ret, nil
}

func doMinus(env *Env, node *Node) (*Node, error) {
	if node.car == nil || (node.car.t != NodeInt && node.car.t != NodeDouble) {
		return nil, errors.New("invalid arguments for -")
	}

	var ret *Node
	curr := node
	if curr.cdr == nil {
		ret = &Node{
			t: NodeInt,
			v: int64(0),
		}
	} else {
		ret = &Node{
			t: node.car.t,
			v: node.car.v,
		}
		curr = curr.cdr
	}
	for curr != nil {
		switch ret.t {
		case NodeInt:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(int64) - curr.car.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) - curr.car.v.(float64)
				ret.t = NodeDouble
			default:
				return nil, errors.New("invalid arguments for -")
			}
		case NodeDouble:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(float64) - float64(curr.car.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) - curr.car.v.(float64)
			default:
				return nil, errors.New("invalid arguments for -")
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doMul(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return &Node{
			t: NodeInt,
			v: int64(1),
		}, nil
	}
	if node.car.t != NodeInt && node.car.t != NodeDouble {
		return nil, errors.New("invalid arguments for *")
	}

	ret := &Node{
		t: node.car.t,
		v: node.car.v,
	}
	curr := node.cdr
	for curr != nil {
		switch ret.t {
		case NodeInt:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(int64) * curr.car.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) * curr.car.v.(float64)
				ret.t = NodeDouble
			default:
				return nil, errors.New("invalid arguments for *")
			}
		case NodeDouble:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(float64) * float64(curr.car.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) * curr.car.v.(float64)
			default:
				return nil, errors.New("invalid arguments for *")
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doDiv(env *Env, node *Node) (*Node, error) {
	if node.car == nil || (node.car.t != NodeInt && node.car.t != NodeDouble) {
		return nil, errors.New("invalid arguments for /")
	}
	if node.cdr == nil {
		switch node.car.t {
		case NodeInt:
			return &Node{
				t: NodeInt,
				v: 1 / node.car.v.(int64),
			}, nil
		case NodeDouble:
			return &Node{
				t: NodeDouble,
				v: 1.0 / node.car.v.(float64),
			}, nil
		default:
			return nil, errors.New("invalid arguments for /")
		}
	}

	ret := &Node{
		t: node.car.t,
		v: node.car.v,
	}
	curr := node.cdr
	for curr != nil {
		switch ret.t {
		case NodeInt:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(int64) / curr.car.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) / curr.car.v.(float64)
				ret.t = NodeDouble
			default:
				return nil, errors.New("invalid arguments for /")
			}
		case NodeDouble:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(float64) / float64(curr.car.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) / curr.car.v.(float64)
			default:
				return nil, errors.New("invalid arguments for /")
			}
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doEqual(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

	var b bool
	switch lhs.t {
	case NodeInt:
		f1 := lhs.v.(int64)
		switch rhs.t {
		case NodeInt:
			b = f1 == rhs.v.(int64)
		case NodeDouble:
			b = f1 == int64(rhs.v.(float64))
		}
	case NodeDouble:
		f1 := lhs.v.(float64)
		switch rhs.t {
		case NodeInt:
			b = f1 == float64(rhs.v.(int64))
		case NodeDouble:
			b = f1 == rhs.v.(float64)
		}
	case NodeString:
		f1 := lhs.v.(string)
		switch rhs.t {
		case NodeString:
			b = f1 == rhs.v.(string)
		}
	}

	if b {
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

func doGt(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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

	if f1 > f2 {
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

func doGe(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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

	if f1 >= f2 {
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

func doLt(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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

func doLe(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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

	if f1 <= f2 {
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
	if node.car == nil || node.car.cdr == nil {
		return nil, errors.New("invalid arguments for if")
	}

	v, err := eval(env, node.car)
	if err != nil {
		return nil, err
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
		if node.car.cdr != nil {
			v, err = eval(env, node.car.cdr.car)
			if err != nil {
				return nil, err
			}
		} else {
			return &Node{
				t: NodeT,
			}, nil
		}
	} else if node.cdr != nil && node.cdr.cdr != nil {
		if node.cdr != nil && node.cdr.cdr != nil {
			v, err = eval(env, node.cdr.cdr.car)
			if err != nil {
				return nil, err
			}
		} else {
			return &Node{
				t: NodeNil,
			}, nil
		}
	}
	return v, nil
}

func doNot(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for not")
	}

	var b bool
	switch node.car.t {
	case NodeInt:
		b = node.car.v.(int64) != 0
	case NodeDouble:
		b = node.car.v.(float64) != 0
	case NodeT:
		b = true
	case NodeQuote:
		return nil, errors.New("invalid arguments for not")
	}

	if !b {
		if node.car.cdr != nil {
			return eval(env, node.car.cdr.car)
		} else {
			return &Node{
				t: NodeT,
			}, nil
		}
	} else {
		if node.cdr != nil && node.cdr.cdr != nil {
			return eval(env, node.cdr.cdr.car)
		} else {
			return &Node{
				t: NodeNil,
			}, nil
		}
	}
}

func doMod(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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
	lhs := node.car
	rhs := node.cdr.car

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
		b2 = true
	}

	if b1 && b2 {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}

	return &Node{
		t: NodeNil,
	}, nil
}

func doOr(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car

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
		b2 = true
	}

	if b1 || b2 {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}

	return &Node{
		t: NodeNil,
	}, nil
}

func doCond(env *Env, node *Node) (*Node, error) {
	var ret *Node
	var err error

	ret = &Node{
		t: NodeNil,
	}
	if node == nil {
		return ret, nil
	}
	curr := node.car
	for curr != nil && curr.car != nil {
		ret, err = eval(env, curr.car)
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
			if curr.cdr != nil {
				ret, err = eval(env, curr.cdr.car)
				if err != nil {
					return nil, err
				}
			}
			break
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doCons(env *Env, node *Node) (*Node, error) {
	lhs := node.car
	rhs := node.cdr.car
	return &Node{
		t:   NodeCell,
		car: lhs,
		cdr: rhs,
	}, nil
}

func doCar(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.car == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	if node.car.t == NodeQuote {
		return &Node{
			t: NodeIdent,
			v: "quote",
		}, nil
	}
	return node.car.car, nil
}

func doCdr(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	if node.car.t == NodeQuote {
		return &Node{
			t:   NodeCell,
			car: node.car.car.cdr,
		}, nil
	}
	return node.car.cdr, nil
}

func doApply(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.cdr == nil || node.cdr.car == nil {
		return nil, errors.New("invalid arguments for apply")
	}
	arg := node.cdr
	if arg.car.t == NodeQuote {
		arg = arg.car.car
	}
	v := &Node{
		t:   NodeCell,
		car: node.car.car,
		cdr: arg,
	}
	return eval(env, v)
}

func doAref(env *Env, node *Node) (*Node, error) {
	return &Node{
		t:   NodeAref,
		car: node.car,
	}, nil
}

func doConcatenate(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeQuote {
		return nil, errors.New("invalid arguments for concatenate")
	}
	var buf bytes.Buffer
	fmt.Println(node.car)
	curr := node.car
	for curr != nil {
		switch curr.car.t {
		case NodeString:
			buf.WriteString(curr.car.v.(string))
		default:
			return nil, errors.New("invalid arguments for concatenate")
		}
		curr = curr.cdr
	}

	return &Node{
		t: NodeString,
		v: buf.String(),
	}, nil
}

func doDefun(env *Env, node *Node) (*Node, error) {
	v := &Node{
		t: NodeEnv,
		e: env,
		v: node.car.v,
	}
	v.cdr = node.cdr

	global := env
	for global.env != nil {
		global = global.env
	}

	global.fncs[node.car.v.(string)] = v
	return v, nil
}

func doQuote(env *Env, node *Node) (*Node, error) {
	return &Node{
		t: NodeQuote,
		v: node,
	}, nil
}

func doGetenv(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeString {
		return nil, errors.New("invalid arguments for getenv")
	}
	return &Node{
		t: NodeString,
		v: os.Getenv(node.car.v.(string)),
	}, nil
}

func doLength(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for length")
	}
	var l int64
	switch node.car.t {
	case NodeString:
		l = int64(len(node.car.v.(string)))
	case NodeCell:
		curr := node.car
		if curr.t == NodeNil {
			break
		}
		l++
		for curr.cdr != nil && curr.cdr.t != NodeNil {
			l++
			curr = curr.cdr
		}
	case NodeNil:
		l = 0
	}
	return &Node{
		t: NodeInt,
		v: l,
	}, nil
}

func doNull(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for length")
	}
	if node.car.t == NodeNil {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}

	return &Node{
		t: NodeNil,
	}, nil
}
