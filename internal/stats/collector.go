package stats

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/PrismManager/gemstone/internal/process"
	"github.com/PrismManager/gemstone/internal/types"
)

// Collector collects system and process statistics
type Collector struct {
	mu          sync.RWMutex
	manager     *process.Manager
	systemStats []types.SystemStats
	maxHistory  int
	interval    time.Duration
	stopChan    chan struct{}
	running     bool
}

// NewCollector creates a new stats collector
func NewCollector(manager *process.Manager) *Collector {
	return &Collector{
		manager:    manager,
		maxHistory: 1000,
		interval:   10 * time.Second,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the stats collector
func (c *Collector) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.collectLoop()
}

// Stop stops the stats collector
func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.running = false
	close(c.stopChan)
}

func (c *Collector) collectLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately
	c.collect()

	for {
		select {
		case <-ticker.C:
			c.collect()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Collector) collect() {
	// Collect system stats
	sysStats := c.collectSystemStats()

	c.mu.Lock()
	c.systemStats = append(c.systemStats, sysStats)
	if len(c.systemStats) > c.maxHistory {
		c.systemStats = c.systemStats[len(c.systemStats)-c.maxHistory:]
	}
	c.mu.Unlock()

	// Collect process stats
	c.manager.CollectAllStats()
}

func (c *Collector) collectSystemStats() types.SystemStats {
	stats := types.SystemStats{
		Timestamp: time.Now(),
	}

	// CPU
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		stats.CPUPercent = cpuPercent[0]
	}

	// Memory
	if memInfo, err := mem.VirtualMemory(); err == nil {
		stats.MemoryTotal = memInfo.Total
		stats.MemoryUsed = memInfo.Used
		stats.MemoryPercent = memInfo.UsedPercent
	}

	// Disk
	if diskInfo, err := disk.Usage("/"); err == nil {
		stats.DiskTotal = diskInfo.Total
		stats.DiskUsed = diskInfo.Used
		stats.DiskPercent = diskInfo.UsedPercent
	}

	// Load average
	if loadInfo, err := load.Avg(); err == nil {
		stats.LoadAverage = []float64{loadInfo.Load1, loadInfo.Load5, loadInfo.Load15}
	}

	// Uptime
	if uptime, err := host.Uptime(); err == nil {
		stats.Uptime = uptime
	}

	return stats
}

// GetCurrentSystemStats returns current system stats
func (c *Collector) GetCurrentSystemStats() types.SystemStats {
	return c.collectSystemStats()
}

// GetSystemStatsHistory returns historical system stats
func (c *Collector) GetSystemStatsHistory(limit int) []types.SystemStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.systemStats) {
		result := make([]types.SystemStats, len(c.systemStats))
		copy(result, c.systemStats)
		return result
	}

	start := len(c.systemStats) - limit
	result := make([]types.SystemStats, limit)
	copy(result, c.systemStats[start:])
	return result
}
