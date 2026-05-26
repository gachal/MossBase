# MossBase MCP Server

MossBase 提供了 MCP (Model Context Protocol) 服务，使 AI 工具（Claude Desktop、Cursor 等）能直接操作 Wiki 知识库。

## 功能概览

- **11 个 Tools** — 页面 CRUD、空间管理、关键词搜索、语义搜索
- **1 个 Resource Template** — `mossbase://spaces/{spaceID}/pages/{pageID}`
- **4 个 Prompts** — 总结页面、搜索回答、解释内容、大纲扩展

## 快速开始

### 1. 启用 MCP

编辑 `configs/config.yaml`：

```yaml
mcp:
  enabled: true
  transport: "stdio"    # stdio | http | both
  http_port: 8095
  api_keys: []          # 留空则不启用认证（本地开发）
  default_user_id: 1
```

或通过环境变量：

```bash
export MOSS_MCP_ENABLED=true
export MOSS_MCP_TRANSPORT=stdio
export MOSS_MCP_API_KEYS=your-secret-key
```

### 2. 构建

```bash
cd backend
go build -o mossbase-mcp-server ./cmd/mcp-server
```

### 3. 运行

**Stdio 模式**（Claude Desktop / Cursor）：

```bash
./mossbase-mcp-server
```

**HTTP 模式**（远程部署）：

```bash
MOSS_MCP_TRANSPORT=http MOSS_MCP_HTTP_PORT=8095 ./mossbase-mcp-server
```

### 4. 配置 Claude Desktop

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`：

```json
{
  "mcpServers": {
    "mossbase": {
      "command": "/path/to/mossbase-mcp-server",
      "env": {
        "MOSS_MCP_ENABLED": "true",
        "MOSS_MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

## 传输模式

| 模式 | 用途 | 说明 |
|------|------|------|
| `stdio` | 本地 AI 工具 | Claude Desktop、Cursor 等 |
| `http` | 远程部署 | 通过 HTTP Streamable 传输 |
| `both` | 同时支持 | 两种传输同时运行 |

## 认证

- **Stdio 模式**：API Key 可选，适合本地开发
- **HTTP 模式**：建议启用 API Key，通过 `X-API-Key` Header 传递
- API Key 使用 `crypto/subtle.ConstantTimeCompare` 防止时序攻击

## 文档索引

- [工具参考](tools.md) — 11 个 MCP Tool 的详细说明
- [资源参考](resources.md) — 页面资源 URI 模板
- [提示词参考](prompts.md) — 4 个内置 Prompt
- [部署指南](deployment.md) — Docker 部署配置
