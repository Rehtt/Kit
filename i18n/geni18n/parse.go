// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Parse(path string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_gen.go") {
			return nil
		}
		return parseFile(path, result)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseFile(filePath string, m map[string]string) error {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 解析文件失败 %s: %v\n", filePath, err)
		return nil
	}

	i18nPkgName := findI18nImportName(file)
	if i18nPkgName == "" {
		return nil
	}
	ast.Inspect(file, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != i18nPkgName {
			return true
		}

		if selExpr.Sel.Name != "GetText" {
			return true
		}

		if len(callExpr.Args) > 0 {
			if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				value, err := strconv.Unquote(lit.Value)
				if err == nil {
					m[value] = value
				}
			}
		}

		return true
	})

	return nil
}

func findI18nImportName(file *ast.File) string {
	for _, imp := range file.Imports {
		if imp.Path.Value == `"github.com/Rehtt/Kit/i18n"` {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return "i18n"
		}
	}
	return ""
}
