# mini-swe 使用文档

## 概述

本仓库已经集成一个基于 GitHub Actions 的 `mini-swe-agent` 评论工作流。  
当前触发命令是：

```text
/mini-swe
```

它支持两种上下文：

1. 在 Pull Request 评论中触发  
2. 在普通 Issue 评论中触发

当前实现已经调整为 **one-shot review 模式**：

- 不再依赖 `mini-swe-agent` 的多轮 agent loop
- 不再要求模型稳定输出工具调用格式
- 由 workflow 先准备上下文，再调用模型一次性生成中文 Markdown 审查结果

这样做的原因是：在本仓库测试过程中，DeepSeek / OpenAI 兼容模型在多轮 agent 控制流下容易出现格式不稳定、无 payload 或响应对象不兼容问题；one-shot 模式更稳定，也更适合代码审查场景。

## 当前工作流位置

主工作流文件：

```text
.github/workflows/mini-swe-agent.yml
```

## 触发方式

### Pull Request 评论触发

在任意同仓库 PR 下评论：

```text
/mini-swe
```

工作流会：

1. 读取 PR 标题、正文、base/head 信息
2. 计算 PR diff、变更文件、diff 统计
3. 调用模型做一次性审查
4. 以中文 Markdown 评论回帖

### Issue 评论触发

在普通 Issue 下评论：

```text
/mini-swe
```

工作流会：

1. checkout 当前 `main`
2. 收集 `cmd/`、`internal/` 下的源码快照
3. 将 Issue 标题、Issue 正文、触发评论和源码快照交给模型
4. 让模型输出中文分析结果并评论回帖

这条路径适合：

- 让模型根据 Issue 现象定位当前 `main` 上的潜在问题
- 做轻量级 root cause analysis
- 让模型指出疑似 bug 文件和修复方向

## 权限限制

只有以下身份的评论会触发：

- `OWNER`
- `MEMBER`
- `COLLABORATOR`

也就是说，外部陌生用户的评论不会触发模型调用。

## PR 与 Issue 的行为差异

### PR 模式

输入给模型的主要内容包括：

- PR 标题
- PR 描述
- 触发评论
- base 分支名
- 变更文件列表
- diff 统计
- unified diff

PR 模式更适合做：

- 代码审查
- 回归风险识别
- 缺失测试提示

### Issue 模式

输入给模型的主要内容包括：

- Issue 标题
- Issue 正文
- 触发评论
- 当前 `main` 分支名
- `cmd/`、`internal/` 文件列表
- 当前 main 的相关源码快照

Issue 模式更适合做：

- 问题定位
- 当前代码快照的根因分析
- 给出修复建议

## Secrets 配置

当前工作流依赖以下 GitHub Actions Secrets：

```text
MSWEA_MODEL_NAME
THIRD_PARTY_API_KEY
THIRD_PARTY_API_BASE
```

### 当前推荐写法

如果使用 DeepSeek 官方 OpenAI 兼容接口：

```text
MSWEA_MODEL_NAME=openai/deepseek-v4-pro
THIRD_PARTY_API_BASE=https://api.deepseek.com
THIRD_PARTY_API_KEY=<你的密钥>
```

### 为什么 `MSWEA_MODEL_NAME` 要带 `openai/`

当前 workflow 使用的是：

- `litellm`
- `custom_llm_provider="openai"`

因此模型名使用 OpenAI 兼容 provider 前缀最稳：

```text
openai/deepseek-v4-pro
```

## 模型兼容性结论

本仓库已经做过兼容性探测，结论如下：

1. DeepSeek 官方 OpenAI 兼容接口本身可用  
2. LiteLLM 的最小调用可用  
3. 近似 review 风格提示的 LiteLLM 调用也可用  
4. 但 `mini-swe-agent` 原始多轮 agent loop 在该模型上不稳定  

因此最终落地采用 **one-shot review 模式**，而不是原始多轮 agent 模式。

## 当前已验证通过的能力

### PR 评论模式

已验证：

- `/mini-swe` 能在 PR 评论里触发
- workflow 能成功读取 PR diff
- 模型能输出真实中文审查结果
- GitHub Actions bot 能把结果评论回帖

### Issue 评论模式

已验证：

- `/mini-swe` 能在普通 Issue 评论里触发
- workflow 能读取 `main` 当前代码快照
- 模型能定位当前代码中的路由回归问题
- GitHub Actions bot 能把中文分析结果评论回帖

## 当前仓库中故意保留的测试 Bug

为了验证 Issue 模式是否能定位问题，当前仓库主分支故意保留了一个隐蔽回归：

文件：

```text
internal/http/router.go
```

当前代码把：

```go
router.GET("/ping", handler.Ping)
```

改成了：

```go
router.POST("/ping", handler.Ping)
```

这会导致：

- `GET /ping` 返回 `404`
- 健康检查失效

Issue 模式已经成功识别出这个问题，并指出根因在 `internal/http/router.go` 的路由 method 使用错误。

## 评论输出格式

当前模型被要求输出中文 Markdown，格式如下：

```md
## mini-swe-agent 审查
### 发现
- [high|medium|low] path:line - 问题及影响

### 验证
- 已运行的命令，或“未运行”

### 结论
- 一句简短总结
```

如果没有实质性问题，输出：

```md
## mini-swe-agent 审查
### 发现
- 未发现实质性问题。

### 验证
- 未运行

### 结论
- 在当前审查范围内未发现实质性问题。
```

## 工作流执行流程

### PR 路径

1. 接收 `issue_comment` 事件
2. 判断是否为 `/mini-swe`
3. 判断评论者权限
4. 判断是否为 PR
5. 拉取 PR 元信息
6. 若为 fork PR，则跳过
7. checkout PR head
8. 构建 diff 上下文
9. 调用模型一次性生成审查内容
10. 评论回帖

### Issue 路径

1. 接收 `issue_comment` 事件
2. 判断是否为 `/mini-swe`
3. 判断评论者权限
4. 判断不是 PR，而是普通 Issue
5. checkout 当前 `main`
6. 收集 `cmd/`、`internal/` 源码快照
7. 调用模型一次性生成问题分析
8. 评论回帖

## 注意事项

### 1. `pull-requests: write` 是必需的

当前 workflow 需要：

```yaml
pull-requests: write
issues: write
```

原因不是修改代码，而是为了能够在 PR / Issue 中创建评论。

### 2. 只允许同仓库 PR 深度审查

fork PR 会被跳过，原因是：

- 不向 fork 代码暴露模型密钥
- 避免第三方提交内容接触 Secrets

### 3. Issue 模式不看 diff，只看当前 `main`

Issue 模式本质上不是“PR review”，而是“根据 Issue 现象分析当前代码”。  
因此它依赖的是：

- 当前 `main` 源码快照
- Issue 提供的现象描述

### 4. `rg` 不可假设存在

GitHub Actions runner 未必安装 `rg`。  
因此 issue 模式收集文件时不要依赖：

```bash
rg --files
```

当前实现已经改为：

```bash
find cmd internal -type f | sort
```

### 5. workflow 里不要依赖 PR 分支上的 agent 配置

之前已经踩过一个坑：

- 如果 review 配置文件从 PR head 读取
- 老分支会继续带旧配置

当前实现已经避免这个问题，不再依赖 PR 分支里的 `mini-swe-agent` 配置文件做多轮控制。

### 6. one-shot 模式不等于自动修复

当前 `/mini-swe` 的目标是：

- 生成审查/分析评论
- 指出疑似根因
- 提供修复方向

当前不会：

- 自动提交代码
- 自动开 PR
- 自动修 bug

## 常见问题

### Q1：为什么不用原始 `mini-swe-agent` 多轮 agent 模式？

因为在当前模型组合下，多轮 agent loop 存在这些问题：

- 工具调用格式不稳定
- submission 不稳定
- 可能长时间无结果
- 容易出现 fallback 评论

one-shot 模式更稳，也更适合 code review / issue analysis 这种任务。

### Q2：为什么有时会看到 fallback 评论？

fallback 评论通常表示：

- 模型没有返回有效内容
- review 文件未生成
- workflow 某一步出错但最后仍走到了评论回帖

当前文案是：

```md
## mini-swe-agent 审查

本次执行未生成有效审查内容，请查看 workflow 日志获取详情。
```

### Q3：如何验证它是否正常工作？

#### 验证 PR 模式

1. 创建一个测试 PR
2. 在 PR 评论 `/mini-swe`
3. 等待 GitHub Actions 回帖

#### 验证 Issue 模式

1. 创建一个描述 bug 的 Issue
2. 在 Issue 评论 `/mini-swe`
3. 等待 GitHub Actions 回帖

## 建议的后续演进

如果后续还要继续增强，建议优先做这些：

1. 给普通 Go 测试补一个独立 CI workflow  
2. 为 issue 模式增加“只读取指定路径”的控制  
3. 为评论结果增加固定严重级别枚举与更稳定的 markdown 模板  
4. 在 README 中保留简版说明，在 `docs/` 中保留长文档

## 当前结论

当前仓库中的 `/mini-swe` 已经具备可用状态：

- PR 评论触发：可用
- Issue 评论触发：可用
- 中文评论输出：可用
- DeepSeek 官方接口：可用
- one-shot 审查模式：可用

如果只看“实战可用性”，这套配置已经可以投入日常仓库协作使用。
