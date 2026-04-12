package services

import (
	"context"
	"database/sql"
)

type HealthService struct {
	db  *sql.DB
	llm *LLMService
}

type HealthStatus struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services,omitempty"`
}

func NewHealthService(db *sql.DB, llm *LLMService) *HealthService {
	return &HealthService{db: db, llm: llm}
}

func (s *HealthService) Check(ctx context.Context) HealthStatus {
	services := map[string]string{
		"database": "ok",
		"llm":      "ok",
	}

	if err := s.db.PingContext(ctx); err != nil {
		services["database"] = "error"
	}

	if err := s.llm.HealthCheck(ctx); err != nil {
		services["llm"] = "error"
	}

	status := "ok"
	if services["database"] != "ok" || services["llm"] != "ok" {
		status = "degraded"
	}

	return HealthStatus{
		Status:   status,
		Services: services,
	}
}
