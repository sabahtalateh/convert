package internal

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
)

type params struct {
	funcBodyStart token.Position
	funcBodyEnd   token.Position

	createOut bool
	inP       *param
	outP      *param
}

type param struct {
	varName  string
	pointer  bool
	array    bool
	dir      string
	pkgPath  string
	pkgName  string
	typeName string
	typ      *ast.StructType
	field    *ast.Field
	imports  []*ast.ImportSpec
}

func (p *params) in(ctx Context, inField *ast.Field) error {
	if len(inField.Names) == 0 || inField.Names[0].Name == "" {
		return errors.Join(errors.New(NodeToString(inField)), errors.New("parameter should have name"))
	}

	p.inP = new(param)
	p.inP.varName = inField.Names[0].Name
	err := p.inP.fromType(ctx, inField.Type)
	if err != nil {
		return errors.Join(errors.New(NodeToString(inField)), err)
	}

	td, imports, err := findTypeSpecAtDir(p.inP.dir, p.inP.typeName)
	if err != nil {
		return errors.Join(errors.New(NodeToString(inField)), err)
	}

	p.inP.typ = td
	p.inP.field = inField
	p.inP.imports = imports

	return nil
}

func (p *params) out(ctx Context, field *ast.Field) error {
	p.outP = new(param)
	p.outP.field = field
	err := p.outP.fromType(ctx, field.Type)
	if errors.Is(err, errImportDirNotFound) {
		return errors.Join(errors.New(NodeToString(field)), errTypeNotFound)
	}
	if err != nil {
		return errors.Join(errors.New(NodeToString(field)), err)
	}

	td, imports, err := findTypeSpecAtDir(p.outP.dir, p.outP.typeName)
	if err == nil {
		p.outP.typ = td
		p.outP.imports = imports
		return nil
	}

	if errors.Is(errTypeNotFound, err) {
		tr, err := parseParamTypeRef(field.Type)
		if err != nil {
			return errors.Join(errors.New(NodeToString(field)), err)
		}

		if len(tr.path) != 1 {
			return errors.Join(errors.New(NodeToString(field)), errTypeNotFound)
		}

		p.createOut = true
		return nil
	}

	return errors.Join(errors.New(NodeToString(field)), err)
}

func (p *param) fromType(ctx Context, typ ast.Expr) error {
	tr, err := parseParamTypeRef(typ)
	if err != nil {
		return err
	}

	p.pointer = tr.pointer
	p.array = tr.array

	if len(tr.path) == 1 {
		obj := types.Universe.Lookup(tr.path[0])
		if obj != nil {
			if _, ok := obj.Type().(*types.Basic); ok {
				return errBasicType
			}
		}

		p.dir = ctx.WorkDir
		p.pkgPath, err = dirPkgPath(p.dir)
		if err != nil {
			return err
		}
		pkgInf, err := getPackageInfo(p.pkgPath)
		if err != nil {
			return err
		}
		p.pkgName = pkgInf.pkgName
		p.typeName = tr.path[0]
	} else {
		pkgInf, err := pkgInfoByName(ctx.Imports, tr.path[0])
		if err != nil {
			return err
		}
		p.dir = pkgInf.dir
		p.pkgPath = pkgInf.pkgPath
		p.pkgName = pkgInf.pkgName
		p.typeName = tr.path[1]
	}

	return nil
}

func (p *params) addPosition(v int) {
	p.funcBodyStart.Line += v
	if p.funcBodyStart.Line < 0 {
		p.funcBodyStart.Line = 0
	}

	p.funcBodyEnd.Line += v
	if p.funcBodyEnd.Line < 0 {
		p.funcBodyEnd.Line = 0
	}
}
