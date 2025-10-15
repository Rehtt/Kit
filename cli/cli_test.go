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

func TestHelp_PrintsUsageAndSubcommands(t *testing.T) {
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

func TestRun_ExecutesFunc_WhenNoSubcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	var got []string
	root.CommandFunc = func(args []string) error { got = append([]string(nil), args...); return nil }

	root.Parse([]string{"a", "b"})

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected args [a b], got: %#v", got)
	}
}

func TestRun_DispatchesToSubcommand_AndPassesArgs(t *testing.T) {
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

func TestRun_InvalidSubcommand_ShowsHelp(t *testing.T) {
	root, buf := newCLIWithBuf("root", "root desc")
	_ = root.AddCommand(NewCLI("foo", "foo desc"))

	root.Parse([]string{"unknown"})
	out := buf.String()
	if !strings.Contains(out, "Usage: root") || !strings.Contains(out, "Subcommands:") {
		t.Fatalf("expected help output for invalid subcommand, got: %q", out)
	}
}

func TestRun_NestedSubcommand_DispatchAndArgs(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	foo := NewCLI("foo", "foo desc")
	bar := NewCLI("bar", "bar desc")
	var got []string
	bar.CommandFunc = func(args []string) error { got = append([]string(nil), args...); return nil }
	if err := foo.AddCommand(bar); err != nil {
		t.Fatalf("AddCommand bar failed: %v", err)
	}
	if err := root.AddCommand(foo); err != nil {
		t.Fatalf("AddCommand foo failed: %v", err)
	}

	root.Parse([]string{"foo", "bar", "x", "y"})

	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("expected nested subcommand to receive [x y], got: %#v", got)
	}
}

func TestPasswordString_HidesValueInHelp(t *testing.T) {
	cli, buf := newCLIWithBuf("test", "test desc")

	// 测试 PasswordString - 使用 FlagSet 的 PasswordString 方法
	password := cli.PasswordString("password", "secret123", "用户密码")

	// 触发帮助输出
	cli.Help()
	output := buf.String()

	// 验证密码在帮助信息中被隐藏
	if strings.Contains(output, "secret123") {
		t.Errorf("密码值不应该在帮助信息中显示，但找到了: %s", output)
	}

	// 验证实际值正确保存
	if *password != "secret123" {
		t.Errorf("期望密码值为 'secret123'，但得到: %s", *password)
	}
}

func TestPasswordStringVar_HidesValueInHelp(t *testing.T) {
	cli, buf := newCLIWithBuf("test", "test desc")

	// 测试 PasswordStringVar - 使用 FlagSet 的 PasswordStringVar 方法
	var password string
	cli.PasswordStringVar(&password, "password", "mypassword", "用户密码")

	// 触发帮助输出
	cli.Help()
	output := buf.String()

	// 验证密码在帮助信息中被隐藏
	if strings.Contains(output, "mypassword") {
		t.Errorf("密码值不应该在帮助信息中显示，但找到了: %s", output)
	}

	// 验证实际值正确保存
	if password != "mypassword" {
		t.Errorf("期望密码值为 'mypassword'，但得到: %s", password)
	}
}

func TestPasswordString_WithShowNum(t *testing.T) {
	tests := []struct {
		name     string
		password string
		showNum  int
		expected string
	}{
		{"默认显示长度", "secret123", 0, "*********"},
		{"自定义显示3个星号", "secret123", 3, "***"},
		{"自定义显示5个星号", "verylongpassword", 5, "*****"},
		{"空密码", "", 0, ""},
		{"空密码自定义显示", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, buf := newCLIWithBuf("test", "test desc")
			var password *string
			if tt.showNum > 0 {
				password = cli.PasswordString("pass", tt.password, "密码", tt.showNum)
			} else {
				password = cli.PasswordString("pass", tt.password, "密码")
			}

			// 验证值正确设置
			if *password != tt.password {
				t.Errorf("期望密码值为 %q，但得到 %q", tt.password, *password)
			}

			// 触发帮助输出
			cli.Help()
			output := buf.String()

			// 验证密码在帮助中被隐藏
			if tt.password != "" && strings.Contains(output, tt.password) {
				t.Errorf("密码 %q 不应该在帮助信息中显示", tt.password)
			}

			// 验证星号的数量是否正确（通过检查帮助输出）
			if tt.expected != "" && !strings.Contains(output, tt.expected) {
				t.Errorf("期望帮助信息包含 %q，输出: %s", tt.expected, output)
			}
		})
	}
}

func TestPasswordString_ParseFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"解析密码参数", []string{"-password=secret123"}, "secret123"},
		{"解析空格分隔", []string{"-password", "mypass"}, "mypass"},
		{"解析带特殊字符的密码", []string{"-password=P@ssw0rd!"}, "P@ssw0rd!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, _ := newCLIWithBuf("test", "test desc")
			var password string
			cli.PasswordStringVar(&password, "password", "", "用户密码")

			// 设置 CommandFunc 以避免 "no command" 错误
			cli.CommandFunc = func(args []string) error {
				return nil
			}

			// 解析参数
			err := cli.Parse(tt.args)
			if err != nil {
				t.Fatalf("解析参数失败: %v", err)
			}

			// 验证密码被正确解析
			if password != tt.expected {
				t.Errorf("期望密码为 %q，但得到 %q", tt.expected, password)
			}
		})
	}
}

func TestPasswordStringVar_WithShowNum(t *testing.T) {
	cli, buf := newCLIWithBuf("test", "test desc")
	var password string
	cli.PasswordStringVar(&password, "password", "initial", "用户密码", 5)

	// 触发帮助输出
	cli.Help()
	output := buf.String()

	// 验证显示5个星号
	if !strings.Contains(output, "*****") {
		t.Errorf("期望帮助信息包含5个星号，输出: %s", output)
	}

	// 验证实际密码不显示
	if strings.Contains(output, "initial") {
		t.Errorf("密码不应该在帮助信息中显示")
	}

	// 验证实际值正确保存
	if password != "initial" {
		t.Errorf("期望密码值为 'initial'，但得到: %s", password)
	}
}

func TestPasswordString_Integration(t *testing.T) {
	cli, buf := newCLIWithBuf("app", "应用程序")

	var username string
	var password string
	cli.StringVar(&username, "username", "admin", "用户名")
	cli.PasswordStringVar(&password, "password", "secret", "密码")

	var executed bool
	cli.CommandFunc = func(args []string) error {
		executed = true
		return nil
	}

	// 解析参数
	err := cli.Parse([]string{"-username=john", "-password=newpass"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证参数被正确解析
	if username != "john" {
		t.Errorf("期望用户名为 'john'，但得到 %q", username)
	}
	if password != "newpass" {
		t.Errorf("期望密码为 'newpass'，但得到 %q", password)
	}

	// 验证命令被执行
	if !executed {
		t.Error("期望命令被执行")
	}

	// 测试帮助输出
	buf.Reset()
	cli2, buf := newCLIWithBuf("app", "应用程序")
	cli2.StringVar(&username, "username", "admin", "用户名")
	cli2.PasswordStringVar(&password, "password", "secret", "密码")
	cli2.Help()
	output := buf.String()

	// 验证用户名在帮助中正常显示
	if !strings.Contains(output, "admin") {
		t.Error("用户名应该在帮助信息中显示")
	}

	// 验证密码在帮助中被隐藏
	if strings.Contains(output, "secret") {
		t.Error("密码不应该在帮助信息中显示")
	}
}

func TestStrings_Basic(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	// 测试 Strings 方法
	servers := cli.Strings("server", []string{"localhost"}, "服务器地址列表")

	cli.CommandFunc = func(args []string) error {
		return nil
	}

	// 解析参数
	err := cli.Parse([]string{"-server", "192.168.1.1", "-server", "192.168.1.2"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
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

func TestStringsVar_Basic(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	// 测试 StringsVar 方法
	var tags []string
	cli.StringsVar(&tags, "tag", []string{}, "标签列表")

	cli.CommandFunc = func(args []string) error {
		return nil
	}

	// 解析参数
	err := cli.Parse([]string{"-tag", "go", "-tag", "cli", "-tag", "test"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
	expected := []string{"go", "cli", "test"}
	if len(tags) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(tags))
	}

	for i, v := range tags {
		if v != expected[i] {
			t.Errorf("tags[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}

func TestStrings_EmptyDefault(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	// 测试没有默认值的情况
	items := cli.Strings("item", []string{}, "项目列表")

	cli.CommandFunc = func(args []string) error {
		return nil
	}

	// 不传递任何参数
	err := cli.Parse([]string{})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果为空
	if len(*items) != 0 {
		t.Errorf("期望空列表，但得到 %d 个元素", len(*items))
	}
}

func TestStrings_SingleValue(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	paths := cli.Strings("path", []string{}, "路径列表")

	cli.CommandFunc = func(args []string) error {
		return nil
	}

	// 只传递一个值
	err := cli.Parse([]string{"-path", "/usr/local"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证结果
	if len(*paths) != 1 {
		t.Errorf("期望1个元素，但得到 %d 个", len(*paths))
	}
	if (*paths)[0] != "/usr/local" {
		t.Errorf("期望 '/usr/local'，但得到 %q", (*paths)[0])
	}
}

func TestStrings_WithOtherFlags(t *testing.T) {
	cli, _ := newCLIWithBuf("test", "test desc")

	var port int
	var verbose bool
	hosts := cli.Strings("host", []string{}, "主机列表")
	cli.IntVar(&port, "port", 8080, "端口号")
	cli.BoolVar(&verbose, "verbose", false, "详细输出")

	cli.CommandFunc = func(args []string) error {
		return nil
	}

	// 混合使用多种 flag
	err := cli.Parse([]string{
		"-host", "server1.com",
		"-port", "9000",
		"-host", "server2.com",
		"-verbose",
		"-host", "server3.com",
	})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证 hosts
	expectedHosts := []string{"server1.com", "server2.com", "server3.com"}
	if len(*hosts) != len(expectedHosts) {
		t.Errorf("hosts 期望长度为 %d，但得到 %d", len(expectedHosts), len(*hosts))
	}
	for i, v := range *hosts {
		if v != expectedHosts[i] {
			t.Errorf("hosts[%d] = %q，期望 %q", i, v, expectedHosts[i])
		}
	}

	// 验证 port
	if port != 9000 {
		t.Errorf("port 期望为 9000，但得到 %d", port)
	}

	// 验证 verbose
	if !verbose {
		t.Error("verbose 期望为 true")
	}
}

func TestStrings_InSubcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	sub := NewCLI("deploy", "部署命令")

	var targets []string
	sub.StringsVar(&targets, "target", []string{}, "部署目标")

	var executed bool
	sub.CommandFunc = func(args []string) error {
		executed = true
		return nil
	}

	if err := root.AddCommand(sub); err != nil {
		t.Fatalf("添加子命令失败: %v", err)
	}

	// 解析参数
	err := root.Parse([]string{"deploy", "-target", "prod1", "-target", "prod2"})
	if err != nil {
		t.Fatalf("解析参数失败: %v", err)
	}

	// 验证命令被执行
	if !executed {
		t.Error("期望子命令被执行")
	}

	// 验证结果
	expected := []string{"prod1", "prod2"}
	if len(targets) != len(expected) {
		t.Errorf("期望长度为 %d，但得到 %d", len(expected), len(targets))
	}
	for i, v := range targets {
		if v != expected[i] {
			t.Errorf("targets[%d] = %q，期望 %q", i, v, expected[i])
		}
	}
}
