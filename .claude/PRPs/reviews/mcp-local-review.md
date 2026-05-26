# 本地代码审查 — MCP Server (Phase 4)

**审查日期**: 2026-05-01
**审查范围**: backend/internal/interfaces/mcp/*, cmd/mcp-server/, configs/, docker-compose.yml
**判定**: BLOCK

---

## 汇总

| 级别 | 数量 |
|------|------|
| CRITICAL | 4 |
| HIGH | 5 |
| MEDIUM | 8 |
| LOW | 4 |

---

## CRITICAL

### C1. 认证系统已实现但未接入调用链 — 认证绕过
- **文件**: `cmd/mcp-server/main.go:89-97`, `internal/interfaces/mcp/server.go:26-43`
- `MCPAuth.Authenticate()` 存在但从未被调用。HTTP handler 忽略了 `*http.Request`，stdio 模式也无认证逻辑。所有 MCP 操作无需认证。

### C2. 完全缺失空间成员权限校验 — 越权访问
- **文件**: `tool_page.go`, `tool_space.go`, `tool_search.go`, `resource_page.go` 全部处理器
- 所有 11 个工具 + 资源读取 + 4 个提示词均未验证调用者是否为空间成员。对比 HTTP 层使用 `middleware.SpaceAuth()` 做成员校验，MCP 层完全绕过。

### C3. ListMembers 在无授权下泄露用户邮箱 — 信息泄露
- **文件**: `tool_space.go:69-76`
- `ListMembers` 直接返回 `dto.SpaceMemberResponse`（含 `Email` 字段），且无权限校验。任何能访问 MCP 端口的人都可枚举任意空间成员邮箱。

### C4. HTTP 优雅关机不完整
- **文件**: `cmd/mcp-server/main.go:88-97`
- HTTP 模式下 `http.ListenAndServe` 不感知 context 取消，`cancel()` 无法关闭监听器。进程无法优雅退出。

---

## HIGH

### H1. GetUserID 硬编码回退为管理员 ID 1
- **文件**: `tool_page.go:24,50,74`, `tool_space.go:36`
- 由于认证未注入 context，`GetUserID(ctx, 1)` 始终返回 `1`（通常是系统管理员），所有操作以管理员权限执行。

### H2. 分页/限制参数无上限约束 — DoS 风险
- **文件**: `tool_search.go:23-38`, `tool_space.go:23-56`
- `PageSize` 和 `Limit` 无最大值。攻击者可传 `page_size=1000000` 导致大量数据加载。对比 HTTP 层有 `pageSize > 100` 的校验。

### H3. Search/SemanticSearch 的 query 缺少长度校验
- **文件**: `tool_search.go:33,48`
- 搜索查询直接传入服务层，无长度限制。超长查询可能导致数据库性能问题。

### H4. docker-compose.yml 弱默认凭据
- **文件**: `docker-compose.yml`
- `MYSQL_ROOT_PASSWORD: changeme`, `JWT_SECRET: changeme-in-production`, `MCP_API_KEY: changeme-mcp-key` 等弱默认值。遗漏环境变量配置时服务以弱密码运行无警告。

### H5. HTTP 服务器无请求超时/请求体限制
- **文件**: `cmd/mcp-server/main.go:89-97`
- 使用裸 `http.ListenAndServe`，无 `ReadTimeout`/`WriteTimeout`。攻击者可通过慢速连接耗尽连接池（Slowloris）。

---

## MEDIUM

### M1. 错误信息透传底层细节
- 所有 handler 使用 `fmt.Errorf("...: %w", err)` 透传数据库错误，可能泄露表名、SQL 信息。

### M2. parsePageURI 忽略 spaceID
- `resource_page.go:50-63` — 解析 URI 时提取了 spaceID 但完全忽略，不验证页面与空间的归属关系。

### M3. 认证禁用模式无日志警告
- `auth.go:39-42` — `api_keys` 为空时静默禁用认证，无任何日志警告。

### M4. 缺少速率限制
- `cmd/mcp-server/main.go:89-97` — HTTP 传输无速率限制，可被暴力攻击。

### M5. parsePromptArgs 盲目数值转换
- `prompt_wiki.go:185-209` — 对所有数字字符串做 uint64 转换，可能误转字符串参数。

### M6. toPageTreeOutputs 递归无深度限制
- `types.go:132-144` — 深层嵌套树可能导致栈溢出。

### M7. ListSpaces 使用匿名结构体
- `tool_space.go:23-56` — 同一匿名结构体重复三次，应提取为命名类型。

### M8. config.yaml 默认 mode: debug
- `configs/config.yaml:3` — 生产默认应为 `release`。

---

## LOW

### L1. CreateFromOutline 丢弃 context
- `prompt_wiki.go:154` — `_ context.Context` 忽略了取消信号。

### L2. MovePage 中 `_ = userID` 死代码
- `tool_page.go:74-75` — 提取了 userID 但显式忽略。

### L3. toSpaceOutput 应放在 types.go
- `tool_space.go:79-90` — 与其他转换函数放在一起更好。

### L4. Qdrant 使用 latest 标签
- `docker-compose.yml:31` — 应固定为具体版本号。

---

## 验证结果

| 检查 | 结果 |
|------|------|
| go vet | PASS |
| go test | PASS |
| go build | PASS |

---

## 审查文件清单

| 文件 | 类型 |
|------|------|
| backend/internal/interfaces/mcp/auth.go | 新增 |
| backend/internal/interfaces/mcp/types.go | 新增 |
| backend/internal/interfaces/mcp/tool_page.go | 新增 |
| backend/internal/interfaces/mcp/tool_space.go | 新增 |
| backend/internal/interfaces/mcp/tool_search.go | 新增 |
| backend/internal/interfaces/mcp/resource_page.go | 新增 |
| backend/internal/interfaces/mcp/prompt_wiki.go | 新增 |
| backend/internal/interfaces/mcp/server.go | 新增 |
| backend/internal/interfaces/mcp/tools.go | 新增 |
| backend/internal/interfaces/mcp/resources.go | 新增 |
| backend/internal/interfaces/mcp/prompts.go | 新增 |
| backend/internal/interfaces/mcp/*_test.go (6个) | 新增 |
| backend/cmd/mcp-server/main.go | 新增 |
| backend/configs/config.yaml | 修改 |
| backend/Dockerfile | 修改 |
| docker-compose.yml | 修改 |
| docs/mcp/*.md (5个) | 新增 |
