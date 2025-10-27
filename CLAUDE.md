# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

K8S-TUI is a fast, keyboard-driven terminal user interface for Kubernetes cluster management built with Go. It uses the Bubble Tea TUI framework for rendering and the official Kubernetes client-go library for cluster interaction.

## Build, Test, and Lint Commands

### Building
```bash
# Build the main binary
go build -o k8s-tui cmd/k8s-tui/main.go

# Build for specific platform
go build -v -o k8s-tui.exe ./cmd/k8s-tui  # Windows
go build -v -o k8s-tui ./cmd/k8s-tui     # Linux/macOS

# Run directly
go run cmd/k8s-tui/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector (important for concurrency bugs)
go test -race ./...

# Run with coverage in atomic mode (CI uses this)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run specific package tests
go test ./internal/models -v
go test ./internal/ui/components -v
go test ./internal/ui/styles -v
```

### Linting
```bash
# Run golangci-lint (what CI uses)
golangci-lint run ./...

# Run with timeout (CI uses 5m)
golangci-lint run --timeout=5m ./...

# Format code
go fmt ./...
gofmt -w -s .
```

### Dependency Management
```bash
# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
go mod tidy
```

## Architecture

### Application Structure

The codebase follows a clean architecture pattern:

- **cmd/k8s-tui/**: Application entry point using Cobra for CLI handling
- **internal/app/**: Core Bubble Tea application model (Model-Update-View pattern)
- **internal/k8s/**: Kubernetes client wrapper with connection management
- **internal/models/**: Data models for pod info, status formatting, and age calculations
- **internal/ui/**: UI layer separated into:
  - **components/**: Reusable UI components (Header, Footer, PodList)
  - **keys/**: Keyboard bindings
  - **styles/**: Lipgloss theme and styling

### Bubble Tea Architecture

The app uses the Elm Architecture (TEA) pattern:

1. **Model** (`app.Model`): Holds application state (client, UI components, dimensions, error state)
2. **Update** (`app.Update`): Processes messages (key presses, window size, pod data) and returns new model + commands
3. **View** (`app.View`): Renders the current state to a string

**Message types**:
- `podsLoadedMsg`: Async pod data loaded from Kubernetes
- `tickMsg`: Timer for 5-second auto-refresh
- `tea.KeyMsg`: Keyboard input
- `tea.WindowSizeMsg`: Terminal resize events

### Kubernetes Client Pattern

The `k8s.Client` wraps `kubernetes.Clientset` and provides:
- Kubeconfig loading priority: in-cluster → KUBECONFIG env → ~/.kube/config
- Context and namespace management
- Connection testing with timeout
- Resource fetching methods (GetPods, GetNamespaces, GetPodLogs)

### UI Component Design

Components like `PodList` maintain their own state:
- Selection index and viewport for scrolling
- Search filtering logic
- Size-aware rendering with viewport adjustment
- Table-style layout with status symbols (✓, ✗, ○, ⚠, ⊗)

## Development Workflow

### Branch Strategy
- **main**: Protected branch requiring PR and passing CI
- **dev**: Experimental work branch for ongoing development
- **feature/*** or **hotfix/***: Feature branches for PRs

### Before Creating PR
```bash
# Always run these locally first
go test ./...
go fmt ./...
golangci-lint run --timeout=5m ./...
```

### CI Requirements
All PRs must pass:
- **Test**: Go 1.21, 1.22, 1.23 on Ubuntu, macOS, Windows with race detector
- **Lint**: golangci-lint with 5m timeout
- **Build**: Cross-platform build verification

The CI configuration uses:
- Test coverage reporting to Codecov (Ubuntu + Go 1.23 only)
- Artifact uploads for built binaries (7-day retention)
- Parallel matrix strategy for efficiency

### Commit Message Convention
Follow Conventional Commits format:
```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, perf, test, chore, ci
Examples:
  feat(ui): Add pod log streaming view
  fix(k8s): Handle connection timeout gracefully
  test: Add unit tests for pod list component
```

## Linting Configuration

The project uses golangci-lint with specific settings:
- **Line length**: 140 characters
- **Function length**: 100 lines, 50 statements
- **Complexity**: Max cyclomatic complexity of 15
- **Import ordering**: Local prefix `github.com/williajm/k8s-tui`
- **Test exclusions**: funlen and dupl disabled for `_test.go` files

Key enabled linters: bodyclose, errcheck, gosec, gosimple, ineffassign, staticcheck, unused, revive

## Running the Application

```bash
# Use default kubeconfig
./k8s-tui

# Specify custom kubeconfig
./k8s-tui --kubeconfig ~/.kube/custom-config

# Start in specific namespace
./k8s-tui --namespace production
./k8s-tui -n production

# Use specific context
./k8s-tui --context staging-cluster
```

## Key Dependencies

- **Bubble Tea** (v1.3.10): TUI framework for Model-Update-View architecture
- **Lipgloss** (v1.1.0): Terminal styling and layout
- **Bubbles** (v0.21.0): Pre-built TUI components
- **Cobra** (v1.10.1): CLI argument parsing
- **client-go** (v0.34.1): Official Kubernetes Go client
- **k8s.io/api** (v0.34.1): Kubernetes API types

## Current Phase

The project is in **Phase 1 - Foundation** (v0.1.0 - Read-Only):
- Foundation established with testing and CI/CD
- Working on basic resource viewing functionality
- Focus is on pods display with real-time updates
- Navigation and keyboard shortcuts implemented

Future phases will add log viewing, events, describe functionality, configuration, themes, and eventually write operations.
