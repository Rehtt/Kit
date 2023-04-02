package i18n

import (
	"encoding/json"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
)

var (
	loadLang *language.Tag
	langPath = "i18n"
)

var (
	text = make(map[string]string)
)

func init() {
	SetLang(nil)
}

func SetLang(l *language.Tag) {
	loadLang = l
	var data []byte
	var path = "default.json"
	if l != nil {
		path = l.String() + ".json"
	}
	data, err := os.ReadFile(filepath.Join(langPath, path))
	if err != nil {
		return
	}
	json.Unmarshal(data, &text)
}
func SetPath(path string) {
	langPath = path
}

func GetText(str string, lang ...language.Tag) string {
	if len(lang) != 0 {
		if !(loadLang != nil && loadLang.String() == lang[0].String()) {
			SetLang(&lang[0])
		}
	}
	out, ok := text[str]
	if !ok {
		out = str
	}
	return out
}
