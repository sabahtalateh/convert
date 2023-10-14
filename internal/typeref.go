package internal

import (
	"fmt"
	"go/ast"
	"go/types"
)

type typeRef interface {
	typeRef()

	modifier(string)
	resolve(*param) error
	equals(ref2 typeRef) bool
	deref() typeRef
	code(Context) (string, []string)
}

type rType struct {
	path []string // used when walking ast

	modifiers []string
	pkgPath   string
	pkgName   string
	typeName  string
}

func (t *rType) modifier(m string) {
	t.modifiers = append(t.modifiers, m)
}

type rMap struct {
	modifiers []string
	key       typeRef
	val       typeRef
}

func (m *rMap) modifier(mod string) {
	m.modifiers = append(m.modifiers, mod)
}

type rChan struct {
	modifiers []string
	val       typeRef
}

func (c *rChan) modifier(m string) {
	c.modifiers = append(c.modifiers, m)
}

func (t *rType) typeRef() {}
func (m *rMap) typeRef()  {}
func (c *rChan) typeRef() {}

type typeRefVisitor struct {
	modifiers []string
	typ       rType
	out       typeRef
}

func (t *typeRefVisitor) modifier(m string) {
	t.modifiers = append(t.modifiers, m)
}

func parseTypeRef(e ast.Expr) (typeRef, error) {
	t := new(typeRefVisitor)
	err := t.visitExpr(e)
	if err != nil {
		return nil, err
	}

	var ret typeRef

	if t.out != nil {
		ret = t.out
	} else {
		ret = &t.typ
	}

	return ret, nil
}

func (t *typeRefVisitor) visitExpr(e ast.Expr) error {
	switch ee := e.(type) {
	case *ast.Ident:
		t.typ.path = append(t.typ.path, ee.Name)
		t.typ.modifiers = append(t.typ.modifiers, t.modifiers...)
		t.modifiers = nil
	case *ast.Ellipsis:
		t.modifier("...")
		if err := t.visitExpr(ee.Elt); err != nil {
			return err
		}
	case *ast.SelectorExpr:
		if err := t.visitExpr(ee.X); err != nil {
			return err
		}
		if err := t.visitExpr(ee.Sel); err != nil {
			return err
		}
	case *ast.IndexExpr:
		t.modifier("[T]")
		if err := t.visitExpr(ee.X); err != nil {
			return err
		}
	case *ast.IndexListExpr:
		for range ee.Indices {
			t.modifier("[T]")
		}
		if err := t.visitExpr(ee.X); err != nil {
			return err
		}
	case *ast.SliceExpr:
		t.modifier("[a:b]")
		if err := t.visitExpr(ee.X); err != nil {
			return err
		}
	case *ast.StarExpr:
		t.modifier("*")
		if err := t.visitExpr(ee.X); err != nil {
			return err
		}
	case *ast.ArrayType:
		t.modifier("[]")
		if err := t.visitExpr(ee.Elt); err != nil {
			return err
		}
	case *ast.MapType:
		k, err := parseTypeRef(ee.Key)
		if err != nil {
			return err
		}
		v, err := parseTypeRef(ee.Value)
		if err != nil {
			return err
		}
		m := &rMap{key: k, val: v}
		m.modifiers = append(m.modifiers, t.modifiers...)
		t.modifiers = nil
		t.out = m
	case *ast.ChanType:
		v, err := parseTypeRef(ee.Value)
		if err != nil {
			return err
		}
		c := &rChan{val: v}
		c.modifiers = append(c.modifiers, t.modifiers...)
		t.modifiers = nil
		t.out = c
	default:
		return errUnsupportedTypeRef
	}

	return nil
}

func (t *rType) resolve(p *param) error {
	if len(t.path) == 0 {
		return errUnsupportedType
	}

	if len(t.path) == 1 {
		typeName := t.path[0]
		obj := types.Universe.Lookup(typeName)
		if obj != nil {
			t.typeName = typeName
			return nil
		}

		t.pkgPath = p.pkgPath
		t.pkgName = p.pkgName
		t.typeName = typeName
		return nil
	}

	pkgName := t.path[0]
	pkgInfo, err := pkgInfoByName(p.imports, pkgName)
	if err != nil {
		return err
	}
	t.pkgPath = pkgInfo.pkgPath
	t.pkgName = pkgInfo.pkgName
	t.typeName = t.path[1]

	return nil
}

func (m *rMap) resolve(p *param) error {
	if err := m.key.resolve(p); err != nil {
		return err
	}
	if err := m.val.resolve(p); err != nil {
		return err
	}
	return nil
}

func (c *rChan) resolve(p *param) error {
	return c.val.resolve(p)
}

func (t *rType) equals(t2 typeRef) bool {
	switch r2 := t2.(type) {
	case *rType:
		if len(t.modifiers) != len(r2.modifiers) {
			return false
		}

		for i, m1 := range t.modifiers {
			if m1 != r2.modifiers[i] {
				return false
			}
		}

		if t.pkgPath != r2.pkgPath {
			return false
		}

		if t.pkgName != r2.pkgName {
			return false
		}

		if t.typeName != r2.typeName {
			return false
		}

		return true
	}

	return false
}

func (m *rMap) equals(ref2 typeRef) bool {
	switch r2 := ref2.(type) {
	case *rMap:
		return m.key.equals(r2.key) && m.val.equals(r2.val)
	}
	return false
}

func (c *rChan) equals(ref2 typeRef) bool {
	switch r2 := ref2.(type) {
	case *rChan:
		return c.val.equals(r2.val)
	}
	return false
}

func (t *rType) deref() typeRef {
	t2 := new(rType)

	for i, m := range t.modifiers {
		if i == 0 && m == "*" {
			continue
		}
		t2.modifiers = append(t2.modifiers, m)
	}

	t2.pkgPath = t.pkgPath
	t2.pkgName = t.pkgName
	t2.typeName = t.typeName

	return t2
}

func (m *rMap) deref() typeRef {
	m2 := new(rMap)

	for i, m := range m.modifiers {
		if i == 0 && m == "*" {
			continue
		}
		m2.modifiers = append(m2.modifiers, m)
	}

	m2.key = m.val.deref()
	m2.val = m.val.deref()

	return m2
}

func (c *rChan) deref() typeRef {
	c2 := new(rChan)

	for i, m := range c.modifiers {
		if i == 0 && m == "*" {
			continue
		}
		c2.modifiers = append(c2.modifiers, m)
	}

	c2.val = c.val.deref()

	return c2
}

func (t *rType) code(ctx Context) (string, []string) {
	var mods string
	for _, modif := range t.modifiers {
		if modif == "*" || modif == "[]" {
			mods += modif
		}
	}

	// builtin type
	if t.pkgName == "" {
		return fmt.Sprintf("%s%s", mods, t.typeName), nil
	}

	pkgInf, err := pkgInfoByName(ctx.Imports, t.pkgName)
	if err != nil {
		return fmt.Sprintf("%s%s", mods, t.typeName), nil
	}

	// type from current dir
	if pkgInf.dir == ctx.WorkDir {
		return fmt.Sprintf("%s%s", mods, t.typeName), nil
	}

	return fmt.Sprintf("%s%s.%s", mods, pkgInf.pkgName, t.typeName), []string{pkgInf.pkgPath}
}

func (m *rMap) code(ctx Context) (string, []string) {
	imports := map[string]struct{}{}

	kCode, kImps := m.key.code(ctx)
	for _, imp := range kImps {
		imports[imp] = struct{}{}
	}

	vCode, vImps := m.val.code(ctx)
	for _, imp := range vImps {
		imports[imp] = struct{}{}
	}

	var outImports []string
	for k := range imports {
		outImports = append(outImports, k)
	}

	var mods string
	for _, modif := range m.modifiers {
		if modif == "*" || modif == "[]" {
			mods += modif
		}
	}

	return fmt.Sprintf("%smap[%s]%s", mods, kCode, vCode), outImports
}

func (c *rChan) code(ctx Context) (string, []string) {
	imports := map[string]struct{}{}

	vCode, vImps := c.val.code(ctx)
	for _, imp := range vImps {
		imports[imp] = struct{}{}
	}

	var outImports []string
	for k := range imports {
		outImports = append(outImports, k)
	}

	var mods string
	for _, modif := range c.modifiers {
		if modif == "*" || modif == "[]" {
			mods += modif
		}
	}

	return fmt.Sprintf("%schan %s", mods, vCode), outImports
}
