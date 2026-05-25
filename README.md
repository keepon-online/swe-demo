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

## mini-swe-agent 工作流

本仓库已经集成 `/mini-swe` 评论工作流，支持：

- Pull Request 评论触发审查
- Issue 评论触发问题分析

触发方式：

```text
/mini-swe
```

必需的 GitHub Secrets：

- `MSWEA_MODEL_NAME`
- `THIRD_PARTY_API_KEY`
- `THIRD_PARTY_API_BASE`

推荐模型配置示例：

```text
MSWEA_MODEL_NAME=openai/deepseek-v4-pro
THIRD_PARTY_API_BASE=https://api.deepseek.com
```

完整使用说明见：

- [docs/mini-swe-usage.md](docs/mini-swe-usage.md)
