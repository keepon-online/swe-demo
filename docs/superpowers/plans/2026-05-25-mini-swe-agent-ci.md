# Mini SWE Agent CI 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 为仓库增加一条由 PR 评论 `/mini-swe` 触发的只读 `mini-swe-agent` 审查工作流，并把结果回复到 PR 评论区。

**架构：** 使用 `issue_comment` 事件作为入口，先在 workflow 中做 PR 上下文、成员权限和 fork 安全校验，再 checkout PR head 代码并运行 `mini-swe-agent`。agent 使用仓库内的审查型 YAML 配置，最终把 trajectory 中的 submission 字段回帖到 PR。

**技术栈：** GitHub Actions、`actions/github-script`、`actions/checkout`、`astral-sh/setup-uv`、`uvx mini-swe-agent`

---

### 任务 1：新增 agent 审查配置

**文件：**
- 创建：`.mini-swe-agent/review.yaml`

- [ ] **步骤 1：编写只读审查型 agent 配置**

```yaml
agent:
  instance_template: |
    You are reviewing a pull request in a local git checkout.

    <task>
    {{task}}
    </task>

    Rules:
    - Do not modify tracked repository files.
    - You may create temporary files under /tmp if needed.
    - Focus on correctness bugs, regressions, security risks, and missing tests.
    - Ignore style-only nits unless they hide a real defect.

    Workflow:
    1. Inspect the changed code and relevant surrounding files.
    2. Run focused read-only validation commands if useful.
    3. Prepare a concise markdown review.

    Write the final markdown review to /tmp/mini-swe-review.md using this format:

    ## mini-swe-agent review
    ### Findings
    - [high|medium|low] path:line - issue and why it matters

    ### Tests
    - Commands run or "Not run"

    ### Verdict
    - One short summary sentence

    If you found no material issues, say "No material findings."

    When you are completely done, submit exactly:
    `echo COMPLETE_TASK_AND_SUBMIT_FINAL_OUTPUT && cat /tmp/mini-swe-review.md`

model:
  cost_tracking: ignore_errors
```

- [ ] **步骤 2：人工检查配置是否与只读审查目标一致**

检查点：
- 没有要求 agent 修改源码
- 最终 submission 输出的是 markdown 审查结果
- 成本跟踪容错已启用

### 任务 2：新增 GitHub Actions 工作流

**文件：**
- 创建：`.github/workflows/mini-swe-agent.yml`

- [ ] **步骤 1：编写 workflow 入口与权限**

```yaml
name: mini-swe-agent

on:
  issue_comment:
    types: [created]

permissions:
  contents: read
  issues: write
  pull-requests: read
```

- [ ] **步骤 2：添加触发条件、PR 元数据读取和 fork 防护**

```yaml
jobs:
  review:
    if: >-
      ${{
        github.event.issue.pull_request &&
        startsWith(github.event.comment.body, '/mini-swe') &&
        contains(fromJSON('["OWNER","MEMBER","COLLABORATOR"]'), github.event.comment.author_association)
      }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/github-script@v7
        id: pr
        with:
          script: |
            const {owner, repo} = context.repo;
            const pull_number = context.issue.number;
            const {data: pr} = await github.rest.pulls.get({owner, repo, pull_number});
            core.setOutput("head_sha", pr.head.sha);
            core.setOutput("head_ref", pr.head.ref);
            core.setOutput("head_repo", pr.head.repo.full_name);
            core.setOutput("base_repo", pr.base.repo.full_name);
            core.setOutput("title", pr.title);
            core.setOutput("body", pr.body || "");
      - uses: actions/github-script@v7
        if: steps.pr.outputs.head_repo != steps.pr.outputs.base_repo
        with:
          script: |
            await github.rest.issues.createComment({
              ...context.repo,
              issue_number: context.issue.number,
              body: "mini-swe-agent skipped this PR because it comes from a fork and the workflow does not expose model secrets to forked code."
            });
      - if: steps.pr.outputs.head_repo != steps.pr.outputs.base_repo
        run: exit 0
```

- [ ] **步骤 3：添加 checkout、uv、prompt 构造和 agent 执行**

```yaml
      - uses: actions/checkout@v4
        with:
          ref: ${{ steps.pr.outputs.head_sha }}
          fetch-depth: 0

      - uses: astral-sh/setup-uv@v5

      - name: Build review prompt
        run: |
          mkdir -p .mini-swe-agent/out
          cat <<'EOF' > .mini-swe-agent/out/task.md
          PR #${{ github.event.issue.number }}
          Title: ${{ steps.pr.outputs.title }}

          PR body:
          ${{ steps.pr.outputs.body }}

          Trigger comment:
          ${{ github.event.comment.body }}

          Reviewer request:
          Review the current pull request and return only actionable review findings.
          EOF

      - name: Run mini-swe-agent
        id: run_agent
        continue-on-error: true
        env:
          MSWEA_MODEL_NAME: ${{ secrets.MSWEA_MODEL_NAME }}
          MSWEA_COST_TRACKING: ignore_errors
          OPENAI_API_KEY: ${{ secrets.THIRD_PARTY_API_KEY }}
          THIRD_PARTY_API_BASE: ${{ secrets.THIRD_PARTY_API_BASE }}
        run: |
          TASK="$(cat .mini-swe-agent/out/task.md)"
          uvx --from mini-swe-agent mini \
            -y \
            --exit-immediately \
            -m "$MSWEA_MODEL_NAME" \
            -c mini.yaml \
            -c .mini-swe-agent/review.yaml \
            -c model.model_kwargs.custom_llm_provider=openai \
            -c "model.model_kwargs.api_base=$THIRD_PARTY_API_BASE" \
            -o .mini-swe-agent/out/review.traj.json \
            -t "$TASK" \
            | tee .mini-swe-agent/out/console.log
```

- [ ] **步骤 4：提取 submission 并回帖**

```yaml
      - name: Build comment body
        if: always()
        run: |
          python3 - <<'PY'
          import json
          from pathlib import Path

          traj = Path(".mini-swe-agent/out/review.traj.json")
          comment = Path(".mini-swe-agent/out/comment.md")

          if traj.exists():
              data = json.loads(traj.read_text())
              submission = (data.get("info", {}).get("submission") or "").strip()
              exit_status = data.get("info", {}).get("exit_status") or "unknown"
              if submission:
                  body = submission
              else:
                  body = f"## mini-swe-agent review\n\nAgent finished without a review payload.\n\nExit status: `{exit_status}`"
          else:
              body = "## mini-swe-agent review\n\nAgent did not produce a trajectory file. Check workflow logs for details."

          comment.write_text(body[:60000] + ("\n\n_Truncated by workflow._" if len(body) > 60000 else ""))
          PY

      - uses: actions/github-script@v7
        if: always()
        with:
          script: |
            const fs = require("fs");
            const body = fs.readFileSync(".mini-swe-agent/out/comment.md", "utf8");
            await github.rest.issues.createComment({
              ...context.repo,
              issue_number: context.issue.number,
              body
            });
```

### 任务 3：补充 README 使用说明

**文件：**
- 修改：`README.md`

- [ ] **步骤 1：增加 mini-swe-agent workflow 使用说明**

```md
## mini-swe-agent review workflow

This repository includes a GitHub Actions workflow that runs `mini-swe-agent` for pull request review.

### Trigger

Comment `/mini-swe` on a pull request.

### Required secrets

- `MSWEA_MODEL_NAME`
- `THIRD_PARTY_API_KEY`
- `THIRD_PARTY_API_BASE`

### Notes

- Only `OWNER`, `MEMBER`, and `COLLABORATOR` comments can trigger the workflow.
- Fork pull requests are skipped to avoid exposing model secrets.
- The workflow is read-only and replies with a review comment.
```

- [ ] **步骤 2：人工检查 README 与 workflow 配置名一致**

检查点：
- secrets 名称一致
- 触发命令一致
- fork 行为说明一致

### 任务 4：验证

**文件：**
- 验证：`.github/workflows/mini-swe-agent.yml`
- 验证：`.mini-swe-agent/review.yaml`
- 验证：`README.md`

- [ ] **步骤 1：运行可用的静态校验**

运行：`command -v actionlint`
预期：如果存在则继续用它校验；如果不存在，执行备用校验

- [ ] **步骤 2：校验 workflow YAML 或进行语法回退校验**

运行：`python3 -c "import yaml, pathlib; yaml.safe_load(pathlib.Path('.github/workflows/mini-swe-agent.yml').read_text())"`
预期：成功解析；如果 `yaml` 模块不存在，则记录该限制并改为人工审查

- [ ] **步骤 3：运行现有 Go 测试确认仓库未回归**

运行：`go test ./...`
预期：PASS
