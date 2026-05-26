package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
)

type AdminServiceImpl struct {
	userRepo        repository.UserRepository
	spaceRepo       repository.SpaceRepository
	spaceMemberRepo repository.SpaceMemberRepository
	pageRepo        repository.PageRepository
	configPath      string
}

func NewAdminService(
	userRepo repository.UserRepository,
	spaceRepo repository.SpaceRepository,
	spaceMemberRepo repository.SpaceMemberRepository,
	pageRepo repository.PageRepository,
	configPath string,
) AdminService {
	return &AdminServiceImpl{userRepo: userRepo, spaceRepo: spaceRepo, spaceMemberRepo: spaceMemberRepo, pageRepo: pageRepo, configPath: configPath}
}

func (s *AdminServiceImpl) GetDashboardStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	_, spaceCount, err := s.spaceRepo.ListAll(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("count spaces: %w", err)
	}
	_, pageCount, err := s.pageRepo.ListAll(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("count pages: %w", err)
	}
	return &dto.DashboardStatsResponse{
		TotalUsers:  userCount,
		TotalSpaces: spaceCount,
		TotalPages:  pageCount,
	}, nil
}

func (s *AdminServiceImpl) ListUsers(ctx context.Context, page, pageSize int) ([]dto.AdminUserResponse, int64, error) {
	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	result := make([]dto.AdminUserResponse, len(users))
	for i, u := range users {
		result[i] = toAdminUserResponse(&u)
	}
	return result, total, nil
}

func (s *AdminServiceImpl) UpdateUserRole(ctx context.Context, userID uint64, role string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Role = entity.UserRole(role)
	return s.userRepo.Update(ctx, user)
}

func (s *AdminServiceImpl) UpdateUserStatus(ctx context.Context, userID uint64, status string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Status = entity.UserStatus(status)
	return s.userRepo.Update(ctx, user)
}

func (s *AdminServiceImpl) ListSpaces(ctx context.Context, page, pageSize int) ([]dto.AdminSpaceResponse, int64, error) {
	offset := (page - 1) * pageSize
	spaces, total, err := s.spaceRepo.ListAll(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list spaces: %w", err)
	}
	result := make([]dto.AdminSpaceResponse, len(spaces))
	for i, sp := range spaces {
		result[i] = toAdminSpaceResponse(&sp)
	}
	return result, total, nil
}

func (s *AdminServiceImpl) GetSpaceDetail(ctx context.Context, spaceID uint64) (*dto.AdminSpaceResponse, error) {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("find space: %w", err)
	}
	resp := toAdminSpaceResponse(space)
	return &resp, nil
}

func (s *AdminServiceImpl) DeleteSpace(ctx context.Context, spaceID uint64) error {
	return s.spaceRepo.Delete(ctx, spaceID)
}

func (s *AdminServiceImpl) ListPages(ctx context.Context, page, pageSize int) ([]dto.AdminPageResponse, int64, error) {
	offset := (page - 1) * pageSize
	pages, total, err := s.pageRepo.ListAll(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list pages: %w", err)
	}
	result := make([]dto.AdminPageResponse, len(pages))
	for i, p := range pages {
		result[i] = toAdminPageResponse(&p)
	}
	return result, total, nil
}

func (s *AdminServiceImpl) DeletePage(ctx context.Context, pageID uint64) error {
	return s.pageRepo.Delete(ctx, pageID)
}

func toAdminUserResponse(u *entity.User) dto.AdminUserResponse {
	return dto.AdminUserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Avatar:    u.Avatar,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
	}
}

func toAdminSpaceResponse(sp *entity.Space) dto.AdminSpaceResponse {
	return dto.AdminSpaceResponse{
		ID:          sp.ID,
		Name:        sp.Name,
		Description: sp.Description,
		Visibility:  string(sp.Visibility),
		OwnerID:     sp.OwnerID,
		CreatedAt:   sp.CreatedAt,
		UpdatedAt:   sp.UpdatedAt,
	}
}

func toAdminPageResponse(p *entity.Page) dto.AdminPageResponse {
	return dto.AdminPageResponse{
		ID:        p.ID,
		SpaceID:   p.SpaceID,
		Title:     p.Title,
		Slug:      p.Slug,
		Status:    string(p.Status),
		Version:   p.Version,
		CreatedBy: p.CreatedBy,
		UpdatedBy: p.UpdatedBy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}

func (s *AdminServiceImpl) GetSettings(ctx context.Context) (*dto.SettingsResponse, error) {
	cfg, err := config.Load(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	maskedKeys := make([]string, len(cfg.MCP.APIKeys))
	for i, k := range cfg.MCP.APIKeys {
		maskedKeys[i] = maskSecret(k)
	}
	hasMCPKeys := len(cfg.MCP.APIKeys) > 0

	return &dto.SettingsResponse{
		MCP: dto.MCPSettings{
			Enabled:       cfg.MCP.Enabled,
			Transport:     cfg.MCP.Transport,
			HTTPPort:      cfg.MCP.HTTPPort,
			APIKeys:       maskedKeys,
			APIKeysMasked: hasMCPKeys,
			DefaultUserID: cfg.MCP.DefaultUserID,
		},
		RAG: dto.RAGSettings{
			Enabled:      cfg.RAG.Enabled,
			BaseURL:      cfg.RAG.BaseURL,
			APIKey:       maskSecret(cfg.RAG.APIKey),
			APIKeyMasked: cfg.RAG.APIKey != "",
			Timeout:      cfg.RAG.Timeout,
		},
	}, nil
}

func (s *AdminServiceImpl) UpdateSettings(ctx context.Context, req dto.SettingsRequest) error {
	cfg, err := config.Load(s.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if req.MCP != nil {
		cfg.MCP.Enabled = req.MCP.Enabled
		cfg.MCP.Transport = req.MCP.Transport
		cfg.MCP.HTTPPort = req.MCP.HTTPPort
		cfg.MCP.DefaultUserID = req.MCP.DefaultUserID
		switch req.MCP.APIKeysAction {
		case "clear":
			cfg.MCP.APIKeys = nil
		case "replace":
			cfg.MCP.APIKeys = req.MCP.APIKeys
		default: // "keep" or empty
			// preserve existing keys
		}
	}

	if req.RAG != nil {
		cfg.RAG.Enabled = req.RAG.Enabled
		cfg.RAG.BaseURL = req.RAG.BaseURL
		cfg.RAG.Timeout = req.RAG.Timeout
		switch {
		case req.RAG.APIKey == dto.SentinelUnchanged:
			// keep existing
		case req.RAG.APIKey == "":
			cfg.RAG.APIKey = ""
		default:
			cfg.RAG.APIKey = req.RAG.APIKey
		}
	}

	if err := config.Save(s.configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}

func (s *AdminServiceImpl) TestRAGConnection(ctx context.Context, req dto.TestRAGRequest) (*dto.TestRAGResponse, error) {
	parsed, err := url.Parse(req.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return &dto.TestRAGResponse{Connected: false, Message: "无效的服务地址，仅支持 http/https"}, nil
	}

	apiKey := req.APIKey
	if req.UseSavedKey {
		cfg, cfgErr := config.Load(s.configPath)
		if cfgErr != nil {
			return &dto.TestRAGResponse{Connected: false, Message: "无法读取已保存的配置"}, nil
		}
		apiKey = cfg.RAG.APIKey
		if req.BaseURL == "" {
			req.BaseURL = cfg.RAG.BaseURL
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	targetURL := strings.TrimRight(req.BaseURL, "/") + "/health"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return &dto.TestRAGResponse{Connected: false, Message: "无效的服务地址"}, nil
	}
	if apiKey != "" {
		httpReq.Header.Set("X-API-Key", apiKey)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return &dto.TestRAGResponse{Connected: false, Message: "连接失败，请检查服务地址是否正确"}, nil
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK {
		return &dto.TestRAGResponse{Connected: true, Message: "RAG 服务连接成功"}, nil
	}
	return &dto.TestRAGResponse{Connected: false, Message: fmt.Sprintf("RAG 服务返回状态码: %d", resp.StatusCode)}, nil
}
