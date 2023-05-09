package strings

import (
	"strings"
)

var replacements = map[string]string{
	`\`:  `\\`,
	`'`:  `\'`,
	`"`:  `\"`,
	`%`:  `\%`,
	`_`:  `\_`,
	"\n": "\\n",
	"\r": "\\r",
	"\t": "\\t",
	"\a": "\\a",
	"\f": "\\f",
	"\v": "\\v",
	"\b": "\\b",
}

func Escape(reverse ...bool) *strings.Replacer {
	rep := make([]string, 0, len(replacements)*2)
	for ori, es := range replacements {
		if len(reverse) != 0 && reverse[0] {
			rep = append(rep, es, ori)
		} else {
			rep = append(rep, ori, es)
		}

	}
	return strings.NewReplacer(rep...)
}

func EscapeString(str string, reverse ...bool) string {
	str = Escape(reverse...).Replace(str)
	return str
}
func EscapeStringRepeat(str string, repeat int, reverse ...bool) string {
	for i := 0; i < repeat; i++ {
		str = Escape(reverse...).Replace(str)
	}
	return str
}
