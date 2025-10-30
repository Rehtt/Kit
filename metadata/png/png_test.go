package png

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// createTestPNG 创建一个简单的测试PNG文件
func createTestPNG(filename string) error {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// 填充红色
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func TestWriteAndReadMetadata(t *testing.T) {
	// 创建测试PNG文件
	inputFile := "test_input.png"
	outputFile := "test_output.png"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := createTestPNG(inputFile); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 准备测试元数据
	testMetadata := []Metadata{
		{
			Keyword:           "Author",
			LanguageTag:       "zh-CN",
			TranslatedKeyword: "作者",
			Text:              "张三",
		},
		{
			Keyword:           "Description",
			LanguageTag:       "en-US",
			TranslatedKeyword: "Description",
			Text:              "This is a test image",
		},
	}

	// 写入元数据
	if err := WriteMetadata(inputFile, outputFile, testMetadata); err != nil {
		t.Fatalf("写入元数据失败: %v", err)
	}

	// 读取元数据
	readMetadata, err := ReadMetadata(outputFile)
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	// 验证元数据
	if len(readMetadata) != len(testMetadata) {
		t.Fatalf("元数据数量不匹配: 期望 %d, 实际 %d", len(testMetadata), len(readMetadata))
	}

	for i, expected := range testMetadata {
		actual := readMetadata[i]
		if actual.Keyword != expected.Keyword {
			t.Errorf("Keyword不匹配: 期望 %s, 实际 %s", expected.Keyword, actual.Keyword)
		}
		if actual.LanguageTag != expected.LanguageTag {
			t.Errorf("LanguageTag不匹配: 期望 %s, 实际 %s", expected.LanguageTag, actual.LanguageTag)
		}
		if actual.TranslatedKeyword != expected.TranslatedKeyword {
			t.Errorf("TranslatedKeyword不匹配: 期望 %s, 实际 %s", expected.TranslatedKeyword, actual.TranslatedKeyword)
		}
		if actual.Text != expected.Text {
			t.Errorf("Text不匹配: 期望 %s, 实际 %s", expected.Text, actual.Text)
		}
	}
}

func TestReadMetadataFromNonPNG(t *testing.T) {
	// 创建一个非PNG文件
	filename := "test_not_png.txt"
	defer os.Remove(filename)

	if err := os.WriteFile(filename, []byte("not a png file"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 尝试读取元数据，应该失败
	_, err := ReadMetadata(filename)
	if err == nil {
		t.Fatal("期望读取非PNG文件时返回错误")
	}
}

func TestWriteMetadataToNonPNG(t *testing.T) {
	// 创建一个非PNG文件
	inputFile := "test_not_png.txt"
	outputFile := "test_output.png"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := os.WriteFile(inputFile, []byte("not a png file"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	metadata := []Metadata{
		{
			Keyword: "Test",
			Text:    "Test",
		},
	}

	// 尝试写入元数据，应该失败
	err := WriteMetadata(inputFile, outputFile, metadata)
	if err == nil {
		t.Fatal("期望向非PNG文件写入元数据时返回错误")
	}
}

func TestReadMetadataFromEmptyPNG(t *testing.T) {
	// 创建一个没有iTXt块的PNG文件
	filename := "test_empty_metadata.png"
	defer os.Remove(filename)

	if err := createTestPNG(filename); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 读取元数据
	metadata, err := ReadMetadata(filename)
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	// 应该返回空列表
	if len(metadata) != 0 {
		t.Errorf("期望空元数据列表，实际获得 %d 条", len(metadata))
	}
}

func TestWriteMetadataWithUTF8(t *testing.T) {
	// 创建测试PNG文件
	inputFile := "test_utf8_input.png"
	outputFile := "test_utf8_output.png"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := createTestPNG(inputFile); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 准备包含UTF-8字符的元数据
	testMetadata := []Metadata{
		{
			Keyword:           "Title",
			LanguageTag:       "zh-CN",
			TranslatedKeyword: "标题",
			Text:              "这是一个测试图片 🎨",
		},
		{
			Keyword:           "Author",
			LanguageTag:       "ja-JP",
			TranslatedKeyword: "著者",
			Text:              "山田太郎",
		},
	}

	// 写入元数据
	if err := WriteMetadata(inputFile, outputFile, testMetadata); err != nil {
		t.Fatalf("写入元数据失败: %v", err)
	}

	// 读取元数据
	readMetadata, err := ReadMetadata(outputFile)
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	// 验证UTF-8字符是否正确
	for i, expected := range testMetadata {
		actual := readMetadata[i]
		if actual.Text != expected.Text {
			t.Errorf("UTF-8文本不匹配: 期望 %s, 实际 %s", expected.Text, actual.Text)
		}
		if actual.TranslatedKeyword != expected.TranslatedKeyword {
			t.Errorf("UTF-8翻译关键字不匹配: 期望 %s, 实际 %s", expected.TranslatedKeyword, actual.TranslatedKeyword)
		}
	}
}

func TestWriteEmptyMetadata(t *testing.T) {
	// 创建测试PNG文件
	inputFile := "test_empty_meta_input.png"
	outputFile := "test_empty_meta_output.png"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	if err := createTestPNG(inputFile); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 写入空元数据列表
	if err := WriteMetadata(inputFile, outputFile, []Metadata{}); err != nil {
		t.Fatalf("写入空元数据失败: %v", err)
	}

	// 读取元数据
	metadata, err := ReadMetadata(outputFile)
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	// 应该没有元数据
	if len(metadata) != 0 {
		t.Errorf("期望0条元数据，实际获得 %d 条", len(metadata))
	}
}

// 流式传输测试

func TestReadMetadataFromReader(t *testing.T) {
	// 创建测试PNG文件
	filename := "test_reader.png"
	defer os.Remove(filename)

	if err := createTestPNG(filename); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 先写入一些元数据
	testMetadata := []Metadata{
		{
			Keyword:           "Title",
			LanguageTag:       "zh-CN",
			TranslatedKeyword: "标题",
			Text:              "测试图片",
		},
	}

	outputFile := "test_reader_output.png"
	defer os.Remove(outputFile)

	if err := WriteMetadata(filename, outputFile, testMetadata); err != nil {
		t.Fatalf("写入元数据失败: %v", err)
	}

	// 使用io.Reader读取
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	reader := bytes.NewReader(data)
	metadata, err := ReadMetadataFromReader(reader)
	if err != nil {
		t.Fatalf("从Reader读取元数据失败: %v", err)
	}

	if len(metadata) != 1 {
		t.Fatalf("元数据数量不匹配: 期望 1, 实际 %d", len(metadata))
	}

	if metadata[0].Text != testMetadata[0].Text {
		t.Errorf("Text不匹配: 期望 %s, 实际 %s", testMetadata[0].Text, metadata[0].Text)
	}
}

func TestWriteMetadataToWriter(t *testing.T) {
	// 创建测试PNG文件
	filename := "test_writer_input.png"
	defer os.Remove(filename)

	if err := createTestPNG(filename); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 准备测试元数据
	testMetadata := []Metadata{
		{
			Keyword:           "Author",
			LanguageTag:       "en-US",
			TranslatedKeyword: "Author",
			Text:              "John Doe",
		},
		{
			Keyword:           "Copyright",
			LanguageTag:       "zh-CN",
			TranslatedKeyword: "版权",
			Text:              "© 2024",
		},
	}

	// 读取输入文件到内存
	inputData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("读取输入文件失败: %v", err)
	}

	// 使用io.Writer写入
	var output bytes.Buffer
	reader := bytes.NewReader(inputData)

	if err := WriteMetadataToWriter(reader, &output, testMetadata); err != nil {
		t.Fatalf("写入元数据到Writer失败: %v", err)
	}

	// 从输出中读取元数据验证
	metadata, err := ReadMetadataFromReader(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	if len(metadata) != len(testMetadata) {
		t.Fatalf("元数据数量不匹配: 期望 %d, 实际 %d", len(testMetadata), len(metadata))
	}

	for i, expected := range testMetadata {
		actual := metadata[i]
		if actual.Keyword != expected.Keyword {
			t.Errorf("Keyword不匹配: 期望 %s, 实际 %s", expected.Keyword, actual.Keyword)
		}
		if actual.Text != expected.Text {
			t.Errorf("Text不匹配: 期望 %s, 实际 %s", expected.Text, actual.Text)
		}
	}
}

func TestStreamingWithLargeMetadata(t *testing.T) {
	// 创建测试PNG文件
	filename := "test_large_meta.png"
	defer os.Remove(filename)

	if err := createTestPNG(filename); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	// 创建大量元数据
	largeText := make([]byte, 10000)
	for i := range largeText {
		largeText[i] = byte('A' + (i % 26))
	}

	testMetadata := []Metadata{
		{
			Keyword:           "Description",
			LanguageTag:       "en-US",
			TranslatedKeyword: "Description",
			Text:              string(largeText),
		},
	}

	// 使用流式API处理
	inputData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	var output bytes.Buffer
	if err := WriteMetadataToWriter(bytes.NewReader(inputData), &output, testMetadata); err != nil {
		t.Fatalf("写入大量元数据失败: %v", err)
	}

	// 验证读取
	metadata, err := ReadMetadataFromReader(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatalf("读取元数据失败: %v", err)
	}

	if len(metadata) != 1 {
		t.Fatalf("元数据数量不匹配: 期望 1, 实际 %d", len(metadata))
	}

	if metadata[0].Text != testMetadata[0].Text {
		t.Errorf("大文本不匹配")
	}
}

func TestStreamingMultipleReadWrite(t *testing.T) {
	// 测试多次读写
	filename := "test_multi_stream.png"
	defer os.Remove(filename)

	if err := createTestPNG(filename); err != nil {
		t.Fatalf("创建测试PNG失败: %v", err)
	}

	metadata1 := []Metadata{
		{
			Keyword: "Author",
			Text:    "First Author",
		},
	}

	// 第一次写入
	inputData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	var buffer1 bytes.Buffer
	if err := WriteMetadataToWriter(bytes.NewReader(inputData), &buffer1, metadata1); err != nil {
		t.Fatalf("第一次写入失败: %v", err)
	}

	// 读取验证
	result, err := ReadMetadataFromReader(bytes.NewReader(buffer1.Bytes()))
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}

	if len(result) != 1 || result[0].Text != "First Author" {
		t.Errorf("第一次验证失败")
	}

	// 第二次添加更多元数据（实际上会替换，因为我们从原始PNG开始）
	metadata2 := []Metadata{
		{
			Keyword: "Copyright",
			Text:    "Copyright 2024",
		},
	}

	var buffer2 bytes.Buffer
	if err := WriteMetadataToWriter(bytes.NewReader(inputData), &buffer2, metadata2); err != nil {
		t.Fatalf("第二次写入失败: %v", err)
	}

	result2, err := ReadMetadataFromReader(bytes.NewReader(buffer2.Bytes()))
	if err != nil {
		t.Fatalf("第二次读取失败: %v", err)
	}

	if len(result2) != 1 || result2[0].Text != "Copyright 2024" {
		t.Errorf("第二次验证失败")
	}
}
