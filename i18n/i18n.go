package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

var (
	loadLang *language.Tag
	langPath = "i18n"
	mu       sync.RWMutex
)

type textMap map[string]string

var texts = make(map[string]textMap)

type GoTextProvider func() textMap

var goProviders = make(map[string]GoTextProvider)

func init() {
	SetLang(nil)
}

func RegisterGoProvider(lang string, provider GoTextProvider) {
	mu.Lock()
	defer mu.Unlock()
	goProviders[lang] = provider
}

func RegisterGoTexts(lang string, t textMap) {
	mu.Lock()
	defer mu.Unlock()
	goProviders[lang] = func() textMap {
		return t
	}
}

func SetLang(l *language.Tag) error {
	text, err := setLang(l)
	if err != nil {
		return err
	}
	texts["default"] = text
	return nil
}

func SetLangByLocalEnv() error {
	tag, err := GetLocalLang()
	if err != nil {
		return err
	}
	return SetLang(tag)
}

func GetLocalLang() (*language.Tag, error) {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		return nil, nil
	}

	langCode := strings.Split(langEnv, ".")[0]
	tag, err := language.Parse(langCode)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func cleanTextMap(m textMap) textMap {
	for k, v := range m {
		if k == v {
			delete(m, k)
		}
	}
	return m
}

func setLang(l *language.Tag) (t textMap, err error) {
	loadLang = l

	langKey := "default"
	if l != nil {
		langKey = l.String()
	}

	// 优先使用Go文件模式
	mu.RLock()
	provider, hasGoProvider := goProviders[langKey]
	mu.RUnlock()

	if hasGoProvider {
		t = cleanTextMap(provider())
		return
	}
	defer func() {
		if provider, hasGoProvider := goProviders["default"]; hasGoProvider && err != nil {
			t = cleanTextMap(provider())
			err = nil
		}
	}()

	// 回退到JSON文件模式
	filename := "default.json"
	if l != nil {
		filename = l.String() + ".json"
	}
	data, err := os.ReadFile(filepath.Join(langPath, filename))
	if err != nil {
		return
	}

	out := make(textMap)
	if err = json.Unmarshal(data, &out); err != nil {
		return
	}
	t = cleanTextMap(out)
	return
}

func SetPath(path string) {
	langPath = path
}

// GetAvailableLanguages 获取所有可用的语言
func GetAvailableLanguages() []string {
	mu.RLock()
	defer mu.RUnlock()

	langs := make([]string, 0, len(goProviders)+len(texts))

	// 添加Go文件模式的语言
	for lang := range goProviders {
		langs = append(langs, lang)
	}

	// 添加已加载的语言
	for lang := range texts {
		if !slices.Contains(langs, lang) {
			langs = append(langs, lang)
		}
	}

	return langs
}

// IsLanguageAvailable 检查指定语言是否可用
func IsLanguageAvailable(lang string) bool {
	mu.RLock()
	defer mu.RUnlock()

	// 检查Go文件模式
	if _, exists := goProviders[lang]; exists {
		return true
	}

	// 检查已加载的语言
	if _, exists := texts[lang]; exists {
		return true
	}

	// 检查JSON文件是否存在
	filename := lang + ".json"
	if lang == "default" {
		filename = "default.json"
	}

	_, err := os.Stat(filepath.Join(langPath, filename))
	return err == nil
}

func GetText(key string, lang ...language.Tag) string {
	if len(lang) > 0 {
		langKey := lang[0].String()

		mu.RLock()
		if text, exists := texts[langKey]; exists {
			if val, found := text[key]; found {
				mu.RUnlock()
				return val
			}
			mu.RUnlock()
			return key
		}
		mu.RUnlock()

		// 尝试加载指定语言
		newText, err := setLang(&lang[0])
		if err == nil {
			mu.Lock()
			texts[langKey] = newText
			mu.Unlock()

			if val, found := newText[key]; found {
				return val
			}
		}
		return key
	}

	// 使用默认语言
	mu.RLock()
	useText := texts["default"]
	if text, exists := useText[key]; exists {
		mu.RUnlock()
		return text
	}
	mu.RUnlock()
	return key
}
