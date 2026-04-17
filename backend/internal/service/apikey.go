package service

import (
	"codemind/internal/config"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// APIKeyService handles API key management.
type APIKeyService struct {
	keyRepo   *repository.APIKeyRepository
	auditRepo *repository.AuditRepository
	rdb       *redis.Client
	logger    *zap.Logger
	encryptor *crypto.Encryptor
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(
	keyRepo *repository.APIKeyRepository,
	auditRepo *repository.AuditRepository,
	rdb *redis.Client,
	logger *zap.Logger,
	encryptor *crypto.Encryptor,
) *APIKeyService {
	return &APIKeyService{
		keyRepo:   keyRepo,
		auditRepo: auditRepo,
		rdb:       rdb,
		logger:    logger,
		encryptor: encryptor,
	}
}

// Create creates a new API key (returns full key only once).
func (s *APIKeyService) Create(req *dto.CreateAPIKeyRequest, userID int64, clientIP string) (*dto.APIKeyCreateResponse, error) {
	count, err := s.keyRepo.CountByUserID(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	cfg := config.Get()
	if int(count) >= cfg.System.MaxKeysPerUser {
		return nil, errcode.ErrAPIKeyLimit
	}

	fullKey, prefix, keyHash, err := crypto.GenerateAPIKey()
	if err != nil {
		s.logger.Error("failed to generate API key", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	keyEncrypted := ""
	if s.encryptor != nil {
		keyEncrypted, err = s.encryptor.Encrypt(fullKey)
		if err != nil {
			s.logger.Error("failed to encrypt API key", zap.Error(err))
			return nil, errcode.ErrInternal
		}
	}

	key := &model.APIKey{
		UserID:       userID,
		Name:         req.Name,
		KeyPrefix:    prefix,
		KeyHash:      keyHash,
		KeyEncrypted: keyEncrypted,
		Status:       model.StatusEnabled,
		ExpiresAt:    req.ExpiresAt,
	}

	if err := s.keyRepo.Create(key); err != nil {
		s.logger.Error("failed to create API key", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	s.recordAudit(userID, model.AuditActionCreateKey, model.AuditTargetAPIKey, &key.ID,
		map[string]string{"name": req.Name, "prefix": prefix}, clientIP)

	return &dto.APIKeyCreateResponse{
		ID:        key.ID,
		Name:      key.Name,
		Key:       fullKey,
		KeyPrefix: prefix,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: key.CreatedAt,
	}, nil
}

// Copy decrypts and returns the full API key.
func (s *APIKeyService) Copy(keyID int64, operatorID int64, operatorRole string, clientIP string) (*dto.APIKeyCopyResponse, error) {
	key, err := s.keyRepo.FindByID(keyID)
	if err != nil {
		return nil, errcode.ErrAPIKeyNotFound
	}

	// Regular users can only copy their own keys
	if operatorRole == model.RoleUser && key.UserID != operatorID {
		return nil, errcode.ErrForbidden
	}

	if s.encryptor == nil {
		s.logger.Error("API key encryptor not initialized")
		return nil, errcode.ErrInternal
	}
	if key.KeyEncrypted == "" {
		return nil, errcode.ErrAPIKeyNotCopyable
	}

	fullKey, err := s.encryptor.Decrypt(key.KeyEncrypted)
	if err != nil {
		s.logger.Error("failed to decrypt API key", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	s.recordAudit(operatorID, model.AuditActionCopyKey, model.AuditTargetAPIKey, &keyID,
		map[string]string{"name": key.Name, "prefix": key.KeyPrefix}, clientIP)

	return &dto.APIKeyCopyResponse{Key: fullKey}, nil
}

// List returns user's API keys.
func (s *APIKeyService) List(userID int64) ([]dto.APIKeyResponse, error) {
	keys, err := s.keyRepo.ListByUserID(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	var resp []dto.APIKeyResponse
	for _, k := range keys {
		resp = append(resp, dto.APIKeyResponse{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			Status:     k.Status,
			LastUsedAt: k.LastUsedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedAt:  k.CreatedAt,
		})
	}

	if resp == nil {
		resp = []dto.APIKeyResponse{}
	}
	return resp, nil
}

// UpdateStatus toggles API key status.
func (s *APIKeyService) UpdateStatus(keyID int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	key, err := s.keyRepo.FindByID(keyID)
	if err != nil {
		return errcode.ErrAPIKeyNotFound
	}

	// Regular users can only manage their own keys
	if operatorRole == model.RoleUser && key.UserID != operatorID {
		return errcode.ErrForbidden
	}

	if err := s.keyRepo.UpdateStatus(keyID, status); err != nil {
		return errcode.ErrDatabase
	}

	// Invalidate cache to ensure status change takes effect immediately
	s.invalidateKeyCache(key.KeyHash)

	action := model.AuditActionEnableKey
	if status == model.StatusDisabled {
		action = model.AuditActionDisableKey
	}
	s.recordAudit(operatorID, action, model.AuditTargetAPIKey, &keyID, nil, clientIP)

	return nil
}

// Delete removes an API key.
func (s *APIKeyService) Delete(keyID int64, operatorID int64, operatorRole string, clientIP string) error {
	key, err := s.keyRepo.FindByID(keyID)
	if err != nil {
		return errcode.ErrAPIKeyNotFound
	}

	if operatorRole == model.RoleUser && key.UserID != operatorID {
		return errcode.ErrForbidden
	}

	if err := s.keyRepo.Delete(keyID); err != nil {
		return errcode.ErrDatabase
	}

	s.invalidateKeyCache(key.KeyHash)

	s.recordAudit(operatorID, model.AuditActionDeleteKey, model.AuditTargetAPIKey, &keyID,
		map[string]string{"name": key.Name, "prefix": key.KeyPrefix}, clientIP)

	return nil
}

func (s *APIKeyService) invalidateKeyCache(keyHash string) {
	if s.rdb == nil || keyHash == "" {
		return
	}
	cacheKey := fmt.Sprintf("codemind:apikey:%s", keyHash)
	if err := s.rdb.Del(context.Background(), cacheKey).Err(); err != nil {
		s.logger.Error("failed to clear API key cache", zap.String("key_hash_prefix", keyHash[:16]+"..."), zap.Error(err))
	}
}

// recordAudit records an audit log entry for API key operations.
func (s *APIKeyService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
	var detailJSON json.RawMessage
	if detail != nil {
		data, _ := json.Marshal(detail)
		detailJSON = data
	}

	log := &model.AuditLog{
		OperatorID: operatorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detailJSON,
		ClientIP:   &clientIP,
	}

	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error("failed to record audit log", zap.Error(err), zap.String("action", action))
	}
}
