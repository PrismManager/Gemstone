package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PrismManager/gemstone/internal/config"
	"github.com/PrismManager/gemstone/internal/types"
)

// Client is the CLI client for communicating with the daemon
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// StartRequest mirrors types.StartRequest for the CLI
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
}

// NewClient creates a new CLI client
func NewClient() (*Client, error) {
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		// Use default config
		cfg = config.DefaultConfig()
	}

	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d/api/v1", cfg.API.Host, cfg.API.Port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authToken: cfg.API.AuthToken,
	}, nil
}

func (c *Client) doRequest(method, path string, body interface{}) (*types.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	var response types.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Start starts a new process
func (c *Client) Start(req *StartRequest) (*types.ProcessInfo, error) {
	resp, err := c.doRequest("POST", "/processes", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var info types.ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Stop stops a process
func (c *Client) Stop(idOrName string) error {
	resp, err := c.doRequest("POST", "/processes/"+idOrName+"/stop", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// Restart restarts a process
func (c *Client) Restart(idOrName string) error {
	resp, err := c.doRequest("POST", "/processes/"+idOrName+"/restart", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// Delete deletes a process
func (c *Client) Delete(idOrName string) error {
	resp, err := c.doRequest("DELETE", "/processes/"+idOrName, nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// List lists all processes
func (c *Client) List() ([]*types.ProcessInfo, error) {
	resp, err := c.doRequest("GET", "/processes", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var processes []*types.ProcessInfo
	if err := json.Unmarshal(data, &processes); err != nil {
		return nil, err
	}

	return processes, nil
}

// Get gets a process by ID or name
func (c *Client) Get(idOrName string) (*types.ProcessInfo, error) {
	resp, err := c.doRequest("GET", "/processes/"+idOrName, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var info types.ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetLogs gets logs for a process
func (c *Client) GetLogs(idOrName string, lines int, logType string) ([]string, error) {
	path := fmt.Sprintf("/processes/%s/logs?lines=%d", idOrName, lines)
	if logType != "" {
		path += "&type=" + logType
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var logs []string
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetAllStats gets stats for all running processes
func (c *Client) GetAllStats() ([]*types.ProcessStats, error) {
	// Get all processes first
	processes, err := c.List()
	if err != nil {
		return nil, err
	}

	var stats []*types.ProcessStats
	for _, p := range processes {
		if p.Status != types.StatusRunning {
			continue
		}

		resp, err := c.doRequest("GET", "/processes/"+p.ID+"/stats", nil)
		if err != nil {
			continue
		}

		if !resp.Success {
			continue
		}

		data, err := json.Marshal(resp.Data)
		if err != nil {
			continue
		}

		var s types.ProcessStats
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}

		stats = append(stats, &s)
	}

	return stats, nil
}

// GetSystemInfo gets system information
func (c *Client) GetSystemInfo() (*types.DaemonInfo, error) {
	resp, err := c.doRequest("GET", "/system", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var info types.DaemonInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}
