package database

import (
	"context"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// HealthChecker monitors database connection health
type HealthChecker struct {
	manager  *ConnectionManager
	interval time.Duration
	status   *HealthStatus
	stopChan chan struct{}
	running  bool
	mutex    sync.RWMutex
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(manager *ConnectionManager, interval time.Duration) *HealthChecker {
	if interval == 0 {
		interval = 30 * time.Second
	}

	hc := &HealthChecker{
		manager:  manager,
		interval: interval,
		status: &HealthStatus{
			Status:      "unknown",
			LastChecked: time.Now(),
			Connections: make(map[string]*ConnectionHealth),
		},
		stopChan: make(chan struct{}),
	}

	// Start monitoring
	go hc.start()

	return hc
}

// GetStatus returns the current health status
func (hc *HealthChecker) GetStatus() *HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	// Create a deep copy to avoid race conditions
	status := &HealthStatus{
		Status:      hc.status.Status,
		LastChecked: hc.status.LastChecked,
		Connections: make(map[string]*ConnectionHealth),
	}

	for k, v := range hc.status.Connections {
		status.Connections[k] = &ConnectionHealth{
			Status:       v.Status,
			LastChecked:  v.LastChecked,
			ResponseTime: v.ResponseTime,
			Error:        v.Error,
		}
	}

	return status
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	if hc.running {
		close(hc.stopChan)
		hc.running = false
	}
}

// start runs the health monitoring loop
func (hc *HealthChecker) start() {
	hc.mutex.Lock()
	hc.running = true
	hc.mutex.Unlock()

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// Initial health check
	hc.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			hc.performHealthCheck()
		case <-hc.stopChan:
			return
		}
	}
}

// performHealthCheck checks the health of all database connections
func (hc *HealthChecker) performHealthCheck() {
	now := time.Now()
	connectionHealth := make(map[string]*ConnectionHealth)
	overallHealthy := true

	// Check shared database
	if hc.manager.sharedDB != nil {
		health := hc.checkConnection(hc.manager.sharedDB, "shared")
		connectionHealth["shared"] = health
		if health.Status != ConnectionStatusActive {
			overallHealthy = false
		}
	}

	// Check tenant databases
	hc.manager.mutex.RLock()
	for tenantID, db := range hc.manager.tenantDBs {
		health := hc.checkConnection(db, tenantID)
		connectionHealth[tenantID] = health
		if health.Status != ConnectionStatusActive {
			overallHealthy = false
		}
	}
	hc.manager.mutex.RUnlock()

	// Update status
	hc.mutex.Lock()
	hc.status.LastChecked = now
	hc.status.Connections = connectionHealth
	if overallHealthy {
		hc.status.Status = "healthy"
	} else {
		hc.status.Status = "unhealthy"
	}
	hc.mutex.Unlock()
}

// checkConnection performs a health check on a single database connection
func (hc *HealthChecker) checkConnection(db *sqlx.DB, identifier string) *ConnectionHealth {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := &ConnectionHealth{
		LastChecked: start,
	}

	err := db.PingContext(ctx)
	health.ResponseTime = time.Since(start)

	if err != nil {
		health.Status = ConnectionStatusUnhealthy
		health.Error = err.Error()
	} else {
		health.Status = ConnectionStatusActive
	}

	return health
}
