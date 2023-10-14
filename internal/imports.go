package internal

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strconv"

	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/ast/astutil"
)

func addImports(file string, imports []string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return err
	}

	var fileImports []string
	for _, spec := range f.Imports {
		impVal, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return err
		}
		fileImports = append(fileImports, impVal)
	}

	var toAdd []string
	for _, imp := range imports {
		if !slices.Contains(fileImports, imp) {
			toAdd = append(toAdd, imp)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	for _, a := range toAdd {
		astutil.AddImport(fset, f, a)
	}

	var output []byte
	buffer := bytes.NewBuffer(output)
	if err = printer.Fprint(buffer, fset, f); err != nil {
		return err
	}

	return os.WriteFile(file, buffer.Bytes(), os.ModePerm)
}
