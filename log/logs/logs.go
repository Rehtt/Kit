package logs

import "github.com/Rehtt/Kit/log"

var logs *log.Log

func init() {
	logs = log.NewLog()
}

func Debug(format string, a ...any) {
	logs.Debug(format, a...)
}

func Info(format string, a ...any) {
	logs.Info(format, a...)
}
func Warn(format string, a ...any) {
	logs.Warn(format, a...)
}
func Fatal(format string, a ...any) {
	logs.Fatal(format, a...)
}
func Panic(format string, a ...any) {
	logs.Panic(format, a...)
}

// Apply 更新配置
func Apply(config *log.Config) {
	logs.Apply(config)
}
