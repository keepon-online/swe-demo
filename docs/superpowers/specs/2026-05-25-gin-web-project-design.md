# Gin Web Project Design

**目标：** 在空目录中初始化一个基于 Gin 的最小可运行 Go Web 项目，作为后续业务开发的标准起点。

## 范围

本次仅实现最小骨架：

- 使用 Gin 作为 HTTP 框架
- 使用 `github.com/keepon-online/swe-demo` 作为 Go module 名
- 提供 `GET /ping` 健康检查接口
- 从环境变量 `PORT` 读取监听端口，默认 `8080`
- 使用 `http.Server` 启动并支持优雅退出
- 提供基础项目说明和忽略规则

本次明确不包含：

- 数据库接入
- 业务分层扩展（service/repository）
- 第三方日志框架
- 中间件体系
- Docker、热更新、CI

## 目录结构

项目采用最小但可扩展的目录布局：

- `cmd/server/main.go`
  入口文件，负责启动应用
- `internal/app/app.go`
  负责应用装配、HTTP Server 创建和优雅退出
- `internal/config/config.go`
  负责读取运行配置
- `internal/http/router.go`
  负责构建 Gin 路由
- `internal/http/handler/health.go`
  负责健康检查接口

同时补充：

- `go.mod`
- `README.md`
- `.gitignore`

## 运行设计

应用启动时先加载配置，再创建 Gin Engine 和 `http.Server`。服务默认监听 `:8080`，当设置 `PORT` 时监听 `:<PORT>`。收到 `SIGINT` 或 `SIGTERM` 后，服务在有限超时时间内执行优雅关闭。

`GET /ping` 返回固定 JSON 响应，用于确认服务存活。

## 测试设计

测试保持最小化，覆盖两个关键行为：

- 配置在未设置 `PORT` 时回退到 `8080`
- 路由 `GET /ping` 返回 `200` 和预期 JSON

## 成功标准

满足以下条件即视为完成：

- `go test ./...` 通过
- 目录结构清晰，可直接作为后续项目起点
- README 说明如何启动服务和调用 `/ping`
