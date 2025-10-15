package cli

import (
	"encoding"
	"flag"
	"time"
)

func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.BoolVar(p, name, value, usage)
}

func Bool(name string, value bool, usage string) *bool { return CommandLine.Bool(name, value, usage) }

func StringVar(p *string, name string, value string, usage string) {
	CommandLine.StringVar(p, name, value, usage)
}

func String(name string, value string, usage string) *string {
	return CommandLine.String(name, value, usage)
}

func IntVar(p *int, name string, value int, usage string) { CommandLine.IntVar(p, name, value, usage) }

func Int(name string, value int, usage string) *int { return CommandLine.Int(name, value, usage) }

func Int64Var(p *int64, name string, value int64, usage string) {
	CommandLine.Int64Var(p, name, value, usage)
}

func Int64(name string, value int64, usage string) *int64 {
	return CommandLine.Int64(name, value, usage)
}

func UintVar(p *uint, name string, value uint, usage string) {
	CommandLine.UintVar(p, name, value, usage)
}

func Uint(name string, value uint, usage string) *uint { return CommandLine.Uint(name, value, usage) }

func Uint64Var(p *uint64, name string, value uint64, usage string) {
	CommandLine.Uint64Var(p, name, value, usage)
}

func Uint64(name string, value uint64, usage string) *uint64 {
	return CommandLine.Uint64(name, value, usage)
}

func Float64Var(p *float64, name string, value float64, usage string) {
	CommandLine.Float64Var(p, name, value, usage)
}

func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64(name, value, usage)
}

func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	CommandLine.DurationVar(p, name, value, usage)
}

func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.Duration(name, value, usage)
}

func TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	CommandLine.TextVar(p, name, value, usage)
}

func Func(name, usage string, fn func(string) error) { CommandLine.Func(name, usage, fn) }

func BoolFunc(name, usage string, fn func(string) error) { CommandLine.BoolFunc(name, usage, fn) }

func Var(p flag.Value, name string, usage string) { CommandLine.Var(p, name, usage) }

func VisitAll(fn func(*flag.Flag)) {
	CommandLine.VisitAll(fn)
}

func Visit(fn func(*flag.Flag)) {
	CommandLine.Visit(fn)
}

func Lookup(name string) *flag.Flag {
	return CommandLine.Lookup(name)
}

func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

func PasswordStringVar(p *string, name string, value string, usage string, showNum ...int) {
	CommandLine.PasswordStringVar(p, name, value, usage, showNum...)
}

func PasswordString(name string, value string, usage string, showNum ...int) *string {
	p := new(string)
	CommandLine.PasswordStringVar(p, name, value, usage, showNum...)
	return p
}

func StringsVar(p *[]string, name string, value []string, usage string) {
	CommandLine.StringsVar(p, name, value, usage)
}

func Strings(name string, value []string, usage string) *[]string {
	return CommandLine.Strings(name, value, usage)
}

func Alias(alias, original string) {
	CommandLine.Alias(alias, original)
}

func StringVarShortLong(p *string, short, long string, value string, usage string) {
	CommandLine.StringVarShortLong(p, short, long, value, usage)
}

func StringShortLong(short, long string, value string, usage string) *string {
	return CommandLine.StringShortLong(short, long, value, usage)
}

func IntVarShortLong(p *int, short, long string, value int, usage string) {
	CommandLine.IntVarShortLong(p, short, long, value, usage)
}

func IntShortLong(short, long string, value int, usage string) *int {
	return CommandLine.IntShortLong(short, long, value, usage)
}

func BoolVarShortLong(p *bool, short, long string, value bool, usage string) {
	CommandLine.BoolVarShortLong(p, short, long, value, usage)
}

func BoolShortLong(short, long string, value bool, usage string) *bool {
	return CommandLine.BoolShortLong(short, long, value, usage)
}

func StringsVarShortLong(p *[]string, short, long string, value []string, usage string) {
	CommandLine.StringsVarShortLong(p, short, long, value, usage)
}

func StringsShortLong(short, long string, value []string, usage string) *[]string {
	return CommandLine.StringsShortLong(short, long, value, usage)
}

func PasswordStringVarShortLong(p *string, short, long string, value string, usage string, showNum ...int) {
	CommandLine.PasswordStringVarShortLong(p, short, long, value, usage, showNum...)
}

func PasswordStringShortLong(short, long string, value string, usage string, showNum ...int) *string {
	return CommandLine.PasswordStringShortLong(short, long, value, usage, showNum...)
}

func Int64VarShortLong(p *int64, short, long string, value int64, usage string) {
	CommandLine.Int64VarShortLong(p, short, long, value, usage)
}

func Int64ShortLong(short, long string, value int64, usage string) *int64 {
	return CommandLine.Int64ShortLong(short, long, value, usage)
}

func UintVarShortLong(p *uint, short, long string, value uint, usage string) {
	CommandLine.UintVarShortLong(p, short, long, value, usage)
}

func UintShortLong(short, long string, value uint, usage string) *uint {
	return CommandLine.UintShortLong(short, long, value, usage)
}

func Uint64VarShortLong(p *uint64, short, long string, value uint64, usage string) {
	CommandLine.Uint64VarShortLong(p, short, long, value, usage)
}

func Uint64ShortLong(short, long string, value uint64, usage string) *uint64 {
	return CommandLine.Uint64ShortLong(short, long, value, usage)
}

func Float64VarShortLong(p *float64, short, long string, value float64, usage string) {
	CommandLine.Float64VarShortLong(p, short, long, value, usage)
}

func Float64ShortLong(short, long string, value float64, usage string) *float64 {
	return CommandLine.Float64ShortLong(short, long, value, usage)
}

func DurationVarShortLong(p *time.Duration, short, long string, value time.Duration, usage string) {
	CommandLine.DurationVarShortLong(p, short, long, value, usage)
}

func DurationShortLong(short, long string, value time.Duration, usage string) *time.Duration {
	return CommandLine.DurationShortLong(short, long, value, usage)
}
