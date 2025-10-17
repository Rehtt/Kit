// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CLIInfo 存储解析出的 CLI 信息
type CLIInfo struct {
	Commands     []*CommandInfo
	RootFlags    []*FlagInfo
	CommandName  string
	PackageName  string
	BinaryName   string
	MainFilePath string
}

// CommandInfo 命令信息
type CommandInfo struct {
	Name         string
	Use          string
	Instruction  string
	Flags        []*FlagInfo
	SubCommands  []*CommandInfo
	Hidden       bool
	VarName      string
	ParentVarRef string
}

// FlagInfo flag 信息
type FlagInfo struct {
	Short       string
	Long        string
	Type        string
	Usage       string
	DefValue    string
	ItemType    string
	SelectNodes []string
}

// Parser AST 解析器
type Parser struct {
	dir       string
	recursive bool
	verbose   bool
	fset      *token.FileSet
	varMap    map[string]*CommandInfo
}

// NewParser 创建解析器
func NewParser(dir string, recursive, verbose bool) *Parser {
	return &Parser{
		dir:       dir,
		recursive: recursive,
		verbose:   verbose,
		fset:      token.NewFileSet(),
		varMap:    make(map[string]*CommandInfo),
	}
}

// Parse 解析目录中的 Go 文件
func (p *Parser) Parse() (*CLIInfo, error) {
	info := &CLIInfo{
		Commands:  make([]*CommandInfo, 0),
		RootFlags: make([]*FlagInfo, 0),
	}

	err := filepath.Walk(p.dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			if !p.recursive && path != p.dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if p.verbose {
			fmt.Fprintf(os.Stderr, "解析文件: %s\n", path)
		}

		if err := p.parseFile(path, info); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "警告: 解析文件 %s 失败: %v\n", path, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	p.buildCommandRelations(info)

	return info, nil
}

// parseFile 解析单个 Go 文件
func (p *Parser) parseFile(path string, info *CLIInfo) error {
	file, err := parser.ParseFile(p.fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	if file.Name.Name == "main" {
		info.PackageName = "main"
		info.MainFilePath = path
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			p.parseAssignment(node, info)
		case *ast.CallExpr:
			p.parseCallExpr(node, info)
		}
		return true
	})

	return nil
}

// parseAssignment 解析赋值语句，查找 cli.NewCLI 调用
func (p *Parser) parseAssignment(stmt *ast.AssignStmt, info *CLIInfo) {
	for i, rhs := range stmt.Rhs {
		callExpr, ok := rhs.(*ast.CallExpr)
		if !ok {
			continue
		}

		if !p.isCliNewCLI(callExpr) {
			continue
		}

		var varName string
		if i < len(stmt.Lhs) {
			if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
				varName = ident.Name
			}
		}

		cmd := p.parseNewCLI(callExpr, varName)
		if cmd != nil {
			p.varMap[varName] = cmd
			if p.verbose {
				fmt.Fprintf(os.Stderr, "  发现命令: %s (变量: %s)\n", cmd.Use, varName)
			}
		}
	}
}

// parseCallExpr 解析函数调用表达式
func (p *Parser) parseCallExpr(callExpr *ast.CallExpr, info *CLIInfo) {
	if p.isAddCommand(callExpr) {
		p.parseAddCommand(callExpr, info)
		return
	}

	if flagInfo := p.parseFlagRegistration(callExpr); flagInfo != nil {
		if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok {
				if cmd, exists := p.varMap[x.Name]; exists {
					cmd.Flags = append(cmd.Flags, flagInfo)
					if p.verbose {
						fmt.Fprintf(os.Stderr, "    发现 flag: -%s/--%s (命令: %s)\n",
							flagInfo.Short, flagInfo.Long, cmd.Use)
					}
				}
			}
		}
	}
}

// isCliNewCLI 检查是否是 cli.NewCLI 调用
func (p *Parser) isCliNewCLI(callExpr *ast.CallExpr) bool {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "NewCLI" {
		return false
	}

	if x, ok := sel.X.(*ast.Ident); ok {
		return x.Name == "cli"
	}

	return false
}

// parseNewCLI 解析 NewCLI 调用参数
func (p *Parser) parseNewCLI(callExpr *ast.CallExpr, varName string) *CommandInfo {
	cmd := &CommandInfo{
		VarName:     varName,
		Flags:       make([]*FlagInfo, 0),
		SubCommands: make([]*CommandInfo, 0),
	}

	if len(callExpr.Args) >= 1 {
		if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
			cmd.Use = strings.Trim(lit.Value, `"`)
			cmd.Name = cmd.Use
		}
	}

	if len(callExpr.Args) >= 2 {
		if lit, ok := callExpr.Args[1].(*ast.BasicLit); ok {
			cmd.Instruction = strings.Trim(lit.Value, `"`)
		}
	}

	return cmd
}

// isAddCommand 检查是否是 AddCommand 调用
func (p *Parser) isAddCommand(callExpr *ast.CallExpr) bool {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "AddCommand"
}

// parseAddCommand 解析 AddCommand 调用
func (p *Parser) parseAddCommand(callExpr *ast.CallExpr, info *CLIInfo) {
	var parentVar string
	if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			parentVar = x.Name
		}
	}

	for _, arg := range callExpr.Args {
		var subCmdVar string

		if ident, ok := arg.(*ast.Ident); ok {
			subCmdVar = ident.Name
		}

		if subCmdVar != "" && parentVar != "" {
			if subCmd, exists := p.varMap[subCmdVar]; exists {
				subCmd.ParentVarRef = parentVar
			}
		}
	}
}

// parseFlagRegistration 解析 Flag 注册调用
func (p *Parser) parseFlagRegistration(callExpr *ast.CallExpr) *FlagInfo {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	methodName := sel.Sel.Name

	if !p.isFlagMethod(methodName) {
		return nil
	}

	flag := &FlagInfo{}

	switch {
	case strings.HasSuffix(methodName, "ShortLong"):
		flag.Type = p.extractFlagType(methodName)
		p.parseShortLongFlag(callExpr, flag)

	case strings.HasPrefix(methodName, "String"), strings.HasPrefix(methodName, "Int"),
		strings.HasPrefix(methodName, "Bool"), strings.HasPrefix(methodName, "Float"),
		strings.HasPrefix(methodName, "Uint"), strings.HasPrefix(methodName, "Duration"):
		flag.Type = p.extractFlagType(methodName)
		p.parseStandardFlag(callExpr, flag)

	default:
		return nil
	}

	return flag
}

// isFlagMethod 检查方法名是否是 Flag 注册方法
func (p *Parser) isFlagMethod(methodName string) bool {
	flagMethods := []string{
		"StringVar", "String", "IntVar", "Int", "BoolVar", "Bool",
		"Int64Var", "Int64", "UintVar", "Uint", "Uint64Var", "Uint64",
		"Float64Var", "Float64", "DurationVar", "Duration",
		"StringsVar", "Strings", "PasswordStringVar", "PasswordString",
		"StringVarShortLong", "StringShortLong",
		"IntVarShortLong", "IntShortLong",
		"BoolVarShortLong", "BoolShortLong",
		"Int64VarShortLong", "Int64ShortLong",
		"UintVarShortLong", "UintShortLong",
		"Uint64VarShortLong", "Uint64ShortLong",
		"Float64VarShortLong", "Float64ShortLong",
		"DurationVarShortLong", "DurationShortLong",
		"StringsVarShortLong", "StringsShortLong",
		"PasswordStringVarShortLong", "PasswordStringShortLong",
	}

	for _, m := range flagMethods {
		if methodName == m {
			return true
		}
	}
	return false
}

// extractFlagType 从方法名提取 flag 类型
func (p *Parser) extractFlagType(methodName string) string {
	methodName = strings.TrimSuffix(methodName, "Var")
	methodName = strings.TrimSuffix(methodName, "ShortLong")
	methodName = strings.ToLower(methodName)

	switch {
	case strings.Contains(methodName, "string"):
		return "string"
	case strings.Contains(methodName, "int64"):
		return "int64"
	case strings.Contains(methodName, "int"):
		return "int"
	case strings.Contains(methodName, "uint64"):
		return "uint64"
	case strings.Contains(methodName, "uint"):
		return "uint"
	case strings.Contains(methodName, "float"):
		return "float64"
	case strings.Contains(methodName, "bool"):
		return "bool"
	case strings.Contains(methodName, "duration"):
		return "duration"
	default:
		return "string"
	}
}

// parseShortLongFlag 解析 ShortLong 类型的 flag
func (p *Parser) parseShortLongFlag(callExpr *ast.CallExpr, flag *FlagInfo) {
	argOffset := 0
	if strings.Contains(callExpr.Fun.(*ast.SelectorExpr).Sel.Name, "Var") {
		argOffset = 1
	}

	if len(callExpr.Args) > argOffset {
		if lit, ok := callExpr.Args[argOffset].(*ast.BasicLit); ok {
			flag.Short = strings.Trim(lit.Value, `"`)
		}
	}

	if len(callExpr.Args) > argOffset+1 {
		if lit, ok := callExpr.Args[argOffset+1].(*ast.BasicLit); ok {
			flag.Long = strings.Trim(lit.Value, `"`)
		}
	}

	if len(callExpr.Args) > argOffset+2 {
		flag.DefValue = p.extractValue(callExpr.Args[argOffset+2])
	}

	if len(callExpr.Args) > argOffset+3 {
		if lit, ok := callExpr.Args[argOffset+3].(*ast.BasicLit); ok {
			flag.Usage = strings.Trim(lit.Value, `"`)
		}
	}

	if len(callExpr.Args) > argOffset+4 {
		p.parseFlagItem(callExpr.Args[argOffset+4:], flag)
	}
}

// parseStandardFlag 解析标准 flag
func (p *Parser) parseStandardFlag(callExpr *ast.CallExpr, flag *FlagInfo) {
	argOffset := 0
	if strings.Contains(callExpr.Fun.(*ast.SelectorExpr).Sel.Name, "Var") {
		argOffset = 1
	}

	if len(callExpr.Args) > argOffset {
		if lit, ok := callExpr.Args[argOffset].(*ast.BasicLit); ok {
			name := strings.Trim(lit.Value, `"`)
			if len(name) == 1 {
				flag.Short = name
			} else {
				flag.Long = name
			}
		}
	}

	if len(callExpr.Args) > argOffset+1 {
		flag.DefValue = p.extractValue(callExpr.Args[argOffset+1])
	}

	if len(callExpr.Args) > argOffset+2 {
		if lit, ok := callExpr.Args[argOffset+2].(*ast.BasicLit); ok {
			flag.Usage = strings.Trim(lit.Value, `"`)
		}
	}

	if len(callExpr.Args) > argOffset+3 {
		p.parseFlagItem(callExpr.Args[argOffset+3:], flag)
	}
}

// extractValue 从 AST 节点提取值
func (p *Parser) extractValue(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(v.Value, `"`)
	case *ast.Ident:
		if v.Name == "true" || v.Name == "false" {
			return v.Name
		}
		return v.Name
	default:
		return ""
	}
}

// parseFlagItem 解析 FlagItem 参数
func (p *Parser) parseFlagItem(args []ast.Expr, flag *FlagInfo) {
	for _, arg := range args {
		callExpr, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}

		var funcName string
		switch fun := callExpr.Fun.(type) {
		case *ast.Ident:
			funcName = fun.Name
		case *ast.SelectorExpr:
			funcName = fun.Sel.Name
		}

		switch funcName {
		case "NewFlagItemFile":
			flag.ItemType = "file"
		case "NewFlagItemDir":
			flag.ItemType = "dir"
		case "NewFlagItemSelect", "NewFlagItemSelectString":
			flag.ItemType = "select"
			flag.SelectNodes = p.parseSelectNodes(callExpr)
		}
	}
}

// parseSelectNodes 解析 select 类型的选项
func (p *Parser) parseSelectNodes(callExpr *ast.CallExpr) []string {
	var nodes []string

	for _, arg := range callExpr.Args {
		switch v := arg.(type) {
		case *ast.BasicLit:
			nodes = append(nodes, strings.Trim(v.Value, `"`))

		case *ast.CompositeLit:
			for _, elt := range v.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Value" {
						if lit, ok := kv.Value.(*ast.BasicLit); ok {
							nodes = append(nodes, strings.Trim(lit.Value, `"`))
						}
					}
				}
			}
		}
	}

	return nodes
}

// buildCommandRelations 建立命令之间的父子关系
func (p *Parser) buildCommandRelations(info *CLIInfo) {
	for _, cmd := range p.varMap {
		if cmd.ParentVarRef == "" {
			info.Commands = append(info.Commands, cmd)
			if cmd.VarName == "CommandLine" || cmd.VarName == "app" || cmd.VarName == "rootCmd" {
				info.RootFlags = cmd.Flags
			}
		}
	}

	for _, cmd := range p.varMap {
		if cmd.ParentVarRef != "" {
			if parent, exists := p.varMap[cmd.ParentVarRef]; exists {
				parent.SubCommands = append(parent.SubCommands, cmd)
			}
		}
	}
}

// InferCommandName 从代码信息推断命令名称
func (info *CLIInfo) InferCommandName() string {
	if len(info.Commands) > 0 {
		for _, cmd := range info.Commands {
			if cmd.Use != "" && cmd.Use != "CommandLine" {
				return cmd.Use
			}
		}
	}

	if info.MainFilePath != "" {
		dir := filepath.Dir(info.MainFilePath)
		return filepath.Base(dir)
	}

	return "app"
}
