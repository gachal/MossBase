# MCP Server 部署指南

## Docker Compose 部署

MossBase MCP Server 已集成到 Docker Compose 配置中。

### 1. 配置环境变量

编辑项目根目录的 `.env` 文件：

```bash
# MCP 配置
MCP_API_KEY=your-secure-api-key
MYSQL_ROOT_PASSWORD=your-mysql-password
JWT_SECRET=your-jwt-secret
RAG_API_KEY=your-rag-api-key
```

### 2. 启动服务

```bash
docker-compose up mcp-server
```

MCP Server 默认以 HTTP 模式运行，监听 8095 端口。

### 3. 验证

```bash
curl -X POST http://localhost:8095/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-api-key" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}'
```

## Docker 服务配置

`docker-compose.yml` 中的 MCP Server 服务：

```yaml
mcp-server:
  build:
    context: ./backend
    dockerfile: Dockerfile
  command: ["/usr/local/bin/mossbase-mcp-server"]
  ports:
    - "8095:8095"
  environment:
    MOSS_MCP_ENABLED: "true"
    MOSS_MCP_TRANSPORT: http
    MOSS_MCP_HTTP_PORT: 8095
    MOSS_MCP_API_KEYS: ${MCP_API_KEY:-changeme-mcp-key}
    # ... 数据库和 RAG 配置
  depends_on:
    mysql:
      condition: service_healthy
```

## 本地开发部署

### Stdio 模式（Claude Desktop）

1. 构建：

```bash
cd backend
go build -o mossbase-mcp-server ./cmd/mcp-server
```

2. 配置 Claude Desktop（`~/Library/Application Support/Claude/claude_desktop_config.json`）：

```json
{
  "mcpServers": {
    "mossbase": {
      "command": "/absolute/path/to/mossbase-mcp-server",
      "env": {
        "MOSS_MCP_ENABLED": "true",
        "MOSS_MCP_TRANSPORT": "stdio",
        "MOSS_DATABASE_HOST": "127.0.0.1",
        "MOSS_DATABASE_PORT": "3306",
        "MOSS_DATABASE_USERNAME": "root",
        "MOSS_DATABASE_PASSWORD": "",
        "MOSS_DATABASE_DBNAME": "mossbase"
      }
    }
  }
}
```

3. 重启 Claude Desktop。

### Cursor 配置

在 Cursor 的 MCP 设置中添加：

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

## 环境变量参考

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MOSS_MCP_ENABLED` | 是否启用 MCP | `false` |
| `MOSS_MCP_TRANSPORT` | 传输模式：stdio/http/both | `stdio` |
| `MOSS_MCP_HTTP_PORT` | HTTP 监听端口 | `8095` |
| `MOSS_MCP_API_KEYS` | API Key 列表（逗号分隔） | 空（不启用认证） |
| `MOSS_MCP_DEFAULT_USER_ID` | 默认用户 ID | `1` |

## 安全注意事项

- 生产环境务必设置 `MOSS_MCP_API_KEYS`
- HTTP 模式建议配合 HTTPS（反向代理）
- API Key 不会出现在日志中
- 支持多个 API Key 便于轮换
