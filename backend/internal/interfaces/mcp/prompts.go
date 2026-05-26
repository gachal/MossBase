package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPrompts registers MCP prompt handlers on the server.
func registerPrompts(s *mcpsdk.Server, wikiH *WikiPromptHandler) {
	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "summarize_page",
		Description: "总结指定知识库页面的核心内容",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "page_id", Description: "页面 ID", Required: true},
			{Name: "space_id", Description: "空间 ID", Required: true},
		},
	}, wikiH.SummarizePage)

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "search_and_answer",
		Description: "在知识库中搜索相关内容并回答问题",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "question", Description: "用户问题", Required: true},
			{Name: "space_id", Description: "空间 ID", Required: true},
		},
	}, wikiH.SearchAndAnswer)

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "explain_page",
		Description: "用通俗语言解释知识库页面内容",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "page_id", Description: "页面 ID", Required: true},
			{Name: "space_id", Description: "空间 ID", Required: true},
			{Name: "audience", Description: "受众级别：beginner/intermediate/expert", Required: false},
		},
	}, wikiH.ExplainPage)

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "create_from_outline",
		Description: "根据大纲创建结构化页面内容",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "space_id", Description: "空间 ID", Required: true},
			{Name: "outline", Description: "Markdown 大纲", Required: true},
			{Name: "parent_id", Description: "父页面 ID", Required: false},
		},
	}, wikiH.CreateFromOutline)
}
