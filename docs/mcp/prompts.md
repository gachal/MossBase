# MCP Prompts 参考

MossBase MCP Server 提供 4 个内置提示词，帮助 AI 工具更好地利用知识库。

## summarize_page

总结指定知识库页面的核心内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | string | 是 | 页面 ID |
| `space_id` | string | 是 | 空间 ID |

**效果：** 获取页面内容，生成中文要点总结。

## search_and_answer

在知识库中搜索相关内容并回答问题。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `question` | string | 是 | 用户问题 |
| `space_id` | string | 是 | 空间 ID |

**效果：** 先搜索相关页面，将搜索结果作为上下文生成回答。

## explain_page

用通俗语言解释知识库页面内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page_id` | string | 是 | 页面 ID |
| `space_id` | string | 是 | 空间 ID |
| `audience` | string | 否 | 受众级别：beginner/intermediate/expert（默认 beginner） |

**效果：** 根据受众级别调整解释深度和用词。

## create_from_outline

根据 Markdown 大纲扩展为完整的页面内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `space_id` | string | 是 | 空间 ID |
| `outline` | string | 是 | Markdown 大纲 |
| `parent_id` | string | 否 | 父页面 ID |

**效果：** 将大纲的每个标题展开为完整段落，生成结构化的 Markdown 内容。
