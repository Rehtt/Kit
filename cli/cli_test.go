package cli

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func newCLIWithBuf(use, instruction string) (*CLI, *bytes.Buffer) {
	cli := NewCLI(use, instruction, flag.ExitOnError)
	var buf bytes.Buffer
	cli.SetOutput(&buf)
	return cli, &buf
}

func TestHelp_PrintsUsageAndSubcommands(t *testing.T) {
	root, buf := newCLIWithBuf("root", "root desc")
	_ = root.AddCommand(
		NewCLI("foo", "foo desc", flag.ExitOnError),
		NewCLI("bar", "bar desc", flag.ExitOnError),
	)

	root.Help()
	out := buf.String()
	if !strings.Contains(out, "Usage of root:") {
		t.Fatalf("expected usage header, got: %q", out)
	}
	if !strings.Contains(out, "Subcommands:") {
		t.Fatalf("expected subcommands section, got: %q", out)
	}
	if !strings.Contains(out, "foo\tfoo desc") || !strings.Contains(out, "bar\tbar desc") {
		t.Fatalf("expected subcommands listed with instructions, got: %q", out)
	}
}

func TestRun_ExecutesFunc_WhenNoSubcommand(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	var got []string
	root.CommandFunc = func(args []string) { got = append([]string(nil), args...) }

	root.Parse([]string{"a", "b"})

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected args [a b], got: %#v", got)
	}
}

func TestRun_DispatchesToSubcommand_AndPassesArgs(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	sub := NewCLI("foo", "foo desc", flag.ExitOnError)
	var got []string
	sub.CommandFunc = func(args []string) { got = append([]string(nil), args...) }
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
	_ = root.AddCommand(NewCLI("foo", "foo desc", flag.ExitOnError))

	root.Parse([]string{"unknown"})
	out := buf.String()
	if !strings.Contains(out, "Usage of root:") || !strings.Contains(out, "Subcommands:") {
		t.Fatalf("expected help output for invalid subcommand, got: %q", out)
	}
}

func TestRun_NestedSubcommand_DispatchAndArgs(t *testing.T) {
	root, _ := newCLIWithBuf("root", "root desc")
	foo := NewCLI("foo", "foo desc", flag.ExitOnError)
	bar := NewCLI("bar", "bar desc", flag.ExitOnError)
	var got []string
	bar.CommandFunc = func(args []string) { got = append([]string(nil), args...) }
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
