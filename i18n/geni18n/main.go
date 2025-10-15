package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Rehtt/Kit/cli"
)

var (
	sourcePath  string
	outputDir   string
	outputFile  string
	indent      bool
	verbose     bool
	outputType  string // "json" 或 "go"
	packageName string // Go文件的包名
)

func main() {
	app := cli.NewCLI("geni18n", "从 Go 源代码中提取 i18n.GetText() 调用的字符串并生成 JSON 或 Go 文件")
	app.Usage = "[选项] [源路径]"

	app.StringVarShortLong(&sourcePath, "p", "path", ".", "源代码路径")
	app.StringVarShortLong(&outputDir, "d", "output-dir", "i18n", "输出目录")
	app.StringVarShortLong(&outputFile, "f", "output-file", "default.json", "输出文件名")
	app.StringVarShortLong(&outputType, "t", "type", "json", "输出类型: json 或 go")
	app.StringVarShortLong(&packageName, "", "package", "i18n", "Go文件的包名 (仅当type=go时使用)")
	app.BoolVarShortLong(&indent, "i", "indent", true, "格式化输出 JSON")
	app.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
	app.CommandFunc = func(args []string) error {
		if len(args) > 0 {
			sourcePath = args[0]
		}

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return fmt.Errorf("源路径不存在: %s", sourcePath)
		}

		// 验证输出类型
		if outputType != "json" && outputType != "go" {
			return fmt.Errorf("不支持的输出类型: %s，仅支持 json 或 go", outputType)
		}

		// 如果是Go文件模式，调整输出文件扩展名
		if outputType == "go" && filepath.Ext(outputFile) != ".go" {
			outputFile = filepath.Base(outputFile)
			if filepath.Ext(outputFile) == ".json" {
				outputFile = outputFile[:len(outputFile)-5] // 移除.json
			}
			outputFile += ".go"
		}

		if verbose {
			fmt.Printf("配置信息:\n")
			fmt.Printf("  源路径: %s\n", sourcePath)
			fmt.Printf("  输出目录: %s\n", outputDir)
			fmt.Printf("  输出文件: %s\n", outputFile)
			fmt.Printf("  输出类型: %s\n", outputType)
			if outputType == "go" {
				fmt.Printf("  包名: %s\n", packageName)
			}
			fmt.Printf("  格式化: %t\n", indent)
			fmt.Printf("\n")
		}

		fmt.Printf("正在扫描: %s\n", sourcePath)
		values, err := Parse(sourcePath)
		if err != nil {
			return fmt.Errorf("解析错误: %v", err)
		}

		if len(values) == 0 {
			fmt.Println("警告: 未找到任何 i18n.GetText() 调用")
			return nil
		}

		fmt.Printf("找到 %d 个翻译键\n", len(values))

		if verbose {
			fmt.Println("\n找到的翻译键:")
			for key := range values {
				fmt.Printf("  - %q\n", key)
			}
			fmt.Println()
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		outputPath := filepath.Join(outputDir, outputFile)

		var fileData []byte
		if outputType == "go" {
			fileData, err = generateGoFile(values, packageName)
			if err != nil {
				return fmt.Errorf("生成Go文件错误: %v", err)
			}
		} else {
			// JSON模式
			if indent {
				fileData, err = json.MarshalIndent(values, "", "  ")
			} else {
				fileData, err = json.Marshal(values)
			}
			if err != nil {
				return fmt.Errorf("JSON 序列化错误: %v", err)
			}
		}

		if err := os.WriteFile(outputPath, fileData, 0o644); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		fmt.Printf("成功生成: %s\n", outputPath)
		if verbose {
			fmt.Printf("文件大小: %d 字节\n", len(fileData))
		}

		return nil
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

// generateGoFile 生成Go文件内容
func generateGoFile(values map[string]string, packageName string) ([]byte, error) {
	var content strings.Builder

	// 包声明
	content.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// 导入语句
	content.WriteString("import (\n")
	content.WriteString("\t\"github.com/Rehtt/Kit/i18n\"\n")
	content.WriteString(")\n\n")

	// 生成变量名（基于文件名）
	varName := generateVarName(outputFile)

	// 生成map变量
	content.WriteString(fmt.Sprintf("var %s = map[string]string{\n", varName))

	// 按键名排序以保证输出一致性
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}

	// 排序以保证输出一致性
	sort.Strings(keys)

	// 生成键值对
	for _, key := range keys {
		value := values[key]
		content.WriteString(fmt.Sprintf("\t%q: %q,\n", key, value))
	}

	content.WriteString("}\n\n")

	// 生成init函数进行注册
	langKey := generateLangKey(outputFile)

	content.WriteString("func init() {\n")
	content.WriteString(fmt.Sprintf("\ti18n.RegisterGoTexts(%q, %s)\n", langKey, varName))
	content.WriteString("\ti18n.SetLangByLocalEnv()\n")
	content.WriteString("}\n")

	return []byte(content.String()), nil
}

func generateVarName(filename string) string {
	if filename == "" {
		return "texts"
	}
	base := filepath.Base(filename)
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base + "Texts"
}

func generateLangKey(filename string) string {
	if filename == "" {
		return "default"
	}
	base := filepath.Base(filename)
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	if base == "default" {
		return "default"
	}
	return base
}
