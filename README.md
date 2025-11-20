# Kit

[![Go Report Card](https://goreportcard.com/badge/github.com/Rehtt/Kit)](https://goreportcard.com/report/github.com/Rehtt/Kit)
[![Go Reference](https://pkg.go.dev/badge/github.com/Rehtt/Kit.svg)](https://pkg.go.dev/github.com/Rehtt/Kit)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/Rehtt/Kit.svg)](https://github.com/Rehtt/Kit/releases)
[![Go version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Rehtt/Kit)

🛠️ **Go Universal Toolkit** - A feature-rich, high-performance Go toolkit collection that provides simple, efficient, and practical tool modules to help developers quickly build high-quality projects.

[中文文档](./README_cn.md) | [English](./README.md)

## ✨ Features

- 🚀 **High Performance**: Performance-optimized implementations
- 🧩 **Modular Design**: Independent modules, use as needed
- 🔧 **Easy to Use**: Simple API design, quick to get started
- 🛡️ **Type Safe**: Full utilization of Go's type system
- 📦 **Zero Dependencies**: Most modules have no external dependencies
- 🔄 **Actively Maintained**: Active development and maintenance

## 📦 Installation

```shell
# Latest version (recommended)
go get github.com/Rehtt/Kit@latest

# Compatible with older Go versions
go get github.com/Rehtt/Kit@go1.17
```

**Requirements**: Go 1.21+ (recommended) or Go 1.17+

## 🚀 Quick Start

### Web Server Example

```go
package main

import (
	"fmt"
	"github.com/Rehtt/Kit/web"
	"net/http"
)

func main() {
	// Create web instance
	app := web.New()

	// Route definition
	app.Get("/hello/:name", func(ctx *web.Context) {
		name := ctx.GetUrlPathParam("name")
		ctx.JSON(200, map[string]string{
			"message": fmt.Sprintf("Hello, %s!", name),
			"status":  "success",
		})
	})

	// Middleware support
	app.Use(func(ctx *web.Context) {
		ctx.Writer.Header().Set("X-Powered-By", "Kit")
		ctx.Next()
	})

	// Start server
	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", app)
}
```

### Logging Example

```go
package main

import "github.com/Rehtt/Kit/log"

func main() {
	// Basic logging
	log.Info("Application started")
	log.Error("Error occurred", "error", err)
	
	// Structured logging
	log.With("user_id", 123).Info("User logged in")
}
```

### Cache Example

```go
package main

import (
	"github.com/Rehtt/Kit/cache"
	"time"
)

func main() {
	// Create cache instance
	c := cache.New()
	
	// Set cache
	c.Set("key", "value", 5*time.Minute)
	
	// Get cache
	if value, ok := c.Get("key"); ok {
		fmt.Println("Cache value:", value)
	}
}
```

## 🧩 Module Overview

`Kit` contains multiple independent modules that you can use as needed:

### Core Modules

| Module | Description | Features |
|--------|-------------|----------|
| [web](./web) | Lightweight web framework | Routing, middleware, JSON support |
| [log](./log) | High-performance logging library | Structured logging, multiple levels, high performance |
| [cache](./cache) | Universal cache interface | In-memory cache, TTL support |
| [db](./db) | Database tools | Condition building, query construction |
| [http](./http) | HTTP client | Request wrapper, retry mechanism |

### Utility Modules

| Module | Description | Features |
|--------|-------------|----------|
| [file](./file) | File operation tools | File I/O, path handling |
| [strings](./strings) | String utilities | High-performance conversion, string processing |
| [random](./random) | Random number generation | Random strings, number generation |
| [util](./util) | General utilities | Snowflake ID, time tools |
| [struct](./struct) | Struct tools | Comparison, tag reading, conversion |

### Data Structure Modules

| Module | Description | Features |
|--------|-------------|----------|
| [maps](./maps) | Thread-safe Map | Concurrent safe, high performance |
| [queue](./queue) | Queue implementation | Multiple queue types |
| [slice](./slice) | Slice utilities | Slice operations, functional programming |
| [heap](./heap) | Heap data structure | Priority queue, heap sort |
| [expiredMap](./expiredMap) | Expiring Map | Auto expiration, memory management |

### Network & System Modules

| Module | Description | Features |
|--------|-------------|----------|
| [net](./net) | Network tools | Network detection, IP processing |
| [host](./host) | Host information | System info retrieval |
| [browser](./browser) | Browser tools | Browser launch, control |
| [wireguard](./wireguard) | WireGuard tools | VPN configuration, management |

### Other Utility Modules

| Module | Description | Features |
|--------|-------------|----------|
| [i18n](./i18n) | Internationalization support | Multi-language, localization |
| [yaml](./yaml) | YAML processing | Configuration file parsing |
| [cli](./cli) | Command line tools | CLI application building |
| [generate](./generate) | Code generation | Template generation, code generation |
| [vt](./vt) | Virtual terminal | Terminal control, color output |
| [multiplex](./multiplex) | Multiplexing | Connection multiplexing, load balancing |

## 📖 Documentation

Each module has detailed documentation and examples. Visit the corresponding module directory:

- 📚 [API Documentation](https://pkg.go.dev/github.com/Rehtt/Kit)
- 🔧 [Usage Examples](./examples) 
- 📝 [Changelog](./CHANGELOG.md)

## 🛠️ Development Guide

### Requirements

- Go 1.21+ (recommended)
- Git

### Local Development

```shell
# Clone the project
git clone https://github.com/Rehtt/Kit.git
cd Kit

# Run tests
go test ./...

# Build project
go build ./...
```

### Code Standards

- Follow Go official code standards
- Use `gofmt` to format code
- Write unit tests
- Add necessary documentation comments

## 🤝 Contributing

We welcome and appreciate all forms of contributions!

### How to Contribute

1. 🍴 Fork this project
2. 🔧 Create a feature branch (`git checkout -b feature/amazing-feature`)
3. 💾 Commit your changes (`git commit -m 'Add some amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🔄 Create a Pull Request

### Contribution Guidelines

- Submit [Issues](https://github.com/Rehtt/Kit/issues) to report bugs or suggest new features
- Create [Pull Requests](https://github.com/Rehtt/Kit/pulls) to contribute code
- Improve documentation and examples
- Share usage experiences and best practices

## 📊 Project Status

- ✅ **Actively Maintained**: Regular updates and maintenance
- 🔄 **Continuous Integration**: Automated testing and building
- 📈 **Continuous Improvement**: Ongoing optimization based on community feedback
- 🛡️ **Stable & Reliable**: Proven in production environments

## 🙏 Acknowledgments

Thanks to all developers who have contributed to the `Kit` project!

## 📄 License

This project is open source under the [MIT License](./LICENSE).

---

<div align="center">
  <p>If this project helps you, please give us a ⭐️ Star!</p>
  <p>Made with ❤️ by <a href="https://github.com/Rehtt">Rehtt</a></p>
</div>

