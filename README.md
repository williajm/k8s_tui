# K8S-TUI

[![CI](https://github.com/williajm/k8s_tui/actions/workflows/ci.yml/badge.svg)](https://github.com/williajm/k8s_tui/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/williajm/k8s-tui)](https://goreportcard.com/report/github.com/williajm/k8s-tui)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/williajm/k8s_tui)](https://github.com/williajm/k8s_tui)

A fast, keyboard-driven terminal user interface for Kubernetes cluster management.

## Features

- **Multi-Resource Support**: View and inspect Pods, Services, Deployments, ConfigMaps, Secrets, and more
- **Real-time Updates**: Automatic refresh with Kubernetes watch API integration
- **Keyboard-Driven**: Complete navigation without mouse, vim-style keybindings
- **Multi-Context**: Switch between different clusters without restarting
- **Namespace Filtering**: Quickly switch between or view all namespaces
- **Log Streaming**: View pod logs directly in the TUI
- **Fast & Lightweight**: Single binary, minimal resource usage

## Installation

### From Source
```bash
git clone https://github.com/williajm/k8s-tui.git
cd k8s-tui
go build -o k8s-tui cmd/k8s-tui/main.go
```

### Prerequisites
- Go 1.21 or higher
- Access to a Kubernetes cluster
- Valid kubeconfig file

## Usage

### Basic Usage
```bash
# Use default kubeconfig
./k8s-tui

# Specify custom kubeconfig
./k8s-tui --kubeconfig ~/.kube/custom-config

# Start in specific namespace
./k8s-tui --namespace production

# Use specific context
./k8s-tui --context staging-cluster
```

### Keyboard Shortcuts

#### Global Navigation
- `Tab` / `Shift+Tab` - Switch between panes
- `1-9` - Quick switch to numbered tab
- `/` - Search/filter in current list
- `Esc` - Cancel/go back
- `?` - Show help screen
- `q` - Quit application

#### List Navigation
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` / `→` / `l` - View details
- `PgUp` / `PgDn` - Page up/down
- `g` / `G` - Go to top/bottom

#### Resource Actions
- `n` - Change namespace
- `c` - Change context
- `r` - Refresh current view
- `L` - View pod logs (when pod selected)
- `E` - View events for resource
- `Y` - View resource as YAML
- `D` - Describe resource

## Configuration

Configuration file location: `~/.k8s-tui/config.yaml`

```yaml
ui:
  theme: dark                # Options: dark, light, auto
  refresh_interval: 5s       # Auto-refresh interval
  show_system_pods: false    # Show kube-system pods
  sidebar_width: 30          # Sidebar width percentage

performance:
  max_list_items: 500       # Maximum items in lists
  cache_ttl: 30s           # Resource cache duration

keybindings:              # Customize key bindings
  quit: ["q", "ctrl+c"]
  help: ["?"]
  search: ["/"]
```

## Development

### Building from Source
```bash
# Clone repository
git clone https://github.com/williajm/k8s-tui.git
cd k8s-tui

# Install dependencies
go mod download

# Build binary
go build -o k8s-tui cmd/k8s-tui/main.go

# Run tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with race detector
go run -race cmd/k8s-tui/main.go
```

### Testing

The project includes comprehensive unit tests for core functionality:

- **Models**: 67.3% coverage - Pod info parsing, status symbols, age formatting
- **UI Components**: 32.0% coverage - List navigation, filtering, pagination
- **Styles**: 100% coverage - Theme colors, status styles, rendering helpers

Run tests locally:
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# Run specific package
go test ./internal/models -v
```

### CI/CD

GitHub Actions automatically runs on every push and pull request:
- ✅ **Tests** on Ubuntu, macOS, and Windows with Go 1.21, 1.22, 1.23
- ✅ **Linting** with golangci-lint
- ✅ **Build** verification on all platforms
- ✅ **Coverage** reporting to Codecov

See [.github/workflows/ci.yml](.github/workflows/ci.yml) for details.

### Branch Protection

The `main` branch is protected with the following rules:
- 🔒 Require pull requests before merging (0 approvals for solo dev)
- ✅ Require status checks to pass (test, lint, build)
- 📝 Require linear history (rebase/squash only)
- 🚫 Prevent force pushes and branch deletion

**Development Workflow:**
```bash
# Work on feature branches
git checkout -b feature/my-feature
git push origin feature/my-feature

# Create PR on GitHub
# CI runs automatically
# Merge when checks pass
```

See [.github/DEVELOPMENT_WORKFLOW.md](.github/DEVELOPMENT_WORKFLOW.md) for detailed workflows.

### Project Structure
```
k8s-tui/
├── cmd/k8s-tui/       # Application entry point
├── internal/          # Internal packages
│   ├── app/          # Main application logic
│   ├── k8s/          # Kubernetes client
│   ├── ui/           # TUI components
│   └── models/       # Data models
└── pkg/              # Public packages
```

## Roadmap

### Current Phase (v0.1.0) - Read-Only
- [x] Project setup
- [ ] Basic resource viewing
- [ ] Navigation and search
- [ ] Real-time updates

### Future Releases
- [ ] v0.2.0 - Enhanced viewing (logs, events, describe)
- [ ] v0.3.0 - Configuration and themes
- [ ] v0.4.0 - Performance optimizations
- [ ] v1.0.0 - Production ready
- [ ] v1.1.0 - Write operations (scale, delete, edit)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Kubernetes client via [client-go](https://github.com/kubernetes/client-go)
- Inspired by [k9s](https://k9scli.io/) and [kubectl](https://kubernetes.io/docs/reference/kubectl/)