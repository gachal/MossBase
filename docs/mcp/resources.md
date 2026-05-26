# MCP Resources 参考

MossBase MCP Server 提供页面资源，允许 AI 工具直接读取知识库内容。

## 资源模板

### mossbase://spaces/{spaceID}/pages/{pageID}

读取指定页面的 Markdown 内容。

**URI 格式：**

```
mossbase://spaces/<spaceID>/pages/<pageID>
```

**示例：**

```
mossbase://spaces/10/pages/42
```

**MIME Type：** `text/markdown`

**返回格式：**

```markdown
# 页面标题

页面正文内容（Markdown 格式）
```

## 资源发现

AI 工具可以通过 MCP 协议的 `resources/list` 和 `resources/templates/list` 方法发现可用的资源模板。

## 使用场景

- AI 工具读取页面内容进行分析
- 基于知识库内容回答用户问题
- 自动生成页面摘要或翻译
