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

## mini-swe-agent review workflow

This repository includes a GitHub Actions workflow that runs `mini-swe-agent` for pull request review.

### Trigger

Comment `/mini-swe` on a pull request.

### Required secrets

- `MSWEA_MODEL_NAME`
- `THIRD_PARTY_API_KEY`
- `THIRD_PARTY_API_BASE`

### Notes

- Set `MSWEA_MODEL_NAME` with the provider prefix, for example `openai/your-model-name`.
- Only `OWNER`, `MEMBER`, and `COLLABORATOR` comments can trigger the workflow.
- Fork pull requests are skipped to avoid exposing model secrets.
- The workflow is read-only and replies with a review comment.
