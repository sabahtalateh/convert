package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

func Convert(ctx Context) error {
	funcParameters := new(params)

	fn, fset, imports, err := findFunc(ctx)
	if err != nil {
		return err
	}

	selfPackagePath, err := dirPkgPath(filepath.Dir(ctx.Location.File))
	if err != nil {
		return err
	}

	fmt.Printf("%s.%s\n", selfPackagePath, fn.Name)

	funcParameters.funcBodyStart = fset.Position(fn.Body.Pos())
	funcParameters.funcBodyEnd = fset.Position(fn.Body.End())

	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return fmt.Errorf("function should have at least 1 parameter")
	}

	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return fmt.Errorf("function should return at least 1 result")
	}

	err = funcParameters.in(ctx.WithImports(imports), fn.Type.Params.List[0])
	if err != nil {
		return err
	}

	err = funcParameters.out(ctx.WithImports(imports), fn.Type.Results.List[0])
	if err != nil {
		return err
	}

	err = updateFile(ctx, funcParameters)
	if err != nil {
		return err
	}

	return nil
}

func findFunc(ctx Context) (*ast.FuncDecl, *token.FileSet, []*ast.ImportSpec, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, ctx.Location.File, nil, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		fun   *ast.FuncDecl
		gDecl *ast.GenDecl
	)

	ast.Inspect(file, func(n ast.Node) bool {
		if fun != nil || gDecl != nil {
			return false
		}
		if n == nil {
			return true
		}
		switch nn := n.(type) {
		case *ast.GenDecl:
			if fset.Position(n.Pos()).Line > ctx.Location.Line {
				gDecl = nn
				return false
			}
		case *ast.FuncDecl:
			if fset.Position(n.Pos()).Line > ctx.Location.Line {
				fun = nn
				return false
			}
		}

		return true
	})

	if gDecl != nil || fun == nil {
		return nil, nil, nil, errors.New("convert should be defined over func")
	}

	return fun, fset, file.Imports, nil
}
