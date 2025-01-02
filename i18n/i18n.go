package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/text/language"
)

var (
	loadLang *language.Tag
	langPath = "i18n"
)

type textMap map[string]string

var texts = make(map[string]textMap)

func init() {
	SetLang(nil)
}

func SetLang(l *language.Tag) error {
	text, err := setLang(l)
	if err != nil {
		return err
	}
	texts["default"] = text
	return nil
}

func setLang(l *language.Tag) (map[string]string, error) {
	loadLang = l
	var data []byte
	path := "default.json"
	if l != nil {
		path = l.String() + ".json"
	}
	data, err := os.ReadFile(filepath.Join(langPath, path))
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	if err = json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	for k, v := range out {
		if k == v {
			delete(out, k)
		}
	}
	return out, nil
}

func SetPath(path string) {
	langPath = path
}

func GetText(str string, lang ...language.Tag) string {
	useText := texts["default"]
	if len(lang) != 0 {
		if !(loadLang != nil && loadLang.String() == lang[0].String()) {
			text, ok := texts[lang[0].String()]
			if !ok {
				useText, _ = setLang(&lang[0])
				texts[lang[0].String()] = useText
			} else {
				useText = text
			}
		}
	}
	out, ok := useText[str]
	if !ok {
		out = str
	}
	return out
}
