# 🛠️ Development Guide

## Quick Start

### Prerequisites

Ensure you have these tools installed:

- **Go 1.24+** - Backend development
- **Node.js 18+** & **npm/yarn** - Frontend development  
- **PostgreSQL 16+** - Primary database
- **ClickHouse 24+** - Analytics database
- **Redis 7+** - Caching and messaging
- **Docker & Docker Compose** - Container development
- **Make** - Build automation

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/brokle-ai/brokle-platform.git
cd brokle-platform

# Setup development environment (this will take a few minutes)
make setup

# Verify setup
make health
```

The `make setup` command will:
1. Install Go dependencies
2. Install Node.js dependencies for frontend
3. Start databases with Docker Compose
4. Run database migrations
5. Seed databases with development data
6. Verify all services are healthy

### Development Workflow

```bash
# Start all development servers
make dev

# Or start services individually
make dev-backend    # Go API server on :8080
make dev-frontend   # Next.js dashboard on :3000
```

Access the application:
- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **WebSocket**: ws://localhost:8080/ws
- **API Documentation**: http://localhost:8080/docs

## Project Structure

```
brokle/
├── cmd/                    # Application entry points
│   ├── server/            # Main API server
│   ├── migrate/           # Database migration tool
│   └── seed/              # Database seeding tool
├── internal/              # Private application code
│   ├── core/              # Business logic
│   │   ├── domain/        # Domain models and interfaces
│   │   └── services/      # Business logic implementation
│   ├── infrastructure/    # External dependencies
│   ├── transport/         # HTTP/WebSocket handlers
│   ├── workers/           # Background job workers
│   └── config/            # Configuration management
├── web/                   # Next.js frontend
├── pkg/                   # Reusable Go packages
├── migrations/            # Database migrations
├── seeders/              # Database seeders
└── docs/                 # Documentation
```

## Backend Development

### Domain-Driven Design Structure

The backend follows Domain-Driven Design principles:

#### Domain Layer (`internal/core/domain/`)

Each domain contains:
- **Entity** - Core business object
- **Repository Interface** - Data access contract
- **Service Interface** - Business logic contract

Example domain structure:
```go
// internal/core/domain/user/user.go
type User struct {
    ID        ulid.ULID `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

// internal/core/domain/user/repository.go
type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id ulid.ULID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
}

// internal/core/domain/user/service.go
type Service interface {
    Register(ctx context.Context, email, name, password string) (*User, error)
    GetProfile(ctx context.Context, userID ulid.ULID) (*User, error)
}
```

#### Service Layer (`internal/core/services/`)

Business logic implementation:
```go
// internal/core/services/user/user_service.go
type service struct {
    repo user.Repository
}

func NewService(repo user.Repository) user.Service {
    return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, email, name, password string) (*user.User, error) {
    // Business logic implementation
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    // ... validation, duplicate checking, etc.
    
    return s.repo.Create(ctx, &user.User{
        Email: email,
        Name:  name,
        // ... other fields
    })
}
```

#### Infrastructure Layer (`internal/infrastructure/`)

External integrations:
- **Database repositories** - PostgreSQL and ClickHouse implementations
- **External APIs** - OpenAI, Anthropic, Stripe clients
- **Caching** - Redis client and cache service
- **Email** - SMTP client and templating
- **Monitoring** - Prometheus metrics and Jaeger tracing

### Adding New Features

#### 1. Define Domain Model

Create the domain in `internal/core/domain/{domain}/`:
```go
// internal/core/domain/project/project.go
type Project struct {
    ID             ulid.ULID `json:"id"`
    Name           string    `json:"name"`
    OrganizationID ulid.ULID `json:"organization_id"`
    // ... other fields
}

// internal/core/domain/project/repository.go
type Repository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, id ulid.ULID) (*Project, error)
    ListByOrganization(ctx context.Context, orgID ulid.ULID) ([]*Project, error)
}

// internal/core/domain/project/service.go
type Service interface {
    CreateProject(ctx context.Context, name string, orgID ulid.ULID) (*Project, error)
    GetProject(ctx context.Context, projectID ulid.ULID) (*Project, error)
}
```

#### 2. Implement Repository

Create PostgreSQL implementation in `internal/infrastructure/database/postgres/repository/`:
```go
// internal/infrastructure/database/postgres/repository/project_repo.go
type ProjectRepository struct {
    db *sql.DB
}

func NewProjectRepository(db *sql.DB) project.Repository {
    return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *project.Project) error {
    query := `INSERT INTO projects (id, name, organization_id, created_at) VALUES ($1, $2, $3, $4)`
    _, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.OrganizationID, time.Now())
    return err
}
```

#### 3. Implement Service

Create service implementation in `internal/core/services/project/`:
```go
// internal/core/services/project/project_service.go
type service struct {
    repo project.Repository
}

func NewService(repo project.Repository) project.Service {
    return &service{repo: repo}
}

func (s *service) CreateProject(ctx context.Context, name string, orgID ulid.ULID) (*project.Project, error) {
    // Validation
    if name == "" {
        return nil, errors.New("project name is required")
    }
    
    p := &project.Project{
        ID:             ulid.New(),
        Name:           name,
        OrganizationID: orgID,
    }
    
    return p, s.repo.Create(ctx, p)
}
```

#### 4. Add HTTP Handlers

Create HTTP handlers in `internal/transport/http/handlers/`:
```go
// internal/transport/http/handlers/project.go
type ProjectHandler struct {
    service project.Service
}

func NewProjectHandler(service project.Service) *ProjectHandler {
    return &ProjectHandler{service: service}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name           string    `json:"name"`
        OrganizationID ulid.ULID `json:"organization_id"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    project, err := h.service.CreateProject(r.Context(), req.Name, req.OrganizationID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(project)
}
```

#### 5. Wire Dependencies

Update the dependency injection container in `internal/app/container.go`:
```go
func (c *Container) initRepositories() error {
    // ... existing repositories
    c.ProjectRepo = repository.NewProjectRepository(c.DB)
    return nil
}

func (c *Container) initServices() error {
    // ... existing services
    c.ProjectService = project.NewService(c.ProjectRepo)
    return nil
}

func (c *Container) initHandlers() error {
    // ... existing handlers
    c.ProjectHandler = handlers.NewProjectHandler(c.ProjectService)
    return nil
}
```

#### 6. Add Routes

Update the router in `internal/transport/http/router.go`:
```go
func (r *Router) setupRoutes() {
    // ... existing routes
    r.mux.HandleFunc("/api/projects", r.container.ProjectHandler.CreateProject).Methods("POST")
    r.mux.HandleFunc("/api/projects/{id}", r.container.ProjectHandler.GetProject).Methods("GET")
}
```

### Database Operations

#### Running Migrations

```bash
# Run all pending migrations
make migrate-up

# Rollback one migration
make migrate-down

# Reset database (WARNING: destroys data)
make db-reset

# Check migration status
make migrate-status
```

#### Creating New Migrations

```bash
# Create PostgreSQL migration
make create-migration DB=postgres NAME=add_projects_table

# Create ClickHouse migration  
make create-migration DB=clickhouse NAME=add_metrics_index
```

This creates timestamped files:
- `migrations/postgres/20240101120000_add_projects_table.up.sql`
- `migrations/postgres/20240101120000_add_projects_table.down.sql`

Example migration files:
```sql
-- migrations/postgres/20240101120000_add_projects_table.up.sql
CREATE TABLE IF NOT EXISTS projects (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_projects_organization_id ON projects(organization_id);
CREATE INDEX idx_projects_created_at ON projects(created_at);
```

```sql
-- migrations/postgres/20240101120000_add_projects_table.down.sql
DROP TABLE IF EXISTS projects;
```

#### Seeding System Data

```bash
# Seed system data (permissions, roles, pricing)
make seed

# Or use CLI directly with options
go run cmd/migrate/main.go seed -verbose    # With detailed output
go run cmd/migrate/main.go seed -reset      # Reset and reseed
go run cmd/migrate/main.go seed-rbac        # RBAC only
go run cmd/migrate/main.go seed-pricing     # Pricing only
```

Seed files are in `seeds/`:
- `permissions.yaml` - 63 system permissions
- `roles.yaml` - 4 role templates
- `pricing.yaml` - 20 AI models, 78 prices

### Testing

#### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/core/services/user/...
```

Example unit test:
```go
// internal/core/services/user/user_service_test.go
func TestUserService_Register(t *testing.T) {
    // Setup
    mockRepo := &mocks.MockUserRepository{}
    service := NewService(mockRepo)
    
    // Mock expectations
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
    
    // Execute
    user, err := service.Register(context.Background(), "test@example.com", "Test User", "password")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
    mockRepo.AssertExpectations(t)
}
```

#### Integration Tests

```bash
# Run integration tests
make test-integration

# Run with test database
TEST_DATABASE_URL=postgres://test:test@localhost:5433/test_db make test-integration
```

#### Load Tests

```bash
# Run load tests
make load-test

# Run specific load test
make load-test TEST=api_endpoints
```

### Code Quality

#### Error Handling Standards

The platform follows industrial-grade error handling patterns. **Always reference these guides** when implementing features:

- **[ERROR_HANDLING_GUIDE.md](./development/ERROR_HANDLING_GUIDE.md)** - Complete implementation guide across Repository → Service → Handler layers
- **[DOMAIN_ALIAS_PATTERNS.md](./development/DOMAIN_ALIAS_PATTERNS.md)** - Professional import patterns for clean, conflict-free code  
- **[ERROR_HANDLING_QUICK_REFERENCE.md](./development/ERROR_HANDLING_QUICK_REFERENCE.md)** - Quick reference for daily development

These patterns ensure:
- Clean architecture with proper error propagation
- Professional domain separation with aliases
- Consistent error context and debugging capabilities
- Industrial Go standards across all layers

#### Linting

```bash
# Run Go linter
make lint-go

# Run frontend linter
make lint-frontend

# Run all linters
make lint
```

#### Code Formatting

```bash
# Format Go code
make fmt

# Format frontend code
make fmt-frontend
```

#### Security Scanning

```bash
# Run security scans
make security-scan

# Check for vulnerabilities
make vuln-check
```

## Frontend Development

### Next.js Structure

The frontend uses Next.js 14+ with App Router:

```
web/src/
├── app/                   # App Router pages
│   ├── (auth)/           # Auth route group
│   ├── (dashboard)/      # Dashboard route group
│   └── layout.tsx        # Root layout
├── components/           # React components
│   ├── ui/              # Base UI components (shadcn/ui)
│   ├── auth/            # Auth components
│   ├── analytics/       # Analytics components
│   └── layout/          # Layout components
├── hooks/               # Custom React hooks
├── lib/                 # Utilities and configurations
├── store/               # State management (Zustand)
└── types/               # TypeScript type definitions
```

### Development Server

```bash
# Start Next.js development server
cd web
npm run dev

# Or use Make command
make dev-frontend
```

The development server includes:
- **Hot reloading** - Instant updates on file changes
- **TypeScript checking** - Real-time type validation
- **API proxy** - Proxy `/api/*` requests to Go backend
- **Error overlay** - Enhanced error display

### API Integration

#### HTTP Client

```typescript
// web/src/lib/api-client.ts
class ApiClient {
  private baseURL: string
  
  constructor(baseURL: string) {
    this.baseURL = baseURL
  }
  
  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...this.getAuthHeaders(),
        ...options.headers,
      },
      ...options,
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    return response.json()
  }
  
  private getAuthHeaders() {
    const token = localStorage.getItem('auth_token')
    return token ? { Authorization: `Bearer ${token}` } : {}
  }
}

export const apiClient = new ApiClient(
  process.env.NODE_ENV === 'development' 
    ? 'http://localhost:8080/api'
    : '/api'
)
```

#### React Query Integration

```typescript
// web/src/hooks/useAnalytics.ts
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'

export function useMetrics(timeRange: string) {
  return useQuery({
    queryKey: ['metrics', timeRange],
    queryFn: () => apiClient.request(`/analytics/metrics?range=${timeRange}`),
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}
```

### Real-time Integration

#### WebSocket Hook

```typescript
// web/src/hooks/useWebSocket.ts
import { useEffect, useRef, useState } from 'react'

export function useWebSocket(url: string) {
  const ws = useRef<WebSocket | null>(null)
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected')
  const [lastMessage, setLastMessage] = useState<any>(null)
  
  useEffect(() => {
    ws.current = new WebSocket(url)
    setConnectionStatus('connecting')
    
    ws.current.onopen = () => setConnectionStatus('connected')
    ws.current.onclose = () => setConnectionStatus('disconnected')
    ws.current.onmessage = (event) => {
      const message = JSON.parse(event.data)
      setLastMessage(message)
    }
    
    return () => {
      ws.current?.close()
    }
  }, [url])
  
  const sendMessage = (message: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message))
    }
  }
  
  return { connectionStatus, lastMessage, sendMessage }
}
```

#### Real-time Components

```typescript
// web/src/components/analytics/RealtimeMetrics.tsx
import { useWebSocket } from '@/hooks/useWebSocket'

export function RealtimeMetrics() {
  const { lastMessage, connectionStatus } = useWebSocket('ws://localhost:8080/ws')
  const [metrics, setMetrics] = useState([])
  
  useEffect(() => {
    if (lastMessage?.type === 'metrics.updated') {
      setMetrics(lastMessage.data)
    }
  }, [lastMessage])
  
  return (
    <div>
      <div className="flex items-center gap-2">
        <div className={`h-2 w-2 rounded-full ${
          connectionStatus === 'connected' ? 'bg-green-500' : 'bg-red-500'
        }`} />
        <span>Real-time Updates: {connectionStatus}</span>
      </div>
      
      {metrics.map((metric) => (
        <MetricCard key={metric.id} metric={metric} />
      ))}
    </div>
  )
}
```

### State Management

Using Zustand for global state:

```typescript
// web/src/store/authStore.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  login: (token: string, user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      
      login: (token, user) => set({ 
        token, 
        user, 
        isAuthenticated: true 
      }),
      
      logout: () => set({ 
        token: null, 
        user: null, 
        isAuthenticated: false 
      }),
    }),
    {
      name: 'auth-storage',
    }
  )
)
```

### Building for Production

```bash
# Build frontend for production
make build-frontend

# Build optimized production bundle
cd web
npm run build
```

The production build includes:
- **Static optimization** - Pre-rendered pages
- **Bundle optimization** - Tree shaking and minification
- **Image optimization** - Next.js image optimization
- **Performance budgets** - Bundle size monitoring

## Debugging

### Backend Debugging

#### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start server with debugger
dlv debug ./cmd/server/main.go

# Set breakpoints and debug
(dlv) b main.main
(dlv) c
```

#### Debugging with VS Code

Create `.vscode/launch.json`:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "./cmd/server/main.go",
            "env": {
                "ENV": "development"
            }
        }
    ]
}
```

#### Logging and Observability

```go
// Use structured logging
import "github.com/sirupsen/logrus"

logger := logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "request_id": requestID,
})
logger.Info("Processing user request")
```

### Frontend Debugging

#### Browser DevTools

- **React DevTools** - Component inspection and profiling
- **Network tab** - API request monitoring
- **Console** - JavaScript debugging and logging
- **Performance tab** - Performance profiling

#### Next.js Debugging

```bash
# Enable debug mode
DEBUG=* npm run dev

# Debug specific modules
DEBUG=next:* npm run dev
```

## Performance Optimization

### Backend Performance

#### Database Query Optimization

```go
// Use proper indexing
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_projects_org_id ON projects(organization_id);

// Use prepared statements
stmt, err := db.Prepare("SELECT * FROM users WHERE email = $1")
defer stmt.Close()
```

#### Caching Strategy

```go
// Redis caching
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
    // Check cache first
    if cached, err := s.cache.Get(ctx, "user:"+id); err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // Fallback to database
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    userJSON, _ := json.Marshal(user)
    s.cache.Set(ctx, "user:"+id, string(userJSON), time.Hour)
    
    return user, nil
}
```

### Frontend Performance

#### Code Splitting

```typescript
// Dynamic imports for code splitting
const AnalyticsDashboard = dynamic(
  () => import('@/components/analytics/AnalyticsDashboard'),
  { loading: () => <p>Loading...</p> }
)
```

#### Image Optimization

```typescript
// Next.js Image component
import Image from 'next/image'

<Image
  src="/logo.png"
  alt="Brokle Logo"
  width={200}
  height={100}
  priority // Load above-the-fold images first
/>
```

## Troubleshooting

### Common Issues

#### Database Connection Issues

```bash
# Check database status
make db-status

# Reset databases
make db-reset

# Check logs
make logs-db
```

#### Frontend Build Issues

```bash
# Clear Next.js cache
rm -rf web/.next/

# Reinstall dependencies
cd web
rm -rf node_modules/
npm install
```

#### Port Conflicts

```bash
# Check what's running on ports
lsof -ti:8080
lsof -ti:3000

# Kill processes
kill -9 $(lsof -ti:8080)
```

### Getting Help

- **Documentation**: Check all docs in `/docs/`
- **Issues**: Create GitHub issues for bugs
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Join our Discord server

---

This development guide should get you up and running with the Brokle platform. For additional help, refer to the other documentation files or reach out to the community.