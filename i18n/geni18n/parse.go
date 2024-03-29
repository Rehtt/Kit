package main

import (
	"bytes"
	strings2 "github.com/Rehtt/Kit/strings"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Parse(path string) (map[string]string, error) {
	parse := make(map[string]string)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_gen.go") {
			return nil
		}
		return parseFile(path, parse)
	})
	if err != nil {
		return nil, err
	}
	return parse, nil
}
func parseFile(filePath string, m map[string]string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var (
		fset    = token.NewFileSet()
		i18nKey string
	)
	p, err := parser.ParseFile(fset, "", fileData, parser.AllErrors)
	if err != nil {
		return err
	}
	for _, imp := range p.Imports {
		if strings.Contains(imp.Path.Value, "\"github.com/Rehtt/Kit/i18n\"") {
			i18nKey = "i18n"
			if imp.Name != nil {
				i18nKey = imp.Name.String()
			}
		}
	}
	if i18nKey == "" {
		return nil
	}

	for _, decl := range p.Decls {
		var tmp bytes.Buffer
		ast.Fprint(&tmp, fset, decl, ast.NotNilFilter)
		if strings.Contains(tmp.String(), "Name: \""+i18nKey+"\"") && strings.Contains(tmp.String(), "Name: \"GetText\"") {
			s := strings.Split(tmp.String(), "\n")
			for i := 0; i < len(s); i++ {
				if strings.HasSuffix(s[i], "Name: \""+i18nKey+"\"") && i+12 < len(s) {
					if strings.HasSuffix(s[i+4], "Name: \"GetText\"") && strings.Contains(s[i+12], "Value: \"\\\"") {
						values := strings.Split(s[i+12], "\\\"")
						v := strings.Join(values[1:len(values)-1], "\\\"")
						v = strings2.EscapeStringRepeat(v, 2, true)
						m[v] = v
					}
				}
			}
		}
	}
	return nil
}
