package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sabahtalateh/mod"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

var (
	errImportDirNotFound = errors.New("directory for import not found")
)

type module struct {
	root       string
	modFile    *modfile.File
	discovered map[string]packageInfo
}

type packageInfo struct {
	pkgPath string
	pkgName string
	dir     string
}

var moduleG *module

func InitMod(workDir string) error {
	var err error

	moduleG, err = initMod(workDir)
	if err != nil {
		return err
	}

	return nil
}

func initMod(dir string) (*module, error) {
	modFilePath, err := mod.ModFilePath(dir)
	if errors.Is(err, mod.ErrNotFound) {
		return nil, errors.Join(fmt.Errorf("not resides within module: %s", dir), err)
	}
	if err != nil {
		return nil, err
	}

	bb, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("malformed go.mod: %s", modFilePath), err)
	}

	modF, err := modfile.Parse("go.mod", bb, nil)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("malformed go.mod: %s", modFilePath), err)
	}

	return &module{root: filepath.Dir(modFilePath), modFile: modF, discovered: map[string]packageInfo{}}, nil
}

func pkgInfoByName(imports []*ast.ImportSpec, alias string) (packageInfo, error) {
	if moduleG == nil {
		return packageInfo{}, errors.New("module not initialized. call internal.InitMod")
	}

	for _, imp := range imports {
		val, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return packageInfo{}, err
		}

		pkgInfo, err := getPackageInfo(val)
		if err != nil {
			return packageInfo{}, err
		}

		if pkgInfo.pkgName == alias {
			return pkgInfo, nil
		}
	}

	return packageInfo{}, errors.Join(errors.New(alias), errImportDirNotFound)
}

func dirPkgPath(dir string) (string, error) {
	if moduleG == nil {
		return "", errors.New("module not initialized. call internal.InitMod")
	}

	if !strings.HasPrefix(dir, moduleG.root) {
		return "", fmt.Errorf("not resides within module: %s", dir)
	}

	return strings.Replace(dir, moduleG.root, moduleG.modFile.Module.Mod.Path, 1), nil
}

func getPackageInfo(pkgPath string) (packageInfo, error) {
	pkgs, err := packages.Load(nil, pkgPath)
	if err != nil {
		return packageInfo{}, err
	}
	if len(pkgs) == 0 {
		return packageInfo{}, fmt.Errorf("malformed import: %s", pkgPath)
	}
	pkg := pkgs[0]
	if len(pkg.Errors) != 0 {
		return packageInfo{}, errors.Join(fmt.Errorf("malformed import: %s", pkgPath), pkg.Errors[0])
	}
	inf := packageInfo{pkgName: pkg.Name, pkgPath: pkg.ID, dir: filepath.Dir(pkg.GoFiles[0])}
	moduleG.discovered[pkgPath] = inf
	return inf, nil
}
