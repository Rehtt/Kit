# Kit

[![Go Report Card](https://goreportcard.com/badge/github.com/Rehtt/Kit)](https://goreportcard.com/report/github.com/Rehtt/Kit)
[![Go Reference](https://pkg.go.dev/badge/github.com/Rehtt/Kit.svg)](https://pkg.go.dev/github.com/Rehtt/Kit)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/Rehtt/Kit.svg)](https://github.com/Rehtt/Kit/releases)
[![Go version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Rehtt/Kit)

🛠️ **Go 通用基础库** - 一个功能丰富、高性能的 Go 工具包集合，旨在提供简单、高效、实用的工具模块，帮助开发者快速构建高质量的项目。

[中文文档](./README_cn.md) | [English](./README.md)

## ✨ 特性

- 🚀 **高性能**: 针对性能优化的实现
- 🧩 **模块化设计**: 独立的模块设计，按需使用
- 🔧 **易于使用**: 简洁的 API 设计，快速上手
- 🛡️ **类型安全**: 充分利用 Go 的类型系统
- 📦 **零依赖**: 大部分模块无外部依赖
- 🔄 **积极维护**: 活跃的开发和维护

## 📦 安装

```shell
# 最新版本 (推荐)
go get github.com/Rehtt/Kit@latest

# 兼容旧版本 Go
go get github.com/Rehtt/Kit@go1.17
```

**系统要求**: Go 1.21+ (推荐) 或 Go 1.17+

## 🚀 快速开始

### Web 服务器示例

```go
package main

import (
	"fmt"
	"github.com/Rehtt/Kit/web"
	"net/http"
)

func main() {
	// 创建 Web 实例
	app := web.New()

	// 路由定义
	app.Get("/hello/:name", func(ctx *web.Context) {
		name := ctx.GetUrlPathParam("name")
		ctx.JSON(200, map[string]string{
			"message": fmt.Sprintf("你好, %s!", name),
			"status":  "success",
		})
	})

	// 中间件支持
	app.Use(func(ctx *web.Context) {
		ctx.Writer.Header().Set("X-Powered-By", "Kit")
		ctx.Next()
	})

	// 启动服务器
	fmt.Println("服务器启动在 :8080")
	http.ListenAndServe(":8080", app)
}
```

### 日志记录示例

```go
package main

import "github.com/Rehtt/Kit/log"

func main() {
	// 基础日志
	log.Info("应用启动")
	log.Error("发生错误", "error", err)
	
	// 结构化日志
	log.With("user_id", 123).Info("用户登录")
}
```

### 缓存使用示例

```go
package main

import (
	"github.com/Rehtt/Kit/cache"
	"time"
)

func main() {
	// 创建缓存实例
	c := cache.New()
	
	// 设置缓存
	c.Set("key", "value", 5*time.Minute)
	
	// 获取缓存
	if value, ok := c.Get("key"); ok {
		fmt.Println("缓存值:", value)
	}
}
```

## 🧩 模块概览

`Kit` 包含多个独立的模块，您可以根据需要选择使用：

### 核心模块

| 模块 | 描述 | 特性 |
|------|------|------|
| [web](./web) | 轻量级 Web 框架 | 路由、中间件、JSON 支持 |
| [log](./log) | 高性能日志库 | 结构化日志、多级别、高性能 |
| [cache](./cache) | 通用缓存接口 | 内存缓存、TTL 支持 |
| [db](./db) | 数据库工具 | 条件构造、查询构建 |
| [http](./http) | HTTP 客户端 | 请求封装、重试机制 |

### 工具模块

| 模块 | 描述 | 特性 |
|------|------|------|
| [file](./file) | 文件操作工具 | 文件读写、路径处理 |
| [strings](./strings) | 字符串工具 | 高性能转换、字符串处理 |
| [random](./random) | 随机数生成 | 随机字符串、数字生成 |
| [util](./util) | 通用工具 | Snowflake ID、时间工具 |
| [struct](./struct) | 结构体工具 | 比较、标签读取、转换 |

### 数据结构模块

| 模块 | 描述 | 特性 |
|------|------|------|
| [maps](./maps) | 线程安全 Map | 并发安全、高性能 |
| [queue](./queue) | 队列实现 | 多种队列类型 |
| [slice](./slice) | 切片工具 | 切片操作、函数式编程 |
| [heap](./heap) | 堆数据结构 | 优先队列、堆排序 |
| [expiredMap](./expiredMap) | 过期 Map | 自动过期、内存管理 |

### 网络与系统模块

| 模块 | 描述 | 特性 |
|------|------|------|
| [net](./net) | 网络工具 | 网络检测、IP 处理 |
| [host](./host) | 主机信息 | 系统信息获取 |
| [browser](./browser) | 浏览器工具 | 浏览器启动、控制 |
| [wireguard](./wireguard) | WireGuard 工具 | VPN 配置、管理 |

### 其他实用模块

| 模块 | 描述 | 特性 |
|------|------|------|
| [i18n](./i18n) | 国际化支持 | 多语言、本地化 |
| [yaml](./yaml) | YAML 处理 | 配置文件解析 |
| [cli](./cli) | 命令行工具 | CLI 应用构建 |
| [generate](./generate) | 代码生成 | 模板生成、代码生成 |
| [vt](./vt) | 虚拟终端 | 终端控制、颜色输出 |
| [multiplex](./multiplex) | 多路复用 | 连接复用、负载均衡 |

## 📖 详细文档

每个模块都有详细的文档和示例，请访问对应的模块目录查看：

- 📚 [API 文档](https://pkg.go.dev/github.com/Rehtt/Kit)
- 🔧 [使用示例](./examples) 
- 📝 [更新日志](./CHANGELOG.md)

## 🛠️ 开发指南

### 环境要求

- Go 1.21+ (推荐)
- Git

### 本地开发

```shell
# 克隆项目
git clone https://github.com/Rehtt/Kit.git
cd Kit

# 运行测试
go test ./...

# 构建项目
go build ./...
```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试
- 添加必要的文档注释

## 🤝 贡献

我们欢迎并感谢所有形式的贡献！

### 如何贡献

1. 🍴 Fork 本项目
2. 🔧 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 💾 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 📤 推送到分支 (`git push origin feature/amazing-feature`)
5. 🔄 创建 Pull Request

### 贡献指南

- 提交 [Issues](https://github.com/Rehtt/Kit/issues) 报告 bug 或建议新功能
- 创建 [Pull Requests](https://github.com/Rehtt/Kit/pulls) 贡献代码
- 改进文档和示例
- 分享使用经验和最佳实践

## 📊 项目状态

- ✅ **活跃维护**: 定期更新和维护
- 🔄 **持续集成**: 自动化测试和构建
- 📈 **不断改进**: 基于社区反馈持续优化
- 🛡️ **稳定可靠**: 在生产环境中经过验证

## 🙏 致谢

感谢所有为 `Kit` 项目做出贡献的开发者们！

## 📄 许可证

本项目基于 [MIT 许可证](./LICENSE) 开源。

---

<div align="center">
  <p>如果这个项目对您有帮助，请给我们一个 ⭐️ Star！</p>
  <p>Made with ❤️ by <a href="https://github.com/Rehtt">Rehtt</a></p>
</div> 
