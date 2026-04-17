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

// ThirdPartyProviderService handles third-party provider management.
type ThirdPartyProviderService struct {
	repo        *repository.ThirdPartyProviderRepository
	backendRepo *repository.LLMBackendRepository
	encryptor   *crypto.Encryptor
	rdb         *redis.Client
	logger      *zap.Logger
}

// NewThirdPartyProviderService creates a new service instance.
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

// CreateTemplate creates a third-party service template.
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
		s.logger.Error("failed to create third-party template", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	return template, nil
}

// ListTemplates returns all templates (admin).
func (s *ThirdPartyProviderService) ListTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	return s.repo.ListTemplates()
}

// ListActiveTemplates returns enabled templates for user selection.
func (s *ThirdPartyProviderService) ListActiveTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	return s.repo.ListActiveTemplates()
}

// UpdateTemplate updates a template.
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

	if err := validateBaseURLsByFormat(template.Format, template.OpenAIBaseURL, template.AnthropicBaseURL); err != nil {
		return err
	}

	if err := s.repo.UpdateTemplate(template); err != nil {
		s.logger.Error("failed to update third-party template", zap.Error(err))
		return errcode.ErrInternal
	}

	return nil
}

// DeleteTemplate deletes a template.
func (s *ThirdPartyProviderService) DeleteTemplate(id int64) error {
	_, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}
	return s.repo.DeleteTemplate(id)
}

// CreateProvider creates a user's third-party provider.
func (s *ThirdPartyProviderService) CreateProvider(
	userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey, format string,
	models []string, templateID *int64,
) (*model.UserThirdPartyProvider, error) {
	if err := validateBaseURLsByFormat(format, openAIBaseURL, anthropicBaseURL); err != nil {
		return nil, err
	}

	exists, _ := s.repo.ExistsProviderName(userID, name, 0)
	if exists {
		return nil, errcode.ErrProviderNameExists
	}

	if err := s.checkModelConflict(userID, models, 0); err != nil {
		return nil, err
	}

	encrypted, err := s.encryptor.Encrypt(apiKey)
	if err != nil {
		s.logger.Error("failed to encrypt third-party API key", zap.Error(err))
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
		s.logger.Error("failed to create third-party provider", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	s.invalidateRouteCache(userID)

	return provider, nil
}

// ListProviders returns all providers for a user.
func (s *ThirdPartyProviderService) ListProviders(userID int64) ([]model.UserThirdPartyProvider, error) {
	return s.repo.ListProvidersByUserID(userID)
}

// UpdateProvider updates a third-party provider.
func (s *ThirdPartyProviderService) UpdateProvider( //nolint:gocyclo // complex business logic.
	id, userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey *string,
	models []string, format *string, status *int16,
) error {
	provider, err := s.repo.GetProviderByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

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
			s.logger.Error("failed to encrypt third-party API key", zap.Error(err))
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

	if err := validateBaseURLsByFormat(provider.Format, provider.OpenAIBaseURL, provider.AnthropicBaseURL); err != nil {
		return err
	}

	if err := s.repo.UpdateProvider(provider); err != nil {
		s.logger.Error("failed to update third-party provider", zap.Error(err))
		return errcode.ErrInternal
	}

	s.invalidateRouteCache(userID)
	return nil
}

// UpdateProviderStatus toggles provider status.
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

// DeleteProvider deletes a third-party provider.
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

const (
	thirdPartyRouteCachePrefix = "codemind:tp_route:"
	thirdPartyRouteCacheTTL    = 2 * time.Minute
)

// ResolveThirdPartyModel resolves third-party service routing by model name.
// Returns nil if the model is not a third-party model.
func (s *ThirdPartyProviderService) ResolveThirdPartyModel(
	ctx context.Context, userID int64, modelName string, _ string,
) *model.ThirdPartyRouteInfo {
	cacheKey := fmt.Sprintf("%s%d", thirdPartyRouteCachePrefix, userID)
	cached, err := s.rdb.HGet(ctx, cacheKey, modelName).Result()
	if err == nil && cached != "" {
		var info model.ThirdPartyRouteInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			return &info
		}
	}

	exists, _ := s.rdb.Exists(ctx, cacheKey).Result()
	if exists > 0 {
		return nil
	}

	providers, err := s.repo.ListActiveProvidersByUserID(userID)
	if err != nil {
		s.logger.Error("failed to load third-party providers", zap.Error(err), zap.Int64("user_id", userID))
		return nil
	}

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

	if len(routeMap) > 0 {
		s.rdb.HSet(ctx, cacheKey, routeMap)
	} else {
		s.rdb.HSet(ctx, cacheKey, "__loaded__", "1")
	}
	s.rdb.Expire(ctx, cacheKey, thirdPartyRouteCacheTTL)

	return matched
}

// DecryptAPIKey decrypts a third-party service API key.
func (s *ThirdPartyProviderService) DecryptAPIKey(encrypted string) (string, error) {
	return s.encryptor.Decrypt(encrypted)
}

// RecordThirdPartyUsage records third-party service usage asynchronously.
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
		s.logger.Error("failed to record third-party usage", zap.Error(err))
	}
}

func (s *ThirdPartyProviderService) checkModelConflict(userID int64, models []string, excludeID int64) error {
	existing, err := s.repo.ListActiveProvidersByUserID(userID)
	if err != nil {
		return errcode.ErrInternal
	}

	usedModels := make(map[string]string)
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
				fmt.Sprintf("model %q is already used in provider %q", m, providerName))
		}
	}

	return nil
}

// ListPlatformModels returns available platform models for users.
func (s *ThirdPartyProviderService) ListPlatformModels() ([]dto.PlatformModelInfo, error) {
	backends, err := s.backendRepo.ListAll()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	result := make([]dto.PlatformModelInfo, 0, len(backends))
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

func validateBaseURLsByFormat(format, openAIBaseURL, anthropicBaseURL string) error {
	switch format {
	case "openai":
		if openAIBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("OpenAI Base URL is required for OpenAI format")
		}
	case "anthropic":
		if anthropicBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("Anthropic Base URL is required for Anthropic format")
		}
	case "all":
		if openAIBaseURL == "" || anthropicBaseURL == "" {
			return errcode.ErrInvalidParams.WithMessage("both OpenAI and Anthropic Base URLs are required for all formats")
		}
	}
	return nil
}

func (s *ThirdPartyProviderService) invalidateRouteCache(userID int64) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", thirdPartyRouteCachePrefix, userID)
	s.rdb.Del(ctx, key)
}
