package cli

import (
	"bytes"
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
