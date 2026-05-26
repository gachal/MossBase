# RAG 服务部署指南

## 1. Docker Compose 部署（推荐）

### 1.1 完整服务架构

Docker Compose 编排包含以下服务：

```
┌──────────────────────────────────────────────────────────┐
│                  Docker Compose 服务                      │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  mossbase-frontend  (:80)    ← Nginx 反向代理            │
│        │                                                 │
│        v                                                 │
│  mossbase-backend   (:8033)  ← Go API 服务               │
│        │                                                 │
│        ├──> mossbase-mysql   (:3306)  ← MySQL 8.0        │
│        │                                                 │
│        └──> mossbase-rag     (:8090)  ← RAG 微服务        │
│                  │                                       │
│                  ├──> mossbase-qdrant (:6333/:6334)      │
│                  │                                       │
│                  └──> mossbase-redis  (:6379)            │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 1.2 前置准备

**创建 `.env` 文件**（项目根目录）：

```bash
# --- 必填 ---
# OpenAI API Key（用于文档嵌入向量化）
OPENAI_API_KEY=sk-your-api-key-here

# --- 可选 ---
# 使用国内中转或自部署的嵌入服务时修改此项
OPENAI_BASE_URL=https://api.openai.com/v1
```

> 如果不使用 OpenAI 官方 API，可将 `OPENAI_BASE_URL` 改为兼容服务的地址。详见 [第 4 节：Embedding 提供商配置](#4-embedding-提供商配置)。

### 1.3 启动所有服务

```bash
# 在项目根目录执行
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看 RAG 服务日志
docker-compose logs -f rag
```

### 1.4 验证部署

```bash
# 1. 检查 RAG 服务健康状态
curl http://localhost:8090/health

# 预期输出:
# {"status":"healthy","components":{"qdrant":"healthy","redis":"healthy","embedding":"healthy"},...}

# 2. 检查 Qdrant Dashboard
# 浏览器访问 http://localhost:6333/dashboard

# 3. 检查 Backend 连接
curl http://localhost:8033/api/v1/health
```

### 1.5 停止服务

```bash
# 停止所有服务（保留数据）
docker-compose down

# 停止并删除数据卷（重置所有数据）
docker-compose down -v
```

## 2. 环境变量说明

### 2.1 全局环境变量（`.env` 文件）

| 变量名              | 必填 | 默认值                          | 说明                               |
|---------------------|------|---------------------------------|------------------------------------|
| `OPENAI_API_KEY`    | 是   |                                 | OpenAI API Key 或兼容服务的密钥    |
| `OPENAI_BASE_URL`   | 否   | `https://api.openai.com/v1`     | 嵌入 API 的基础 URL                |

### 2.2 RAG 服务环境变量

在 `docker-compose.yml` 的 `rag` 服务中配置：

| 变量名                         | 默认值                          | 说明                                     |
|--------------------------------|---------------------------------|------------------------------------------|
| `RAG_SERVER_PORT`              | `8090`                          | 服务监听端口                             |
| `RAG_SERVER_MODE`              | `release`                       | 运行模式：`debug` / `release`           |
| `RAG_QDRANT_HOST`              | `qdrant`                        | Qdrant 主机（Docker 网络中用服务名）     |
| `RAG_QDRANT_PORT`              | `6334`                          | Qdrant gRPC 端口                         |
| `RAG_QDRANT_COLLECTION`        | `mossbase_documents`            | Collection 名称                          |
| `RAG_QDRANT_API_KEY`           | (空)                            | Qdrant API Key（可选）                   |
| `RAG_REDIS_ADDR`               | `redis:6379`                    | Redis 地址（Docker 网络中用服务名）      |
| `RAG_REDIS_PASSWORD`           | (空)                            | Redis 密码                               |
| `RAG_REDIS_DB`                 | `0`                             | Redis 数据库编号                         |
| `RAG_EMBEDDING_API_KEY`        | (必填，通常引用 `${OPENAI_API_KEY}`) | 嵌入 API 密钥                     |
| `RAG_EMBEDDING_BASE_URL`       | (必填，通常引用 `${OPENAI_BASE_URL}`) | 嵌入 API 基础 URL                 |
| `RAG_EMBEDDING_MODEL`          | `text-embedding-3-small`        | 嵌入模型名称                             |
| `RAG_EMBEDDING_DIMENSIONS`     | `1536`                          | 向量维度                                 |
| `RAG_CHUNK_SIZE`               | `500`                           | 文本分块大小（字符数）                   |
| `RAG_CHUNK_OVERLAP`            | `50`                            | 分块重叠字符数                           |
| `RAG_SEARCH_TOP_K`             | `5`                             | 默认搜索结果数量                         |
| `RAG_SEARCH_MIN_SCORE`         | `0.5`                           | 最低相似度阈值                           |
| `RAG_AUTH_API_KEYS`            | (必填)                          | API 鉴权密钥，逗号分隔                   |
| `RAG_LOG_LEVEL`                | `info`                          | 日志级别                                 |
| `RAG_ASYNC_WORKERS`            | `4`                             | 异步任务并发数                           |
| `RAG_EMBEDDING_BATCH_SIZE`     | `20`                            | 嵌入 API 批量大小                        |
| `RAG_EMBEDDING_TIMEOUT`        | `30`                            | 嵌入 API 超时（秒）                      |

### 2.3 Backend RAG 相关环境变量

在 `docker-compose.yml` 的 `backend` 服务中配置：

| 变量名              | 默认值                          | 说明                   |
|---------------------|---------------------------------|------------------------|
| `MOSS_RAG_ENABLED`  | `true`                          | 是否启用 RAG 功能      |
| `MOSS_RAG_BASE_URL` | `http://rag:8090`               | RAG 服务地址           |
| `MOSS_RAG_API_KEY`  | `mossbase-rag-internal-key`     | RAG API Key            |
| `MOSS_RAG_TIMEOUT`  | `30`                            | 请求超时（秒）         |

> 注意：Docker 网络中 Backend 调用 RAG 使用服务名 `rag` 而非 `localhost`。

## 3. 本地开发部署

如需在本地分别运行各服务（用于开发调试），按以下步骤操作。

### 3.1 启动依赖服务

只需通过 Docker 启动 Qdrant 和 Redis：

```bash
# 启动 Qdrant
docker run -d \
  --name mossbase-qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v qdrant_data:/qdrant/storage \
  qdrant/qdrant:latest

# 启动 Redis
docker run -d \
  --name mossbase-redis \
  -p 6379:6379 \
  redis:7-alpine
```

### 3.2 启动 RAG 服务

```bash
cd services/rag

# 设置环境变量
export RAG_SERVER_PORT=8090
export RAG_QDRANT_HOST=127.0.0.1
export RAG_QDRANT_PORT=6334
export RAG_REDIS_ADDR=127.0.0.1:6379
export RAG_EMBEDDING_API_KEY=sk-your-api-key
export RAG_EMBEDDING_BASE_URL=https://api.openai.com/v1
export RAG_EMBEDDING_MODEL=text-embedding-3-small
export RAG_AUTH_API_KEYS=mossbase-rag-default-key

# 启动
go run ./cmd/server/main.go
```

### 3.3 启动 Backend

```bash
cd backend

# 修改 configs/config.yaml 中的 rag 配置
# rag:
#   enabled: true
#   base_url: "http://127.0.0.1:8090"
#   api_key: "mossbase-rag-default-key"

# 或通过环境变量
export MOSS_RAG_ENABLED=true
export MOSS_RAG_BASE_URL=http://127.0.0.1:8090
export MOSS_RAG_API_KEY=mossbase-rag-default-key

go run ./cmd/server/main.go
```

### 3.4 启动 Frontend

```bash
cd frontend
npm install
npm run dev
```

## 4. Embedding 提供商配置

RAG 服务使用 OpenAI 兼容 API 进行文本嵌入，支持多种提供商。

### 4.1 OpenAI 官方

```bash
RAG_EMBEDDING_API_KEY=sk-your-openai-api-key
RAG_EMBEDDING_BASE_URL=https://api.openai.com/v1
RAG_EMBEDDING_MODEL=text-embedding-3-small
RAG_EMBEDDING_DIMENSIONS=1536
```

### 4.2 国内中转服务

使用国内 API 中转服务（如 API2D、OhMyGPT、AI.HPC 等）以避免网络问题：

```bash
RAG_EMBEDDING_API_KEY=your-proxy-api-key
RAG_EMBEDDING_BASE_URL=https://api.api2d.com/v1
RAG_EMBEDDING_MODEL=text-embedding-3-small
RAG_EMBEDDING_DIMENSIONS=1536
```

### 4.3 Azure OpenAI

```bash
RAG_EMBEDDING_API_KEY=your-azure-api-key
RAG_EMBEDDING_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment
RAG_EMBEDDING_MODEL=text-embedding-ada-002
RAG_EMBEDDING_DIMENSIONS=1536
```

### 4.4 LocalAI（本地部署）

```bash
RAG_EMBEDDING_API_KEY=no-key-needed
RAG_EMBEDDING_BASE_URL=http://localhost:8080/v1
RAG_EMBEDDING_MODEL=text-embedding-ada-002
RAG_EMBEDDING_DIMENSIONS=1536
```

### 4.5 Ollama（本地部署）

```bash
RAG_EMBEDDING_API_KEY=ollama
RAG_EMBEDDING_BASE_URL=http://localhost:11434/v1
RAG_EMBEDDING_MODEL=nomic-embed-text
RAG_EMBEDDING_DIMENSIONS=768
```

> 使用 Ollama 时需先拉取嵌入模型：`ollama pull nomic-embed-text`

### 4.6 模型选择建议

| 场景                 | 推荐模型                      | 维度  | 说明                 |
|----------------------|-------------------------------|-------|----------------------|
| 通用场景（推荐）     | `text-embedding-3-small`      | 1536  | 性价比最优           |
| 高精度需求           | `text-embedding-3-large`      | 3072  | 检索精度更高         |
| 纯本地部署           | `nomic-embed-text` (Ollama)   | 768   | 无需外部 API         |
| 中文优化             | `bge-m3` (Ollama)             | 1024  | 中文语义理解更好     |

## 5. Qdrant 数据持久化与备份

### 5.1 Docker 数据卷

`docker-compose.yml` 中已配置 `qdrant_data` 数据卷：

```yaml
volumes:
  qdrant_data:  # Qdrant 数据持久化

services:
  qdrant:
    volumes:
      - qdrant_data:/qdrant/storage
```

数据卷位于 Docker 默认存储路径（通常为 `/var/lib/docker/volumes/`）。

### 5.2 手动备份

```bash
# 创建快照
curl -X POST http://localhost:6333/collections/mossbase_documents/snapshots

# 下载快照文件
curl -o backup_snapshot.tar \
  http://localhost:6333/collections/mossbase_documents/snapshots/{snapshot_name}

# 或直接复制数据卷
docker cp mossbase-qdrant:/qdrant/storage ./qdrant-backup-$(date +%Y%m%d)
```

### 5.3 数据恢复

```bash
# 从快照恢复
curl -X PUT \
  http://localhost:6333/collections/mossbase_documents/snapshots/upload \
  -F "snapshot=@backup_snapshot.tar"

# 或将数据卷复制回去
docker cp ./qdrant-backup-20260430 mossbase-qdrant:/qdrant/storage
docker restart mossbase-qdrant
```

### 5.4 定期备份脚本

```bash
#!/bin/bash
# backup-qdrant.sh
# 建议加入 crontab: 0 2 * * * /path/to/backup-qdrant.sh

BACKUP_DIR="/data/backups/qdrant"
COLLECTION="mossbase_documents"
QDRANT_URL="http://localhost:6333"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# 创建快照
SNAPSHOT_RESPONSE=$(curl -s -X POST "$QDRANT_URL/collections/$COLLECTION/snapshots")
SNAPSHOT_NAME=$(echo "$SNAPSHOT_RESPONSE" | jq -r '.result.name')

if [ "$SNAPSHOT_NAME" = "null" ] || [ -z "$SNAPSHOT_NAME" ]; then
    echo "[$(date)] Failed to create snapshot" >> "$BACKUP_DIR/backup.log"
    exit 1
fi

# 下载快照
curl -o "$BACKUP_DIR/${TIMESTAMP}_${SNAPSHOT_NAME}" \
    "$QDRANT_URL/collections/$COLLECTION/snapshots/$SNAPSHOT_NAME"

echo "[$(date)] Backup completed: ${TIMESTAMP}_${SNAPSHOT_NAME}" >> "$BACKUP_DIR/backup.log"

# 保留最近 7 天的备份
find "$BACKUP_DIR" -name "*.tar" -mtime +7 -delete
```

## 6. 监控

### 6.1 健康检查

RAG 服务提供 `GET /health` 端点，可用于：

**手动检查：**

```bash
curl -s http://localhost:8090/health | jq .
```

**Docker 健康检查（在 `docker-compose.yml` 中配置）：**

```yaml
rag:
  # ... 其他配置 ...
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s
```

**外部监控集成：**

```bash
# Prometheus 抓取示例（需要 RAG 服务暴露 /metrics 端点）
# - job_name: 'mossbase-rag'
#   static_configs:
#     - targets: ['localhost:8090']
```

### 6.2 Qdrant Dashboard

Qdrant 内置 Web 管理界面，可用于：

- 查看 Collection 列表和向量数量
- 浏览向量 payload 数据
- 手动执行搜索测试
- 管理快照

访问地址：`http://localhost:6333/dashboard`

### 6.3 日志查看

```bash
# RAG 服务日志
docker-compose logs -f rag

# 只看错误日志
docker-compose logs -f rag 2>&1 | grep -i error

# Qdrant 日志
docker-compose logs -f qdrant

# 所有服务日志
docker-compose logs -f
```

### 6.4 关键监控指标

| 指标                        | 来源          | 告警阈值               |
|-----------------------------|---------------|------------------------|
| RAG 服务健康状态            | `/health`     | status != "healthy"    |
| Qdrant 连接状态             | `/health`     | qdrant != "healthy"    |
| Redis 连接状态              | `/health`     | redis != "healthy"     |
| Embedding API 可达性        | `/health`     | embedding != "healthy" |
| 向量数量                    | Qdrant API    | 异常下降               |
| 索引任务积压                | Redis 队列    | 队列长度 > 100         |

## 7. 常见问题排查

### 7.1 RAG 服务无法启动

**症状**：`docker-compose up rag` 后容器立即退出。

**排查步骤**：

```bash
# 查看容器日志
docker-compose logs rag

# 常见错误：
# 1. "connection refused" -> Qdrant 或 Redis 未就绪
# 2. "invalid api key" -> 检查 RAG_AUTH_API_KEYS 配置
# 3. "embedding api error" -> 检查 OPENAI_API_KEY 和网络连通性
```

**解决方案**：

- 确保 Qdrant 和 Redis 先于 RAG 服务启动（`depends_on` 配置）
- 添加健康检查依赖（`condition: service_healthy`）
- 检查 `.env` 文件中的 `OPENAI_API_KEY` 是否有效

### 7.2 文档索引失败

**症状**：调用 `POST /documents` 后，文档未能被搜索到。

**排查步骤**：

```bash
# 1. 检查 RAG 健康状态
curl http://localhost:8090/health

# 2. 检查 Embedding API 连通性
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# 3. 检查 Qdrant 中的向量数量
curl http://localhost:6333/collections/mossbase_documents

# 4. 查看 RAG 服务错误日志
docker-compose logs rag 2>&1 | grep -i "embed\|error\|failed"
```

**常见原因**：

| 原因                        | 解决方案                                     |
|-----------------------------|----------------------------------------------|
| Embedding API Key 无效      | 检查 `.env` 中的 `OPENAI_API_KEY`           |
| 网络不通（国内访问 OpenAI） | 配置 `OPENAI_BASE_URL` 使用中转服务         |
| API 额度不足                | 检查 OpenAI 账户余额                         |
| 文档内容为空                | 确保传入的 `content` 字段非空                |

### 7.3 搜索结果质量差

**症状**：搜索返回的结果与查询不相关。

**排查步骤**：

```bash
# 1. 尝试降低 min_score 查看更多结果
curl -X POST http://localhost:8090/api/v1/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mossbase-rag-internal-key" \
  -d '{"query": "测试查询", "min_score": 0.3, "top_k": 10}'
```

**调优建议**：

- **分块参数**：增大 `RAG_CHUNK_SIZE`（如 1000）以保留更多上下文
- **重叠窗口**：增大 `RAG_CHUNK_OVERLAP`（如 100）以减少信息丢失
- **嵌入模型**：升级到 `text-embedding-3-large` 以获得更好的语义表示
- **最低分数**：根据实际情况调整 `RAG_SEARCH_MIN_SCORE`（0.3-0.7）
- **中文优化**：考虑使用中文优化模型（如 `bge-m3`）

### 7.4 Qdrant 连接被拒绝

**症状**：RAG 服务日志中出现 `dial tcp: connection refused`。

```bash
# 检查 Qdrant 是否运行
docker-compose ps qdrant

# 检查端口是否可达
curl http://localhost:6333/collections

# 检查 Docker 网络连通性（从 RAG 容器内）
docker-compose exec rag curl -s http://qdrant:6334
```

**解决方案**：

- 确保 Qdrant 容器正在运行：`docker-compose restart qdrant`
- 检查 `RAG_QDRANT_HOST` 配置：Docker 网络中使用 `qdrant`，本地开发使用 `127.0.0.1`
- 检查端口映射：确保 `6333`（HTTP）和 `6334`（gRPC）端口已映射

### 7.5 磁盘空间不足

**症状**：Qdrant 或 Docker 磁盘空间告警。

```bash
# 查看 Docker 磁盘使用
docker system df

# 查看 Qdrant 数据卷大小
docker volume inspect mossbase_qdrant_data

# 清理未使用的 Docker 资源
docker system prune -a --volumes
# 注意：此命令会删除所有未使用的数据卷，请确认无重要数据
```

### 7.6 Redis 连接问题

**症状**：RAG 服务日志中出现 Redis 连接错误。

```bash
# 检查 Redis 是否运行
docker-compose ps redis

# 测试 Redis 连接
docker-compose exec redis redis-cli ping

# 查看 Redis 队列状态
docker-compose exec redis redis-cli LLEN "asynq:{default}:pending"
```

**解决方案**：

- 重启 Redis：`docker-compose restart redis`
- 检查 `RAG_REDIS_ADDR` 配置是否正确
- 如果 Redis 需要密码，确保 `RAG_REDIS_PASSWORD` 配置正确

## 8. 生产环境建议

### 8.1 安全配置

- **修改默认 API Key**：将 `RAG_AUTH_API_KEYS` 改为强随机字符串
- **启用 Qdrant API Key**：配置 `RAG_QDRANT_API_KEY` 保护 Qdrant 端口
- **限制端口暴露**：生产环境中不对外暴露 Qdrant（6333/6334）和 Redis（6379）端口
- **使用 HTTPS**：在 RAG 服务前部署 Nginx/Caddy 反向代理并启用 TLS
- **网络安全**：使用 Docker 内部网络隔离服务，仅暴露必要端口

### 8.2 性能配置

```yaml
# 生产环境推荐配置
rag:
  environment:
    RAG_SERVER_MODE: release
    RAG_ASYNC_WORKERS: 8           # 根据 CPU 核心数调整
    RAG_EMBEDDING_BATCH_SIZE: 50   # 根据 API 限制调整
    RAG_EMBEDDING_TIMEOUT: 60      # 增加超时
    RAG_SEARCH_TOP_K: 10          # 更多的候选结果
```

### 8.3 资源规划

| 服务      | 最小配置       | 推荐配置       | 说明                         |
|-----------|---------------|---------------|------------------------------|
| RAG       | 1 CPU / 512MB | 2 CPU / 1GB   | 嵌入计算密集                 |
| Qdrant    | 1 CPU / 1GB   | 2 CPU / 4GB   | 取决于向量数量               |
| Redis     | 0.5 CPU / 256MB | 1 CPU / 512MB | 队列数据量通常较小           |
| MySQL     | 1 CPU / 1GB   | 2 CPU / 2GB   | 取决于文档数量               |

### 8.4 扩展方案

- **RAG 服务水平扩展**：可部署多个 RAG 实例，通过负载均衡分发请求
- **Qdrant 集群**：对于大规模数据，可部署 Qdrant 集群模式
- **Redis Sentinel**：提高 Redis 可用性
