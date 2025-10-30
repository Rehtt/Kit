# PNG Metadata

PNG 图片 iTXt 元数据读写，支持 UTF-8 和流式传输。

## 安装

```bash
go get github.com/rehtt/Kit/metadata/png
```

## 快速开始

### 文件操作

```go
import "github.com/rehtt/Kit/metadata/png"

// 写入
metadata := []png.Metadata{
    {Keyword: "Author", Text: "张三"},
    {Keyword: "Title", Text: "测试图片"},
}
png.WriteMetadata("input.png", "output.png", metadata)

// 读取
data, _ := png.ReadMetadata("output.png")
for _, m := range data {
    fmt.Printf("%s: %s\n", m.Keyword, m.Text)
}
```

### 流式操作

```go
// 从 Reader 读取
reader := bytes.NewReader(pngData)
metadata, _ := png.ReadMetadataFromReader(reader)

// 写入到 Writer
var output bytes.Buffer
png.WriteMetadataToWriter(reader, &output, metadata)
```

### HTTP 场景

```go
func handler(w http.ResponseWriter, r *http.Request) {
    file, _, _ := r.FormFile("image")
    defer file.Close()
    
    metadata := []png.Metadata{
        {Keyword: "UploadTime", Text: time.Now().String()},
    }
    
    png.WriteMetadataToWriter(file, w, metadata)
}
```

## API

### Metadata 结构

```go
type Metadata struct {
    Keyword           string  // 关键字
    LanguageTag       string  // 语言标签 (如: zh-CN, en-US)
    TranslatedKeyword string  // 翻译关键字
    Text              string  // 文本内容
}
```

### 函数

```go
// 文件操作
func ReadMetadata(filename string) ([]Metadata, error)
func WriteMetadata(inputFile, outputFile string, metadata []Metadata) error

// 流式操作
func ReadMetadataFromReader(r io.Reader) ([]Metadata, error)
func WriteMetadataToWriter(r io.Reader, w io.Writer, metadata []Metadata) error
```

## 测试

```bash
go test -v
```

## License

MIT
