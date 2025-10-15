package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rehtt/Kit/cli"
)

var (
	sourcePath string
	outputDir  string
	outputFile string
	indent     bool
	verbose    bool
)

func main() {
	app := cli.NewCLI("geni18n", "从 Go 源代码中提取 i18n.GetText() 调用的字符串并生成 JSON 文件")
	app.Usage = "[选项] [源路径]"

	app.StringVarShortLong(&sourcePath, "p", "path", ".", "源代码路径")
	app.StringVarShortLong(&outputDir, "d", "output-dir", "i18n", "输出目录")
	app.StringVarShortLong(&outputFile, "f", "output-file", "default.json", "输出文件名")
	app.BoolVarShortLong(&indent, "i", "indent", true, "格式化输出 JSON")
	app.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
	app.CommandFunc = func(args []string) error {
		if len(args) > 0 {
			sourcePath = args[0]
		}

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return fmt.Errorf("源路径不存在: %s", sourcePath)
		}

		if verbose {
			fmt.Printf("配置信息:\n")
			fmt.Printf("  源路径: %s\n", sourcePath)
			fmt.Printf("  输出目录: %s\n", outputDir)
			fmt.Printf("  输出文件: %s\n", outputFile)
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

		var jsonData []byte
		if indent {
			jsonData, err = json.MarshalIndent(values, "", "  ")
		} else {
			jsonData, err = json.Marshal(values)
		}
		if err != nil {
			return fmt.Errorf("JSON 序列化错误: %v", err)
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		outputPath := filepath.Join(outputDir, outputFile)
		if err := os.WriteFile(outputPath, jsonData, 0o644); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		fmt.Printf("成功生成: %s\n", outputPath)
		if verbose {
			fmt.Printf("文件大小: %d 字节\n", len(jsonData))
		}

		return nil
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
