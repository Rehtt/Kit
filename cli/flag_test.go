package cli

import (
	"bytes"
	"flag"
	"fmt"
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

	err := fs.Parse([]string{"--pass", "mypassword"})
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

	err := fs.Parse([]string{"--item", "apple", "--item", "banana"})
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

func TestFlagSet_Single(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}

	var host string
	var port int
	var verbose bool
	fs.StringVar(&host, "host", "localhost", "主机地址")
	fs.IntVar(&port, "port", 8080, "端口号")
	fs.BoolVar(&verbose, "verbose", false, "详细输出")

	err := fs.Parse([]string{"--host", "127.0.0.1", "--port", "9000", "--verbose"})
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

	err = fs.Parse([]string{"-host", "127.0.0.1"})
	if err == nil {
		t.Fatalf("期望解析失败")
	}
	singleLongFlagError := fmt.Sprintf(SingleLongFlagError, "host", "host")
	if err.Error() != singleLongFlagError {
		t.Fatalf("期望 %s ,但得到 %v", singleLongFlagError, err)
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

	if !strings.Contains(output, "-c") || !strings.Contains(output, "--config") {
		t.Errorf("期望帮助信息包含 '-c,\t--config'，但输出: %s", output)
	}
	if !strings.Contains(output, "-p") || !strings.Contains(output, "--port") {
		t.Errorf("期望帮助信息包含 '-p,\t--port'，但输出: %s", output)
	}
	if !strings.Contains(output, "-v") || !strings.Contains(output, "--verbose") {
		t.Errorf("期望帮助信息包含 '-v,\t--verbose'，但输出: %s", output)
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

	if !strings.Contains(output, "-i") || !strings.Contains(output, "--items") {
		t.Errorf("期望帮助信息包含 '-i,\t--items'，但输出: %s", output)
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

	if !strings.Contains(output, "-p") || !strings.Contains(output, "--password") {
		t.Errorf("期望帮助信息包含 '-p,\t--password'，但输出: %s", output)
	}

	t.Logf("PasswordStringVarShortLong 帮助信息显示效果:\n%s", output)
}

func TestFlagSet_CombinedShortFlags(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *FlagSet
		args     []string
		validate func(*testing.T, *FlagSet)
	}{
		{
			name: "组合布尔 flags: -abc",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				fs.BoolVarShortLong(new(bool), "c", "", false, "选项 c")
				return fs
			},
			args: []string{"-abc"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "true" {
					t.Error("期望 b = true")
				}
				if fs.Lookup("c").Value.String() != "true" {
					t.Error("期望 c = true")
				}
			},
		},
		{
			name: "组合布尔 flags 带非布尔 flag: -abf file.txt",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				fs.StringVarShortLong(new(string), "f", "file", "", "文件")
				return fs
			},
			args: []string{"-abf", "file.txt"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "true" {
					t.Error("期望 b = true")
				}
				if fs.Lookup("f").Value.String() != "file.txt" {
					t.Errorf("期望 f = 'file.txt'，但得到 %q", fs.Lookup("f").Value.String())
				}
			},
		},
		{
			name: "单个短 flag 不受影响: -a",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				return fs
			},
			args: []string{"-a"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "false" {
					t.Error("期望 b = false")
				}
			},
		},
		{
			name: "负数不被展开: -123",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.IntVarShortLong(new(int), "n", "num", 0, "数字")
				return fs
			},
			args: []string{"-n", "-123"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("n").Value.String() != "-123" {
					t.Errorf("期望 n = '-123'，但得到 %q", fs.Lookup("n").Value.String())
				}
			},
		},
		{
			name: "长 flag 不受影响: --abc",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVar(new(bool), "abc", false, "选项 abc")
				return fs
			},
			args: []string{"--abc"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("abc").Value.String() != "true" {
					t.Error("期望 abc = true")
				}
			},
		},
		{
			name: "不存在的 flag 组合不展开",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				var buf bytes.Buffer
				fs.SetOutput(&buf) // 抑制错误输出
				return fs
			},
			args: []string{"-ax"}, // x 不存在
			validate: func(t *testing.T, fs *FlagSet) {
				// 这种情况下应该报错，不会成功解析
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setup()
			err := fs.Parse(tt.args)

			// 对于不存在的 flag，期望解析失败
			if tt.name == "不存在的 flag 组合不展开" {
				if err == nil {
					t.Error("期望解析失败，但成功了")
				}
				return
			}

			if err != nil {
				t.Fatalf("解析参数失败: %v", err)
			}
			tt.validate(t, fs)
		})
	}
}

func TestFlagSet_CombinedFlagsWithRemainingArgs(t *testing.T) {
	fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	fs.BoolVarShortLong(new(bool), "v", "verbose", false, "详细输出")
	fs.BoolVarShortLong(new(bool), "x", "", false, "选项 x")
	fs.StringVarShortLong(new(string), "f", "file", "", "文件")

	err := fs.Parse([]string{"-vxf", "test.txt", "arg1", "arg2"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if fs.Lookup("v").Value.String() != "true" {
		t.Error("期望 v = true")
	}
	if fs.Lookup("x").Value.String() != "true" {
		t.Error("期望 x = true")
	}
	if fs.Lookup("f").Value.String() != "test.txt" {
		t.Errorf("期望 f = 'test.txt'，但得到 %q", fs.Lookup("f").Value.String())
	}

	args := fs.Args()
	if len(args) != 2 || args[0] != "arg1" || args[1] != "arg2" {
		t.Errorf("期望剩余参数为 [arg1 arg2]，但得到 %v", args)
	}
}

func TestFlagSet_CombinedFlagsInlineValue(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *FlagSet
		args     []string
		validate func(*testing.T, *FlagSet)
	}{
		{
			name: "组合 flags 后面跟内联值: -abcasd",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				fs.StringVarShortLong(new(string), "c", "", "", "选项 c")
				return fs
			},
			args: []string{"-abcasd"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "true" {
					t.Error("期望 b = true")
				}
				if fs.Lookup("c").Value.String() != "asd" {
					t.Errorf("期望 c = 'asd'，但得到 %q", fs.Lookup("c").Value.String())
				}
			},
		},
		{
			name: "单个 flag 后面跟内联值: -c123",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.StringVarShortLong(new(string), "c", "", "", "选项 c")
				return fs
			},
			args: []string{"-c123"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("c").Value.String() != "123" {
					t.Errorf("期望 c = '123'，但得到 %q", fs.Lookup("c").Value.String())
				}
			},
		},
		{
			name: "组合 flags 内联数字值: -abO2",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				fs.StringVarShortLong(new(string), "O", "", "", "优化级别")
				return fs
			},
			args: []string{"-abO2"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "true" {
					t.Error("期望 b = true")
				}
				if fs.Lookup("O").Value.String() != "2" {
					t.Errorf("期望 O = '2'，但得到 %q", fs.Lookup("O").Value.String())
				}
			},
		},
		{
			name: "整数类型内联值: -n5",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.IntVarShortLong(new(int), "n", "", 0, "数字")
				return fs
			},
			args: []string{"-n5"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("n").Value.String() != "5" {
					t.Errorf("期望 n = '5'，但得到 %q", fs.Lookup("n").Value.String())
				}
			},
		},
		{
			name: "组合后带内联值再带其他参数: -abctest arg1",
			setup: func() *FlagSet {
				fs := &FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
				fs.BoolVarShortLong(new(bool), "a", "", false, "选项 a")
				fs.BoolVarShortLong(new(bool), "b", "", false, "选项 b")
				fs.StringVarShortLong(new(string), "c", "", "", "选项 c")
				return fs
			},
			args: []string{"-abctest", "arg1"},
			validate: func(t *testing.T, fs *FlagSet) {
				if fs.Lookup("a").Value.String() != "true" {
					t.Error("期望 a = true")
				}
				if fs.Lookup("b").Value.String() != "true" {
					t.Error("期望 b = true")
				}
				if fs.Lookup("c").Value.String() != "test" {
					t.Errorf("期望 c = 'test'，但得到 %q", fs.Lookup("c").Value.String())
				}
				args := fs.Args()
				if len(args) != 1 || args[0] != "arg1" {
					t.Errorf("期望剩余参数为 [arg1]，但得到 %v", args)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setup()
			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("解析参数失败: %v", err)
			}
			tt.validate(t, fs)
		})
	}
}
