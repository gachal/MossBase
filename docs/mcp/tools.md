# MCP Tools 参考

MossBase MCP Server 提供 11 个工具，分为三类：页面操作、空间管理、搜索。

## 页面工具

### page_create

在指定知识空间中创建新页面。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 目标空间 ID |
| `title` | string | 是 | 页面标题（最大 200 字符） |
| `content` | string | 否 | Markdown 内容 |
| `parent_id` | uint64 | 否 | 父页面 ID，为空则作为根页面 |

**示例：**

```json
{
  "space_id": 1,
  "title": "新页面",
  "content": "# 标题\n\n正文内容",
  "parent_id": 5
}
```

### page_get

根据 ID 获取页面详情，包含完整 Markdown 内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | uint64 | 是 | 页面 ID |

### page_update

更新页面的标题和/或内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | uint64 | 是 | 页面 ID |
| `title` | string | 否 | 新标题 |
| `content` | string | 否 | 新 Markdown 内容 |

### page_delete

删除指定页面。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | uint64 | 是 | 页面 ID |

**返回：**

```json
{
  "success": true,
  "message": "page deleted"
}
```

### page_move

移动页面到新的父级或在同级中的位置。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | uint64 | 是 | 页面 ID |
| `parent_id` | uint64 | 否 | 新父页面 ID，0 表示根级 |
| `position` | int | 否 | 在同级中的位置 |

### page_tree

获取指定空间的完整页面树结构。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 空间 ID |

**返回：** 嵌套的树形结构，每个节点包含 `id`、`title`、`slug`、`status`、`children`。

## 空间工具

### space_list

列出当前用户可访问的所有知识空间。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码，默认 1 |
| `page_size` | int | 否 | 每页数量，默认 20 |

### space_get

获取空间详情。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 空间 ID |

### space_members

列出空间的成员及其角色。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 空间 ID |

## 搜索工具

### search

在指定空间中按关键词搜索页面（MySQL FULLTEXT）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 空间 ID |
| `query` | string | 是 | 搜索关键词 |
| `page` | int | 否 | 页码，默认 1 |
| `page_size` | int | 否 | 每页数量，默认 20 |

### semantic_search

使用 RAG 进行语义搜索（需要启用 RAG 服务）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | uint64 | 是 | 空间 ID |
| `query` | string | 是 | 搜索查询 |
| `limit` | int | 否 | 最大结果数，默认 10 |
