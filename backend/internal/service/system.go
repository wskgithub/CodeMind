package service

import (
	"encoding/json"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// SystemService 系统管理业务逻辑
type SystemService struct {
	configRepo *repository.SystemRepository
	auditRepo  *repository.AuditRepository
	annRepo    *repository.AnnouncementRepository
	logger     *zap.Logger
}

// NewSystemService 创建系统管理服务
func NewSystemService(
	configRepo *repository.SystemRepository,
	auditRepo *repository.AuditRepository,
	annRepo *repository.AnnouncementRepository,
	logger *zap.Logger,
) *SystemService {
	return &SystemService{
		configRepo: configRepo,
		auditRepo:  auditRepo,
		annRepo:    annRepo,
		logger:     logger,
	}
}

// ──────────────────────────────────
// 系统配置
// ──────────────────────────────────

// GetConfigs 获取所有系统配置
func (s *SystemService) GetConfigs() ([]model.SystemConfig, error) {
	return s.configRepo.ListAll()
}

// UpdateConfigs 批量更新系统配置
func (s *SystemService) UpdateConfigs(req *dto.UpdateConfigsRequest, operatorID int64, clientIP string) error {
	configs := make([]model.SystemConfig, 0, len(req.Configs))
	for _, item := range req.Configs {
		configs = append(configs, model.SystemConfig{
			ConfigKey:   item.Key,
			ConfigValue: item.Value,
		})
	}

	if err := s.configRepo.BatchUpsert(configs); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionUpdateConfig, model.AuditTargetConfig, nil,
		map[string]interface{}{"keys": req.Configs}, clientIP)

	return nil
}

// ──────────────────────────────────
// 公告管理
// ──────────────────────────────────

// ListAnnouncements 获取公告列表
func (s *SystemService) ListAnnouncements(isAdmin bool) ([]model.Announcement, error) {
	if isAdmin {
		return s.annRepo.ListAll()
	}
	return s.annRepo.ListPublished()
}

// CreateAnnouncement 创建公告
func (s *SystemService) CreateAnnouncement(req *dto.CreateAnnouncementRequest, authorID int64, clientIP string) (*model.Announcement, error) {
	ann := &model.Announcement{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
		Status:   req.Status,
		Pinned:   req.Pinned,
	}

	if err := s.annRepo.Create(ann); err != nil {
		return nil, errcode.ErrDatabase
	}

	s.recordAudit(authorID, model.AuditActionCreateAnnounce, model.AuditTargetAnnouncement, &ann.ID,
		map[string]string{"title": req.Title}, clientIP)

	return ann, nil
}

// UpdateAnnouncement 更新公告
func (s *SystemService) UpdateAnnouncement(id int64, req *dto.UpdateAnnouncementRequest, operatorID int64, clientIP string) error {
	ann, err := s.annRepo.FindByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.Pinned != nil {
		fields["pinned"] = *req.Pinned
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}

	if err := s.annRepo.UpdateFields(id, fields); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionUpdateAnnounce, model.AuditTargetAnnouncement, &id,
		map[string]interface{}{"old_title": ann.Title, "changes": fields}, clientIP)

	return nil
}

// DeleteAnnouncement 删除公告
func (s *SystemService) DeleteAnnouncement(id int64, operatorID int64, clientIP string) error {
	if _, err := s.annRepo.FindByID(id); err != nil {
		return errcode.ErrRecordNotFound
	}

	if err := s.annRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionDeleteAnnounce, model.AuditTargetAnnouncement, &id, nil, clientIP)
	return nil
}

// ──────────────────────────────────
// 审计日志
// ──────────────────────────────────

// ListAuditLogs 查询审计日志
func (s *SystemService) ListAuditLogs(query *dto.AuditLogQuery) ([]model.AuditLog, int64, error) {
	filters := map[string]interface{}{
		"action":      query.Action,
		"operator_id": query.OperatorID,
	}

	if query.StartDate != "" {
		if t, err := time.Parse("2006-01-02", query.StartDate); err == nil {
			filters["start_date"] = t
		}
	}
	if query.EndDate != "" {
		if t, err := time.Parse("2006-01-02", query.EndDate); err == nil {
			filters["end_date"] = t.Add(24 * time.Hour)
		}
	}

	return s.auditRepo.List(query.GetPage(), query.GetPageSize(), filters)
}

// recordAudit 记录审计日志
func (s *SystemService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
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
		s.logger.Error("记录审计日志失败", zap.Error(err))
	}
}
