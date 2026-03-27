Here's the updated README for sysix with all the professional elements:

```markdown
# sysix

[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Language](https://img.shields.io/badge/language-Go-00ADD8)](https://golang.org)
[![Status](https://img.shields.io/badge/status-In%20Development-yellow)](#)

A powerful system monitor with a live terminal UI and web dashboard. Real-time monitoring of your system performance, running processes, network activity, and open ports.

## Features

- **Live Terminal UI** - Real-time system monitoring in your terminal with `sysix watch`
- **Web Dashboard** - Beautiful web interface for system metrics with `sysix serve`
- **System Monitoring** - CPU, memory, disk usage, and more
- **Process Tracking** - Monitor running processes in real-time
- **Network Insights** - Track network activity and connections
- **Port Monitoring** - See which ports are open and listening
- **Configurable** - Customize refresh rates and visible panels

## Installation

### From Source

```bash
git clone https://github.com/System9-Software/sysix.git
cd sysix
go build -o sysix ./cmd
```

### Using Go Install

```bash
go install github.com/System9-Software/sysix@latest
```

## Quick Start

### Launch the Terminal UI

```bash
sysix watch
```

Monitor your system in real-time with an interactive terminal interface.

### Launch the Web Dashboard

```bash
sysix serve
```

Access the web dashboard at `http://localhost:8080`

### Get a Quick Snapshot

```bash
# Basic system status
sysix status

# Include process information
sysix status --procs

# Include open ports
sysix status --ports

# Include everything
sysix status --procs --ports
```

## Configuration

Create a `config.yaml` file in your sysix directory to customize behavior:

```yaml
# Refresh rate in milliseconds
refresh_rate: 1000

# Enable/disable panels
panels:
  cpu: true
  memory: true
  disk: true
  processes: true
  network: true
  ports: true
```

Edit `config.yaml` to adjust refresh rate and visible panels according to your needs.

## Building from Source

### Requirements
- Go 1.18 or higher

### Build Steps

```bash
git clone https://github.com/System9-Software/sysix.git
cd sysix
go build -o sysix ./cmd
./sysix watch
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

© System9 Software

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues, questions, or suggestions, please open an [issue](https://github.com/System9-Software/sysix/issues) on GitHub.
```
