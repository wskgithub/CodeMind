package service

import (
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"
	"codemind/pkg/llm"
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// LLMBackendService handles LLM backend node management.
type LLMBackendService struct {
	backendRepo  *repository.LLMBackendRepository
	auditRepo    *repository.AuditRepository
	loadBalancer *llm.LoadBalancer
	logger       *zap.Logger
}

// NewLLMBackendService creates a new LLM backend service.
func NewLLMBackendService(
	backendRepo *repository.LLMBackendRepository,
	auditRepo *repository.AuditRepository,
	loadBalancer *llm.LoadBalancer,
	logger *zap.Logger,
) *LLMBackendService {
	return &LLMBackendService{
		backendRepo:  backendRepo,
		auditRepo:    auditRepo,
		loadBalancer: loadBalancer,
		logger:       logger,
	}
}

// List returns all backend nodes.
func (s *LLMBackendService) List() ([]model.LLMBackend, error) {
	return s.backendRepo.ListAll()
}

// Create creates a backend node.
func (s *LLMBackendService) Create(req *dto.CreateLLMBackendRequest, operatorID int64, clientIP string) error {
	backend := &model.LLMBackend{
		Name:                 req.Name,
		DisplayName:          req.DisplayName,
		BaseURL:              req.BaseURL,
		APIKey:               req.APIKey,
		Format:               req.Format,
		Weight:               req.Weight,
		MaxConcurrency:       req.MaxConcurrency,
		HealthCheckURL:       req.HealthCheckURL,
		TimeoutSeconds:       req.TimeoutSeconds,
		StreamTimeoutSeconds: req.StreamTimeoutSeconds,
		ModelPatterns:        req.ModelPatterns,
		Status:               model.LLMBackendEnabled,
	}

	if backend.Weight == 0 {
		backend.Weight = 100
	}
	if backend.MaxConcurrency == 0 {
		backend.MaxConcurrency = 100
	}
	if backend.TimeoutSeconds == 0 {
		backend.TimeoutSeconds = 300
	}
	if backend.StreamTimeoutSeconds == 0 {
		backend.StreamTimeoutSeconds = 600
	}
	if backend.ModelPatterns == "" {
		backend.ModelPatterns = "*"
	}

	if err := s.backendRepo.Create(backend); err != nil {
		return errcode.ErrDatabase
	}

	s.refreshLoadBalancer()

	s.recordAudit(operatorID, "create_llm_backend", "llm_backend", &backend.ID,
		map[string]interface{}{"name": req.Name, "base_url": req.BaseURL}, clientIP)

	return nil
}

// Update updates a backend node.
func (s *LLMBackendService) Update(id int64, req *dto.UpdateLLMBackendRequest, operatorID int64, clientIP string) error {
	backend, err := s.backendRepo.FindByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	if req.DisplayName != nil {
		backend.DisplayName = *req.DisplayName
	}
	if req.BaseURL != nil {
		backend.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		backend.APIKey = *req.APIKey
	}
	if req.Format != nil {
		backend.Format = *req.Format
	}
	if req.Weight != nil {
		backend.Weight = *req.Weight
	}
	if req.MaxConcurrency != nil {
		backend.MaxConcurrency = *req.MaxConcurrency
	}
	if req.Status != nil {
		backend.Status = *req.Status
	}
	if req.HealthCheckURL != nil {
		backend.HealthCheckURL = *req.HealthCheckURL
	}
	if req.TimeoutSeconds != nil {
		backend.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.StreamTimeoutSeconds != nil {
		backend.StreamTimeoutSeconds = *req.StreamTimeoutSeconds
	}
	if req.ModelPatterns != nil {
		backend.ModelPatterns = *req.ModelPatterns
	}

	if err := s.backendRepo.Update(backend); err != nil {
		return errcode.ErrDatabase
	}

	s.refreshLoadBalancer()

	s.recordAudit(operatorID, "update_llm_backend", "llm_backend", &id,
		map[string]interface{}{"name": backend.Name}, clientIP)

	return nil
}

// Delete deletes a backend node.
func (s *LLMBackendService) Delete(id int64, operatorID int64, clientIP string) error {
	backend, err := s.backendRepo.FindByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	if err := s.backendRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	s.refreshLoadBalancer()

	s.recordAudit(operatorID, "delete_llm_backend", "llm_backend", &id,
		map[string]interface{}{"name": backend.Name}, clientIP)

	return nil
}

// RefreshLoadBalancer reloads backends and refreshes load balancer.
func (s *LLMBackendService) RefreshLoadBalancer() {
	s.refreshLoadBalancer()
}

func (s *LLMBackendService) refreshLoadBalancer() {
	if s.loadBalancer == nil {
		return
	}

	backends, err := s.backendRepo.ListEnabled()
	if err != nil {
		s.logger.Error("failed to load LLM backends", zap.Error(err))
		return
	}

	nodes := make([]*llm.LBNode, 0, len(backends))
	for _, b := range backends {
		var provider llm.Provider
		switch b.Format {
		case "anthropic":
			client := llm.NewAnthropicClient(b.BaseURL, b.APIKey, b.TimeoutSeconds, b.StreamTimeoutSeconds)
			provider = llm.NewAnthropicProvider(b.Name, client)
		default:
			client := llm.NewClient(b.BaseURL, b.APIKey, b.TimeoutSeconds, b.StreamTimeoutSeconds)
			provider = llm.NewOpenAIProvider(b.Name, client)
		}

		patterns := strings.Split(b.ModelPatterns, ",")
		nodes = append(nodes, &llm.LBNode{
			ID:            b.ID,
			Name:          b.Name,
			Provider:      provider,
			Weight:        b.Weight,
			MaxConn:       b.MaxConcurrency,
			ModelPatterns: patterns,
		})
	}

	s.loadBalancer.UpdateNodes(nodes)
}

func (s *LLMBackendService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
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
		s.logger.Error("failed to record audit log", zap.Error(err))
	}
}
