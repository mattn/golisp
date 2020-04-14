# golisp

![CI](https://github.com/mattn/golisp/workflows/CI/badge.svg)

Lisp Interpreter

## Usage

```shell
$ golisp < foo.lisp
```

## Installation

```shell
$ go get github.com/mattn/golisp/cmd/golisp
```

## Features

### Call Go functions.

Print random ints.

```lisp
(setq time (go:import 'time))
(setq rand (go:import 'math/rand))
(.Seed rand (.UnixNano (.Now time)))
(print (.Int rand))
```

### Use goroutine/channel

```lisp
(setq time (go:import time))
(let ((ch (go:make-chan string 1)))
  (go
    (.Sleep time 1e9)
    (go:chan-send ch "3"
    (.Sleep time 1e9)
    (go:chan-send ch "2"
    (.Sleep time 1e9)
    (go:chan-send ch "1"
    (.Sleep time 1e9)
    (go:chan-send ch "Fire!"
  )
  (print (car (go:chan-recv ch)))
  (print (car (go:chan-recv ch)))
  (print (car (go:chan-recv ch)))
  (print (car (go:chan-recv ch)))
)
```

## TODO

* macro

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
