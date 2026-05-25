# Gin Web Project 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 在空目录中初始化一个基于 Gin 的最小可运行 Go Web 项目。

**架构：** 入口层放在 `cmd/server`，应用装配和优雅退出逻辑集中在 `internal/app`，配置和 HTTP 路由分别独立到 `internal/config` 与 `internal/http`。只保留 `/ping` 健康接口和 `PORT` 配置，避免过早引入业务抽象。

**技术栈：** Go、Gin、标准库 `net/http`、`httptest`、`os/signal`

---

### 任务 1：编写最小测试

**文件：**
- 创建：`internal/config/config_test.go`
- 创建：`internal/http/router_test.go`

- [ ] **步骤 1：编写配置默认值失败测试**

```go
package config

import "testing"

func TestLoadUsesDefaultPortWhenEnvMissing(t *testing.T) {
	t.Setenv("PORT", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}
```

- [ ] **步骤 2：编写 `/ping` 路由失败测试**

```go
package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterServesPing(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["message"] != "pong" {
		t.Fatalf("expected message pong, got %q", body["message"])
	}
}
```

- [ ] **步骤 3：运行测试验证失败**

运行：`go test ./internal/...`
预期：FAIL，报错 `undefined: Load` 或 `undefined: NewRouter`

### 任务 2：实现配置与路由

**文件：**
- 创建：`internal/config/config.go`
- 创建：`internal/http/router.go`
- 创建：`internal/http/handler/health.go`

- [ ] **步骤 1：实现配置加载**

```go
package config

import "os"

type Config struct {
	Port string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{Port: port}
}
```

- [ ] **步骤 2：实现健康检查处理器**

```go
package handler

import "github.com/gin-gonic/gin"

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
```

- [ ] **步骤 3：实现路由注册**

```go
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/keepon-online/swe-demo/internal/http/handler"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", handler.Ping)
	return router
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：`go test ./internal/...`
预期：PASS

### 任务 3：实现启动入口与应用装配

**文件：**
- 创建：`internal/app/app.go`
- 创建：`cmd/server/main.go`

- [ ] **步骤 1：实现应用启动与优雅退出**

```go
package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keepon-online/swe-demo/internal/config"
	apphttp "github.com/keepon-online/swe-demo/internal/http"
)

func Run() error {
	cfg := config.Load()
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: apphttp.NewRouter(),
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
```

- [ ] **步骤 2：实现入口函数**

```go
package main

import (
	"log"

	"github.com/keepon-online/swe-demo/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **步骤 3：运行整体测试**

运行：`go test ./...`
预期：PASS

### 任务 4：补充项目元信息

**文件：**
- 创建：`go.mod`
- 创建：`README.md`
- 创建：`.gitignore`

- [ ] **步骤 1：初始化模块和依赖**

```go
module github.com/keepon-online/swe-demo

go 1.23.0

require github.com/gin-gonic/gin v1.10.1
```

- [ ] **步骤 2：补充 README**

```md
# swe-demo

一个基于 Gin 的最小 Go Web 项目骨架。

## Run

```bash
go run ./cmd/server
```

设置端口：

```bash
PORT=9090 go run ./cmd/server
```

健康检查：

```bash
curl http://localhost:8080/ping
```
```

- [ ] **步骤 3：补充忽略规则**

```gitignore
bin/
dist/
.DS_Store
coverage.out
```

- [ ] **步骤 4：格式化并验证**

运行：`gofmt -w ./cmd ./internal`

运行：`go test ./...`
预期：PASS
