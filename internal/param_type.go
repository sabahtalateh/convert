package internal

import (
	"errors"
	"go/ast"
)

type paramTypeRef struct {
	pointer bool
	array   bool
	path    []string
	err     error
}

func parseParamTypeRef(e ast.Expr) (*paramTypeRef, error) {
	t := new(paramTypeRef)
	err := t.visitParamExpr(e)
	if err != nil {
		return nil, err
	}

	if len(t.path) == 0 {
		return nil, errors.New("malformed expression")
	}

	return t, nil
}

func (t *paramTypeRef) visitParamExpr(e ast.Expr) error {
	switch ee := e.(type) {
	case *ast.StarExpr:
		return t.visitParamStarExpt(ee)
	case *ast.ArrayType:
		return t.visitParamArrayType(ee)
	case *ast.Ident:
		return t.visitParamIdent(ee)
	case *ast.SelectorExpr:
		return t.visitParamSelectorExpr(ee)
	default:
		return errUnsupportedType
	}
}

func (t *paramTypeRef) visitParamStarExpt(e *ast.StarExpr) error {
	if t.pointer {
		return errUnsupportedType
	}

	t.pointer = true
	return t.visitParamExpr(e.X)
}

func (t *paramTypeRef) visitParamArrayType(e *ast.ArrayType) error {
	if t.pointer || t.array {
		return errUnsupportedType
	}

	t.array = true
	return t.visitParamExpr(e.Elt)
}

func (t *paramTypeRef) visitParamIdent(e *ast.Ident) error {
	t.path = append(t.path, e.Name)
	return nil
}

func (t *paramTypeRef) visitParamSelectorExpr(e *ast.SelectorExpr) error {
	if err := t.visitParamExpr(e.X); err != nil {
		return err
	}
	t.path = append(t.path, e.Sel.Name)
	return nil
}
