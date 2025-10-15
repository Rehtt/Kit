package cli

import (
	"bytes"
	"flag"
	"strings"
	"testing"
	"time"
)

func TestFlagSet_PasswordString(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	password := fs.PasswordString("pass", "secret", "密码", 5)

	if *password != "secret" {
		t.Errorf("期望默认值为 'secret'，但得到 %q", *password)
	}

	err := fs.Parse([]string{"-pass", "mypassword"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if *password != "mypassword" {
		t.Errorf("期望密码为 'mypassword'，但得到 %q", *password)
	}
}

func TestFlagSet_Strings(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	values := fs.Strings("item", []string{}, "项目列表")

	err := fs.Parse([]string{"-item", "apple", "-item", "banana"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

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

func TestFlagSet_Alias(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var config string
	fs.StringVar(&config, "config", "", "配置文件")
	fs.Alias("c", "config")

	err := fs.Parse([]string{"-c", "app.json"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if config != "app.json" {
		t.Errorf("期望 config = 'app.json'，但得到 %q", config)
	}
}

func TestFlagSet_ShortLong(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var host string
	var port int
	var verbose bool
	fs.StringVarShortLong(&host, "h", "host", "localhost", "主机地址")
	fs.IntVarShortLong(&port, "p", "port", 8080, "端口号")
	fs.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")

	err := fs.Parse([]string{"-h", "127.0.0.1", "--port", "9000", "-v"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if host != "127.0.0.1" {
		t.Errorf("期望 host = '127.0.0.1'，但得到 %q", host)
	}
	if port != 9000 {
		t.Errorf("期望 port = 9000，但得到 %d", port)
	}
	if !verbose {
		t.Error("期望 verbose = true")
	}
}

func TestFlagSet_AllNativeTypes(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var (
		str string
		i   int
		i64 int64
		u   uint
		u64 uint64
		f64 float64
		dur time.Duration
		b   bool
	)

	fs.StringVarShortLong(&str, "s", "string", "", "字符串")
	fs.IntVarShortLong(&i, "i", "int", 0, "整数")
	fs.Int64VarShortLong(&i64, "l", "long", 0, "长整数")
	fs.UintVarShortLong(&u, "u", "uint", 0, "无符号整数")
	fs.Uint64VarShortLong(&u64, "U", "uint64", 0, "64位无符号整数")
	fs.Float64VarShortLong(&f64, "f", "float", 0.0, "浮点数")
	fs.DurationVarShortLong(&dur, "d", "duration", 0, "持续时间")
	fs.BoolVarShortLong(&b, "v", "verbose", false, "详细输出")

	err := fs.Parse([]string{
		"-s", "hello",
		"--int", "42",
		"-l", "1024",
		"--uint", "100",
		"-U", "2048",
		"--float", "3.14",
		"-d", "5s",
		"--verbose",
	})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if str != "hello" || i != 42 || i64 != 1024 || u != 100 || u64 != 2048 || f64 != 3.14 || dur != 5*time.Second || !b {
		t.Error("参数解析结果不正确")
	}
}

func TestFlagSet_ShortLongHelp(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var config string
	var port int
	var verbose bool

	fs.StringVarShortLong(&config, "c", "config", "", "配置文件路径")
	fs.IntVarShortLong(&port, "p", "port", 8080, "监听端口")
	fs.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")

	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.PrintDefaults()

	output := buf.String()

	if !strings.Contains(output, "-c/--config") {
		t.Errorf("期望帮助信息包含 '-c/--config'，但输出: %s", output)
	}
	if !strings.Contains(output, "-p/--port") {
		t.Errorf("期望帮助信息包含 '-p/--port'，但输出: %s", output)
	}
	if !strings.Contains(output, "-v/--verbose") {
		t.Errorf("期望帮助信息包含 '-v/--verbose'，但输出: %s", output)
	}

	configCount := strings.Count(output, "配置文件路径")
	if configCount != 1 {
		t.Errorf("期望 '配置文件路径' 只出现1次，但出现了 %d 次", configCount)
	}

	t.Logf("帮助信息显示效果:\n%s", output)
}

func TestFlagSet_StringsShortLongHelp(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var items []string
	fs.StringsVarShortLong(&items, "i", "items", []string{}, "项目列表")

	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.PrintDefaults()

	output := buf.String()

	if !strings.Contains(output, "-i/--items") {
		t.Errorf("期望帮助信息包含 '-i/--items'，但输出: %s", output)
	}

	t.Logf("StringsVarShortLong 帮助信息显示效果:\n%s", output)
}

func TestFlagSet_PasswordStringShortLongHelp(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var password string
	fs.PasswordStringVarShortLong(&password, "p", "password", "", "用户密码", 5)

	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.PrintDefaults()

	output := buf.String()

	if !strings.Contains(output, "-p/--password") {
		t.Errorf("期望帮助信息包含 '-p/--password'，但输出: %s", output)
	}

	t.Logf("PasswordStringVarShortLong 帮助信息显示效果:\n%s", output)
}
