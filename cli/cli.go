package cli

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct {
	Use         string
	Instruction string
	Func        func(args []string, cli *CLI)
	floor       int
	clis        []*CLI
}

func (c *CLI) AddCommand(cli ...*CLI) {
	c.clis = append(c.clis, cli...)
}

func (c *CLI) Run() {
	flag.Parse()
	cli := findCLI(c)
	if cli == nil {
		os.Exit(0)
	}
	if cli.Func != nil {
		cli.Func(flag.Args()[cli.floor:], cli)
	}
}

func (c *CLI) Help() {
	fmt.Println("参数：")
	for _, v := range c.clis {
		fmt.Println(" ", v.Use, "\t", v.Instruction)
	}
}

func findCLI(c *CLI) *CLI {
	if len(c.clis) == 0 {
		return c
	}
	for _, v := range c.clis {
		v.floor = c.floor + 1
		if v.floor > len(flag.Args()) {
			break
		}
		if flag.Arg(v.floor-1) == v.Use {
			if len(flag.Args()) > v.floor {
				return findCLI(v)
			} else {
				return v
			}
		}
	}
	c.Help()
	return nil
}
