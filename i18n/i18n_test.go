package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

func TestRegisterGoTexts(t *testing.T) {
	// 清理测试环境
	goProviders = make(map[string]GoTextProvider)
	texts = make(map[string]textMap)

	// 注册测试数据
	testTexts := textMap{
		"hello":   "Hello",
		"world":   "World",
		"goodbye": "Goodbye",
	}
	RegisterGoTexts("test", testTexts)

	// 验证注册成功
	if len(goProviders) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(goProviders))
	}

	provider, exists := goProviders["test"]
	if !exists {
		t.Error("Provider not found")
	}

	result := provider()
	if len(result) != 3 {
		t.Errorf("Expected 3 texts, got %d", len(result))
	}

	if result["hello"] != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", result["hello"])
	}
}

func TestGetText(t *testing.T) {
	// 清理测试环境
	goProviders = make(map[string]GoTextProvider)
	texts = make(map[string]textMap)

	// 注册测试数据
	RegisterGoTexts("default", textMap{
		"hello": "Hello",
		"world": "World",
	})

	RegisterGoTexts("zh", textMap{
		"hello": "你好",
		"world": "世界",
	})

	// 初始化
	SetLang(nil)

	// 测试默认语言
	if GetText("hello") != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", GetText("hello"))
	}

	// 测试指定语言
	zh := language.Chinese
	t.Logf("Chinese language tag: %s", zh.String())
	result := GetText("hello", zh)
	t.Logf("GetText result: %s", result)
	if result != "你好" {
		t.Errorf("Expected '你好', got '%s'", result)
	}

	// 测试不存在的键
	if GetText("nonexistent") != "nonexistent" {
		t.Errorf("Expected 'nonexistent', got '%s'", GetText("nonexistent"))
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	// 清理测试环境
	goProviders = make(map[string]GoTextProvider)
	texts = make(map[string]textMap)

	// 注册测试数据
	RegisterGoTexts("default", textMap{"hello": "Hello"})
	RegisterGoTexts("zh", textMap{"hello": "你好"})

	langs := GetAvailableLanguages()
	if len(langs) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(langs))
	}

	// 检查是否包含预期的语言
	hasDefault := false
	hasZh := false
	for _, lang := range langs {
		if lang == "default" {
			hasDefault = true
		}
		if lang == "zh" {
			hasZh = true
		}
	}

	if !hasDefault {
		t.Error("Expected 'default' language")
	}
	if !hasZh {
		t.Error("Expected 'zh' language")
	}
}

func TestIsLanguageAvailable(t *testing.T) {
	// 清理测试环境
	goProviders = make(map[string]GoTextProvider)
	texts = make(map[string]textMap)

	// 注册测试数据
	RegisterGoTexts("test-lang", textMap{"hello": "Hello"})

	if !IsLanguageAvailable("test-lang") {
		t.Error("Expected 'test-lang' to be available")
	}

	if IsLanguageAvailable("nonexistent") {
		t.Error("Expected 'nonexistent' to be unavailable")
	}
}

func TestCleanTextMap(t *testing.T) {
	testMap := textMap{
		"hello": "Hello",
		"world": "World",
		"same":  "same", // 这个应该被清理
		"empty": "",
		"space": " ",
	}

	result := cleanTextMap(testMap)

	// "same" 应该被删除
	if _, exists := result["same"]; exists {
		t.Error("Expected 'same' key to be removed")
	}

	// 其他键应该保留
	if result["hello"] != "Hello" {
		t.Error("Expected 'hello' key to be preserved")
	}
}
