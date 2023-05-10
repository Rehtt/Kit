package logs

import "github.com/Rehtt/Kit/log"

var logs *log.Log

func init() {
	logs = log.NewLog()
}

func Debug(format string, a ...interface{}) {
	logs.Debug(format, a...)
}

func Info(format string, a ...interface{}) {
	logs.Info(format, a...)
}
func Warn(format string, a ...interface{}) {
	logs.Warn(format, a...)
}
func Fatal(format string, a ...interface{}) {
	logs.Fatal(format, a...)
}
func Panic(format string, a ...interface{}) {
	logs.Panic(format, a...)
}

// Apply 更新配置
func Apply(config *log.Config) {
	logs.Apply(config)
}
