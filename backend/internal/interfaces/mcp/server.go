package mcp

import (
	"context"
	"fmt"

	"github.com/gachal/mossbase/backend/internal/domain/repository"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gachal/mossbase/backend/internal/application/service"
)

type spaceAuthorizer struct {
	memberRepo repository.SpaceMemberRepository
	auth       *MCPAuth
}

func (a *spaceAuthorizer) checkWrite(ctx context.Context, spaceID uint64) error {
	member, err := a.memberRepo.FindBySpaceAndUser(ctx, spaceID, a.auth.UserID())
	if err != nil {
		return fmt.Errorf("access denied: not a member of space %d", spaceID)
	}
	if string(member.Role) == "viewer" {
		return fmt.Errorf("access denied: viewer cannot modify space %d", spaceID)
	}
	return nil
}

func (a *spaceAuthorizer) checkRead(ctx context.Context, spaceID uint64) error {
	_, err := a.memberRepo.FindBySpaceAndUser(ctx, spaceID, a.auth.UserID())
	if err != nil {
		return fmt.Errorf("access denied: not a member of space %d", spaceID)
	}
	return nil
}

// MCPServer wires together all MCP tools, resources, and prompts.
type MCPServer struct {
	pageSvc  service.PageService
	spaceSvc service.SpaceService
	auth     *MCPAuth
	authz    *spaceAuthorizer
}

// NewMCPServer creates a new MCPServer with the given dependencies.
func NewMCPServer(pageSvc service.PageService, spaceSvc service.SpaceService, auth *MCPAuth, memberRepo repository.SpaceMemberRepository) *MCPServer {
	authz := &spaceAuthorizer{memberRepo: memberRepo, auth: auth}
	return &MCPServer{
		pageSvc:  pageSvc,
		spaceSvc: spaceSvc,
		auth:     auth,
		authz:    authz,
	}
}

// Setup creates the MCP server, registers all features, and returns it.
func (s *MCPServer) Setup() *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "mossbase",
		Version: "v1.0.0",
	}, nil)

	pageH := NewPageToolHandler(s.pageSvc, s.auth, s.authz)
	spaceH := NewSpaceToolHandler(s.spaceSvc, s.auth, s.authz)
	searchH := NewSearchToolHandler(s.pageSvc, s.auth, s.authz)
	pageRH := NewPageResourceHandler(s.pageSvc, s.authz)
	wikiH := NewWikiPromptHandler(s.pageSvc, s.authz)

	registerTools(server, pageH, spaceH, searchH)
	registerResources(server, pageRH)
	registerPrompts(server, wikiH)

	return server
}
