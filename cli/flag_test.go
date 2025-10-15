package cli

import (
	"flag"
	"testing"
)

func TestFlagSet_PasswordStringVar(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var password string
	fs.PasswordStringVar(&password, "password", "default123", "用户密码")

	// 测试解析参数
	err := fs.Parse([]string{"-password", "newpass"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证密码被正确设置
	if password != "newpass" {
		t.Errorf("期望密码为 'newpass'，但得到 %q", password)
	}
}

func TestFlagSet_PasswordString(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	password := fs.PasswordString("pass", "secret", "密码", 5)

	// 验证默认值
	if *password != "secret" {
		t.Errorf("期望默认值为 'secret'，但得到 %q", *password)
	}

	// 测试解析参数
	err := fs.Parse([]string{"-pass", "mypassword"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证密码被正确更新
	if *password != "mypassword" {
		t.Errorf("期望密码为 'mypassword'，但得到 %q", *password)
	}
}

func TestFlagSet_StringsVar(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var values []string
	fs.StringsVar(&values, "value", []string{"default"}, "值列表")

	// 验证默认值
	if len(values) != 1 || values[0] != "default" {
		t.Errorf("期望默认值为 ['default']，但得到 %v", values)
	}

	// 测试解析参数
	err := fs.Parse([]string{"-value", "a", "-value", "b", "-value", "c"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果（包含默认值）
	expected := []string{"default", "a", "b", "c"}
	if len(values) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(values))
	}

	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestFlagSet_Strings(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	values := fs.Strings("item", []string{}, "项目列表")

	// 验证初始为空
	if len(*values) != 0 {
		t.Errorf("期望初始为空，但得到 %d 个元素", len(*values))
	}

	// 测试解析参数
	err := fs.Parse([]string{"-item", "apple", "-item", "banana"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
	expected := []string{"apple", "banana"}
	if len(*values) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(*values))
	}

	for i, v := range *values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestFlagSet_Strings_EqualsSyntax(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	values := fs.Strings("opt", []string{}, "选项")

	// 测试使用 = 号语法
	err := fs.Parse([]string{"-opt=value1", "-opt=value2"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
	expected := []string{"value1", "value2"}
	if len(*values) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(*values))
	}

	for i, v := range *values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestFlagSet_Strings_MixedWithOtherFlags(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var count int
	var verbose bool
	files := fs.Strings("file", []string{}, "文件列表")
	fs.IntVar(&count, "count", 0, "数量")
	fs.BoolVar(&verbose, "v", false, "详细模式")

	// 混合使用各种类型的 flag
	err := fs.Parse([]string{
		"-file", "a.txt",
		"-count", "10",
		"-file", "b.txt",
		"-v",
		"-file", "c.txt",
	})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证 files
	expectedFiles := []string{"a.txt", "b.txt", "c.txt"}
	if len(*files) != len(expectedFiles) {
		t.Errorf("files 期望长度为 %d，但得到 %d", len(expectedFiles), len(*files))
	}
	for i, v := range *files {
		if v != expectedFiles[i] {
			t.Errorf("files[%d] = %q，期望 %q", i, v, expectedFiles[i])
		}
	}

	// 验证 count
	if count != 10 {
		t.Errorf("count 期望为 10，但得到 %d", count)
	}

	// 验证 verbose
	if !verbose {
		t.Error("verbose 期望为 true")
	}
}

func TestFlagSet_Strings_EmptyStrings(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	values := fs.Strings("value", []string{}, "值")

	// 测试传递空字符串
	err := fs.Parse([]string{"-value", "", "-value", "nonempty", "-value", ""})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果（空字符串也应该被接受）
	expected := []string{"", "nonempty", ""}
	if len(*values) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(*values))
	}

	for i, v := range *values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestFlagSet_Strings_SpecialCharacters(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	values := fs.Strings("arg", []string{}, "参数")

	// 测试包含特殊字符的字符串
	err := fs.Parse([]string{
		"-arg", "hello world",
		"-arg", "path/to/file",
		"-arg", "key=value",
		"-arg", "中文字符",
	})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
	expected := []string{"hello world", "path/to/file", "key=value", "中文字符"}
	if len(*values) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(*values))
	}

	for i, v := range *values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestFlagSet_MultipleStringsFlagsIndependent(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	sources := fs.Strings("src", []string{}, "源文件")
	targets := fs.Strings("dst", []string{}, "目标文件")

	// 测试多个独立的字符串数组 flag
	err := fs.Parse([]string{
		"-src", "file1.go",
		"-dst", "out1.bin",
		"-src", "file2.go",
		"-dst", "out2.bin",
	})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证 sources
	expectedSrc := []string{"file1.go", "file2.go"}
	if len(*sources) != len(expectedSrc) {
		t.Errorf("sources 期望长度为 %d，但得到 %d", len(expectedSrc), len(*sources))
	}
	for i, v := range *sources {
		if v != expectedSrc[i] {
			t.Errorf("sources[%d] = %q，期望 %q", i, v, expectedSrc[i])
		}
	}

	// 验证 targets
	expectedDst := []string{"out1.bin", "out2.bin"}
	if len(*targets) != len(expectedDst) {
		t.Errorf("targets 期望长度为 %d，但得到 %d", len(expectedDst), len(*targets))
	}
	for i, v := range *targets {
		if v != expectedDst[i] {
			t.Errorf("targets[%d] = %q，期望 %q", i, v, expectedDst[i])
		}
	}
}
