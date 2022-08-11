package golisp

//go:genearte stringer -t nodetype .

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/mattn/golisp/gopkg"
	_ "github.com/mattn/golisp/statik"
)

type Ft int

const (
	FtSpecial Ft = iota
	FtBuiltin
)

type Fn func(*Env, *Node) (*Node, error)

type FnInfo struct {
	ft Ft
	fn Fn
}

var ops map[string]FnInfo

func makeFn(ft Ft, fn Fn) FnInfo {
	return FnInfo{ft: ft, fn: fn}
}

func init() {
	ops = make(map[string]FnInfo)
	ops["dotimes"] = makeFn(FtSpecial, doDotimes)
	ops["prin1"] = makeFn(FtBuiltin, doPrin1)
	ops["print"] = makeFn(FtBuiltin, doPrint)
	ops["princ"] = makeFn(FtBuiltin, doPrinc)
	ops["let"] = makeFn(FtSpecial, doLet)
	ops["let*"] = makeFn(FtSpecial, doLetStar)
	ops["setq"] = makeFn(FtSpecial, doSetq)
	ops["1+"] = makeFn(FtBuiltin, doPlusOne)
	ops["1-"] = makeFn(FtBuiltin, doMinusOne)
	ops["+"] = makeFn(FtBuiltin, doPlus)
	ops["-"] = makeFn(FtBuiltin, doMinus)
	ops["*"] = makeFn(FtBuiltin, doMul)
	ops["/"] = makeFn(FtBuiltin, doDiv)
	ops["<"] = makeFn(FtBuiltin, doLt)
	ops["<="] = makeFn(FtBuiltin, doLe)
	ops[">"] = makeFn(FtBuiltin, doGt)
	ops[">="] = makeFn(FtBuiltin, doGe)
	ops["="] = makeFn(FtBuiltin, doEqual)
	ops["if"] = makeFn(FtSpecial, doIf)
	ops["not"] = makeFn(FtBuiltin, doNot)
	ops["mod"] = makeFn(FtBuiltin, doMod)
	ops["%"] = makeFn(FtBuiltin, doMod)
	ops["and"] = makeFn(FtSpecial, doAnd)
	ops["or"] = makeFn(FtSpecial, doOr)
	ops["cond"] = makeFn(FtSpecial, doCond)
	ops["cons"] = makeFn(FtBuiltin, doCons)
	ops["car"] = makeFn(FtBuiltin, doCar)
	ops["cdr"] = makeFn(FtBuiltin, doCdr)
	ops["rest"] = makeFn(FtBuiltin, doCdr)
	ops["apply"] = makeFn(FtBuiltin, doApply)
	ops["concatenate"] = makeFn(FtBuiltin, doConcatenate)
	ops["defun"] = makeFn(FtSpecial, doDefun)
	ops["float"] = makeFn(FtBuiltin, doFloat)
	ops["while"] = makeFn(FtSpecial, doWhile)
	ops["quote"] = makeFn(FtSpecial, doQuote)
	ops["getenv"] = makeFn(FtBuiltin, doGetenv)
	ops["length"] = makeFn(FtBuiltin, doLength)
	ops["null"] = makeFn(FtBuiltin, doNull)
	ops["list"] = makeFn(FtBuiltin, doList)
	ops["make-string"] = makeFn(FtBuiltin, doMakeString)
	ops["progn"] = makeFn(FtBuiltin, doProgn)
	ops["eval"] = makeFn(FtBuiltin, doEval)
	ops["consp"] = makeFn(FtBuiltin, doConsp)
	ops["oddp"] = makeFn(FtBuiltin, doOddp)
	ops["evenp"] = makeFn(FtBuiltin, doEvenp)
	ops["load"] = makeFn(FtBuiltin, doLoad)
	ops["funcall"] = makeFn(FtBuiltin, doFuncall)
	ops["lambda"] = makeFn(FtSpecial, doLambda)
	ops["type-of"] = makeFn(FtBuiltin, doTypeOf)
	ops["labels"] = makeFn(FtSpecial, doLabels)
	ops["flet"] = makeFn(FtSpecial, doFlet)
	ops["rplaca"] = makeFn(FtBuiltin, doRplaca)
	ops["rplacd"] = makeFn(FtBuiltin, doRplacd)
	ops["nconc"] = makeFn(FtBuiltin, doNconc)
	ops["defmacro"] = makeFn(FtSpecial, doDefmacro)

	ops["go:import"] = makeFn(FtSpecial, doGoImport)
	ops["go:make-chan"] = makeFn(FtSpecial, doGoMakeChan)
	ops["go:chan-recv"] = makeFn(FtBuiltin, doGoChanRecv)
	ops["go:chan-send"] = makeFn(FtBuiltin, doGoChanSend)
	ops["go"] = makeFn(FtSpecial, doGo)
}

type Env struct {
	vars map[string]*Node
	fncs map[string]*Node
	mcrs map[string]*Node
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
		mcrs: make(map[string]*Node),
		env:  env,
		out:  out,
	}
}

func (e *Env) SetOut(o io.Writer) {
	e.out = o
}

func (n *Node) CarIsNil() bool {
	return n.car == nil || n.car.t == NodeNil
}

func (n *Node) CdrIsNil() bool {
	return n.cdr == nil || n.cdr.t == NodeNil
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

func evalList(env *Env, node *Node) (*Node, error) {
	var head, prev *Node
	curr := node
	for curr != nil && curr.car != nil {
		vv, err := eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		if vv == nil {
			vv = &Node{
				t: NodeNil,
			}
		}
		vvv := &Node{
			t:   NodeCell,
			car: vv,
		}
		if prev != nil {
			prev.cdr = vvv
		} else {
			head = vvv
		}
		prev = vvv
		curr = curr.cdr
	}
	if head == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	return head, nil
}

func call(env *Env, node *Node) (*Node, error) {
	var macro bool

	if node.car != nil && node.car.t == NodeIdent {
		name := node.car.v.(string)
		if ft, ok := ops[name]; ok {
			return ft.fn(env, node.cdr)
		}
		e := env
		var fn, alist *Node
		var ok bool
		for e != nil {
			fn, ok = e.fncs[name]
			if ok {
				alist = node.cdr
				break
			}
			fn, ok = e.vars[name]
			if ok && fn.t == NodeLambda {
				alist = node.cdr
				break
			}
			e = e.env
		}
		if fn == nil {
			global := env
			for global.env != nil {
				global = global.env
			}
			fn, ok = global.mcrs[name]
			if ok {
				alist = node.cdr
				macro = true
			}
			if fn == nil {
				return nil, fmt.Errorf("invalid op: %v", name)
			}
		}
		if !macro && fn.t != NodeLambda {
			ev := &Node{
				t:   NodeCell,
				car: fn,
				cdr: alist,
			}
			return eval(env, ev)
		}
		if macro {
			node = &Node{
				t: NodeCell,
				car: &Node{
					t:   NodeEnv,
					v:   name,
					e:   fn.e,
					cdr: fn.cdr,
				},
				cdr: node.cdr,
			}
		} else {
			node = &Node{
				t: NodeCell,
				car: &Node{
					t:   NodeEnv,
					v:   name,
					e:   fn.e,
					cdr: fn,
				},
				cdr: node.cdr,
			}
		}
	} else if node.car != nil && node.car.t == NodeLambda {
		node = &Node{
			t: NodeCell,
			car: &Node{
				t:   NodeEnv,
				e:   node.car.e,
				car: node.car,
				cdr: node.car,
			},
			cdr: node.cdr,
		}
	}

	if node.car == nil || node.car.t != NodeEnv {
		return node, nil
	}

	scope := NewEnv(node.car.e)
	var code *Node
	if node.car.cdr.car != nil {
		arg := node.car.cdr.car
		val := node.cdr
		for arg != nil && arg.t != NodeNil {
			var name string
			if arg.car != nil {
				name = arg.car.v.(string)
			} else {
				name = arg.v.(string)
			}
			var vv *Node
			var err error
			if name == "&rest" {
				arg = arg.cdr
				name = arg.car.v.(string)
				vv, err = evalList(env, val)
			} else if macro {
				vv, err = val.car, nil
			} else if arg.car == nil {
				vv, err = evalList(env, val)
			} else {
				vv, err = eval(env, val.car)
			}
			if err != nil {
				return nil, err
			}
			scope.vars[name] = vv
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

	var ret *Node
	var err error
	for code != nil && code.car != nil {
		ret, err = eval(scope, code.car)
		if err != nil {
			return nil, err
		}
		code = code.cdr
	}
	return ret, nil
}

func eval(env *Env, node *Node) (*Node, error) {
	var ret *Node
	switch node.t {
	case NodeLambda:
		return call(env, node)
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
		if node.car.t == NodeIdent {
			ft, ok := ops[node.car.v.(string)]
			if ok {
				if ft.ft == FtBuiltin {
					alist, err := evalList(env, node.cdr)
					if err != nil {
						return nil, err
					}
					return ft.fn(env, alist)
				}
				alist := node.cdr
				if alist == nil {
					alist = &Node{
						t: NodeNil,
					}
				}
				return ft.fn(env, alist)
			}
			fn := node.car.v.(string)
			if fn[0] == '.' && node.cdr != nil {
				if len(fn) == 1 {
					return doGoField(env, node)
				}
				return doGoMethodCall(env, node)
			}
		}
		if node.car.t == NodeCell {
			if node.car.car != nil && node.car.car.t == NodeIdent && node.car.car.v.(string) == "lambda" {
				node = &Node{
					t: NodeCell,
					car: &Node{
						t:   NodeEnv,
						car: node.car.car,
						cdr: node.car.cdr,
					},
					cdr: node.cdr,
				}
			}
		}
		return call(env, node)
	case NodeQuote:
		ret = node.car
	case NodeBquote:
		return doBquote(env, node)
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
		fmt.Fprintln(env.out, node.car.v)
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

func doPrinc(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for print")
	}
	if node.car.t == NodeNil {
		fmt.Fprint(env.out, "nil")
	} else if node.car.t == NodeT {
		fmt.Fprint(env.out, "t")
	} else if node.car.t == NodeQuote {
		fmt.Fprint(env.out, node.car)
	} else if node.car.t == NodeCell {
		fmt.Fprint(env.out, node.car)
	} else {
		fmt.Fprint(env.out, node.car.v)
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
	count, err := eval(env, node.car.cdr.car)
	if err != nil {
		return nil, err
	}
	c := count.v.(int64)

	scope := NewEnv(env)
	vv := &Node{
		t: NodeInt,
		v: int64(0),
	}
	scope.vars[v] = vv

	cond := node.cdr
	var i int64
	for i = int64(0); i < c; i++ {
		vv = &Node{
			t: NodeInt,
			v: int64(i),
		}
		scope.vars[v] = vv
		if cond != nil {
			curr := cond
			for curr != nil {
				_, err = eval(scope, curr.car)
				if err != nil {
					return nil, err
				}
				curr = curr.cdr
			}
		}
	}
	vv = &Node{
		t: NodeInt,
		v: int64(i),
	}
	scope.vars[v] = vv

	if node.car.cdr.cdr != nil {
		return eval(scope, node.car.cdr.cdr.car)
	}

	return &Node{
		t: NodeNil,
	}, nil
}

func doLet(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for let")
	}
	if node.car.t == NodeNil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	scope := NewEnv(env)

	var ret, vv *Node
	var err error
	curr := node.car
	for curr != nil {
		if curr.car.cdr == nil {
			scope.vars[curr.car.v.(string)] = &Node{
				t: NodeNil,
			}
		} else {
			vv, err = eval(env, curr.car.cdr.car)
			if err != nil {
				return nil, err
			}
			switch curr.car.car.t {
			case NodeCell:
				obj, err := eval(env, curr.car.car.cdr.car)
				if err != nil {
					return nil, err
				}
				if obj.t == NodeCell && obj.car != nil && obj.car.t == NodeGoValue {
					obj = obj.car
				}
				fld := obj.v.(reflect.Value).Elem().FieldByName(curr.car.car.cdr.cdr.car.v.(string))
				fld.Set(vv.car.v.(reflect.Value))
			case NodeIdent:
				scope.vars[curr.car.car.v.(string)] = vv
			}
		}
		curr = curr.cdr
	}

	curr = node.cdr
	for curr != nil {
		ret, err = eval(scope, curr.car)
		if err != nil {
			return nil, err
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doLetStar(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for let")
	}
	if node.car.t == NodeNil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	scope := NewEnv(env)

	var ret, vv *Node
	var err error
	curr := node.car
	for curr != nil {
		if curr.car.cdr == nil {
			scope.vars[curr.car.v.(string)] = &Node{
				t: NodeNil,
			}
		} else {
			vv, err = eval(env, curr.car.cdr.car)
			if err != nil {
				return nil, err
			}
			scope.vars[curr.car.car.v.(string)] = vv
		}
		env = scope
		scope = NewEnv(env)
		curr = curr.cdr
	}

	curr = node.cdr
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
	if node.CarIsNil() {
		return node, nil
	}
	if node.car.t != NodeIdent {
		return nil, errors.New("invalid arguments for setq")
	}

	var ret *Node
	var err error
	curr := node
	for curr != nil {
		name := curr.car.v.(string)
		ret, err = eval(env, curr.cdr.car)
		if err != nil {
			return nil, err
		}
		e := env
		for e != nil {
			_, ok := e.vars[name]
			if ok {
				e.vars[name] = ret
				break
			}
			e = e.env
		}
		if e == nil {
			env.vars[name] = ret
		}
		curr = curr.cdr.cdr
	}
	return ret, nil
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
	if node.car == nil || node.car.t == NodeNil {
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
	for curr != nil && curr.car != nil {
		switch ret.t {
		case NodeInt:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(int64) + curr.car.v.(int64)
			case NodeDouble:
				ret.v = float64(ret.v.(int64)) + curr.car.v.(float64)
				ret.t = NodeDouble
			case NodeNil:
			default:
				return nil, errors.New("invalid arguments for +")
			}
		case NodeDouble:
			switch curr.car.t {
			case NodeInt:
				ret.v = ret.v.(float64) + float64(curr.car.v.(int64))
			case NodeDouble:
				ret.v = ret.v.(float64) + curr.car.v.(float64)
			case NodeNil:
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
			v, err = eval(env, node.cdr.car)
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
	if node.car == nil || node.car.t == NodeNil {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}
	return &Node{
		t: NodeNil,
	}, nil
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
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
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
	lhs, err := eval(env, node.car)
	if err != nil {
		return nil, err
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
	curr := node
	for curr != nil && curr.car != nil && curr.car.t != NodeNil {
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
			if curr.car.cdr != nil {
				curr = curr.car.cdr
				for curr != nil {
					ret, err = eval(env, curr.car)
					if err != nil {
						return nil, err
					}
					curr = curr.cdr
				}
			} else {
				ret = &Node{
					t: NodeT,
					v: true,
				}
			}
			break
		}
		curr = curr.cdr
	}
	return ret, nil
}

func doCons(env *Env, node *Node) (*Node, error) {
	var rhs *Node
	lhs := node.car
	rhs = node.cdr.car
	return &Node{
		t:   NodeCell,
		car: lhs,
		cdr: rhs,
	}, nil
}

func doCar(env *Env, node *Node) (*Node, error) {
	curr := node.car
	if curr.t == NodeQuote {
		return &Node{
			t: NodeIdent,
			v: "quote",
		}, nil
	}
	if curr.t != NodeCell && curr.t != NodeNil {
		return nil, fmt.Errorf("arguments should be list: %v", curr)
	}

	curr = curr.car
	if curr == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	return curr, nil
}

func doCdr(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	curr := node.car
	if curr.t == NodeQuote {
		return &Node{
			t:   NodeCell,
			car: curr.car,
		}, nil
	}

	curr = curr.cdr
	if curr == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	return curr, nil
}

func doFirst(env *Env, node *Node) (*Node, error) {
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

func doSecond(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.car == nil || node.car.car.car == nil {
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
	return node.car.car.car, nil
}

func doThird(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.car == nil || node.car.car.car == nil || node.car.car.car.car == nil {
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
	return node.car.car.car.car, nil
}

func doApply(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.cdr == nil || node.cdr.car == nil {
		return nil, errors.New("invalid arguments for apply")
	}

	var head, x *Node
	curr := node.cdr
	for curr != nil && curr.cdr != nil && curr.cdr.t != NodeNil {
		nn := &Node{
			t:   NodeCell,
			car: curr.car,
		}
		if head != nil {
			x.cdr = nn
			x = nn
		} else {
			head, x = nn, nn
		}
		curr = curr.cdr
	}

	if curr.car != nil && curr.car.t != NodeNil && curr.car.t != NodeCell {
		return nil, fmt.Errorf("last argument should be list: %v", node.car)
	}
	if head != nil {
		x.cdr = curr.car
	} else {
		head = curr.car
	}

	vv := &Node{
		t:   NodeCell,
		car: node.car,
		cdr: head,
	}
	return call(env, vv)
}

func doAref(env *Env, node *Node) (*Node, error) {
	return &Node{
		t:   NodeAref,
		car: node.car,
	}, nil
}

func doConcatenate(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeIdent {
		return nil, errors.New("invalid arguments for concatenate")
	}
	var buf bytes.Buffer
	curr := node.cdr
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

func doFloat(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for float")
	}
	var ret float64
	switch node.car.t {
	case NodeInt:
		ret = float64(node.car.v.(int64))
	case NodeDouble:
		ret = node.car.v.(float64)
	default:
		return nil, errors.New("invalid arguments for float")
	}

	return &Node{
		t: NodeDouble,
		v: ret,
	}, nil
}

func doWhile(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for while")
	}

	scope := NewEnv(env)

	for {
		ret, err := eval(env, node.car)
		if err != nil {
			return nil, err
		}
		if ret.t != NodeT {
			break
		}
		ret, err = doProgn(scope, node)
		if err != nil {
			return nil, err
		}
	}

	return &Node{
		t: NodeNil,
		v: nil,
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
	return node.car, nil
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

func doList(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return &Node{
			t: NodeNil,
		}, nil
	}
	return node, nil
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

func doMakeString(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeInt {
		return nil, errors.New("invalid arguments for make-string")
	}

	return &Node{
		t: NodeString,
		v: strings.Repeat(" ", int(node.car.v.(int64))),
	}, nil
}

func doProgn(env *Env, node *Node) (*Node, error) {
	ret := &Node{
		t: NodeNil,
	}
	var err error
	curr := node.cdr
	for curr != nil {
		ret, err = eval(env, curr.car)
		if err != nil {
			return nil, err
		}
		curr = curr.cdr
	}

	return ret, nil
}

func doEval(env *Env, node *Node) (*Node, error) {
	return eval(env, node.car)
}

func doConsp(env *Env, node *Node) (*Node, error) {
	switch node.car.t {
	case NodeQuote, NodeBquote, NodeCell:
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}
	return &Node{
		t: NodeNil,
	}, nil
}

func doLoad(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeString {
		return nil, errors.New("invalid arguments for load")
	}

	f, err := os.Open(node.car.v.(string))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	curr, err := NewParser(f).Parse()
	if err != nil {
		return nil, err
	}
	return env.Eval(curr)
}

func doFuncall(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for funcall")
	}
	v := &Node{
		t:   NodeCell,
		car: node.car,
		cdr: node.cdr,
	}
	return eval(env, v)
}

func doLambda(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for lambda")
	}
	return &Node{
		t:   NodeLambda,
		e:   env,
		car: node.car,
		cdr: node.cdr,
	}, nil
}

func doOddp(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for oddp")
	}

	var b bool
	switch node.car.t {
	case NodeInt:
		b = node.car.v.(int64)%2 != 0
	case NodeDouble:
		b = int64(node.car.v.(float64))%2 != 0
	}
	if b {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}
	return &Node{
		t: NodeNil,
	}, nil
}

func doEvenp(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for evenp")
	}

	var b bool
	switch node.car.t {
	case NodeInt:
		b = node.car.v.(int64)%2 == 0
	case NodeDouble:
		b = int64(node.car.v.(float64))%2 == 0
	}
	if b {
		return &Node{
			t: NodeT,
			v: true,
		}, nil
	}
	return &Node{
		t: NodeNil,
	}, nil
}

func doTypeOf(env *Env, node *Node) (*Node, error) {
	t := "unknown"

	if node.car == nil {
		return nil, errors.New("invalid arguments for type-of")
	}

	curr := node.car
	switch curr.t {
	case NodeNil:
		t = "null"
	case NodeT:
		t = "boolean"
	case NodeInt:
		t = "int"
	case NodeDouble:
		t = "float"
	case NodeString:
		t = "string"
	case NodeQuote:
		t = "cons"
	case NodeBquote:
		t = "cons"
	case NodeCell:
		t = "cons"
	case NodeAref:
		t = "aref"
	case NodeBuiltinfunc:
	case NodeSpecial:
	case NodeLambda:
		t = "function"
	case NodeIdent:
		t = "symbol"
	case NodeEnv:
		t = "environment"
	case NodeGoValue:
		t = "go:" + reflect.TypeOf(curr.v).String()
	case NodeError:
		t = "error"
	}

	return &Node{
		t: NodeString,
		v: t,
	}, nil
}

func doLabels(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for labels")
	}

	curr := node.car
	scope := NewEnv(env)

	for curr != nil && curr.car != nil && curr.car.car != nil {
		if curr.car == nil || curr.car.car.t != NodeIdent || curr.car.cdr.t != NodeCell {
			return nil, errors.New("invalid arguments for labels")
		}
		vv := &Node{
			t:   NodeEnv,
			e:   scope,
			v:   curr.car.car.v.(string),
			car: curr.car.cdr.car,
			cdr: curr.car.cdr.cdr,
		}

		nn := &Node{
			t:   NodeLambda,
			e:   scope,
			car: vv.car,
			cdr: vv.cdr,
		}

		scope.fncs[curr.car.car.v.(string)] = nn
		curr = curr.cdr
	}

	return doProgn(scope, node)
}

func doFlet(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for flet")
	}

	curr := node.car
	scope := NewEnv(env)

	for curr != nil && curr.car != nil && curr.car.car != nil {
		if curr.car == nil || curr.car.car.t != NodeIdent || curr.car.cdr.t != NodeCell {
			return nil, errors.New("invalid arguments for flet")
		}
		vv := &Node{
			t:   NodeEnv,
			e:   env,
			v:   curr.car.car.v.(string),
			car: curr.car.cdr.car,
			cdr: curr.car.cdr.cdr,
		}

		nn := &Node{
			t:   NodeLambda,
			e:   env,
			car: vv.car,
			cdr: vv.cdr,
		}

		scope.fncs[curr.car.car.v.(string)] = nn
		curr = curr.cdr
	}

	return doProgn(scope, node)
}

func doBquote(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for bquote")
	}
	curr := node

	var bq bool
	if curr.t == NodeBquote {
		curr = curr.car
		bq = true
	}

	var l, rr *Node
	for curr != nil {
		expand := 0
		var v *Node
		var ok bool
		var err error
		if bq && curr.car.t == NodeIdent {
			name := curr.car.v.(string)
			if strings.HasPrefix(name, "@") {
				expand = 2
				v, ok = env.vars[name[1:]]
				if !ok {
					return nil, errors.New("invalid arguments for bquote")
				}
			} else {
				expand = 1
				v, ok = env.vars[name]
				if !ok {
					return nil, errors.New("invalid arguments for bquote")
				}
			}
		} else {
			v, err = eval(env, curr.car)
			if err != nil {
				return nil, err
			}
		}
		if l == nil {
			switch expand {
			case 0, 1:
				l = &Node{
					t:   NodeCell,
					car: v,
				}
				rr = l
			case 2:
				rr, l = v.car, v.car
			}
		} else {
			switch expand {
			case 0, 1:
				l.cdr = &Node{
					t:   NodeCell,
					car: v,
				}
				l = l.cdr
				break
			case 2:
				l.cdr = v
				for l.cdr != nil {
					l = l.cdr
				}
			}
		}
		curr = curr.cdr
	}

	return rr, nil
}

func doRplaca(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeCell {
		return nil, errors.New("invalid arguments for rplaca")
	}

	node.car.car = node.cdr.car
	return node.car, nil
}

func doRplacd(env *Env, node *Node) (*Node, error) {
	if node.car == nil || node.car.t != NodeCell {
		return nil, errors.New("invalid arguments for rplacd")
	}

	lhs := node.car
	rhs := node.cdr.car
	lhs.cdr = rhs
	return lhs, nil
}

func doNconc(env *Env, node *Node) (*Node, error) {
	curr := node
	for curr != nil && curr.cdr != nil {
		if curr.car != nil && curr.car.t != NodeCell && curr.car.t != NodeNil {
			return nil, errors.New("invalid arguments for nconc")
		}
		curr = curr.cdr
	}

	var head, nn *Node
	curr = node

	for curr != nil && curr.t != NodeNil {
		if curr.CarIsNil() && !curr.CdrIsNil() {
			curr = curr.cdr
			continue
		}

		if head != nil {
			for nn.cdr != nil && nn.cdr.t == NodeCell {
				nn = nn.cdr
			}
			nn.cdr = curr.car
		} else {
			head, nn = curr.car, curr.car
		}
		curr = curr.cdr
	}
	return head, nil
}

func doDefmacro(env *Env, node *Node) (*Node, error) {
	if node.car == nil {
		return nil, errors.New("invalid arguments for defmacro")
	}
	nn := &Node{
		t:   NodeLambda,
		e:   env,
		car: node.car,
		cdr: node.cdr,
	}

	global := env
	for global.env != nil {
		global = global.env
	}

	global.mcrs[node.car.v.(string)] = nn

	return nn, nil
}

func doGoMethodCall(env *Env, node *Node) (rret *Node, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rerr = errors.New(fmt.Sprint(err))
		}
	}()
	name := node.car.v.(string)[1:]
	obj, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}
	if obj.t == NodeCell && obj.car != nil && obj.car.t == NodeGoValue {
		obj = obj.car
	}

	var rv reflect.Value
	var rrv []reflect.Value
	pkg, ok := obj.v.(map[string]reflect.Value)
	if ok {
		rv, ok := pkg[name]
		if !ok {
			return nil, fmt.Errorf("invalid symbol name: %v", name)
		}
		rt := rv.Type()
		numIn := rt.NumIn()

		curr := node.cdr.cdr
		args := []reflect.Value{}
		in := 0
		for curr != nil && curr.car != nil {
			arg, err := eval(env, curr.car)
			if err != nil {
				return nil, err
			}
			if arg.t == NodeCell && arg.car != nil && arg.car.t == NodeGoValue {
				arg = arg.car
			}
			var rav reflect.Value
			at := rt.In(in)
			if arg.t == NodeGoValue {
				rav = arg.v.(reflect.Value)
				args = append(args, rav.Convert(at))
			} else if arg.v == nil {
				args = append(args, reflect.Zero(at))
			} else {
				rav = reflect.ValueOf(arg.v)
				args = append(args, rav.Convert(at))
			}
			curr = curr.cdr

			in++
			if in == numIn {
				break
			}
		}
		rrv = rv.Call(args)
	} else {
		rv, ok = obj.v.(reflect.Value)
		if !ok {
			rv = reflect.ValueOf(obj.v)
		}
		method := rv.MethodByName(name)
		curr := node.cdr.cdr
		args := []reflect.Value{}
		for curr != nil {
			args = append(args, reflect.ValueOf(curr.car))
			curr = curr.cdr
		}
		rrv = method.Call(args)
	}

	rets := &Node{
		t: NodeCell,
	}

	head := rets
	for _, ret := range rrv {
		x := &Node{
			t: NodeCell,
			car: &Node{
				t: NodeGoValue,
				v: ret,
			},
		}
		rets.cdr = x
		rets = x
	}

	return head.cdr, nil
}

func doGoField(env *Env, node *Node) (rret *Node, rerr error) {
	if node.car == nil || node.cdr == nil || node.cdr.cdr == nil || node.cdr.cdr.car == nil {
		return nil, errors.New("invalid arguments for .")
	}

	defer func() {
		if err := recover(); err != nil {
			rerr = errors.New(fmt.Sprint(err))
		}
	}()
	obj, err := eval(env, node.cdr.car)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprint(node.cdr.cdr.car.v)

	if obj.t == NodeCell && obj.car != nil && obj.car.t == NodeGoValue {
		obj = obj.car
	}

	var rv reflect.Value
	pkg, ok := obj.v.(map[string]reflect.Value)
	if ok {
		rv, ok = pkg[name]
		if !ok {
			return nil, fmt.Errorf("invalid symbol name: %v", name)
		}
	} else {
		rv, ok = obj.v.(reflect.Value)
		if !ok {
			rv = reflect.ValueOf(obj.v)
		}
		rv = rv.FieldByName(name)
	}

	return &Node{
		t: NodeCell,
		car: &Node{
			t: NodeGoValue,
			v: rv,
		},
	}, nil
}

func doGoImport(env *Env, node *Node) (*Node, error) {
	var name string
	if node.car.car != nil {
		name = fmt.Sprint(node.car.car.v)
	} else {
		name = fmt.Sprint(node.car.v)
	}
	pkg, ok := gopkg.Packages[name]
	if !ok {
		return nil, fmt.Errorf("invalid package name: %v", name)
	}
	return &Node{
		t: NodeGoValue,
		v: pkg,
	}, nil
}

func doGoMakeChan(env *Env, node *Node) (*Node, error) {
	name := fmt.Sprint(node.car.v)
	typ, ok := gopkg.BasicTypes[name]
	if !ok {
		return nil, fmt.Errorf("invalid type name: %v", name)
	}
	size := 0
	if node.car.car != nil {
		if node.car.car.t != NodeInt {
			return nil, fmt.Errorf("invalid size name: %v", node.car.car.v)
		}
		size = int(node.car.car.v.(int64))
	}
	return &Node{
		t: NodeGoValue,
		v: reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), size),
	}, nil
}

func doGoChanSend(env *Env, node *Node) (rret *Node, rerr error) {
	/*
		defer func() {
			if err := recover(); err != nil {
				rerr = errors.New(fmt.Sprint(err))
			}
		}()
	*/
	if node.car == nil || node.cdr == nil || node.cdr.car == nil || node.car.t != NodeGoValue {
		fmt.Println(node.car)
		fmt.Println(node.car.cdr)
		fmt.Println(node.car.cdr.car)
		fmt.Println(node.car.t)
		return nil, errors.New("invalid arguments for go:chan-send")
	}

	ch := node.car.v.(reflect.Value)

	var rv reflect.Value
	if node.cdr.car.t == NodeGoValue {
		rv = node.car.car.v.(reflect.Value)
	} else {
		rv = reflect.ValueOf(node.cdr.car.v)
	}

	ch.Send(rv)

	return node.car.car, nil
}

func doGoChanRecv(env *Env, node *Node) (rret *Node, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rerr = errors.New(fmt.Sprint(err))
		}
	}()
	if node.car == nil || node.car.t != NodeGoValue {
		return nil, errors.New("invalid arguments for go:chan-recv")
	}

	ch := node.car.v.(reflect.Value)

	rv, ok := ch.Recv()

	var res *Node
	if ok {
		res = &Node{
			t: NodeT,
			v: true,
		}
	} else {
		res = &Node{
			t: NodeNil,
		}
	}

	return &Node{
		t: NodeCell,
		car: &Node{
			t: NodeGoValue,
			v: rv,
		},
		cdr: res,
	}, nil
}

func doGo(env *Env, node *Node) (rret *Node, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rerr = errors.New(fmt.Sprint(err))
		}
	}()
	if node.car == nil || node.car.t != NodeCell {
		return nil, errors.New("invalid arguments for go")
	}

	go func(env *Env) {
		defer func() {
			recover()
		}()
		curr := node
		for curr != nil {
			_, err := eval(env, curr.car)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			curr = curr.cdr
		}
	}(env)

	return &Node{
		t: NodeNil,
	}, nil
}
