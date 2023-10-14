package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
)

func updateFile(ctx Context, pp *params) error {
	var (
		err     error
		imports []string
	)

	err = checkSlice(pp)
	if err != nil {
		return err
	}

	funcBodyLines, err := funcBody(ctx, pp)
	if err != nil {
		return err
	}

	bb, err := os.ReadFile(ctx.Location.File)
	if err != nil {
		return err
	}

	fileLines := strings.Split(string(bb), "\n")

	if pp.createOut {
		var added int
		fileLines, imports, added = insertOutStructure(ctx, fileLines, ctx.Location.Line-1, pp)
		pp.addPosition(added)
		ctx = ctx.AddLocationLine(added)
	}

	// delete `//go:generate convert ..`
	fileLines = delLine(fileLines, ctx.Location.Line-1)
	pp.addPosition(-1)
	ctx = ctx.AddLocationLine(-1)

	// delete existing function body
	fileLines = delFuncBody(fileLines, pp.funcBodyStart, pp.funcBodyEnd)

	// insert new function body
	fileLines = insertFuncBody(fileLines, funcBodyLines, pp.funcBodyStart.Line)

	err = os.WriteFile(ctx.Location.File, []byte(strings.Join(fileLines, "\n")), os.ModePerm)
	if err != nil {
		return err
	}

	return addImports(ctx.Location.File, imports)
}

func funcBody(ctx Context, pp *params) ([]string, error) {
	var (
		lines []string
	)

	outP := pp.outP

	if outP.array {
		outArrVar := "var out []"
		if outP.pointer {
			outArrVar += "*"
		}
		if ctx.WorkDir != outP.dir {
			outArrVar += outP.pkgName + "."
		}
		outArrVar += outP.typeName
		lines = append(lines, outArrVar)
		lines = append(lines, "")

		cycle := fmt.Sprintf("for _, x := range %s {", pp.inP.varName)
		lines = append(lines, cycle)

		cycleVar := "\tvar out2 "
		if outP.pointer {
			cycleVar += "*"
		}
		if ctx.WorkDir != outP.dir {
			cycleVar += outP.pkgName + "."
		}
		cycleVar += outP.typeName
		lines = append(lines, cycleVar)

		fieldsLines, err := outFields("x", "out2", "\t", pp)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fieldsLines...)

		lines = append(lines, "")
		lines = append(lines, "\tout = append(out, out2)")

		lines = append(lines, "}")
		lines = append(lines, "")
		lines = append(lines, "return out")
	} else {
		outVar := "var out "
		if outP.pointer {
			outVar += "*"
		}
		if ctx.WorkDir != outP.dir {
			outVar += outP.pkgName + "."
		}
		outVar += outP.typeName
		lines = append(lines, outVar)
		lines = append(lines, "")

		fieldsLines, err := outFields(pp.inP.varName, "out", "", pp)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fieldsLines...)

		lines = append(lines, "")
		lines = append(lines, "return out")
	}

	return lines, nil
}

func insertFuncBody(fileLines, bodyLines []string, atLine int) []string {
	head := fileLines[:atLine]

	var tail []string
	for i := atLine; i < len(fileLines); i++ {
		tail = append(tail, fileLines[i])
	}

	for i := 0; i < len(bodyLines); i++ {
		bodyLines[i] = fmt.Sprintf("\t%s", bodyLines[i])
	}

	out := append(head, bodyLines...)
	out = append(out, tail...)

	return out
}

func delFuncBody(lines []string, start token.Position, end token.Position) []string {
	var out []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineNo := i + 1

		if lineNo < start.Line {
			out = append(out, line)
			continue
		}

		if lineNo > end.Line {
			out = append(out, line)
			continue
		}

		if lineNo == start.Line {
			out = append(out, line[:strings.Index(line, "{")+1])
		}

		if lineNo == end.Line {
			out = append(out, "}")
		}
	}

	return out
}

func insertOutStructure(ctx Context, fileLines []string, beforeLine int, pp *params) ([]string, []string, int) {
	var (
		StructLines []string
		imports     = map[string]struct{}{}
	)

	StructLines = append(StructLines, fmt.Sprintf("type %s struct {", pp.outP.typeName))

	var fields []*ast.Field

	if pp.inP.typ.Fields.List != nil {
		fields = pp.inP.typ.Fields.List
	}

	for _, field := range fields {
		resolved, err := resolveTypeRef(pp.inP, field.Type)
		if err != nil {
			continue
		}

		var names []string
		for _, name := range field.Names {
			if !ast.IsExported(name.Name) {
				continue
			}
			names = append(names, name.Name)
		}

		if len(field.Names) != 0 && len(names) == 0 {
			continue
		}

		selfImport := &ast.ImportSpec{Name: &ast.Ident{Name: pp.inP.pkgName}, Path: &ast.BasicLit{Value: pp.inP.pkgPath}}
		fieldCode, fieldImports := resolved.code(ctx.WithImports(append(pp.inP.imports, selfImport)))
		for _, fImp := range fieldImports {
			imports[fImp] = struct{}{}
		}
		StructLines = append(StructLines, fmt.Sprintf("\t%s %s", strings.Join(names, ","), fieldCode))
	}

	StructLines = append(StructLines, "}")
	StructLines = append(StructLines, "")

	head := fileLines[:beforeLine]

	var tail []string
	for i := beforeLine; i < len(fileLines); i++ {
		tail = append(tail, fileLines[i])
	}

	fileLines = append(head, StructLines...)
	fileLines = append(fileLines, tail...)

	var importOut []string
	for k := range imports {
		importOut = append(importOut, k)
	}

	return fileLines, importOut, len(StructLines)
}

func outFields(inVar, outVar, tabs string, pp *params) ([]string, error) {
	var lines []string

	var fields []*ast.Field
	if pp.createOut {
		if pp.inP.typ.Fields.List != nil {
			fields = pp.inP.typ.Fields.List
			for i, field := range fields {
				var exportedNames []*ast.Ident
				for _, name := range field.Names {
					if ast.IsExported(name.Name) {
						exportedNames = append(exportedNames, name)
					}
				}
				fields[i].Names = exportedNames
			}
		}
	} else {
		if pp.outP.typ.Fields.List != nil {
			fields = pp.outP.typ.Fields.List
		}
	}

	for _, field := range fields {
		ll, err := outField(inVar, outVar, tabs, field, pp)
		if err != nil {
			return nil, err
		}
		lines = append(lines, ll...)
	}

	return lines, nil
}

func outField(inVar, outVar, tabs string, outF *ast.Field, pp *params) ([]string, error) {
	if len(outF.Names) == 0 {
		return nil, nil
	}

	var lines []string

	for _, outFieldName := range outF.Names {
		if !ast.IsExported(outFieldName.Name) {
			continue
		}

		inFieldType, ok := findFieldType(pp.inP.typ, outFieldName.Name)
		if !ok {
			lines = append(lines, fmt.Sprintf("%s// no field: %s.%s", tabs, pp.inP.typeName, outFieldName))
			continue
		}

		resolvedInTypeRef, err := resolveTypeRef(pp.inP, inFieldType)
		if err != nil {
			lines = append(lines, fmt.Sprintf("%s// unsupported field type: %s.%s %s",
				tabs, pp.outP.typeName, outFieldName, NodeToString(inFieldType)))
			continue
		}

		var resolvedOutTypeRef typeRef
		if pp.createOut {
			resolvedOutTypeRef, err = resolveTypeRef(pp.inP, outF.Type)
			if err != nil {
				lines = append(lines, fmt.Sprintf("%s// unsupported field type: %s.%s %s",
					tabs, pp.inP.typeName, outFieldName, NodeToString(outF.Type)))
				continue
			}
		} else {
			resolvedOutTypeRef, err = resolveTypeRef(pp.outP, outF.Type)
			if err != nil {
				lines = append(lines, fmt.Sprintf("%s// unsupported field type: %s.%s %s",
					tabs, pp.outP.typeName, outFieldName, NodeToString(outF.Type)))
				continue
			}
		}

		if resolvedInTypeRef.equals(resolvedOutTypeRef) {
			lines = append(lines, fmt.Sprintf("%s%s.%s = %s.%s", tabs, outVar, outFieldName, inVar, outFieldName))
			continue
		}

		if resolvedInTypeRef.deref().equals(resolvedOutTypeRef) {
			lines = append(lines, fmt.Sprintf("%s%s.%s = *%s.%s", tabs, outVar, outFieldName, inVar, outFieldName))
			continue
		}

		if resolvedInTypeRef.equals(resolvedOutTypeRef.deref()) {
			lines = append(lines, fmt.Sprintf("%s%s.%s = &%s.%s", tabs, outVar, outFieldName, inVar, outFieldName))
			continue
		}

		lines = append(lines,
			fmt.Sprintf("%s// %s.%s = %s.%s not assignable", tabs, outVar, outFieldName, inVar, outFieldName))
	}

	return lines, nil
}

func findFieldType(Struct *ast.StructType, name string) (ast.Expr, bool) {
	if Struct.Fields == nil || Struct.Fields.List == nil {
		println(123)
		return nil, false
	}
	for _, f := range Struct.Fields.List {
		for _, n := range f.Names {
			if n.Name == name {
				return f.Type, true
			}
		}
	}
	return nil, false
}

func checkSlice(p *params) error {
	check := 0
	if p.inP.array {
		check += 1
	}
	if p.outP.array {
		check -= 1
	}
	if check != 0 {
		return errors.Join(
			fmt.Errorf("%s, %s", NodeToString(p.inP.field), NodeToString(p.outP.field)),
			errors.New("parameter and return types should be both slices or both not slices"),
		)
	}

	return nil
}

func delLine(lines []string, lineNo int) []string {
	return append(lines[:lineNo], lines[lineNo+1:]...)
}
