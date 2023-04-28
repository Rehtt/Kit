package log

import (
	"github.com/Rehtt/Kit/vt/color"
	"testing"
)

func TestNewLog(t *testing.T) {
	l := NewLog()
	l.TimeLayout = "20060102150405"
	l.WarnColor = color.NewColors(color.BgCyan)
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Fatal("fatal")
	l.Panic("panic")

	Debug("debug")
	Info("info")
	Warn("warn")
	Fatal("fatal")
	Panic("panic")
}
