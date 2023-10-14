package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func resolveTypeRef(p *param, e ast.Expr) (typeRef, error) {
	out, err := parseTypeRef(e)
	if err != nil {
		return nil, err
	}
	if err = out.resolve(p); err != nil {
		return nil, err
	}

	return out, nil
}

func findTypeSpecAtDir(dir string, Type string) (*ast.StructType, []*ast.ImportSpec, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, nil, err
	}
	for pkgName, pkg := range pkgs {
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}

		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gen.Specs {
					t, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if t.Name == nil {
						continue
					}

					if t.Name.Name != Type {
						continue
					}

					Struct, ok := t.Type.(*ast.StructType)
					if !ok {
						return nil, nil, errNotStruct
					}

					return Struct, file.Imports, nil
				}
			}
		}
	}

	return nil, nil, errTypeNotFound
}
