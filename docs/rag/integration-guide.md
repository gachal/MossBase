# RAG 集成指南：外部 LLM 接入

## 概述

MossBase RAG 服务的核心价值在于为外部大语言模型（LLM）提供知识库上下文。本文档介绍如何将 RAG 搜索 API 与各种 LLM 服务集成，实现「知识库问答」功能。

## 集成模式

```
┌──────────┐    ┌──────────────┐    ┌──────────┐    ┌──────────────┐
│  用户提问 │───>│  RAG 搜索    │───>│  上下文   │───>│  LLM 生成    │
│          │    │  (检索相关片段)│    │  注入提示 │    │  (最终回答)  │
└──────────┘    └──────────────┘    └──────────┘    └──────────────┘
```

核心流程（Retrieval-Augmented Generation）：

1. **用户提问**：获取用户输入的查询文本
2. **RAG 搜索**：调用 `POST /api/v1/search` 检索知识库中最相关的文档片段
3. **上下文注入**：将搜索结果作为上下文信息拼接到 LLM 的系统提示词或用户消息中
4. **LLM 生成**：将包含上下文的完整提示词发送给 LLM，获取最终回答

## Python 示例

### 基本集成（使用 OpenAI SDK）

```python
import os
import httpx
from openai import OpenAI

# --- 配置 ---
RAG_BASE_URL = os.getenv("MOSS_RAG_BASE_URL", "http://localhost:8090")
RAG_API_KEY = os.getenv("MOSS_RAG_API_KEY", "mossbase-rag-internal-key")
LLM_API_KEY = os.getenv("OPENAI_API_KEY")
LLM_BASE_URL = os.getenv("OPENAI_BASE_URL", "https://api.openai.com/v1")
LLM_MODEL = os.getenv("LLM_MODEL", "gpt-4o")


def search_knowledge_base(query: str, space_id: str | None = None, top_k: int = 5) -> list[dict]:
    """调用 RAG 搜索 API，返回相关文档片段。"""
    payload: dict = {
        "query": query,
        "top_k": top_k,
        "min_score": 0.5,
    }
    if space_id:
        payload["space_id"] = space_id

    response = httpx.post(
        f"{RAG_BASE_URL}/api/v1/search",
        json=payload,
        headers={
            "Content-Type": "application/json",
            "X-API-Key": RAG_API_KEY,
        },
        timeout=30.0,
    )
    response.raise_for_status()
    body = response.json()

    if body["code"] != 0:
        raise RuntimeError(f"RAG search failed: {body['message']}")

    return body["data"]["results"]


def build_context(results: list[dict]) -> str:
    """将搜索结果格式化为上下文字符串。"""
    if not results:
        return "未找到相关知识库内容。"

    context_parts = []
    for i, result in enumerate(results, 1):
        source = result.get("title", "未知文档")
        content = result.get("content", "")
        score = result.get("score", 0)
        context_parts.append(
            f"【来源 {i}】{source}（相关度: {score:.2f}）\n{content}"
        )

    return "\n\n---\n\n".join(context_parts)


def chat_with_knowledge_base(user_query: str, space_id: str | None = None) -> str:
    """完整的 RAG 问答流程。"""
    # 1. 检索相关文档
    results = search_knowledge_base(user_query, space_id=space_id)

    # 2. 构建上下文
    context = build_context(results)

    # 3. 注入上下文并调用 LLM
    client = OpenAI(api_key=LLM_API_KEY, base_url=LLM_BASE_URL)

    system_prompt = (
        "你是 MossBase 知识库助手。请根据以下参考资料回答用户的问题。\n"
        "如果参考资料中没有相关信息，请明确说明，不要编造内容。\n"
        "回答时请引用来源编号。\n\n"
        f"--- 参考资料 ---\n{context}"
    )

    response = client.chat.completions.create(
        model=LLM_MODEL,
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_query},
        ],
        temperature=0.3,
        max_tokens=2048,
    )

    return response.choices[0].message.content


# --- 使用示例 ---
if __name__ == "__main__":
    answer = chat_with_knowledge_base(
        "MossBase 的技术架构是怎样的？",
        space_id="space-abc123",
    )
    print(answer)
```

### 多轮对话集成

```python
def chat_session_with_rag(space_id: str | None = None):
    """支持多轮对话的 RAG 问答会话。"""
    client = OpenAI(api_key=LLM_API_KEY, base_url=LLM_BASE_URL)
    conversation_history: list[dict] = []

    print("MossBase 知识库助手（输入 'quit' 退出）")
    print("-" * 50)

    while True:
        user_input = input("\n用户: ").strip()
        if user_input.lower() == "quit":
            break

        # 每轮对话都检索最新上下文
        results = search_knowledge_base(user_input, space_id=space_id)
        context = build_context(results)

        # 系统提示词（保持不变）
        system_prompt = (
            "你是 MossBase 知识库助手。根据参考资料回答问题。\n"
            "如果资料中没有相关信息，请明确说明。\n"
            f"--- 参考资料 ---\n{context}"
        )

        # 更新对话历史
        conversation_history.append({"role": "user", "content": user_input})

        # 调用 LLM（每次都带上完整历史）
        messages = [{"role": "system", "content": system_prompt}] + conversation_history
        response = client.chat.completions.create(
            model=LLM_MODEL,
            messages=messages,
            temperature=0.3,
            max_tokens=2048,
        )

        assistant_message = response.choices[0].message.content
        conversation_history.append({"role": "assistant", "content": assistant_message})

        print(f"\n助手: {assistant_message}")
```

## Go 示例

### 基本集成

```go
package ragllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config RAG + LLM 集成配置
type Config struct {
	RAGBaseURL string
	RAGAPIKey  string
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
}

// SearchResult RAG 搜索结果
type SearchResult struct {
	DocumentID string                 `json:"document_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SearchResponse RAG 搜索响应
type SearchResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    struct {
		Results []SearchResult `json:"results"`
		Total   int            `json:"total"`
	} `json:"data"`
}

// LLMMessage LLM 对话消息
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SearchKnowledgeBase 调用 RAG 搜索 API
func SearchKnowledgeBase(ctx context.Context, cfg Config, query string, spaceID string, topK int) ([]SearchResult, error) {
	payload := map[string]interface{}{
		"query":     query,
		"top_k":     topK,
		"min_score": 0.5,
	}
	if spaceID != "" {
		payload["space_id"] = spaceID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		cfg.RAGBaseURL+"/api/v1/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", cfg.RAGAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rag search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rag search failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if searchResp.Code != 0 {
		return nil, fmt.Errorf("rag search error: %s", searchResp.Message)
	}

	return searchResp.Data.Results, nil
}

// BuildContext 将搜索结果格式化为上下文字符串
func BuildContext(results []SearchResult) string {
	if len(results) == 0 {
		return "未找到相关知识库内容。"
	}

	var b strings.Builder
	for i, result := range results {
		if i > 0 {
			b.WriteString("\n\n---\n\n")
		}
		fmt.Fprintf(&b, "【来源 %d】%s（相关度: %.2f）\n%s",
			i+1, result.Title, result.Score, result.Content)
	}

	return b.String()
}

// ChatWithKnowledgeBase 完整的 RAG 问答流程
func ChatWithKnowledgeBase(ctx context.Context, cfg Config, query string, spaceID string) (string, error) {
	// 1. 检索相关文档
	results, err := SearchKnowledgeBase(ctx, cfg, query, spaceID, 5)
	if err != nil {
		return "", fmt.Errorf("search: %w", err)
	}

	// 2. 构建上下文
	context := BuildContext(results)

	// 3. 调用 LLM（此处为示例，实际可用 OpenAI Go SDK）
	systemPrompt := fmt.Sprintf(
		"你是 MossBase 知识库助手。请根据以下参考资料回答用户的问题。\n"+
			"如果参考资料中没有相关信息，请明确说明，不要编造内容。\n"+
			"回答时请引用来源编号。\n\n"+
			"--- 参考资料 ---\n%s", context)

	// 返回结构化结果供上层调用 LLM SDK
	_ = systemPrompt // 实际使用时传给 LLM SDK

	return context, nil
}
```

### 在 MossBase Backend 中集成

```go
// MossBase Backend 内部的 RAG + LLM 调用示例
// 位于 internal/application/service/ 下的 chat_service

func (s *ChatServiceImpl) AnswerQuestion(ctx context.Context, query string, spaceID string) (string, error) {
	// 1. 调用 RAG 服务搜索
	searchReq := map[string]interface{}{
		"query":     query,
		"space_id":  spaceID,
		"top_k":     5,
		"min_score": 0.5,
	}

	results, err := s.ragClient.Search(ctx, searchReq)
	if err != nil {
		// 降级：RAG 不可用时仍可调用 LLM，只是没有上下文
		s.logger.Warn("RAG search failed, falling back to LLM without context",
			zap.Error(err))
		results = nil
	}

	// 2. 构建带上下文的提示词
	context := rag.BuildContext(results)
	systemPrompt := s.buildSystemPrompt(context)

	// 3. 调用 LLM
	answer, err := s.llmClient.Chat(ctx, systemPrompt, query)
	if err != nil {
		return "", fmt.Errorf("llm call failed: %w", err)
	}

	return answer, nil
}
```

## 上下文注入最佳实践

### 1. 系统提示词模板

推荐的系统提示词结构：

```
你是 [产品名] 知识库助手。你的职责是根据提供的参考资料准确回答用户问题。

## 行为准则

1. 仅根据参考资料中的信息回答问题
2. 如果参考资料中没有相关信息，明确告知用户"参考资料中未找到相关信息"
3. 不要编造、推测或使用训练数据中的知识来补充回答
4. 引用来源时使用 [来源 N] 的格式

## 参考资料

{context}
```

### 2. Token 预算管理

不同 LLM 的上下文窗口大小不同，需要合理分配 Token 预算：

```
总上下文窗口: 128K tokens (GPT-4o)
├── 系统提示词:  ~500 tokens
├── 参考资料:    ~4,000 - 8,000 tokens (建议上限)
├── 对话历史:    ~2,000 tokens
├── 用户问题:    ~200 tokens
└── 预留回答:    ~2,000 tokens
```

控制参考资料 Token 的策略：

- 设置合理的 `top_k`（建议 3-5）
- 设置 `min_score` 过滤低质量结果（建议 0.5-0.7）
- 对每个搜索结果的 `content` 截断到最大长度（如 1000 字符）
- 使用 Token 计数器精确控制（推荐 `tiktoken` 库）

```python
import tiktoken

def truncate_to_tokens(text: str, max_tokens: int = 1000, model: str = "gpt-4o") -> str:
    """将文本截断到指定 Token 数。"""
    encoding = tiktoken.encoding_for_model(model)
    tokens = encoding.encode(text)
    if len(tokens) <= max_tokens:
        return text
    return encoding.decode(tokens[:max_tokens])
```

### 3. 搜索策略优化

**多查询扩展**：将用户问题改写为多个搜索查询以提高召回率。

```python
def expand_query(user_query: str, llm_client) -> list[str]:
    """使用 LLM 将用户问题扩展为多个搜索查询。"""
    response = llm_client.chat.completions.create(
        model="gpt-4o-mini",
        messages=[
            {"role": "system", "content": "将用户问题改写为 3 个不同的搜索查询，每行一个，不要编号。"},
            {"role": "user", "content": user_query},
        ],
        temperature=0.5,
        max_tokens=200,
    )
    queries = response.choices[0].message.content.strip().split("\n")
    return [q.strip() for q in queries if q.strip()]
```

**混合搜索**：结合 RAG 语义搜索和传统关键词搜索。

```python
def hybrid_search(query: str, space_id: str) -> list[dict]:
    """混合搜索：语义搜索 + 关键词搜索，去重合并结果。"""
    # 语义搜索
    semantic_results = search_knowledge_base(query, space_id, top_k=5)

    # 数据库关键词搜索（LIKE）
    keyword_results = db_keyword_search(query, space_id, limit=5)

    # 合并去重（按 document_id + chunk_index 去重）
    seen = set()
    merged = []
    for result in semantic_results + keyword_results:
        key = (result["document_id"], result.get("chunk_index", 0))
        if key not in seen:
            seen.add(key)
            merged.append(result)

    # 按相关度排序
    merged.sort(key=lambda x: x.get("score", 0), reverse=True)
    return merged[:10]
```

### 4. 流式响应集成

对于支持流式输出的 LLM，推荐使用流式响应提升用户体验：

```python
def stream_chat_with_rag(user_query: str, space_id: str | None = None):
    """流式 RAG 问答。"""
    # 搜索和上下文构建（同步完成）
    results = search_knowledge_base(user_query, space_id=space_id)
    context = build_context(results)

    client = OpenAI(api_key=LLM_API_KEY, base_url=LLM_BASE_URL)

    system_prompt = f"你是知识库助手。\n--- 参考资料 ---\n{context}"

    # 流式调用
    stream = client.chat.completions.create(
        model=LLM_MODEL,
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_query},
        ],
        temperature=0.3,
        max_tokens=2048,
        stream=True,
    )

    for chunk in stream:
        delta = chunk.choices[0].delta
        if delta.content:
            yield delta.content
```

### 5. 错误处理与降级

```python
def resilient_chat_with_rag(user_query: str) -> str:
    """带降级策略的 RAG 问答。"""
    # 尝试 RAG 搜索
    try:
        results = search_knowledge_base(user_query, top_k=3)
        context = build_context(results)
    except (httpx.ConnectError, httpx.TimeoutException) as e:
        # RAG 服务不可用，降级为无上下文模式
        print(f"[WARN] RAG unavailable: {e}")
        context = "（知识库搜索不可用，请根据你的通用知识回答。）"

    system_prompt = f"你是知识库助手。\n--- 参考资料 ---\n{context}"

    try:
        client = OpenAI(api_key=LLM_API_KEY, base_url=LLM_BASE_URL)
        response = client.chat.completions.create(
            model=LLM_MODEL,
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_query},
            ],
            temperature=0.3,
            max_tokens=2048,
        )
        return response.choices[0].message.content
    except Exception as e:
        return f"抱歉，AI 服务暂时不可用：{e}"
```

## 与不同 LLM 提供商的集成

### OpenAI (GPT-4o / GPT-4o-mini)

```python
from openai import OpenAI

client = OpenAI()  # 自动读取 OPENAI_API_KEY 环境变量
```

### Anthropic (Claude)

```python
import anthropic

client = anthropic.Anthropic()
response = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=2048,
    system=system_prompt,
    messages=[{"role": "user", "content": user_query}],
)
```

### 国内模型（通义千问 / 智谱 / DeepSeek）

```python
# 以 DeepSeek 为例，使用 OpenAI 兼容接口
from openai import OpenAI

client = OpenAI(
    api_key="your-deepseek-api-key",
    base_url="https://api.deepseek.com/v1",
)
# 调用方式与 OpenAI 完全一致
```

### Ollama（本地部署）

```python
from openai import OpenAI

client = OpenAI(
    api_key="ollama",  # Ollama 不需要真实 API Key
    base_url="http://localhost:11434/v1",
)
```
