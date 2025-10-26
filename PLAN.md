# Kubernetes TUI - Detailed Project Plan

## Project Overview
A terminal user interface (TUI) application for Kubernetes cluster management, built in Go. Initially read-only, designed for fast navigation and real-time monitoring of Kubernetes resources.

## Architecture

### Component Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                         Main App                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                   TUI Controller                      │  │
│  │  (Bubble Tea Model & Update Loop)                    │  │
│  └─────────────────────────────────────────────────────┘  │
│                              │                              │
│  ┌──────────────┬────────────┴──────────┬──────────────┐  │
│  │              │                        │              │  │
│  ▼              ▼                        ▼              ▼  │
│ ┌────┐ ┌──────────────┐ ┌──────────┐ ┌──────────────┐    │
│ │View│ │Resource      │ │K8s       │ │State         │    │
│ │    │ │Components    │ │Service   │ │Manager       │    │
│ │    │ ├──────────────┤ │          │ │              │    │
│ │    │ │• List View   │ │          │ │• Namespace   │    │
│ │    │ │• Detail View │ │          │ │• Context     │    │
│ │    │ │• Log View    │ │          │ │• Filters     │    │
│ │    │ │• Event View  │ │          │ │• Selection   │    │
│ └────┘ └──────────────┘ └──────────┘ └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Kubernetes API   │
                    └───────────────────┘
```

### Data Flow
1. **Initialization**: Load kubeconfig → Create client → Connect to cluster
2. **Resource Loading**: Fetch resources → Cache locally → Display in UI
3. **User Interaction**: Keyboard input → Update selection → Refresh view
4. **Real-time Updates**: Watch API → Receive events → Update cache → Refresh UI

## Detailed Component Specifications

### 1. Core Application Structure

#### Directory Layout
```
k8s-tui/
├── cmd/
│   └── k8s-tui/
│       └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   └── app.go               # Main application model
│   ├── k8s/
│   │   ├── client.go            # Kubernetes client wrapper
│   │   ├── resources.go         # Resource fetching logic
│   │   └── watch.go             # Watch/event handling
│   ├── ui/
│   │   ├── components/
│   │   │   ├── list.go          # List view component
│   │   │   ├── detail.go        # Detail view component
│   │   │   ├── header.go        # Header component
│   │   │   └── footer.go        # Footer/help component
│   │   ├── styles/
│   │   │   └── theme.go         # Color schemes and styles
│   │   └── keys/
│   │       └── bindings.go      # Keyboard shortcuts
│   ├── models/
│   │   ├── resource.go          # Resource data models
│   │   └── state.go             # Application state
│   └── config/
│       └── config.go            # Configuration management
├── pkg/
│   └── utils/
│       ├── format.go            # Formatting utilities
│       └── table.go             # Table rendering
├── go.mod
├── go.sum
├── README.md
├── PLAN.md
└── Makefile
```

### 2. Resource Types and Views

#### Supported Resources (Phase 1)
| Resource Type | List Columns | Detail Sections |
|--------------|--------------|-----------------|
| **Pods** | Name, Ready, Status, Restarts, Age, IP | Containers, Conditions, Events, Logs |
| **Services** | Name, Type, ClusterIP, Port(s), Age | Endpoints, Selectors, Ports |
| **Deployments** | Name, Ready, Up-to-date, Available, Age | Replicas, Strategy, Conditions, Events |
| **StatefulSets** | Name, Ready, Age | Replicas, Update Strategy, Volume Claims |
| **ConfigMaps** | Name, Data Count, Age | Data Keys, Values (truncated) |
| **Secrets** | Name, Type, Data Count, Age | Keys only (no values) |
| **Nodes** | Name, Status, Roles, Age, Version | Capacity, Allocatable, Conditions, System Info |
| **Namespaces** | Name, Status, Age | Resource Quotas, Labels |

### 3. UI Layout Specifications

#### Main Layout Grid
```
┌─[Header]────────────────────────────────────────────────┐ 1 row
│ K8S-TUI | Context: minikube | NS: default | ◉ Connected │
├─[Tabs]──────────────────────────────────────────────────┤ 1 row
│ Workloads | Services | Config | Storage | Cluster      │
├─[Content]───────────────────────────────────────────────┤ dynamic
│ ┌─[Sidebar]─────┐ ┌─[Main]─────────────────────────┐  │
│ │ ▾ Pods     3  │ │ Name:    nginx-6799fc88d8-x2p  │  │
│ │   nginx    ✓  │ │ Status:  Running                │  │
│ │   redis    ✓  │ │ Ready:   1/1                    │  │
│ │   postgres ⚠  │ │ Age:     2 days                 │  │
│ │               │ │                                 │  │
│ │ ▸ Deployments │ │ [Containers]                    │  │
│ │ ▸ Services    │ │ Name     Image      CPU  Memory │  │
│ │               │ │ nginx    nginx:1.21  10m  128Mi │  │
│ └───────────────┘ └─────────────────────────────────┘  │
├─[Footer]────────────────────────────────────────────────┤ 2 rows
│ [↑↓] Navigate [Enter] Select [Tab] Switch [/] Search   │
│ [n] Namespace [c] Context [r] Refresh [?] Help [q] Quit │
└──────────────────────────────────────────────────────────┘
```

### 4. Keyboard Navigation

#### Global Shortcuts
| Key | Action | Context |
|-----|--------|---------|
| `q`, `Ctrl+C` | Quit application | Global |
| `?` | Show help screen | Global |
| `Tab`, `Shift+Tab` | Switch between panes | Global |
| `/` | Enter search mode | List view |
| `Esc` | Cancel/Back | Any mode |
| `r`, `F5` | Force refresh | Global |
| `n` | Namespace selector | Global |
| `c` | Context selector | Global |
| `1-9` | Quick switch tabs | Global |

#### Navigation Shortcuts
| Key | Action | Context |
|-----|--------|---------|
| `↑`, `k` | Move up | List/Detail |
| `↓`, `j` | Move down | List/Detail |
| `PgUp`, `Ctrl+U` | Page up | List/Detail |
| `PgDn`, `Ctrl+D` | Page down | List/Detail |
| `Home`, `g` | Go to top | List/Detail |
| `End`, `G` | Go to bottom | List/Detail |
| `Enter`, `→`, `l` | Open detail/expand | List |
| `←`, `h`, `Backspace` | Go back/collapse | Detail |

#### Resource-Specific Shortcuts
| Key | Action | Context |
|-----|--------|---------|
| `L` | View logs | Pod detail |
| `E` | View events | Any resource |
| `Y` | View YAML | Any resource |
| `D` | Describe resource | Any resource |
| `Ctrl+S` | Export to file | YAML view |

### 5. Technical Implementation Details

#### Kubernetes Client Configuration
```go
// Client initialization with multiple contexts
type K8sClient struct {
    clientset     *kubernetes.Clientset
    config        *rest.Config
    namespace     string
    contexts      []string
    currentContext string
}

// Features to implement:
// - Auto-detect kubeconfig from:
//   * KUBECONFIG env var
//   * ~/.kube/config
//   * In-cluster config
// - Context switching without restart
// - Namespace caching for performance
// - Connection retry with exponential backoff
```

#### State Management
```go
type AppState struct {
    // Navigation
    CurrentTab      int
    CurrentResource string
    SelectedItem    int

    // View state
    ViewMode        ViewMode  // List, Detail, Logs, YAML
    SplitRatio      float64   // Sidebar width

    // Filters
    Namespace       string
    SearchQuery     string
    ShowSystemPods  bool

    // Cache
    Resources       map[string][]runtime.Object
    LastUpdate      time.Time

    // Settings
    RefreshInterval time.Duration
    Theme           Theme
}
```

#### Resource Watching Strategy
```go
// Implement efficient watching with:
// 1. Shared informers for resource caching
// 2. Rate limiting to prevent API overload
// 3. Selective watching based on current view
// 4. Graceful degradation on watch failure

type ResourceWatcher struct {
    informerFactory informers.SharedInformerFactory
    stopCh         chan struct{}
    handlers       map[string]cache.ResourceEventHandler
}
```

### 6. Performance Considerations

#### Optimization Strategies
1. **Lazy Loading**: Only fetch details when selected
2. **Pagination**: Limit list results (configurable)
3. **Caching**:
   - Resource lists cached for 30 seconds
   - Namespace list cached for 5 minutes
   - Node info cached for 1 minute
4. **Debouncing**: Search input debounced by 300ms
5. **Virtual Scrolling**: For large lists (>1000 items)

#### Resource Limits
- Max pods per list: 500 (configurable)
- Max log lines: 1000 (tail)
- Max YAML size: 1MB
- Watch timeout: 30 seconds
- API request timeout: 10 seconds

### 7. Error Handling

#### Error Categories
1. **Connection Errors**: Display reconnection UI
2. **Permission Errors**: Show required RBAC permissions
3. **Resource Not Found**: Graceful message, auto-refresh
4. **API Rate Limiting**: Implement backoff, show warning
5. **Parsing Errors**: Log to debug file, show generic message

### 8. Configuration

#### Config File (`~/.k8s-tui/config.yaml`)
```yaml
# Display settings
ui:
  theme: dark                # dark, light, auto
  refresh_interval: 5s       # Auto-refresh interval
  show_system_pods: false    # Show kube-system pods
  sidebar_width: 30          # Percentage

# Performance
performance:
  max_list_items: 500
  cache_ttl: 30s
  watch_timeout: 30s

# Shortcuts (customizable)
keybindings:
  quit: ["q", "ctrl+c"]
  help: ["?", "h"]
  search: ["/"]

# Features
features:
  auto_refresh: true
  watch_events: true
  syntax_highlighting: true
```

### 9. Testing Strategy

#### Test Coverage Goals
- Unit tests: 80% coverage minimum (current: 24.1%)
- Integration tests for K8s client
- Mock K8s API server for testing
- TUI component testing with snapshot tests

#### Test Structure
```
internal/
├── models/
│   └── resource_test.go          ✅ (67.3% coverage)
├── ui/
│   ├── components/
│   │   └── list_test.go          ✅ (32.0% coverage)
│   └── styles/
│       └── theme_test.go         ✅ (100% coverage)
└── k8s/
    └── client_test.go            ⏳ TODO

tests/                            ⏳ TODO
├── integration/
│   └── k8s_client_test.go
└── e2e/
    └── navigation_test.go
```

#### CI/CD Infrastructure ✅
- [x] GitHub Actions workflow (.github/workflows/ci.yml)
- [x] Test job (3 OS × 3 Go versions = 9 matrix jobs)
- [x] Lint job (golangci-lint)
- [x] Build job (cross-platform verification)
- [x] Coverage reporting (Codecov integration)
- [x] Branch protection documentation
- [x] Automated setup scripts (PowerShell + Bash)

### 10. Development Phases

#### Phase 1: Foundation (Week 1-2)
- [x] Project structure setup
- [x] Basic TUI with Bubble Tea
- [x] Kubernetes client connection
- [x] Pod list view
- [x] Basic navigation

#### Phase 2: Core Features (Week 3-4)
- [ ] All workload resources
- [ ] Service resources
- [ ] Detail views
- [ ] Namespace switching
- [ ] Search functionality

#### Phase 3: Advanced Features (Week 5-6)
- [ ] Real-time updates via watch
- [ ] Log streaming
- [ ] YAML view
- [ ] Event view
- [ ] Context switching

#### Phase 4: Polish (Week 7-8)
- [ ] Configuration file
- [ ] Themes
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] Documentation

### 11. Future Enhancements (Post-MVP)

#### Read-Write Operations
- Scale deployments
- Delete pods
- Edit ConfigMaps/Secrets
- Apply YAML manifests

#### Advanced Features
- Port forwarding UI
- Exec into containers
- Resource graph visualization
- Metrics integration (CPU/Memory graphs)
- CRD support
- Multi-cluster view
- Resource diff view
- Helm release management

### 12. Dependencies

```go
// go.mod
module github.com/williajm/k8s-tui

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.17.1
    github.com/charmbracelet/lipgloss v0.9.1
    k8s.io/client-go v0.29.0
    k8s.io/apimachinery v0.29.0
    k8s.io/api v0.29.0
    github.com/spf13/cobra v1.8.0      // CLI framework
    github.com/spf13/viper v1.18.2     // Configuration
    github.com/olekukonko/tablewriter v0.0.5  // Table rendering
    github.com/muesli/termenv v0.15.2  // Terminal detection
)
```

## Success Metrics
- Startup time: <1 second
- Resource list load: <500ms
- Memory usage: <100MB for typical cluster
- Keyboard navigation: All features accessible without mouse
- Zero external dependencies for runtime (single binary)