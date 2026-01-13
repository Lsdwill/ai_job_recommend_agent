package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 性能指标收集器
type Metrics struct {
	// 请求计数
	totalRequests  uint64
	activeRequests int64
	failedRequests uint64
	streamRequests uint64

	// 延迟统计
	requestLatency sync.Map // map[string]*LatencyStats

	// 系统指标
	startTime time.Time

	mu sync.RWMutex
}

// LatencyStats 延迟统计
type LatencyStats struct {
	count uint64
	sum   uint64
	min   uint64
	max   uint64
	mu    sync.RWMutex
}

var globalMetrics = &Metrics{
	startTime: time.Now(),
}

// GetGlobalMetrics 获取全局指标收集器
func GetGlobalMetrics() *Metrics {
	return globalMetrics
}

// IncTotalRequests 增加总请求数
func (m *Metrics) IncTotalRequests() {
	atomic.AddUint64(&m.totalRequests, 1)
}

// IncActiveRequests 增加活跃请求数
func (m *Metrics) IncActiveRequests() {
	atomic.AddInt64(&m.activeRequests, 1)
}

// DecActiveRequests 减少活跃请求数
func (m *Metrics) DecActiveRequests() {
	atomic.AddInt64(&m.activeRequests, -1)
}

// IncFailedRequests 增加失败请求数
func (m *Metrics) IncFailedRequests() {
	atomic.AddUint64(&m.failedRequests, 1)
}

// IncStreamRequests 增加流式请求数
func (m *Metrics) IncStreamRequests() {
	atomic.AddUint64(&m.streamRequests, 1)
}

// RecordLatency 记录请求延迟
func (m *Metrics) RecordLatency(endpoint string, duration time.Duration) {
	durationMs := uint64(duration.Milliseconds())

	val, _ := m.requestLatency.LoadOrStore(endpoint, &LatencyStats{
		min: durationMs,
		max: durationMs,
	})

	stats := val.(*LatencyStats)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.count++
	stats.sum += durationMs
	if durationMs < stats.min {
		stats.min = durationMs
	}
	if durationMs > stats.max {
		stats.max = durationMs
	}
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	totalReq := atomic.LoadUint64(&m.totalRequests)
	activeReq := atomic.LoadInt64(&m.activeRequests)
	failedReq := atomic.LoadUint64(&m.failedRequests)
	streamReq := atomic.LoadUint64(&m.streamRequests)

	uptime := time.Since(m.startTime)
	qps := float64(totalReq) / uptime.Seconds()

	latencyStats := make(map[string]interface{})
	m.requestLatency.Range(func(key, value interface{}) bool {
		endpoint := key.(string)
		stats := value.(*LatencyStats)

		stats.mu.RLock()
		defer stats.mu.RUnlock()

		avg := uint64(0)
		if stats.count > 0 {
			avg = stats.sum / stats.count
		}

		latencyStats[endpoint] = map[string]interface{}{
			"count":  stats.count,
			"avg_ms": avg,
			"min_ms": stats.min,
			"max_ms": stats.max,
		}
		return true
	})

	return map[string]interface{}{
		"requests": map[string]interface{}{
			"total":   totalReq,
			"active":  activeReq,
			"failed":  failedReq,
			"stream":  streamReq,
			"success": totalReq - failedReq,
		},
		"performance": map[string]interface{}{
			"qps":     qps,
			"uptime":  uptime.String(),
			"latency": latencyStats,
		},
		"system": map[string]interface{}{
			"goroutines":      runtime.NumGoroutine(),
			"cpu_cores":       runtime.NumCPU(),
			"gomaxprocs":      runtime.GOMAXPROCS(0),
			"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
			"memory_sys_mb":   memStats.Sys / 1024 / 1024,
			"memory_heap_mb":  memStats.HeapAlloc / 1024 / 1024,
			"gc_count":        memStats.NumGC,
			"gc_pause_total":  time.Duration(memStats.PauseTotalNs).String(),
		},
	}
}

// Reset 重置统计信息（用于测试）
func (m *Metrics) Reset() {
	atomic.StoreUint64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.activeRequests, 0)
	atomic.StoreUint64(&m.failedRequests, 0)
	atomic.StoreUint64(&m.streamRequests, 0)
	m.requestLatency = sync.Map{}
	m.startTime = time.Now()
}
