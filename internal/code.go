package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

func NodeToString(n ast.Node) string {
	switch field := n.(type) {
	case *ast.Field:
		if len(field.Names) == 0 {
			return NodeToString(field.Type)
		}
		var names []string
		for _, fName := range field.Names {
			names = append(names, NodeToString(fName))
		}
		return fmt.Sprintf("%s %s", strings.Join(names, ", "), NodeToString(field.Type))
	case *ast.BlockStmt:
		// to handle empty function bodies within builtin.go
		if field == nil {
			return ""
		}
	}

	bb := new(bytes.Buffer)
	if err := printer.Fprint(bb, token.NewFileSet(), n); err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	res := bb.String()
	switch nn := n.(type) {
	case *ast.TypeSpec:
		suffix := ""
		_, isStruct := nn.Type.(*ast.StructType)
		if isStruct {
			suffix = "struct {...}"
		}
		_, isInterface := nn.Type.(*ast.InterfaceType)
		if isInterface {
			suffix = "interface {...}"
		}

		// don't trim body for typedef and type alias
		if isStruct || isInterface {
			body := NodeToString(nn.Type)
			res = ApplyFns(
				strings.TrimSuffix(res, body),
				strings.TrimSpace,
				func(x string) string {
					if suffix == "" {
						return x
					}
					return fmt.Sprintf("%s %s", x, suffix)
				},
			)
		}
	case *ast.FuncDecl:
		body := NodeToString(nn.Body)
		res = ApplyFns(strings.TrimSuffix(res, body), strings.TrimSpace)
	}

	return res
}
