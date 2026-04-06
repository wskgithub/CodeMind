package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ThirdPartyProviderService 第三方模型服务业务逻辑
type ThirdPartyProviderService struct {
	repo        *repository.ThirdPartyProviderRepository
	backendRepo *repository.LLMBackendRepository
	encryptor   *crypto.Encryptor
	rdb         *redis.Client
	logger      *zap.Logger
}

// NewThirdPartyProviderService 创建 Service 实例
func NewThirdPartyProviderService(
	repo *repository.ThirdPartyProviderRepository,
	backendRepo *repository.LLMBackendRepository,
	encryptor *crypto.Encryptor,
	rdb *redis.Client,
	logger *zap.Logger,
) *ThirdPartyProviderService {
	return &ThirdPartyProviderService{
		repo:        repo,
		backendRepo: backendRepo,
		encryptor:   encryptor,
		rdb:         rdb,
		logger:      logger,
	}
}

// ──────────────────────────────────
// 模板管理（管理员）
// ──────────────────────────────────

// CreateTemplate 创建第三方服务模板
func (s *ThirdPartyProviderService) CreateTemplate(
	name, openAIBaseURL, anthropicBaseURL, format string,
	models []string, description, icon *string,
	sortOrder int, operatorID int64,
) (*model.ThirdPartyProviderTemplate, error) {
	if err := validateBaseURLsByFormat(format, openAIBaseURL, anthropicBaseURL); err != nil {
		return nil, err
	}

	exists, _ := s.repo.ExistsTemplateName(name, 0)
	if exists {
		return nil, errcode.ErrProviderTemplateNameExists
	}

	template := &model.ThirdPartyProviderTemplate{
		Name:             name,
		OpenAIBaseURL:    openAIBaseURL,
		AnthropicBaseURL: anthropicBaseURL,
		Models:           model.StringSlice(models),
		Format:           format,
		Description:      description,
		Icon:             icon,
		SortOrder:        sortOrder,
		CreatedBy:        operatorID,
	}

	if err := s.repo.CreateTemplate(template); err != nil {
		s.logger.Error("创建第三方服务模板失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	return template, nil
}

// ListTemplates 获取所有模板（管理员）
func (s *ThirdPartyProviderService) ListTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	return s.repo.ListTemplates()
}

// ListActiveTemplates 获取启用的模板（用户选择）
func (s *ThirdPartyProviderService) ListActiveTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	return s.repo.ListActiveTemplates()
}

// UpdateTemplate 更新模板
func (s *ThirdPartyProviderService) UpdateTemplate(
	id int64, name, openAIBaseURL, anthropicBaseURL *string,
	models []string, format *string,
	description, icon *string, sortOrder *int, status *int16,
) error {
	template, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	if name != nil {
		exists, _ := s.repo.ExistsTemplateName(*name, id)
		if exists {
			return errcode.ErrProviderTemplateNameExists
		}
		template.Name = *name
	}
	if openAIBaseURL != nil {
		template.OpenAIBaseURL = *openAIBaseURL
	}
	if anthropicBaseURL != nil {
		template.AnthropicBaseURL = *anthropicBaseURL
	}
	if models != nil {
		template.Models = model.StringSlice(models)
	}
	if format != nil {
		template.Format = *format
	}
	if description != nil {
		template.Description = description
	}
	if icon != nil {
		template.Icon = icon
	}
	if sortOrder != nil {
		template.SortOrder = *sortOrder
	}
	if status != nil {
		template.Status = *status
	}

	// 更新后重新校验 URL 与 format 的一致性
	if err := validateBaseURLsByFormat(template.Format, template.OpenAIBaseURL, template.AnthropicBaseURL); err != nil {
		return err
	}

	if err := s.repo.UpdateTemplate(template); err != nil {
		s.logger.Error("更新第三方服务模板失败", zap.Error(err))
		return errcode.ErrInternal
	}

	return nil
}

// DeleteTemplate 删除模板
func (s *ThirdPartyProviderService) DeleteTemplate(id int64) error {
	_, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}
	return s.repo.DeleteTemplate(id)
}

// ──────────────────────────────────
// 用户第三方服务管理
// ──────────────────────────────────

// CreateProvider 用户创建第三方服务
func (s *ThirdPartyProviderService) CreateProvider(
	userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey, format string,
	models []string, templateID *int64,
) (*model.UserThirdPartyProvider, error) {
	if err := validateBaseURLsByFormat(format, openAIBaseURL, anthropicBaseURL); err != nil {
		return nil, err
	}

	// 校验名称唯一性
	exists, _ := s.repo.ExistsProviderName(userID, name, 0)
	if exists {
		return nil, errcode.ErrProviderNameExists
	}

	// 校验模型名不与该用户已有服务冲突
	if err := s.checkModelConflict(userID, models, 0); err != nil {
		return nil, err
	}

	// 加密 API Key
	encrypted, err := s.encryptor.Encrypt(apiKey)
	if err != nil {
		s.logger.Error("加密第三方 API Key 失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	provider := &model.UserThirdPartyProvider{
		UserID:           userID,
		Name:             name,
		OpenAIBaseURL:    openAIBaseURL,
		AnthropicBaseURL: anthropicBaseURL,
		APIKeyEncrypted:  encrypted,
		Models:           model.StringSlice(models),
		Format:           format,
		TemplateID:       templateID,
	}

	if err := s.repo.CreateProvider(provider); err != nil {
		s.logger.Error("创建第三方服务失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	// 清除路由缓存
	s.invalidateRouteCache(userID)

	return provider, nil
}

// ListProviders 获取用户的所有第三方服务
func (s *ThirdPartyProviderService) ListProviders(userID int64) ([]model.UserThirdPartyProvider, error) {
	return s.repo.ListProvidersByUserID(userID)
}

// UpdateProvider 更新第三方服务
func (s *ThirdPartyProviderService) UpdateProvider(
	id, userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey *string,
	models []string, format *string, status *int16,
) error {
	provider, err := s.repo.GetProviderByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	// 权限校验：只能操作自己的服务
	if provider.UserID != userID {
		return errcode.ErrForbidden
	}

	if name != nil {
		exists, _ := s.repo.ExistsProviderName(userID, *name, id)
		if exists {
			return errcode.ErrProviderNameExists
		}
		provider.Name = *name
	}
	if openAIBaseURL != nil {
		provider.OpenAIBaseURL = *openAIBaseURL
	}
	if anthropicBaseURL != nil {
		provider.AnthropicBaseURL = *anthropicBaseURL
	}
	if apiKey != nil && *apiKey != "" {
		encrypted, err := s.encryptor.Encrypt(*apiKey)
		if err != nil {
			s.logger.Error("加密第三方 API Key 失败", zap.Error(err))
			return errcode.ErrInternal
		}
		provider.APIKeyEncrypted = encrypted
	}
	if models != nil {
		if err := s.checkModelConflict(userID, models, id); err != nil {
			return err
		}
		provider.Models = model.StringSlice(models)
	}
	if format != nil {
		provider.Format = *format
	}
	if status != nil {
		provider.Status = *status
	}

	// 更新后重新校验 URL 与 format 的一致性
	if err := validateBaseURLsByFormat(provider.Format, provider.OpenAIBaseURL, provider.AnthropicBaseURL); err != nil {
		return err
	}

	if err := s.repo.UpdateProvider(provider); err != nil {
		s.logger.Error("更新第三方服务失败", zap.Error(err))
		return errcode.ErrInternal
	}

	s.invalidateRouteCache(userID)
	return nil
}

// UpdateProviderStatus 切换第三方服务状态
func (s *ThirdPartyProviderService) UpdateProviderStatus(id, userID int64, status int16) error {
	provider, err := s.repo.GetProviderByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}
	if provider.UserID != userID {
		return errcode.ErrForbidden
	}

	if err := s.repo.UpdateProviderStatus(id, status); err != nil {
		return errcode.ErrInternal
	}

	s.invalidateRouteCache(userID)
	return nil
}

// DeleteProvider 删除第三方服务
func (s *ThirdPartyProviderService) DeleteProvider(id, userID int64) error {
	provider, err := s.repo.GetProviderByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}
	if provider.UserID != userID {
		return errcode.ErrForbidden
	}

	if err := s.repo.DeleteProvider(id); err != nil {
		return errcode.ErrInternal
	}

	s.invalidateRouteCache(userID)
	return nil
}

// ──────────────────────────────────
// 第三方模型路由
// ──────────────────────────────────

// 路由缓存 key 前缀和 TTL
const (
	thirdPartyRouteCachePrefix = "codemind:tp_route:"
	thirdPartyRouteCacheTTL    = 2 * time.Minute
)

// ResolveThirdPartyModel 根据模型名称和请求协议格式解析第三方服务路由
// requestFormat: "openai" 或 "anthropic"，由调用的端点决定
// 返回 nil 表示该模型不属于任何第三方服务（应走内置路由）
func (s *ThirdPartyProviderService) ResolveThirdPartyModel(
	ctx context.Context, userID int64, modelName string, requestFormat string,
) *model.ThirdPartyRouteInfo {
	// 尝试从 Redis 缓存获取路由映射
	cacheKey := fmt.Sprintf("%s%d", thirdPartyRouteCachePrefix, userID)
	cached, err := s.rdb.HGet(ctx, cacheKey, modelName).Result()
	if err == nil && cached != "" {
		var info model.ThirdPartyRouteInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			// 协议兼容性由 checkAndHandleThirdParty 调用方校验，此处只负责路由查找
			return &info
		}
	}

	// 缓存未命中或格式错误，需要检查整个路由映射是否已缓存
	exists, _ := s.rdb.Exists(ctx, cacheKey).Result()
	if exists > 0 {
		return nil
	}

	// 从数据库加载并构建路由映射
	providers, err := s.repo.ListActiveProvidersByUserID(userID)
	if err != nil {
		s.logger.Error("加载第三方服务列表失败", zap.Error(err), zap.Int64("user_id", userID))
		return nil
	}

	// 构建 model → route 映射并写入缓存
	routeMap := make(map[string]string)
	var matched *model.ThirdPartyRouteInfo

	for _, p := range providers {
		info := model.ThirdPartyRouteInfo{
			ProviderID:       p.ID,
			ProviderName:     p.Name,
			OpenAIBaseURL:    p.OpenAIBaseURL,
			AnthropicBaseURL: p.AnthropicBaseURL,
			APIKeyEncrypted:  p.APIKeyEncrypted,
			Format:           p.Format,
		}
		infoJSON, _ := json.Marshal(info)
		infoStr := string(infoJSON)

		for _, m := range p.Models {
			routeMap[m] = infoStr
			if m == modelName {
				matched = &info
			}
		}
	}

	// 写入缓存（即使为空映射也要写入，标记「已加载」避免重复查库）
	if len(routeMap) > 0 {
		s.rdb.HSet(ctx, cacheKey, routeMap)
	} else {
		s.rdb.HSet(ctx, cacheKey, "__loaded__", "1")
	}
	s.rdb.Expire(ctx, cacheKey, thirdPartyRouteCacheTTL)

	return matched
}

// DecryptAPIKey 解密第三方服务的 API Key
func (s *ThirdPartyProviderService) DecryptAPIKey(encrypted string) (string, error) {
	return s.encryptor.Decrypt(encrypted)
}

// RecordThirdPartyUsage 异步记录第三方服务用量（仅供参考）
func (s *ThirdPartyProviderService) RecordThirdPartyUsage(
	userID, providerID, apiKeyID int64,
	modelName, requestType string,
	promptTokens, completionTokens, totalTokens int,
	cacheCreationTokens, cacheReadTokens int,
	durationMs *int,
) {
	usage := &model.ThirdPartyTokenUsage{
		UserID:                   userID,
		ProviderID:               providerID,
		APIKeyID:                 apiKeyID,
		Model:                    modelName,
		PromptTokens:             promptTokens,
		CompletionTokens:         completionTokens,
		TotalTokens:              totalTokens,
		CacheCreationInputTokens: cacheCreationTokens,
		CacheReadInputTokens:     cacheReadTokens,
		RequestType:              requestType,
		DurationMs:               durationMs,
	}

	if err := s.repo.CreateThirdPartyUsage(usage); err != nil {
		s.logger.Error("记录第三方用量失败", zap.Error(err))
	}
}

// ──────────────────────────────────
// 内部方法
// ──────────────────────────────────

// checkModelConflict 检查模型名是否与用户已有服务冲突
func (s *ThirdPartyProviderService) checkModelConflict(userID int64, models []string, excludeID int64) error {
	existing, err := s.repo.ListActiveProvidersByUserID(userID)
	if err != nil {
		return errcode.ErrInternal
	}

	usedModels := make(map[string]string) // model → provider name
	for _, p := range existing {
		if p.ID == excludeID {
			continue
		}
		for _, m := range p.Models {
			usedModels[m] = p.Name
		}
	}

	for _, m := range models {
		if providerName, ok := usedModels[m]; ok {
			return errcode.ErrInvalidParams.WithMessage(
				fmt.Sprintf("模型 %q 已在服务 %q 中使用", m, providerName))
		}
	}

	return nil
}

// ListPlatformModels 获取 CodeMind 平台可用模型信息（面向普通用户展示）
func (s *ThirdPartyProviderService) ListPlatformModels() ([]dto.PlatformModelInfo, error) {
	backends, err := s.backendRepo.ListAll()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	var result []dto.PlatformModelInfo
	for _, b := range backends {
		if b.Status != model.LLMBackendEnabled {
			continue
		}
		displayName := b.DisplayName
		if displayName == "" {
			displayName = b.Name
		}
		result = append(result, dto.PlatformModelInfo{
			Name:          b.Name,
			DisplayName:   displayName,
			Format:        b.Format,
			ModelPatterns: b.ModelPatterns,
			Status:        b.Status,
		})
	}
	return result, nil
}

// validateBaseURLsByFormat 根据协议格式校验 Base URL 必填性
func validateBaseURLsByFormat(format, openAIBaseURL, anthropicBaseURL string) error {
	switch format {
	case "openai":
		if openAIBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("选择 OpenAI 协议时必须填写 OpenAI Base URL")
		}
	case "anthropic":
		if anthropicBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("选择 Anthropic 协议时必须填写 Anthropic Base URL")
		}
	case "all":
		if openAIBaseURL == "" || anthropicBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("选择全部协议时必须同时填写 OpenAI 和 Anthropic Base URL")
		}
	}
	return nil
}

// invalidateRouteCache 清除用户的第三方路由缓存
func (s *ThirdPartyProviderService) invalidateRouteCache(userID int64) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", thirdPartyRouteCachePrefix, userID)
	s.rdb.Del(ctx, key)
}
