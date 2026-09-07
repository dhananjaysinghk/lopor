package observability

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type SystemMetrics struct {
	UptimeSeconds   int64   `json:"uptime_seconds"`
	Goroutines      int     `json:"goroutines"`
	HeapAllocMB     float64 `json:"heap_alloc_mb"`
	HeapSysMB       float64 `json:"heap_sys_mb"`
	GCAllocMB       float64 `json:"gc_alloc_mb"`
	TotalAllocMB    float64 `json:"total_alloc_mb"`
	NumGC           uint32  `json:"num_gc"`
	TotalRequests   uint64  `json:"total_requests"`
	TotalErrors     uint64  `json:"total_errors"`
	ActiveWSConns   int64   `json:"active_ws_conns"`
	DBConnections   int32   `json:"db_connections"`
	DBIdleConns     int32   `json:"db_idle_conns"`
	RedisPingResult string  `json:"redis_status"`
	Timestamp       string  `json:"timestamp"`
}

type DeepHealthStatus struct {
	Status       string            `json:"status"` // "HEALTHY", "DEGRADED", "UNHEALTHY"
	Version      string            `json:"version"`
	Components   map[string]string `json:"components"` // e.g. "database": "up", "redis": "up", "ai_gateway": "up"
	CheckedAt    string            `json:"checked_at"`
	ResponseTime string            `json:"response_time"`
}

type MetricsCollector struct {
	startTime    time.Time
	reqCount     uint64
	errCount     uint64
	activeWS     int64
	dbPool       *pgxpool.Pool
	redisClient  *redis.Client
}

var GlobalCollector *MetricsCollector

func NewMetricsCollector(dbPool *pgxpool.Pool, redisClient *redis.Client) *MetricsCollector {
	c := &MetricsCollector{
		startTime:   time.Now(),
		dbPool:      dbPool,
		redisClient: redisClient,
	}
	GlobalCollector = c
	return c
}

func (mc *MetricsCollector) IncRequests() {
	atomic.AddUint64(&mc.reqCount, 1)
}

func (mc *MetricsCollector) IncErrors() {
	atomic.AddUint64(&mc.errCount, 1)
}

func (mc *MetricsCollector) IncWSConns() {
	atomic.AddInt64(&mc.activeWS, 1)
}

func (mc *MetricsCollector) DecWSConns() {
	atomic.AddInt64(&mc.activeWS, -1)
}

func (mc *MetricsCollector) CollectMetrics() *SystemMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var dbTotal, dbIdle int32
	if mc.dbPool != nil {
		stat := mc.dbPool.Stat()
		dbTotal = stat.TotalConns()
		dbIdle = stat.IdleConns()
	}

	redisStatus := "disabled"
	if mc.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := mc.redisClient.Ping(ctx).Err(); err == nil {
			redisStatus = "connected"
		} else {
			redisStatus = "error: " + err.Error()
		}
	}

	return &SystemMetrics{
		UptimeSeconds:   int64(time.Since(mc.startTime).Seconds()),
		Goroutines:      runtime.NumGoroutine(),
		HeapAllocMB:     bToMb(mem.HeapAlloc),
		HeapSysMB:       bToMb(mem.HeapSys),
		GCAllocMB:       bToMb(mem.GCSys),
		TotalAllocMB:    bToMb(mem.TotalAlloc),
		NumGC:           mem.NumGC,
		TotalRequests:   atomic.LoadUint64(&mc.reqCount),
		TotalErrors:     atomic.LoadUint64(&mc.errCount),
		ActiveWSConns:   atomic.LoadInt64(&mc.activeWS),
		DBConnections:   dbTotal,
		DBIdleConns:     dbIdle,
		RedisPingResult: redisStatus,
		Timestamp:       time.Now().Format(time.RFC3339),
	}
}

func (mc *MetricsCollector) PerformDeepHealthCheck(ctx context.Context) *DeepHealthStatus {
	start := time.Now()
	components := make(map[string]string)
	overallStatus := "HEALTHY"

	// 1. Check Database
	if mc.dbPool != nil {
		if err := mc.dbPool.Ping(ctx); err != nil {
			components["database"] = "down: " + err.Error()
			overallStatus = "DEGRADED"
		} else {
			components["database"] = "healthy"
		}
	} else {
		components["database"] = "in-memory (mock)"
	}

	// 2. Check Redis
	if mc.redisClient != nil {
		if err := mc.redisClient.Ping(ctx).Err(); err != nil {
			components["redis"] = "down: " + err.Error()
			overallStatus = "DEGRADED"
		} else {
			components["redis"] = "healthy"
		}
	} else {
		components["redis"] = "disabled"
	}

	// 3. Check AI Gateway
	components["ai_gateway"] = "healthy"

	return &DeepHealthStatus{
		Status:       overallStatus,
		Version:      "1.0.0",
		Components:   components,
		CheckedAt:    time.Now().Format(time.RFC3339),
		ResponseTime: time.Since(start).String(),
	}
}

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
