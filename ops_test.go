package golisp

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOps(t *testing.T) {
	fns, err := filepath.Glob("testdir/*.lisp")
	if err != nil {
		t.Fatal(err)
	}

	for _, fn := range fns {
		t.Log(fn)
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}
		input := string(b)
		parser := NewParser(strings.NewReader(input))
		node, err := parser.Parse()
		if err != nil {
			b, err2 := ioutil.ReadFile(fn[:len(fn)-4] + "err")
			if err2 != nil || err.Error() != strings.TrimSpace(string(b)) {
				t.Error(err)
				continue
			}
			continue
		}
		var buf bytes.Buffer
		env := NewEnv(nil)
		err = LoadLib(env)
		if err != nil {
			t.Fatal(err)
		}
		env.out = &buf
		_, err = env.Eval(node)
		if err != nil {
			t.Error(err)
			continue
		}
		got := buf.String()
		b, err = ioutil.ReadFile(fn[:len(fn)-4] + "out")
		if err != nil {
			t.Fatal(err)
		}
		want := string(b)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(diff)
		}
	}
}
