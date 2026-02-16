package service

import (
	"encoding/json"

	"codemind/internal/config"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// APIKeyService API Key 管理业务逻辑
type APIKeyService struct {
	keyRepo   *repository.APIKeyRepository
	auditRepo *repository.AuditRepository
	logger    *zap.Logger
}

// NewAPIKeyService 创建 API Key 服务
func NewAPIKeyService(
	keyRepo *repository.APIKeyRepository,
	auditRepo *repository.AuditRepository,
	logger *zap.Logger,
) *APIKeyService {
	return &APIKeyService{
		keyRepo:   keyRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Create 创建新的 API Key（返回完整 Key，仅此一次）
func (s *APIKeyService) Create(req *dto.CreateAPIKeyRequest, userID int64, clientIP string) (*dto.APIKeyCreateResponse, error) {
	// 检查 Key 数量限制
	count, err := s.keyRepo.CountByUserID(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	cfg := config.Get()
	if int(count) >= cfg.System.MaxKeysPerUser {
		return nil, errcode.ErrAPIKeyLimit
	}

	// 生成 API Key
	fullKey, prefix, keyHash, err := crypto.GenerateAPIKey()
	if err != nil {
		s.logger.Error("生成 API Key 失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	// 存储 Key
	key := &model.APIKey{
		UserID:    userID,
		Name:      req.Name,
		KeyPrefix: prefix,
		KeyHash:   keyHash,
		Status:    model.StatusEnabled,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.keyRepo.Create(key); err != nil {
		s.logger.Error("创建 API Key 失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	// 记录审计日志
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

// List 获取用户的 API Key 列表
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

// UpdateStatus 切换 Key 状态
func (s *APIKeyService) UpdateStatus(keyID int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	key, err := s.keyRepo.FindByID(keyID)
	if err != nil {
		return errcode.ErrAPIKeyNotFound
	}

	// 权限校验：普通用户只能操作自己的 Key
	if operatorRole == model.RoleUser && key.UserID != operatorID {
		return errcode.ErrForbidden
	}

	if err := s.keyRepo.UpdateStatus(keyID, status); err != nil {
		return errcode.ErrDatabase
	}

	action := model.AuditActionEnableKey
	if status == model.StatusDisabled {
		action = model.AuditActionDisableKey
	}
	s.recordAudit(operatorID, action, model.AuditTargetAPIKey, &keyID, nil, clientIP)

	return nil
}

// Delete 删除 API Key
func (s *APIKeyService) Delete(keyID int64, operatorID int64, operatorRole string, clientIP string) error {
	key, err := s.keyRepo.FindByID(keyID)
	if err != nil {
		return errcode.ErrAPIKeyNotFound
	}

	// 权限校验
	if operatorRole == model.RoleUser && key.UserID != operatorID {
		return errcode.ErrForbidden
	}

	if err := s.keyRepo.Delete(keyID); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionDeleteKey, model.AuditTargetAPIKey, &keyID,
		map[string]string{"name": key.Name, "prefix": key.KeyPrefix}, clientIP)

	return nil
}

// recordAudit 记录审计日志
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
		s.logger.Error("记录审计日志失败", zap.Error(err), zap.String("action", action))
	}
}
