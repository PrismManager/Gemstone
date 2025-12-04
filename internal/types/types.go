package types

import (
	"time"
)

// ProcessStatus represents the status of a process
type ProcessStatus string

const (
	StatusStopped    ProcessStatus = "stopped"
	StatusRunning    ProcessStatus = "running"
	StatusStarting   ProcessStatus = "starting"
	StatusStopping   ProcessStatus = "stopping"
	StatusErrored    ProcessStatus = "errored"
	StatusRestarting ProcessStatus = "restarting"
)

// ProcessInfo represents detailed information about a managed process
type ProcessInfo struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Status        ProcessStatus     `json:"status"`
	PID           int               `json:"pid,omitempty"`
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	WorkDir       string            `json:"work_dir,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	AutoStart     bool              `json:"auto_start"`
	AutoRestart   bool              `json:"auto_restart"`
	MaxRestarts   int               `json:"max_restarts"`
	RestartCount  int               `json:"restart_count"`
	User          string            `json:"user,omitempty"`
	Group         string            `json:"group,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	StartedAt     *time.Time        `json:"started_at,omitempty"`
	StoppedAt     *time.Time        `json:"stopped_at,omitempty"`
	Uptime        int64             `json:"uptime,omitempty"` // seconds
	CPU           float64           `json:"cpu,omitempty"`    // percentage
	Memory        uint64            `json:"memory,omitempty"` // bytes
	MemoryPercent float64           `json:"memory_percent,omitempty"`
}

// ProcessStats represents resource usage statistics
type ProcessStats struct {
	ID            string    `json:"id"`
	PID           int       `json:"pid"`
	CPU           float64   `json:"cpu"`
	Memory        uint64    `json:"memory"`
	MemoryPercent float64   `json:"memory_percent"`
	NumThreads    int32     `json:"num_threads"`
	NumFDs        int32     `json:"num_fds"`
	ReadBytes     uint64    `json:"read_bytes"`
	WriteBytes    uint64    `json:"write_bytes"`
	Timestamp     time.Time `json:"timestamp"`
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskTotal     uint64    `json:"disk_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskPercent   float64   `json:"disk_percent"`
	LoadAverage   []float64 `json:"load_average"`
	Uptime        uint64    `json:"uptime"`
	Timestamp     time.Time `json:"timestamp"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        string    `json:"id"`
	ProcessID string    `json:"process_id"`
	Type      string    `json:"type"` // stdout, stderr
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// StartRequest represents a request to start a new process
type StartRequest struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	AutoStart   bool              `json:"auto_start"`
	AutoRestart bool              `json:"auto_restart"`
	MaxRestarts int               `json:"max_restarts"`
	User        string            `json:"user,omitempty"`
	Group       string            `json:"group,omitempty"`
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HistoricalStats represents time-series stats for charts
type HistoricalStats struct {
	ProcessID string         `json:"process_id"`
	Stats     []ProcessStats `json:"stats"`
}

// DaemonInfo represents daemon information
type DaemonInfo struct {
	Version      string      `json:"version"`
	Uptime       int64       `json:"uptime"`
	StartedAt    time.Time   `json:"started_at"`
	ProcessCount int         `json:"process_count"`
	SystemStats  SystemStats `json:"system_stats"`
}
