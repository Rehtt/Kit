package cli

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func newCLIWithBuf(use, instruction string) (*CLI, *bytes.Buffer) {
	cli := NewCLI(use, instruction)
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	return cli, &buf
}

func TestCLI_Help(t *testing.T) {
	root, buf := newCLIWithBuf("root", "root desc")
	_ = root.AddCommand(
		NewCLI("foo", "foo desc"),
		NewCLI("bar", "bar desc"),
	)

	root.Help()
	out := buf.String()
	if !strings.Contains(out, "Usage: root") {
		t.Fatalf("expected usage header, got: %q", out)
	}
	if !strings.Contains(out, "Subcommands:") {
		t.Fatalf("expected subcommands section, got: %q", out)
	}
	if !strings.Contains(out, "foo  foo desc") || !strings.Contains(out, "bar  bar desc") {
		t.Fatalf("expected subcommands listed with instructions, got: %q", out)
	}
}

func TestCLI_ExecuteCommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	var got []string
	root.CommandFunc = func(args []string) error { got = append([]string(nil), args...); return nil }

	root.Parse([]string{"a", "b"})

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected args [a b], got: %#v", got)
	}
}

func TestCLI_Subcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	sub := NewCLI("foo", "foo desc")
	var got []string
	sub.CommandFunc = func(args []string) error { got = append([]string(nil), args...); return nil }
	if err := root.AddCommand(sub); err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}

	root.Parse([]string{"foo", "x", "y"})

	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("expected subcommand to receive [x y], got: %#v", got)
	}
}

func TestCLI_NestedSubcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	foo := NewCLI("foo", "foo desc")
	bar := NewCLI("bar", "bar desc")
	var got []string
	bar.CommandFunc = func(args []string) error { got = append([]string(nil), args...); return nil }
	_ = foo.AddCommand(bar)
	_ = root.AddCommand(foo)

	root.Parse([]string{"foo", "bar", "x", "y"})

	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("expected nested subcommand to receive [x y], got: %#v", got)
	}
}

func TestCLI_PasswordString(t *testing.T) {
	cli, buf := newCLIWithBuf("test", "test desc")
	password := cli.PasswordString("password", "secret123", "用户密码")

	cli.Help()
	output := buf.String()

	if strings.Contains(output, "secret123") {
		t.Errorf("密码值不应该在帮助信息中显示")
	}

	if *password != "secret123" {
		t.Errorf("期望密码值为 'secret123'，但得到: %s", *password)
	}
}

func TestCLI_Strings(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")
	servers := cli.Strings("server", []string{"localhost"}, "服务器地址列表")
	cli.CommandFunc = func(args []string) error { return nil }

	err := cli.Parse([]string{"-server", "192.168.1.1", "-server", "192.168.1.2"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	expected := []string{"localhost", "192.168.1.1", "192.168.1.2"}
	if len(*servers) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(*servers))
	}

	for i, v := range *servers {
		if v != expected[i] {
			t.Errorf("servers[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestCLI_CommandSort(t *testing.T) {
	cli, buf := newCLIWithBuf("test", "test desc")
	_ = cli.AddCommand(
		NewCLI("a", "a desc"),
		NewCLI("c", "c desc"),
		NewCLI("d", "d desc"),
		NewCLI("b", "b desc"),
	)
	cli.SubCommands.SetSort(CommandSortAdded)
	cli.Help()
	out := buf.String()
	if !strings.Contains(out, "a  a desc\n  c  c desc\n  d  d desc\n  b  b desc") {
		t.Fatalf("expected subcommands listed with instructions, got: %q", out)
	}

	buf.Reset()
	cli.SubCommands.SetSort(CommandSortAlphaAsc)
	cli.Help()
	out = buf.String()
	if !strings.Contains(out, "a  a desc\n  b  b desc\n  c  c desc\n  d  d desc") {
		t.Fatalf("expected subcommands listed with instructions, got: %q", out)
	}

	buf.Reset()
	cli.SubCommands.SetSort(CommandSortAlphaDesc)
	cli.Help()
	out = buf.String()
	if !strings.Contains(out, "d  d desc\n  c  c desc\n  b  b desc\n  a  a desc") {
		t.Fatalf("expected subcommands listed with instructions, got: %q", out)
	}
}

func TestCLI_RawCommand(t *testing.T) {
	t.Run("RequiresCommandFunc", func(t *testing.T) {
		cli, _ := newCLIWithBuf("raw", "raw desc")
		cli.Raw = true

		err := cli.Parse([]string{"foo"})
		if err == nil || err.Error() != "raw command must have a command function" {
			t.Fatalf("expected raw command error, got %v", err)
		}
	})

	t.Run("ExecutesWithoutParsingFlags", func(t *testing.T) {
		cli, _ := newCLIWithBuf("raw", "raw desc")
		cli.Raw = true

		var got []string
		cli.CommandFunc = func(args []string) error {
			got = append([]string(nil), args...)
			return nil
		}

		if err := cli.Parse([]string{"--some", "value"}); err != nil {
			t.Fatalf("raw parse failed: %v", err)
		}

		if len(got) != 2 || got[0] != "--some" || got[1] != "value" {
			t.Fatalf("expected raw args to pass through, got %#v", got)
		}
	})
}

func TestCLI_HiddenSubcommands(t *testing.T) {
	cli, buf := newCLIWithBuf("root", "root desc")
	visible := NewCLI("visible", "show me")
	hidden := NewCLI("hidden", "do not show")
	hidden.Hidden = true
	_ = cli.AddCommand(visible, hidden)

	cli.Help()
	out := buf.String()

	if !strings.Contains(out, "visible") {
		t.Fatalf("expected visible command in help, got %q", out)
	}
	if strings.Contains(out, "hidden") {
		t.Fatalf("did not expect hidden command in help, got %q", out)
	}
}

func TestCLI_UnknownSubcommand(t *testing.T) {
	cli, _ := newCLIWithBuf("root", "root desc")
	_ = cli.AddCommand(NewCLI("known", "known desc"))

	err := cli.Parse([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected error wrapping flag.ErrHelp, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand message, got %v", err)
	}
}

func TestCLI_AddCommandDuplicate(t *testing.T) {
	cli, _ := newCLIWithBuf("root", "root desc")

	err := cli.AddCommand(NewCLI("dup", "first"), NewCLI("dup", "second"))
	if err == nil {
		t.Fatal("expected duplicate command error")
	}
	if !strings.Contains(err.Error(), "duplicate command") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCLI_IsCompleteFlag(t *testing.T) {
	cli, _ := newCLIWithBuf("root", "root desc")
	var (
		cfg string
		out string
	)
	cli.StringVarShortLong(&cfg, "c", "config", "", "配置")
	cli.StringVar(&out, "output", "", "输出")

	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"LongName", "--config", true},
		{"ShortName", "-c", true},
		{"Unknown", "--missing", false},
		{"ShortMultiChar", "-co", false},
		{"LongSingleChar", "--c", false},
		{"DefaultFlagLong", "--output", true},
		{"DefaultFlagSingleDash", "-output", false},
	}

	for _, tt := range tests {
		if got := cli.IsCompleteFlag(tt.arg); got != tt.want {
			t.Errorf("%s: IsCompleteFlag(%q) = %v, want %v", tt.name, tt.arg, got, tt.want)
		}
	}
}

func TestCLI_IsCompleteFlagInContext(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	child := NewCLI("child", "child desc")
	var flagVal string
	child.StringVar(&flagVal, "child-flag", "", "child flag")
	_ = root.AddCommand(child)

	if got := root.IsCompleteFlagInContext("--child-flag", []string{"child"}); !got {
		t.Fatal("expected child flag recognized in context")
	}
	if got := root.IsCompleteFlagInContext("--child-flag", []string{"unknown"}); got {
		t.Fatal("expected unknown child to fallback to root and be false")
	}
}

func TestCLI_CombinedShortFlags(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	var verbose bool
	var extract bool
	var file string
	var got []string

	cli.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
	cli.BoolVarShortLong(&extract, "x", "extract", false, "提取")
	cli.StringVarShortLong(&file, "f", "file", "", "文件")

	cli.CommandFunc = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}

	err := cli.Parse([]string{"-vxf", "test.tar.gz", "dest"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if !verbose {
		t.Error("期望 verbose = true")
	}
	if !extract {
		t.Error("期望 extract = true")
	}
	if file != "test.tar.gz" {
		t.Errorf("期望 file = 'test.tar.gz'，但得到 %q", file)
	}
	if len(got) != 1 || got[0] != "dest" {
		t.Errorf("期望剩余参数为 [dest]，但得到 %v", got)
	}
}

func TestCLI_CombinedBoolFlags(t *testing.T) {
	cli, _ := newCLIWithBuf("ls", "list directory")

	var all bool
	var long bool
	var human bool

	cli.BoolVarShortLong(&all, "a", "all", false, "显示所有文件")
	cli.BoolVarShortLong(&long, "l", "long", false, "长格式")
	cli.BoolVarShortLong(&human, "h", "human-readable", false, "人类可读格式")

	cli.CommandFunc = func(args []string) error { return nil }

	// 测试组合: -alh
	err := cli.Parse([]string{"-alh"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if !all {
		t.Error("期望 all = true")
	}
	if !long {
		t.Error("期望 long = true")
	}
	if !human {
		t.Error("期望 human = true")
	}
}

func TestCLI_CombinedFlagsWithSubcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	sub := NewCLI("test", "test subcommand")

	var verbose bool
	var debug bool
	var got []string

	sub.BoolVarShortLong(&verbose, "v", "verbose", false, "详细输出")
	sub.BoolVarShortLong(&debug, "d", "debug", false, "调试模式")
	sub.CommandFunc = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}

	_ = root.AddCommand(sub)

	err := root.Parse([]string{"test", "-vd", "arg1"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if !verbose {
		t.Error("期望 verbose = true")
	}
	if !debug {
		t.Error("期望 debug = true")
	}
	if len(got) != 1 || got[0] != "arg1" {
		t.Errorf("期望剩余参数为 [arg1]，但得到 %v", got)
	}
}

func TestCLI_CombinedFlagsWithInlineValue(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	var a, b bool
	var c string
	var got []string

	cli.BoolVarShortLong(&a, "a", "", false, "选项 a")
	cli.BoolVarShortLong(&b, "b", "", false, "选项 b")
	cli.StringVarShortLong(&c, "c", "", "", "选项 c")

	cli.CommandFunc = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}

	// 测试 -abcasd
	err := cli.Parse([]string{"-abcasd", "arg1"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if !a {
		t.Error("期望 a = true")
	}
	if !b {
		t.Error("期望 b = true")
	}
	if c != "asd" {
		t.Errorf("期望 c = 'asd'，但得到 %q", c)
	}
	if len(got) != 1 || got[0] != "arg1" {
		t.Errorf("期望剩余参数为 [arg1]，但得到 %v", got)
	}
}

func TestCLI_SingleFlagWithInlineValue(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	var c string

	cli.StringVarShortLong(&c, "c", "", "", "选项 c")
	cli.CommandFunc = func(args []string) error { return nil }

	// 测试 -c123
	err := cli.Parse([]string{"-c123"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	if c != "123" {
		t.Errorf("期望 c = '123'，但得到 %q", c)
	}
}
