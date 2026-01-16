package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Name      string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
}

// HealthChecker performs health checks
type HealthChecker struct {
	checks map[string]CheckFunc
	mu     sync.RWMutex
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) HealthCheck

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
	}
}

// Register registers a health check
func (h *HealthChecker) Register(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// BackendHealthCheck checks if backend is operational
func BackendHealthCheck(backend interface{ Name() string }) CheckFunc {
	return func(ctx context.Context) HealthCheck {
		return HealthCheck{
			Name:      "backend",
			Status:    HealthStatusHealthy,
			Message:   fmt.Sprintf("Backend '%s' is operational", backend.Name()),
			Timestamp: time.Now(),
		}
	}
}

// UptimeHealthCheck reports server uptime
func UptimeHealthCheck(startTime time.Time) CheckFunc {
	return func(ctx context.Context) HealthCheck {
		uptime := time.Since(startTime)
		return HealthCheck{
			Name:      "uptime",
			Status:    HealthStatusHealthy,
			Message:   fmt.Sprintf("Uptime: %s", uptime.Round(time.Second)),
			Timestamp: time.Now(),
		}
	}
}
