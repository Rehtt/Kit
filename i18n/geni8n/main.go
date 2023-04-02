package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	flag.Parse()
	path := os.Args[0]
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	values, err := Parse(path)
	if err != nil {
		log.Fatalln(err)
	}
	template, _ := json.Marshal(values)

	os.MkdirAll("i18n", 777)
	os.WriteFile("i18n/default.json", template, 644)

	//f := token.NewFileSet()
	//var list = make([]*ast.Field, 0, len(values))
	//for _, v := range values {
	//
	//	list = append(list, &ast.Field{
	//		Names: []*ast.Ident{
	//			ast.NewIdent(v),
	//		},
	//		Type: ast.NewIdent("string"),
	//		Tag: &ast.BasicLit{
	//			Kind:  token.STRING,
	//			Value: fmt.Sprintf("`json:\"%s\"`", v),
	//		},
	//	})
	//}
	//var file = &ast.File{
	//	Name:  ast.NewIdent("i18n"),
	//	Scope: ast.NewScope(nil),
	//	Decls: []ast.Decl{
	//		ast.Decl(&ast.GenDecl{
	//			Tok: token.TYPE,
	//			Specs: []ast.Spec{
	//				ast.Spec(&ast.TypeSpec{
	//					Name: ast.NewIdent("Text"),
	//					Type: &ast.StructType{
	//						Fields: &ast.FieldList{
	//							List: list,
	//						},
	//					},
	//				}),
	//			},
	//		}),
	//	},
	//}
	//var tmp bytes.Buffer
	//format.Node(&tmp, f, file)
	//os.WriteFile("test.go", tmp.Bytes(), 644)

}
