package strings

import "strings"

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

func EscapeString(str string, reverse ...bool) string {
	rep := make([]string, 0, len(replacements)*2)
	for ori, es := range replacements {
		if len(reverse) != 0 && reverse[0] {
			rep = append(rep, es, ori)
		} else {
			rep = append(rep, ori, es)
		}

	}
	str = strings.NewReplacer(rep...).Replace(str)
	return str
}
