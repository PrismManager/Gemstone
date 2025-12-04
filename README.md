# Gemstone (gem) - Process Manager for Linux

A modern, lightweight process manager for Linux servers written in Go. Similar to PM2 but with native performance and simpler deployment.

## Features

- **Process Management**: Start, stop, restart, and delete processes
- **Auto-restart**: Automatically restart crashed processes
- **Auto-start on boot**: Processes start automatically after system reboot
- **Logging**: Separate stdout/stderr logs with rotation support
- **Resource Monitoring**: CPU, memory, threads, and I/O statistics
- **REST API**: Full API for remote management and web interfaces
- **Historical Stats**: Time-series data for charts and monitoring
- **Systemd Integration**: Native systemd service management

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/PrismManager/gemstone.git
cd gemstone

# Build
make build

# Install (requires root)
sudo make install

# Start the daemon
sudo systemctl start gemstone

# Enable auto-start on boot
sudo systemctl enable gemstone
```

### Basic Usage

```bash
# Start a process
gem start 'node app.js' --name myapp

# Start with options
gem start 'python server.py' --name api --cwd /opt/app --auto-restart
```

## Configuration

Configuration file: `/etc/gemstone/config.yaml`

```yaml
api:
  enabled: true
  port: 9876
  host: "127.0.0.1"
  auth_token: ""  # Set for authentication
  enable_cors: false

logging:
  max_size: 10        # MB
  max_backups: 5
  max_age: 30         # days
  compress: true
  directory: "/var/log/gemstone"
```

## REST API

The daemon exposes a REST API for remote management:

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/system` | System information |
| GET | `/api/v1/system/stats` | Current system stats |
| GET | `/api/v1/system/stats/history` | Historical system stats |
| GET | `/api/v1/processes` | List all processes |
| POST | `/api/v1/processes` | Start a new process |
| GET | `/api/v1/processes/:id` | Get process details |
| DELETE | `/api/v1/processes/:id` | Delete a process |
| POST | `/api/v1/processes/:id/stop` | Stop a process |
| POST | `/api/v1/processes/:id/restart` | Restart a process |
| GET | `/api/v1/processes/:id/stats` | Get process stats |
| GET | `/api/v1/processes/:id/stats/history` | Historical process stats |
| GET | `/api/v1/processes/:id/logs` | Get process logs |

### Example: Start a process via API

```bash
curl -X POST http://localhost:9876/api/v1/processes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "myapp",
    "command": "node",
    "args": ["app.js"],
    "work_dir": "/opt/myapp",
    "auto_start": true,
    "auto_restart": true,
    "max_restarts": 10
  }'
```

### Authentication

Set `auth_token` in config to enable authentication:

```yaml
api:
  auth_token: "your-secret-token"
```

Then include the token in requests:

```bash
curl -H "Authorization: Bearer your-secret-token" http://localhost:9876/api/v1/processes
```

## Directories

| Path | Description |
|------|-------------|
| `/etc/gemstone/` | Configuration files |
| `/var/lib/gemstone/` | Process data (saved state) |
| `/var/log/gemstone/` | Process logs |
| `/run/gemstone/` | Runtime files (socket, PID) |

## Building from Source

### Requirements

- Go 1.21 or later
- Make

### Build

```bash
# Download dependencies
make deps

# Build both binaries
make build

# Run tests
make test

# Format code
make fmt
```

## Web Manager

The web manager is a separate project that provides a web interface for managing processes. It communicates with gemstone via the REST API.

See: [gemstone-web](https://github.com/PrismManager/gemstone-web) (coming soon)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.