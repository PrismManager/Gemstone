package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/PrismManager/gemstone/internal/config"
	"github.com/PrismManager/gemstone/internal/logger"
	"github.com/PrismManager/gemstone/internal/types"
)

// Process represents a managed process
type Process struct {
	mu           sync.RWMutex
	info         *types.ProcessInfo
	cmd          *exec.Cmd
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *logger.ProcessLogger
	statsHistory []types.ProcessStats
	maxHistory   int
}

// New creates a new process from a start request
func New(req *types.StartRequest, logDir string) (*Process, error) {
	id := uuid.New().String()[:8]

	now := time.Now()
	info := &types.ProcessInfo{
		ID:          id,
		Name:        req.Name,
		Status:      types.StatusStopped,
		Command:     req.Command,
		Args:        req.Args,
		WorkDir:     req.WorkDir,
		Env:         req.Env,
		AutoStart:   req.AutoStart,
		AutoRestart: req.AutoRestart,
		MaxRestarts: req.MaxRestarts,
		User:        req.User,
		Group:       req.Group,
		CreatedAt:   now,
	}

	procLogger, err := logger.NewProcessLogger(id, req.Name, logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Process{
		info:       info,
		logger:     procLogger,
		maxHistory: 1000,
	}, nil
}

// FromConfig creates a process from configuration
func FromConfig(cfg *config.Process, logDir string) (*Process, error) {
	req := &types.StartRequest{
		Name:        cfg.Name,
		Command:     cfg.Command,
		Args:        cfg.Args,
		WorkDir:     cfg.WorkDir,
		Env:         cfg.Env,
		AutoStart:   cfg.AutoStart,
		AutoRestart: cfg.AutoRestart,
		MaxRestarts: cfg.MaxRestarts,
		User:        cfg.User,
		Group:       cfg.Group,
	}

	p, err := New(req, logDir)
	if err != nil {
		return nil, err
	}

	if cfg.ID != "" {
		p.info.ID = cfg.ID
	}

	return p, nil
}

// Start starts the process
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.info.Status == types.StatusRunning {
		return fmt.Errorf("process %s is already running", p.info.Name)
	}

	p.info.Status = types.StatusStarting

	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancel = cancel

	cmd := exec.CommandContext(ctx, p.info.Command, p.info.Args...)

	if p.info.WorkDir != "" {
		cmd.Dir = p.info.WorkDir
	}

	cmd.Env = os.Environ()
	for k, v := range p.info.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if p.info.User != "" {
		cred, err := getUserCredentials(p.info.User, p.info.Group)
		if err != nil {
			p.info.Status = types.StatusErrored
			return fmt.Errorf("failed to get user credentials: %w", err)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: cred,
			Setpgid:    true,
		}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.info.Status = types.StatusErrored
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		p.info.Status = types.StatusErrored
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		p.info.Status = types.StatusErrored
		return fmt.Errorf("failed to start process: %w", err)
	}

	p.cmd = cmd
	p.info.PID = cmd.Process.Pid
	p.info.Status = types.StatusRunning
	now := time.Now()
	p.info.StartedAt = &now
	p.info.StoppedAt = nil

	go p.captureOutput(stdout, "stdout")
	go p.captureOutput(stderr, "stderr")
	go p.waitForExit()

	return nil
}

// Stop stops the process
func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.info.Status != types.StatusRunning {
		return fmt.Errorf("process %s is not running", p.info.Name)
	}

	p.info.Status = types.StatusStopping

	if p.cancel != nil {
		p.cancel()
	}

	if p.cmd != nil && p.cmd.Process != nil {
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM)

		go func() {
			time.Sleep(5 * time.Second)
			p.mu.Lock()
			defer p.mu.Unlock()
			if p.info.Status == types.StatusStopping {
				_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
			}
		}()
	}

	return nil
}

// Restart restarts the process
func (p *Process) Restart() error {
	if p.info.Status == types.StatusRunning {
		if err := p.Stop(); err != nil {
			return err
		}
		for i := 0; i < 10; i++ {
			time.Sleep(500 * time.Millisecond)
			p.mu.RLock()
			status := p.info.Status
			p.mu.RUnlock()
			if status == types.StatusStopped {
				break
			}
		}
	}
	return p.Start()
}

// Info returns process information
func (p *Process) Info() *types.ProcessInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info := *p.info

	if info.Status == types.StatusRunning && info.StartedAt != nil {
		info.Uptime = int64(time.Since(*info.StartedAt).Seconds())
	}

	if info.PID > 0 {
		if proc, err := process.NewProcess(int32(info.PID)); err == nil {
			if cpu, err := proc.CPUPercent(); err == nil {
				info.CPU = cpu
			}
			if mem, err := proc.MemoryInfo(); err == nil && mem != nil {
				info.Memory = mem.RSS
			}
			if memPercent, err := proc.MemoryPercent(); err == nil {
				info.MemoryPercent = float64(memPercent)
			}
		}
	}

	return &info
}

// Stats returns current process stats
func (p *Process) Stats() *types.ProcessStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.info.PID <= 0 || p.info.Status != types.StatusRunning {
		return nil
	}

	stats := &types.ProcessStats{
		ID:        p.info.ID,
		PID:       p.info.PID,
		Timestamp: time.Now(),
	}

	proc, err := process.NewProcess(int32(p.info.PID))
	if err != nil {
		return stats
	}

	if cpu, err := proc.CPUPercent(); err == nil {
		stats.CPU = cpu
	}
	if mem, err := proc.MemoryInfo(); err == nil && mem != nil {
		stats.Memory = mem.RSS
	}
	if memPercent, err := proc.MemoryPercent(); err == nil {
		stats.MemoryPercent = float64(memPercent)
	}
	if threads, err := proc.NumThreads(); err == nil {
		stats.NumThreads = threads
	}
	if fds, err := proc.NumFDs(); err == nil {
		stats.NumFDs = fds
	}
	if ioCounters, err := proc.IOCounters(); err == nil && ioCounters != nil {
		stats.ReadBytes = ioCounters.ReadBytes
		stats.WriteBytes = ioCounters.WriteBytes
	}

	return stats
}

// CollectStats collects and stores stats for historical data
func (p *Process) CollectStats() {
	stats := p.Stats()
	if stats == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.statsHistory = append(p.statsHistory, *stats)

	if len(p.statsHistory) > p.maxHistory {
		p.statsHistory = p.statsHistory[len(p.statsHistory)-p.maxHistory:]
	}
}

// GetStatsHistory returns historical stats
func (p *Process) GetStatsHistory(limit int) []types.ProcessStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 || limit > len(p.statsHistory) {
		result := make([]types.ProcessStats, len(p.statsHistory))
		copy(result, p.statsHistory)
		return result
	}

	start := len(p.statsHistory) - limit
	result := make([]types.ProcessStats, limit)
	copy(result, p.statsHistory[start:])
	return result
}

// GetLogs returns recent log entries
func (p *Process) GetLogs(lines int, logType string) ([]string, error) {
	return p.logger.GetLogs(lines, logType)
}

// ToConfig converts process to configuration format
func (p *Process) ToConfig() *config.Process {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &config.Process{
		ID:          p.info.ID,
		Name:        p.info.Name,
		Command:     p.info.Command,
		Args:        p.info.Args,
		WorkDir:     p.info.WorkDir,
		Env:         p.info.Env,
		AutoStart:   p.info.AutoStart,
		AutoRestart: p.info.AutoRestart,
		MaxRestarts: p.info.MaxRestarts,
		User:        p.info.User,
		Group:       p.info.Group,
	}
}

// ID returns the process ID
func (p *Process) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.ID
}

// Name returns the process name
func (p *Process) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Name
}

// Status returns the process status
func (p *Process) Status() types.ProcessStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.Status
}

// ShouldAutoStart returns whether the process should auto-start
func (p *Process) ShouldAutoStart() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info.AutoStart
}

func (p *Process) captureOutput(reader io.Reader, outputType string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		p.logger.Log(outputType, line)
	}
}

func (p *Process) waitForExit() {
	if p.cmd == nil {
		return
	}

	err := p.cmd.Wait()

	p.mu.Lock()
	now := time.Now()
	p.info.StoppedAt = &now
	p.info.PID = 0

	shouldRestart := p.info.AutoRestart && p.info.Status == types.StatusRunning

	if err != nil {
		p.logger.Log("stderr", fmt.Sprintf("Process exited with error: %v", err))
	}

	if shouldRestart && p.info.RestartCount < p.info.MaxRestarts {
		p.info.Status = types.StatusRestarting
		p.info.RestartCount++
		p.mu.Unlock()

		time.Sleep(time.Second)
		_ = p.Start()
		return
	}

	p.info.Status = types.StatusStopped
	p.mu.Unlock()
}

func getUserCredentials(username, groupname string) (*syscall.Credential, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return nil, err
		}
		gid, err = strconv.ParseUint(g.Gid, 10, 32)
		if err != nil {
			return nil, err
		}
	}

	return &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}, nil
}

// Close closes the process and its resources
func (p *Process) Close() error {
	if p.logger != nil {
		return p.logger.Close()
	}
	return nil
}
