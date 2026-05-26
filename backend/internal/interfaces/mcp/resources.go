package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerResources registers MCP resource templates on the server.
func registerResources(s *mcpsdk.Server, pageRH *PageResourceHandler) {
	s.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: "mossbase://spaces/{spaceID}/pages/{pageID}",
		Name:        "MossBase Page",
		Description: "知识库页面内容（Markdown 格式）",
		MIMEType:    "text/markdown",
	}, pageRH.ReadPageResource)
}
