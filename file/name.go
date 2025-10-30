package file

import (
	"regexp"
	"strings"
	"unicode"
)

func SanitizeFilename(s string) string {
	// 1. 去掉控制字符（ASCII 控制区间 + DEL）
	reControl := regexp.MustCompile(`[\x00-\x1F\x7F]+`)
	s = reControl.ReplaceAllString(s, "")

	// 2. 去掉零宽字符和不可见 Unicode 字符
	invisible := []string{
		"\u200B", // ZERO WIDTH SPACE
		"\u200C", // ZERO WIDTH NON-JOINER
		"\u200D", // ZERO WIDTH JOINER
		"\uFEFF", // ZERO WIDTH NO-BREAK SPACE
		"\u2060", // WORD JOINER
	}
	for _, c := range invisible {
		s = strings.ReplaceAll(s, c, "")
	}

	// 3. 去掉不可打印字符
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsPrint(r) {
			out = append(out, r)
		}
	}
	s = string(out)

	// 4. 替换文件系统非法字符（适配 Windows / macOS / Linux）
	reIllegal := regexp.MustCompile(`[<>:"/\\|?*\x00]+`)
	s = reIllegal.ReplaceAllString(s, "_")

	// 5. 去掉前后空白和句点（Windows 不允许以空格或.结尾）
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".")

	// 6. 如果清理后为空，则返回一个默认名称
	if s == "" {
		s = "untitled"
	}

	return s
}
