# Mini SWE Agent CI Design

**目标：** 为当前 Go 项目增加一个基于 `mini-swe-agent` 的 GitHub Actions 工作流，在 PR 评论中通过 `/mini-swe` 命令触发只读代码分析，并将总结结果评论回复到对应 PR。

## 范围

本次仅实现一条受控的 agent 工作流：

- 使用 GitHub Actions `issue_comment` 事件触发
- 仅当评论目标是 Pull Request，且评论内容以 `/mini-swe` 开头时运行
- 仅允许仓库成员或协作者触发
- 仅处理同仓库分支的 PR，不处理 fork PR
- 使用 `mini-swe-agent` 进行只读分析，不修改代码，不创建提交，不创建 PR
- 运行完成后把分析结果作为评论回复到对应 PR

本次明确不包含：

- 普通 Go 测试/构建型 CI
- 自动修复、自动提交、自动开 PR
- fork PR 的 agent 分析支持
- artifact 上传和持久化运行轨迹
- 多命令路由或复杂审批流

## 触发与权限设计

工作流文件位于 `.github/workflows/mini-swe-agent.yml`。

触发条件：

- 事件：`issue_comment`
- 评论正文：以 `/mini-swe` 开头
- 评论对象：必须是 PR，而不是普通 issue

权限控制：

- 只接受 `OWNER`、`MEMBER`、`COLLABORATOR` 角色触发
- 当 PR 来自 fork 仓库时，工作流直接退出并回复一条说明，避免 secrets 暴露

最小 GitHub 权限：

- `contents: read`
- `issues: write`
- `pull-requests: read`

不授予任何代码写入权限。

## Agent 运行设计

工作流使用 `astral-sh/setup-uv` 安装 `uv`，再通过 `uvx mini-swe-agent` 执行 agent，避免在仓库中维护额外 Python 运行环境。

运行前需要：

- checkout 当前仓库
- 获取 PR 的 head 分支信息
- 切换到 PR 对应提交或分支，保证 agent 分析的是 PR 实际代码

agent 的输入包括：

- PR 标题和正文
- 触发评论正文
- PR 基础信息（编号、分支、作者）

agent 的任务目标限定为：

- 审查当前 PR 的实现
- 找出潜在缺陷、风险和缺失测试
- 输出简洁、可执行的审查结论

## 第三方中转模型配置

工作流通过 GitHub Secrets 注入模型配置，不把任何凭据写入仓库。

预期的 secrets/env 包括：

- `MSWEA_MODEL_NAME`
- `THIRD_PARTY_API_KEY`
- `THIRD_PARTY_API_BASE`

若第三方中转是 OpenAI 兼容接口，则通过仓库内的最小 agent 配置文件声明 provider/base URL 映射，使 `mini-swe-agent` 通过兼容接口访问模型。

考虑到部分中转模型缺少完整成本元数据，默认增加容错配置，避免因成本统计失败导致工作流中断。

## 输出设计

工作流结束后生成一条 PR 评论：

- 成功时：返回结构化分析总结
- 失败时：返回失败摘要和建议排查方向

评论内容保持简洁，优先给出高风险问题、行为回归风险和测试缺口，不输出冗长日志。

## 仓库文件设计

计划新增以下文件：

- `.github/workflows/mini-swe-agent.yml`
  GitHub Actions 工作流
- `.mini-swe-agent/config.yaml` 或等价最小配置文件
  保存第三方中转模型接入所需的非敏感配置
- `README.md`
  补充工作流用法、命令格式和 secrets 说明

## 失败处理

工作流应显式处理以下失败路径：

- 评论不是 PR 上下文
- 评论命令不匹配
- 触发者权限不足
- PR 来自 fork
- 模型配置缺失
- agent 执行失败

这些路径都应给出明确日志；需要回复评论的场景只覆盖“进入正式执行但失败”的情况，避免无效评论噪音。

## 测试与验证

本次以工作流静态校验和仓库级可读性验证为主：

- YAML 结构正确
- 条件表达式覆盖评论内容、PR 上下文、权限和 fork 限制
- README 文档足够让维护者完成 secrets 配置和使用

如果本地环境允许，可额外运行 YAML 语法检查；若环境不具备对应工具，则至少通过人工检查保证表达式、权限和步骤顺序自洽。

## 成功标准

满足以下条件即视为完成：

- 仓库存在可读、可维护的 `mini-swe-agent` 工作流
- 工作流只在受控 PR 评论命令下运行
- 工作流不会对 fork PR 暴露模型 secrets
- 运行结果以 PR 评论形式返回
- README 说明清楚触发方式与所需 secrets
