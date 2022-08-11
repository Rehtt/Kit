# cli
简单cli

```go
root := &cli.CLI{}
start := &cli.CLI{
	Use:         "start",
	Instruction: "启动",
}
web := &cli.CLI{
	Use: "web",
	Func: func(args []string, cli *cli.CLI) {
		cli.Help()
	},
}
start.AddCommand(web)
root.AddCommand(start)
root.Run()
```
