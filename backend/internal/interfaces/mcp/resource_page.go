package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gachal/mossbase/backend/internal/application/service"
)

// PageResourceHandler exposes MossBase pages as MCP resources.
type PageResourceHandler struct {
	pageSvc service.PageService
	authz   *spaceAuthorizer
}

// NewPageResourceHandler creates a new PageResourceHandler.
func NewPageResourceHandler(pageSvc service.PageService, authz *spaceAuthorizer) *PageResourceHandler {
	return &PageResourceHandler{pageSvc: pageSvc, authz: authz}
}

// ReadPageResource handles reads for the mossbase://spaces/{spaceID}/pages/{pageID} URI template.
func (h *PageResourceHandler) ReadPageResource(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
	uri := req.Params.URI

	spaceID, pageID, err := parsePageURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid resource URI %q: %w", uri, err)
	}

	if err := h.authz.checkRead(ctx, spaceID); err != nil {
		return nil, err
	}

	resp, err := h.pageSvc.GetByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to read page %d: %w", pageID, err)
	}

	return &mcpsdk.ReadResourceResult{
		Contents: []*mcpsdk.ResourceContents{
			{
				URI:      uri,
				MIMEType: "text/markdown",
				Text:     fmt.Sprintf("# %s\n\n%s", resp.Title, resp.Content),
			},
		},
	}, nil
}

// parsePageURI extracts spaceID and pageID from a mossbase://spaces/{spaceID}/pages/{pageID} URI.
func parsePageURI(uri string) (spaceID uint64, pageID uint64, err error) {
	parts := strings.Split(strings.TrimPrefix(uri, "mossbase://"), "/")
	if len(parts) != 4 || parts[0] != "spaces" || parts[2] != "pages" {
		return 0, 0, fmt.Errorf("expected format mossbase://spaces/{{spaceID}}/pages/{{pageID}}")
	}
	spaceID, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid space ID %q: %w", parts[1], err)
	}
	pageID, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page ID %q: %w", parts[3], err)
	}
	return spaceID, pageID, nil
}
