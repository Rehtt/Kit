package log

import (
	"github.com/Rehtt/Kit/vt/color"
)

type Level int

const (
	DEBUG = Level(iota)
	INFO
	WARN
	FATAL
	PANIC
)

var LevelStr = map[Level]string{
	DEBUG: "debug",
	INFO:  "info",
	WARN:  "warn",
	FATAL: "fatal",
	PANIC: "panic",
}

type Config struct {
	InfoOutFile  string
	ErrorOutFile string
	Level        Level
	// 不显示时间
	NotShowTime bool
	// 时间格式，默认：YYYY-MM-DD hh:mm:ss
	TimeLayout string

	// 不显示颜色，默认：false
	NotShowColor bool
	// Debug 颜色，默认：FgBlue
	DebugColor color.Colors
	// Info 颜色，默认：无
	InfoColor color.Colors
	// Warn 颜色，默认：FgYellow
	WarnColor color.Colors
	// Fatal 颜色，默认：FgRed
	FatalColor color.Colors
	// Panic 颜色，默认：FgHiRed
	PanicColor color.Colors
}

func newConfig() *Config {
	c := new(Config)
	c.init()
	return c
}

// Apply 更新配置
func (c *Config) Apply(config *Config) {
	c.init()
	if config != c {
		*config = *c
	}
}

func (c *Config) init() {
	if !c.NotShowTime && c.TimeLayout == "" {
		c.TimeLayout = "2006-01-02 15:04:05"
	}
	if !c.NotShowColor {
		if !c.DebugColor.HasColors() {
			c.DebugColor = color.NewColors(color.FgBlue)
		}
		if !c.WarnColor.HasColors() {
			c.WarnColor = color.NewColors(color.FgYellow)
		}
		if !c.FatalColor.HasColors() {
			c.FatalColor = color.NewColors(color.FgRed)
		}
		if !c.PanicColor.HasColors() {
			c.PanicColor = color.NewColors(color.FgHiRed)
		}
	}
}
