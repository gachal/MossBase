# RAG 服务架构文档

## 1. 系统概览

MossBase RAG（Retrieval-Augmented Generation）服务是一个独立的微服务，负责文档的索引和语义搜索。它通过 REST API 与 MossBase Backend 交互，使用 Qdrant 作为向量数据库存储文档嵌入向量。

### 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         MossBase 系统架构                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐       ┌──────────────┐       ┌──────────────────┐    │
│  │          │ HTTP  │              │ HTTP  │                  │    │
│  │ Frontend │──────>│   Backend    │──────>│   RAG Service    │    │
│  │  (:80)   │<──────│   (:8033)    │<──────│    (:8090)       │    │
│  │          │       │              │       │                  │    │
│  └──────────┘       └──────┬───────┘       └───┬──────┬───────┘    │
│                            │                   │      │            │
│                            │                   │      │            │
│                            v                   v      v            │
│                     ┌──────────┐        ┌───────┐ ┌──────────┐    │
│                     │          │        │       │ │          │    │
│                     │  MySQL   │        │Qdrant │ │  Redis   │    │
│                     │ (:3306)  │        │(:6333)│ │ (:6379)  │    │
│                     │          │        │       │ │          │    │
│                     └──────────┘        └───────┘ └──────────┘    │
│                                                                     │
│                            ┌──────────────────┐                    │
│                            │ Embedding Provider│                    │
│                            │ (OpenAI / 自定义)  │                    │
│                            └──────────────────┘                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 调用链路

```
用户浏览器 → Frontend (Vue) → Backend (Go/Gin) → RAG Service (Go)
                                                  ├── Qdrant (向量存储)
                                                  ├── Redis (异步任务队列)
                                                  └── Embedding Provider (文本向量化)
```

## 2. 组件说明

### 2.1 RAG Service

独立 Go 微服务，位于 `services/rag/`，职责包括：

- **文档索引**：接收文档内容，进行文本分块（Chunking）、向量化（Embedding）、存储到 Qdrant
- **语义搜索**：接收查询文本，向量化后在 Qdrant 中检索最相关的文档片段
- **文档管理**：支持单文档的创建、删除、更新
- **异步处理**：通过 Asynq + Redis 实现文档索引的异步任务队列

关键特性：
- OpenAI 兼容 API 接口，支持任何兼容 OpenAI API 的嵌入模型提供商
- 可配置的分块策略（按字符数、重叠窗口）
- 自动创建 Qdrant Collection 和索引
- 健康检查接口，支持优雅降级

### 2.2 Qdrant

高性能向量数据库，用于：

- 存储文档块的嵌入向量（默认使用 1536 维，对应 `text-embedding-3-small`）
- 支持余弦相似度（Cosine）搜索
- 提供 gRPC（:6334）和 HTTP（:6333）双协议接口
- 内置 Web Dashboard（`http://localhost:6333/dashboard`）

Collection 结构：

```json
{
  "collection_name": "mossbase_documents",
  "vectors": {
    "size": 1536,
    "distance": "Cosine"
  },
  "payload_schema": {
    "document_id": "keyword",
    "space_id": "keyword",
    "page_id": "keyword",
    "chunk_index": "integer",
    "content": "text",
    "title": "text",
    "created_at": "integer"
  }
}
```

### 2.3 Redis / Asynq

Redis 作为 Asynq 任务队列的后端存储，用于：

- **异步文档索引**：文档分块和嵌入是耗时操作，通过任务队列异步执行
- **任务重试**：嵌入 API 调用失败时自动重试（指数退避）
- **任务状态追踪**：可通过 API 查询索引任务的处理状态

任务类型：

| 任务类型            | 说明                     | 最大重试次数 |
|---------------------|--------------------------|-------------|
| `task:embed_document` | 文档分块 + 嵌入 + 存储  | 3           |
| `task:delete_document`| 从 Qdrant 删除文档向量  | 3           |

### 2.4 Embedding Provider

支持任何 OpenAI 兼容的嵌入模型 API：

- **OpenAI 官方**：`text-embedding-3-small`、`text-embedding-ada-002` 等
- **国内中转服务**：通过配置 `base_url` 指向中转 API
- **自部署模型**：如 LocalAI、Ollama 等提供 OpenAI 兼容接口的服务
- **其他兼容服务**：Azure OpenAI、通义千问等

配置方式通过环境变量：

```bash
RAG_EMBEDDING_API_KEY=sk-xxx
RAG_EMBEDDING_BASE_URL=https://api.openai.com/v1
RAG_EMBEDDING_MODEL=text-embedding-3-small
```

## 3. 数据流

### 3.1 文档索引流程

当用户在 MossBase 中创建或更新页面时，触发以下流程：

```
┌─────────┐    ┌──────────┐    ┌──────────────┐    ┌──────────┐
│  用户    │    │ Backend  │    │ RAG Service  │    │  Qdrant  │
└────┬─────┘    └────┬─────┘    └──────┬───────┘    └────┬─────┘
     │               │                │                  │
     │ 保存页面      │                │                  │
     │──────────────>│                │                  │
     │               │                │                  │
     │               │ POST /documents│                  │
     │               │ (异步调用)      │                  │
     │               │───────────────>│                  │
     │               │                │                  │
     │               │  202 Accepted  │                  │
     │               │<───────────────│                  │
     │               │                │                  │
     │  返回成功      │         ┌──────┴───────┐          │
     │<──────────────│         │ 1. 文本分块    │          │
     │               │         │    (Chunking) │          │
     │               │         │              │          │
     │               │         │ 2. 批量嵌入    │          │
     │               │         │ (Embedding)   │          │
     │               │         │              │          │
     │               │         │ 3. 向量上传    │          │
     │               │         │    (Upsert)   │────────>│
     │               │         └──────────────┘          │
     │               │                                   │
```

详细步骤：

1. **页面保存**：用户在前端编辑并保存页面，Backend 将内容写入 MySQL
2. **触发索引**：Backend 通过 HTTP 调用 RAG 服务的 `POST /api/v1/documents`，传入文档 ID、标题、内容、空间 ID 等信息
3. **返回确认**：RAG 服务接收请求后立即返回 `202 Accepted`，实际处理在后台异步执行
4. **文本分块**：将文档内容按配置的块大小（默认 500 字符）和重叠窗口（默认 50 字符）进行切分
5. **批量嵌入**：将每个文本块调用 Embedding API 转为向量
6. **向量存储**：将向量及其元数据（payload）通过 gRPC upsert 到 Qdrant

### 3.2 语义搜索流程

当用户在 MossBase 中进行搜索时：

```
┌─────────┐    ┌──────────┐    ┌──────────────┐    ┌──────────┐
│  用户    │    │ Backend  │    │ RAG Service  │    │  Qdrant  │
└────┬─────┘    └────┬─────┘    └──────┬───────┘    └────┬─────┘
     │               │                │                  │
     │ 搜索 "xxx"    │                │                  │
     │──────────────>│                │                  │
     │               │                │                  │
     │               │ POST /search   │                  │
     │               │───────────────>│                  │
     │               │                │                  │
     │               │         ┌──────┴───────┐          │
     │               │         │ 1. 查询嵌入    │          │
     │               │         │ (Embedding)   │          │
     │               │         │              │          │
     │               │         │ 2. 向量搜索    │          │
     │               │         │   (Search)    │────────>│
     │               │         │              │<────────│
     │               │         │ 3. 结果排序    │          │
     │               │         └──────────────┘          │
     │               │                │                  │
     │               │  搜索结果       │                  │
     │               │<───────────────│                  │
     │               │                │                  │
     │  返回结果列表  │                │                  │
     │<──────────────│                │                  │
     │               │                │                  │
```

详细步骤：

1. **用户搜索**：用户在前端输入搜索关键词
2. **转发请求**：Backend 调用 RAG 服务的 `POST /api/v1/search`，传入查询文本和过滤条件
3. **查询嵌入**：将查询文本通过 Embedding API 转为向量
4. **向量搜索**：在 Qdrant 中搜索与查询向量最相似的文档块
5. **结果排序**：按相似度得分排序，返回 top-k 结果
6. **返回结果**：RAG 服务将搜索结果返回给 Backend，Backend 再返回给前端

### 3.3 文档删除流程

```
用户删除页面 → Backend 删除 MySQL 记录 → 调用 RAG DELETE /documents/:id
→ RAG 从 Qdrant 中删除该文档所有向量块
```

## 4. 优雅降级策略

RAG 服务作为增强功能，不应影响 MossBase 核心功能的可用性。降级策略如下：

### 4.1 Backend 侧降级

```
┌─────────────────────────────────────────────────┐
│           Backend RAG 调用决策树                  │
├─────────────────────────────────────────────────┤
│                                                  │
│  RAG 是否启用？(MOSS_RAG_ENABLED)               │
│  ├── false → 跳过所有 RAG 调用                  │
│  └── true  → 尝试调用 RAG 服务                   │
│       ├── 连接成功 → 正常处理                    │
│       └── 连接失败 → 记录警告日志，继续正常流程   │
│                                                  │
│  页面保存：                                      │
│  ├── RAG 可用 → 异步调用索引 API                 │
│  └── RAG 不可用 → 仅保存到 MySQL，记录警告       │
│                                                  │
│  搜索：                                          │
│  ├── RAG 可用 → 返回语义搜索结果                 │
│  └── RAG 不可用 → 降级为数据库 LIKE 搜索         │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 4.2 RAG 服务内部降级

- **Embedding API 不可用**：任务进入重试队列，最多重试 3 次（指数退避：1s、4s、16s），全部失败后标记任务为失败状态
- **Qdrant 不可用**：健康检查返回 `degraded` 状态，搜索请求返回 `503 Service Unavailable`
- **Redis 不可用**：降级为同步处理模式（直接在 HTTP 请求中执行嵌入和存储）

### 4.3 状态监控

RAG 服务通过 `GET /health` 端点暴露组件状态：

```json
{
  "status": "healthy",
  "components": {
    "qdrant": "healthy",
    "redis": "healthy",
    "embedding": "healthy"
  }
}
```

可能的组件状态：
- `healthy`：正常工作
- `degraded`：部分功能受限
- `unhealthy`：组件不可用

## 5. 配置参考

### 5.1 RAG 服务环境变量

| 环境变量                      | 默认值                            | 说明                                         |
|-------------------------------|-----------------------------------|----------------------------------------------|
| `RAG_SERVER_PORT`             | `8090`                            | HTTP 服务监听端口                            |
| `RAG_SERVER_MODE`             | `release`                         | 运行模式：`debug` / `release`               |
| `RAG_QDRANT_HOST`             | `127.0.0.1`                       | Qdrant 主机地址                              |
| `RAG_QDRANT_PORT`             | `6334`                            | Qdrant gRPC 端口                             |
| `RAG_QDRANT_COLLECTION`       | `mossbase_documents`              | Qdrant Collection 名称                       |
| `RAG_QDRANT_API_KEY`          | (空)                              | Qdrant API Key（可选，用于鉴权）             |
| `RAG_REDIS_ADDR`              | `127.0.0.1:6379`                  | Redis 地址                                   |
| `RAG_REDIS_PASSWORD`          | (空)                              | Redis 密码（可选）                           |
| `RAG_REDIS_DB`                | `0`                               | Redis 数据库编号                             |
| `RAG_EMBEDDING_API_KEY`       | (必填)                            | Embedding API 密钥                           |
| `RAG_EMBEDDING_BASE_URL`      | `https://api.openai.com/v1`       | Embedding API 基础 URL                       |
| `RAG_EMBEDDING_MODEL`         | `text-embedding-3-small`          | 嵌入模型名称                                 |
| `RAG_EMBEDDING_DIMENSIONS`    | `1536`                            | 向量维度（需与模型匹配）                     |
| `RAG_CHUNK_SIZE`              | `500`                             | 文本分块大小（字符数）                       |
| `RAG_CHUNK_OVERLAP`           | `50`                              | 分块重叠字符数                               |
| `RAG_SEARCH_TOP_K`            | `5`                               | 默认返回搜索结果数量                         |
| `RAG_SEARCH_MIN_SCORE`        | `0.5`                             | 最低相似度阈值（0.0 ~ 1.0）                  |
| `RAG_AUTH_API_KEYS`           | (必填)                            | API 鉴权密钥，多个用逗号分隔                 |
| `RAG_LOG_LEVEL`               | `info`                            | 日志级别：`debug` / `info` / `warn` / `error`|
| `RAG_ASYNC_WORKERS`           | `4`                               | 异步任务处理并发数                           |
| `RAG_EMBEDDING_BATCH_SIZE`    | `20`                              | 单次嵌入 API 调用的最大文本块数              |
| `RAG_EMBEDDING_TIMEOUT`       | `30`                              | 嵌入 API 请求超时（秒）                      |

### 5.2 Backend RAG 相关配置

对应 `backend/configs/config.yaml` 中的 `rag` 部分：

```yaml
rag:
  enabled: false                          # 是否启用 RAG 功能
  base_url: "http://127.0.0.1:8090"      # RAG 服务地址
  api_key: "mossbase-rag-default-key"     # RAG 服务 API Key
  timeout: 30                             # 请求超时时间（秒）
```

对应环境变量：

| 环境变量              | 默认值                           | 说明                  |
|-----------------------|----------------------------------|-----------------------|
| `MOSS_RAG_ENABLED`    | `false`                          | 是否启用 RAG 功能     |
| `MOSS_RAG_BASE_URL`   | `http://127.0.0.1:8090`          | RAG 服务地址          |
| `MOSS_RAG_API_KEY`    | `mossbase-rag-default-key`       | RAG 服务 API Key      |
| `MOSS_RAG_TIMEOUT`    | `30`                             | 请求超时（秒）        |

### 5.3 Embedding 模型与维度对照

| 模型名称                      | 维度  | 提供商   | 备注                     |
|-------------------------------|-------|----------|--------------------------|
| `text-embedding-3-small`      | 1536  | OpenAI   | 推荐，性价比最优         |
| `text-embedding-3-large`      | 3072  | OpenAI   | 更高精度，费用更高       |
| `text-embedding-ada-002`      | 1536  | OpenAI   | 旧版模型                 |
| 自定义模型                     | 自定义 | 自部署   | 需确保 `dimensions` 配置与模型输出维度一致 |
