package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTools registers all MCP tool handlers on the server.
func registerTools(s *mcpsdk.Server, pageH *PageToolHandler, spaceH *SpaceToolHandler, searchH *SearchToolHandler) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_create",
		Description: "在指定知识空间中创建一个新页面",
	}, pageH.CreatePage)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_get",
		Description: "根据 ID 获取页面详情（包含 Markdown 内容）",
	}, pageH.GetPage)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_update",
		Description: "更新页面标题和/或内容",
	}, pageH.UpdatePage)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_delete",
		Description: "删除指定页面",
	}, pageH.DeletePage)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_move",
		Description: "移动页面到新的父级或位置",
	}, pageH.MovePage)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "page_tree",
		Description: "获取指定空间的完整页面树结构",
	}, pageH.GetPageTree)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "space_list",
		Description: "列出当前用户可访问的所有知识空间",
	}, spaceH.ListSpaces)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "space_get",
		Description: "获取空间详情",
	}, spaceH.GetSpace)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "space_members",
		Description: "列出空间的成员及其角色",
	}, spaceH.ListMembers)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "search",
		Description: "在指定空间中按关键词搜索页面",
	}, searchH.Search)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "semantic_search",
		Description: "使用 RAG 进行语义搜索",
	}, searchH.SemanticSearch)
}
