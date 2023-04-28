package log

import (
	"fmt"
	"github.com/Rehtt/Kit/buf"
	"os"
	"runtime/debug"
	"time"
)

type Log struct {
	*Config
}

type Option interface {
	Apply(config *Config)
}

func NewLog(option ...Option) *Log {
	var c = newConfig()
	for _, opt := range option {
		if opt != nil {
			opt.Apply(c)
		}
	}
	return &Log{
		Config: c,
	}
}

func (l Log) Debug(format string, a ...interface{}) {
	if l.Level > DEBUG {
		return
	}
	fmt.Println(l.sprintf(DEBUG, format, a...))
}

func (l Log) Info(format string, a ...interface{}) {
	if l.Level > INFO {
		return
	}
	fmt.Println(l.sprintf(INFO, format, a...))
}

func (l Log) Warn(format string, a ...interface{}) {
	if l.Level > WARN {
		return
	}
	fmt.Println(l.sprintf(WARN, format, a...))
}

func (l Log) Fatal(format string, a ...interface{}) {
	if l.Level > FATAL {
		return
	}
	fmt.Println(l.sprintf(FATAL, format, a...))
	debug.PrintStack()
	os.Exit(1)
}

func (l Log) Panic(format string, a ...interface{}) {
	if l.Level > PANIC {
		return
	}
	fmt.Println(l.sprintf(PANIC, format, a...))
	panic(fmt.Sprintf(format, a...))
}

func (l Log) sprintf(leve Level, format string, a ...interface{}) string {
	var tmp = buf.NewBuf()
	if !l.NotShowTime {
		tmp.WriteString(time.Now().Format(l.TimeLayout))
	}
	tmp.WriteString(" [")
	tmp.WriteString(LevelStr[leve])
	tmp.WriteString("] ")

	tmp.WriteString(fmt.Sprintf(format, a...))

	if l.NotShowColor {
		return tmp.ToString(true)
	}
	switch leve {
	case DEBUG:
		return tmp.ToColorString(l.DebugColor, true)
	case INFO:
		return tmp.ToColorString(l.InfoColor, true)
	case WARN:
		return tmp.ToColorString(l.WarnColor, true)
	case FATAL:
		return tmp.ToColorString(l.FatalColor, true)
	case PANIC:
		return tmp.ToColorString(l.PanicColor, true)
	}
	return ""
}
