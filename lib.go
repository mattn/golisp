package golisp

import (
	"path"

	"github.com/rakyll/statik/fs"
)

//go:generate statik -src=lib

func LoadLib(env *Env) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	dir, err := statikFS.Open("/")
	if err != nil {
		return err
	}
	defer dir.Close()

	fis, err := dir.Readdir(-1)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		f, err := statikFS.Open(path.Join("/", fi.Name()))
		if err != nil {
			return err
		}
		node, err := NewParser(f).Parse()
		if err != nil {
			f.Close()
			return err
		}
		_, err = env.Eval(node)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}
