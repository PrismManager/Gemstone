package daemon

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/PrismManager/gemstone/internal/api"
	"github.com/PrismManager/gemstone/internal/config"
	"github.com/PrismManager/gemstone/internal/process"
	"github.com/PrismManager/gemstone/internal/stats"
)

// Version is the daemon version
const Version = "0.1.0"

// Daemon represents the gemstone daemon
type Daemon struct {
	config         *config.Config
	manager        *process.Manager
	api            *api.Server
	statsCollector *stats.Collector
	startedAt      time.Time
	socketPath     string
}

// New creates a new daemon instance
func New() (*Daemon, error) {
	// Load configuration
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create process manager
	manager, err := process.NewManager(cfg, config.GetDataPath(), config.GetLogPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create process manager: %w", err)
	}

	// Create stats collector
	statsCollector := stats.NewCollector(manager)

	// Create API server
	apiServer := api.NewServer(cfg, manager, statsCollector)

	return &Daemon{
		config:         cfg,
		manager:        manager,
		api:            apiServer,
		statsCollector: statsCollector,
		socketPath:     config.GetSocketPath(),
	}, nil
}

// Run starts the daemon
func (d *Daemon) Run() error {
	d.startedAt = time.Now()

	// Make sure that the socket directory exists
	socketDir := filepath.Dir(d.socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove stale socket file
	if err := os.Remove(d.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}
	defer listener.Close()

	// Set socket permissions
	if err := os.Chmod(d.socketPath, 0666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Start auto-start processes
	d.manager.StartAutoStartProcesses()

	// Start stats collector
	d.statsCollector.Start()

	// Start API server (if enabled)
	if d.config.API.Enabled {
		go d.api.Start()
	}

	// Handle socket connections (for CLI communication)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go d.handleConnection(conn)
	}
}

// Shutdown gracefully shuts down the daemon
func (d *Daemon) Shutdown() {
	// Stop stats collector
	d.statsCollector.Stop()

	// Stop API server
	d.api.Stop()

	// Stop all processes
	d.manager.StopAll()

	// Remove socket file
	os.Remove(d.socketPath)
}

// handleConnection handles a Unix socket connection
func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()
	// This is handled by the API/RPC layer
	// For now, we'll use HTTP API for all communication
}

// GetInfo returns daemon information
func (d *Daemon) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":       Version,
		"uptime":        int64(time.Since(d.startedAt).Seconds()),
		"started_at":    d.startedAt,
		"process_count": d.manager.Count(),
		"running_count": d.manager.RunningCount(),
	}
}
