package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/PrismManager/gemstone/internal/config"
	"github.com/PrismManager/gemstone/internal/types"
)

// Manager manages all processes
type Manager struct {
	mu        sync.RWMutex
	processes map[string]*Process
	config    *config.Config
	dataDir   string
	logDir    string
}

// NewManager creates a new process manager
func NewManager(cfg *config.Config, dataDir, logDir string) (*Manager, error) {
	// Make sure that the directories exist
	for _, dir := range []string{dataDir, logDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	m := &Manager{
		processes: make(map[string]*Process),
		config:    cfg,
		dataDir:   dataDir,
		logDir:    logDir,
	}

	// Load saved processes
	if err := m.loadProcesses(); err != nil {
		return nil, fmt.Errorf("failed to load processes: %w", err)
	}

	return m, nil
}

// Start starts a new process
func (m *Manager) Start(req *types.StartRequest) (*types.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if process with same name exists
	for _, p := range m.processes {
		if p.Name() == req.Name {
			return nil, fmt.Errorf("process with name %s already exists", req.Name)
		}
	}

	proc, err := New(req, m.logDir)
	if err != nil {
		return nil, err
	}

	if err := proc.Start(); err != nil {
		return nil, err
	}

	m.processes[proc.ID()] = proc

	// Save processes
	if err := m.saveProcesses(); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to save processes: %v\n", err)
	}

	return proc.Info(), nil
}

// Stop stops a process by ID or name
func (m *Manager) Stop(idOrName string) error {
	m.mu.RLock()
	proc := m.findProcess(idOrName)
	m.mu.RUnlock()

	if proc == nil {
		return fmt.Errorf("process %s not found", idOrName)
	}

	return proc.Stop()
}

// Restart restarts a process by ID or name
func (m *Manager) Restart(idOrName string) error {
	m.mu.RLock()
	proc := m.findProcess(idOrName)
	m.mu.RUnlock()

	if proc == nil {
		return fmt.Errorf("process %s not found", idOrName)
	}

	return proc.Restart()
}

// Delete removes a process
func (m *Manager) Delete(idOrName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var procID string
	for id, p := range m.processes {
		if id == idOrName || p.Name() == idOrName {
			if p.Status() == types.StatusRunning {
				if err := p.Stop(); err != nil {
					return err
				}
			}
			p.Close()
			procID = id
			break
		}
	}

	if procID == "" {
		return fmt.Errorf("process %s not found", idOrName)
	}

	delete(m.processes, procID)

	return m.saveProcesses()
}

// Get returns process info by ID or name
func (m *Manager) Get(idOrName string) *types.ProcessInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proc := m.findProcess(idOrName)
	if proc == nil {
		return nil
	}

	return proc.Info()
}

// List returns all processes
func (m *Manager) List() []*types.ProcessInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*types.ProcessInfo, 0, len(m.processes))
	for _, p := range m.processes {
		result = append(result, p.Info())
	}

	return result
}

// Stats returns stats for a process
func (m *Manager) Stats(idOrName string) *types.ProcessStats {
	m.mu.RLock()
	proc := m.findProcess(idOrName)
	m.mu.RUnlock()

	if proc == nil {
		return nil
	}

	return proc.Stats()
}

// AllStats returns stats for all running processes
func (m *Manager) AllStats() []*types.ProcessStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*types.ProcessStats, 0)
	for _, p := range m.processes {
		if stats := p.Stats(); stats != nil {
			result = append(result, stats)
		}
	}

	return result
}

// GetStatsHistory returns historical stats for a process
func (m *Manager) GetStatsHistory(idOrName string, limit int) []types.ProcessStats {
	m.mu.RLock()
	proc := m.findProcess(idOrName)
	m.mu.RUnlock()

	if proc == nil {
		return nil
	}

	return proc.GetStatsHistory(limit)
}

// GetLogs returns logs for a process
func (m *Manager) GetLogs(idOrName string, lines int, logType string) ([]string, error) {
	m.mu.RLock()
	proc := m.findProcess(idOrName)
	m.mu.RUnlock()

	if proc == nil {
		return nil, fmt.Errorf("process %s not found", idOrName)
	}

	return proc.GetLogs(lines, logType)
}

// CollectAllStats collects stats for all running processes
func (m *Manager) CollectAllStats() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processes {
		if p.Status() == types.StatusRunning {
			p.CollectStats()
		}
	}
}

// StartAutoStartProcesses starts all processes marked for auto-start
func (m *Manager) StartAutoStartProcesses() {
	m.mu.RLock()
	toStart := make([]*Process, 0)
	for _, p := range m.processes {
		if p.ShouldAutoStart() && p.Status() == types.StatusStopped {
			toStart = append(toStart, p)
		}
	}
	m.mu.RUnlock()

	for _, p := range toStart {
		if err := p.Start(); err != nil {
			fmt.Printf("Failed to auto-start process %s: %v\n", p.Name(), err)
		}
	}
}

// StopAll stops all running processes
func (m *Manager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processes {
		if p.Status() == types.StatusRunning {
			_ = p.Stop()
		}
	}
}

// Count returns the number of managed processes
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processes)
}

// RunningCount returns the number of running processes
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, p := range m.processes {
		if p.Status() == types.StatusRunning {
			count++
		}
	}
	return count
}

func (m *Manager) findProcess(idOrName string) *Process {
	// Try by ID first
	if p, ok := m.processes[idOrName]; ok {
		return p
	}

	// Try by name
	for _, p := range m.processes {
		if p.Name() == idOrName {
			return p
		}
	}

	return nil
}

func (m *Manager) saveProcesses() error {
	configs := make([]*config.Process, 0, len(m.processes))
	for _, p := range m.processes {
		configs = append(configs, p.ToConfig())
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.dataDir, "processes.json"), data, 0644)
}

func (m *Manager) loadProcesses() error {
	path := filepath.Join(m.dataDir, "processes.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var configs []*config.Process
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	for _, cfg := range configs {
		proc, err := FromConfig(cfg, m.logDir)
		if err != nil {
			fmt.Printf("Warning: failed to load process %s: %v\n", cfg.Name, err)
			continue
		}
		m.processes[proc.ID()] = proc
	}

	return nil
}
