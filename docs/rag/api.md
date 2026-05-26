# RAG 服务 API 参考

## 基本信息

| 项目     | 值                                   |
|----------|--------------------------------------|
| 基础 URL | `http://localhost:8090`              |
| 协议     | HTTP/REST                            |
| 认证方式 | `X-API-Key` 请求头                   |
| 内容类型 | `application/json`                   |
| 字符编码 | UTF-8                                |

## 认证

所有 `/api/v1/` 路径下的接口均需要在请求头中携带 API Key：

```
X-API-Key: mossbase-rag-internal-key
```

API Key 通过环境变量 `RAG_AUTH_API_KEYS` 配置，支持多个 Key（用逗号分隔）。未携带或 Key 无效时返回 `401 Unauthorized`。

## 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 40001,
  "message": "错误描述信息"
}
```

### HTTP 状态码

| 状态码 | 含义               | 说明                               |
|--------|--------------------|------------------------------------|
| `200`  | OK                 | 请求成功                           |
| `202`  | Accepted           | 请求已接受，异步处理中             |
| `400`  | Bad Request        | 请求参数错误                       |
| `401`  | Unauthorized       | API Key 缺失或无效                 |
| `404`  | Not Found          | 资源不存在                         |
| `500`  | Internal Error     | 服务内部错误                       |
| `503`  | Service Unavailable| 依赖服务不可用（Qdrant/Redis）     |

### 错误码参考

| 错误码   | 说明                             |
|----------|----------------------------------|
| `0`      | 成功                             |
| `40001`  | 请求参数校验失败                 |
| `40002`  | 文档内容为空                     |
| `40101`  | API Key 缺失                     |
| `40102`  | API Key 无效                     |
| `40401`  | 文档不存在                       |
| `50001`  | Qdrant 连接失败                  |
| `50002`  | Embedding API 调用失败           |
| `50003`  | Redis 连接失败                   |
| `50301`  | 服务降级中，部分功能不可用       |

---

## 接口详情

### 1. 索引文档

将文档内容分块、向量化并存入 Qdrant。异步处理，立即返回。

```
POST /api/v1/documents
```

#### 请求参数

| 字段          | 类型     | 必填 | 说明                                     |
|---------------|----------|------|------------------------------------------|
| `document_id` | string   | 是   | 文档唯一标识（对应 MossBase 的 page_id） |
| `space_id`    | string   | 是   | 空间 ID，用于搜索过滤                    |
| `title`       | string   | 是   | 文档标题                                 |
| `content`     | string   | 是   | 文档内容（Markdown 纯文本）              |
| `metadata`    | object   | 否   | 自定义元数据，会随向量一起存储           |

#### 请求示例

```bash
curl -X POST http://localhost:8090/api/v1/documents \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mossbase-rag-internal-key" \
  -d '{
    "document_id": "page-550e8400-e29b-41d4-a716-446655440000",
    "space_id": "space-abc123",
    "title": "MossBase 架构设计",
    "content": "# MossBase 架构设计\n\n## 概述\n\nMossBase 采用 DDD（领域驱动设计）架构...\n\n## 技术栈\n\n- 后端：Go + Gin + GORM\n- 前端：Vue 3 + TypeScript\n- 数据库：MySQL 8.0\n- 向量数据库：Qdrant",
    "metadata": {
      "author": "admin",
      "tags": ["架构", "设计"]
    }
  }'
```

#### 成功响应（202 Accepted）

```json
{
  "code": 0,
  "message": "document indexing task created",
  "data": {
    "document_id": "page-550e8400-e29b-41d4-a716-446655440000",
    "task_id": "task-7890abcdef",
    "status": "pending"
  }
}
```

#### 参数校验失败响应（400 Bad Request）

```json
{
  "code": 40001,
  "message": "validation failed: document_id is required"
}
```

#### 空内容响应（400 Bad Request）

```json
{
  "code": 40002,
  "message": "document content must not be empty"
}
```

---

### 2. 删除文档

从 Qdrant 中删除指定文档的所有向量块。异步处理。

```
DELETE /api/v1/documents/:id
```

#### 路径参数

| 参数 | 类型   | 说明                          |
|------|--------|-------------------------------|
| `id` | string | 文档 ID（对应索引时的 document_id） |

#### 请求示例

```bash
curl -X DELETE http://localhost:8090/api/v1/documents/page-550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: mossbase-rag-internal-key"
```

#### 成功响应（202 Accepted）

```json
{
  "code": 0,
  "message": "document deletion task created",
  "data": {
    "document_id": "page-550e8400-e29b-41d4-a716-446655440000",
    "task_id": "task-delete-abcdef12",
    "status": "pending"
  }
}
```

#### 文档不存在响应（404 Not Found）

```json
{
  "code": 40401,
  "message": "document not found: page-nonexistent"
}
```

---

### 3. 语义搜索

将查询文本向量化，在 Qdrant 中搜索最相似的文档块。同步返回结果。

```
POST /api/v1/search
```

#### 请求参数

| 字段          | 类型     | 必填 | 说明                                     |
|---------------|----------|------|------------------------------------------|
| `query`       | string   | 是   | 搜索查询文本                             |
| `space_id`    | string   | 否   | 限定搜索空间（不传则搜索所有空间）       |
| `top_k`       | integer  | 否   | 返回结果数量，默认 5，最大 20            |
| `min_score`   | float    | 否   | 最低相似度阈值（0.0 ~ 1.0），默认 0.5    |
| `metadata_filter` | object | 否 | 元数据过滤条件（Qdrant filter 语法）     |

#### 请求示例

**基本搜索：**

```bash
curl -X POST http://localhost:8090/api/v1/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mossbase-rag-internal-key" \
  -d '{
    "query": "MossBase 使用了哪些技术栈？",
    "top_k": 5
  }'
```

**限定空间搜索：**

```bash
curl -X POST http://localhost:8090/api/v1/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mossbase-rag-internal-key" \
  -d '{
    "query": "DDD 架构设计原则",
    "space_id": "space-abc123",
    "top_k": 3,
    "min_score": 0.6
  }'
```

**带元数据过滤的搜索：**

```bash
curl -X POST http://localhost:8090/api/v1/search \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mossbase-rag-internal-key" \
  -d '{
    "query": "部署指南",
    "space_id": "space-abc123",
    "top_k": 5,
    "metadata_filter": {
      "must": [
        {
          "key": "metadata.tags",
          "match": { "value": "运维" }
        }
      ]
    }
  }'
```

#### 成功响应（200 OK）

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "MossBase 使用了哪些技术栈？",
    "results": [
      {
        "document_id": "page-550e8400-e29b-41d4-a716-446655440000",
        "space_id": "space-abc123",
        "title": "MossBase 架构设计",
        "content": "## 技术栈\n\n- 后端：Go + Gin + GORM\n- 前端：Vue 3 + TypeScript\n- 数据库：MySQL 8.0\n- 向量数据库：Qdrant",
        "chunk_index": 2,
        "score": 0.89,
        "metadata": {
          "author": "admin",
          "tags": ["架构", "设计"]
        }
      },
      {
        "document_id": "page-660e8400-e29b-41d4-a716-446655440001",
        "space_id": "space-abc123",
        "title": "技术选型说明",
        "content": "为什么选择 Go 作为后端语言：高性能、编译型、并发友好...",
        "chunk_index": 0,
        "score": 0.75,
        "metadata": {}
      }
    ],
    "total": 2
  }
}
```

#### 无结果响应（200 OK）

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "query": "量子计算入门",
    "results": [],
    "total": 0
  }
}
```

#### Qdrant 不可用响应（503 Service Unavailable）

```json
{
  "code": 50001,
  "message": "qdrant connection failed: dial tcp 127.0.0.1:6334: connect: connection refused"
}
```

---

### 4. 健康检查

检查 RAG 服务及其依赖组件的健康状态。无需认证。

```
GET /health
```

#### 请求示例

```bash
curl http://localhost:8090/health
```

#### 正常响应（200 OK）

```json
{
  "status": "healthy",
  "components": {
    "qdrant": "healthy",
    "redis": "healthy",
    "embedding": "healthy"
  },
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

#### 部分降级响应（200 OK）

```json
{
  "status": "degraded",
  "components": {
    "qdrant": "healthy",
    "redis": "healthy",
    "embedding": "unhealthy"
  },
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

> 当 `embedding` 为 `unhealthy` 时，搜索仍可工作（使用已缓存的向量），但新文档索引会失败。

#### 服务不可用响应（503 Service Unavailable）

```json
{
  "status": "unhealthy",
  "components": {
    "qdrant": "unhealthy",
    "redis": "healthy",
    "embedding": "healthy"
  },
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

---

### 5. 服务信息

获取 RAG 服务的配置和状态信息。需要认证。

```
GET /api/v1/info
```

#### 请求示例

```bash
curl http://localhost:8090/api/v1/info \
  -H "X-API-Key: mossbase-rag-internal-key"
```

#### 成功响应（200 OK）

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "version": "1.0.0",
    "embedding_model": "text-embedding-3-small",
    "embedding_dimensions": 1536,
    "chunk_size": 500,
    "chunk_overlap": 50,
    "qdrant_collection": "mossbase_documents",
    "indexed_documents": 42,
    "total_chunks": 318,
    "async_workers": 4
  }
}
```

---

## 速率限制

当前版本未实施速率限制。建议在反向代理层（如 Nginx）配置请求频率控制：

- 索引接口：建议 10 次/分钟
- 搜索接口：建议 60 次/分钟
- 健康检查：无限制

## 跨域（CORS）

RAG 服务默认不配置 CORS，仅接受来自 Backend 的服务间调用。如需从浏览器直接调用，请在 RAG 服务前部署反向代理并配置 CORS 头。
