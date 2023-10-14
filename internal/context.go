package internal

import (
	"go/ast"
	"go/token"
	"path/filepath"
)

type Context struct {
	Location Location
	WorkDir  string
	Imports  []*ast.ImportSpec
	FSet     *token.FileSet
}

func NewContext(loc Location) Context {
	return Context{Location: loc, WorkDir: filepath.Dir(loc.File)}
}

func (c Context) WithWorkDir(wd string) Context {
	return Context{
		Location: c.Location,
		WorkDir:  wd,
		Imports:  c.Imports,
		FSet:     c.FSet,
	}
}

func (c Context) WithImports(ii []*ast.ImportSpec) Context {
	return Context{
		Location: c.Location,
		WorkDir:  c.WorkDir,
		Imports:  ii,
		FSet:     c.FSet,
	}
}

func (c Context) WithFSet(fset *token.FileSet) Context {
	return Context{
		Location: c.Location,
		WorkDir:  c.WorkDir,
		Imports:  c.Imports,
		FSet:     fset,
	}
}

func (c Context) AddLocationLine(v int) Context {
	locLine := c.Location.Line + v
	if locLine < 0 {
		locLine = 0
	}

	return Context{
		Location: Location{
			File: c.Location.File,
			Line: locLine,
		},
		WorkDir: c.WorkDir,
		Imports: c.Imports,
		FSet:    c.FSet,
	}
}
