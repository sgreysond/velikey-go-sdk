# Contributing to VeliKey Go SDK

We welcome contributions to the VeliKey Go SDK! This document provides guidelines for contributing to the Go client library for VeliKey Aegis.

## Development Environment

### Prerequisites

- **Go**: 1.19+ (latest stable recommended)
- **Git**: For version control
- **Make**: For build automation (optional)
- **VeliKey Aegis**: Access to control plane for testing

### Setup

```bash
# Clone the repository
git clone https://github.com/sgreysond/velikey-go-sdk.git
cd velikey-go-sdk

# Initialize Go module (if needed)
go mod tidy

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Build examples
go build -o bin/example examples/basic/main.go
```

## Development Workflow [[memory:7696176]]

1. **Create a feature branch** from `main`
2. **Make your changes** with appropriate tests
3. **Run quality checks**:
   ```bash
   go fmt ./...
   go vet ./...
   go test -race ./...
   golangci-lint run
   ```
4. **Commit with conventional messages** (feat, fix, chore, docs, refactor)
5. **Open a Pull Request** with clear description
6. **Address review feedback** promptly

## Code Quality Standards

### Testing Requirements [[memory:7696167]]

All contributions must include comprehensive testing:

- **Unit Tests**: Individual function and type testing
- **Integration Tests**: API integration testing
- **End-to-End Tests**: Complete workflow testing
- **Benchmark Tests**: Performance-critical code

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific packages
go test ./pkg/client
go test ./pkg/resources

# Run integration tests (requires build tag)
go test -tags=integration ./...

# Run benchmarks
go test -bench=. ./...

# Race condition detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Code Style

We follow standard Go conventions:

```bash
# Format code
go fmt ./...

# Lint code
go vet ./...

# Advanced linting (requires golangci-lint)
golangci-lint run

# Check for common issues
staticcheck ./...

# Imports formatting (requires goimports)
goimports -w .

# Full quality check
make quality-check  # if Makefile exists
```

### Quality Tools Configuration

```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters-settings:
  goimports:
    local-prefixes: github.com/sgreysond/velikey-go-sdk
  govet:
    check-shadowing: true
  gocyclo:
    min-complexity: 10
  maligned:
    suggest-new: true
  goconst:
    min-len: 3
    min-occurrences: 3

linters:
  enable:
    - goimports
    - gosec
    - govet
    - ineffassign
    - misspell
    - gocyclo
    - goconst
    - gofmt
    - gosimple
    - staticcheck
    - unconvert
    - unused
  disable:
    - errcheck  # We handle this manually for better context
```

## Project Structure

```
.
├── client.go                # Main client
├── types.go                 # Common types  
├── errors.go                # Error types
├── config.go                # Configuration
├── pkg/
│   ├── resources/          # API resources
│   │   ├── agents.go      # Agent operations
│   │   ├── policies.go    # Policy operations
│   │   ├── monitoring.go  # Monitoring operations
│   │   └── compliance.go  # Compliance operations
│   ├── builders/          # Builder patterns
│   │   └── policy.go     # Policy builder
│   └── internal/          # Internal packages
│       ├── http/         # HTTP client
│       └── auth/         # Authentication
├── examples/              # Usage examples
│   ├── basic/
│   ├── kubernetes-operator/
│   └── terraform-provider/
├── test/                 # Test utilities
│   ├── fixtures/        # Test fixtures
│   └── integration/     # Integration tests
└── docs/                # Package documentation
```

## SDK Architecture

### Main Client

```go
package velikey

import (
    "context"
    "net/http"
    "time"
)

// Config holds client configuration
type Config struct {
    APIKey     string
    BaseURL    string
    Timeout    time.Duration
    RetryCount int
    HTTPClient *http.Client
}

// Client is the main VeliKey SDK client
type Client struct {
    config Config
    http   HTTPClient
    
    // Resources
    Agents      *AgentsResource
    Policies    *PoliciesResource
    Monitoring  *MonitoringResource
    Compliance  *ComplianceResource
    Diagnostics *DiagnosticsResource
}

// NewClient creates a new VeliKey client
func NewClient(config Config) *Client {
    if config.BaseURL == "" {
        config.BaseURL = "https://api.velikey.com"
    }
    if config.Timeout == 0 {
        config.Timeout = 30 * time.Second
    }
    if config.HTTPClient == nil {
        config.HTTPClient = &http.Client{Timeout: config.Timeout}
    }

    httpClient := NewHTTPClient(config)

    client := &Client{
        config: config,
        http:   httpClient,
    }

    // Initialize resources
    client.Agents = NewAgentsResource(httpClient)
    client.Policies = NewPoliciesResource(httpClient)
    client.Monitoring = NewMonitoringResource(httpClient)
    client.Compliance = NewComplianceResource(httpClient)
    client.Diagnostics = NewDiagnosticsResource(httpClient)

    return client
}

// QuickSetup provides high-level setup for new customers
func (c *Client) QuickSetup(ctx context.Context, opts SetupOptions) (*SetupResult, error) {
    builder := c.NewPolicyBuilder().
        Name(fmt.Sprintf("%s Policy", opts.ComplianceFramework)).
        ComplianceStandard(opts.ComplianceFramework).
        EnforcementMode(opts.EnforcementMode)
    
    if opts.PostQuantum {
        builder = builder.PostQuantumReady()
    }
    
    policy, err := builder.Create(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create policy: %w", err)
    }
    
    return &SetupResult{
        PolicyName: policy.Name,
        PolicyID:   policy.ID,
        NextSteps:  c.generateNextSteps(policy),
    }, nil
}

// GetSecurityStatus retrieves overall security status
func (c *Client) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
    var status SecurityStatus
    err := c.http.Get(ctx, "/security/status", &status)
    return &status, err
}
```

### Resource Pattern

```go
// BaseResource provides common functionality for all resources
type BaseResource struct {
    http HTTPClient
}

// AgentsResource handles agent-related operations
type AgentsResource struct {
    BaseResource
}

// NewAgentsResource creates a new agents resource
func NewAgentsResource(httpClient HTTPClient) *AgentsResource {
    return &AgentsResource{
        BaseResource: BaseResource{http: httpClient},
    }
}

// List retrieves all agents
func (r *AgentsResource) List(ctx context.Context, opts *ListOptions) ([]*Agent, error) {
    path := "/agents"
    if opts != nil {
        path = r.buildQueryPath(path, opts)
    }
    
    var response struct {
        Agents []*Agent `json:"agents"`
    }
    
    err := r.http.Get(ctx, path, &response)
    return response.Agents, err
}

// Get retrieves a specific agent
func (r *AgentsResource) Get(ctx context.Context, agentID string) (*Agent, error) {
    if agentID == "" {
        return nil, ErrInvalidAgentID
    }
    
    var agent Agent
    err := r.http.Get(ctx, fmt.Sprintf("/agents/%s", agentID), &agent)
    return &agent, err
}

// Create registers a new agent
func (r *AgentsResource) Create(ctx context.Context, req *CreateAgentRequest) (*Agent, error) {
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    var agent Agent
    err := r.http.Post(ctx, "/agents", req, &agent)
    return &agent, err
}

// Update modifies an existing agent
func (r *AgentsResource) Update(ctx context.Context, agentID string, req *UpdateAgentRequest) (*Agent, error) {
    if agentID == "" {
        return nil, ErrInvalidAgentID
    }
    
    var agent Agent
    err := r.http.Put(ctx, fmt.Sprintf("/agents/%s", agentID), req, &agent)
    return &agent, err
}

// Delete removes an agent
func (r *AgentsResource) Delete(ctx context.Context, agentID string) error {
    if agentID == "" {
        return ErrInvalidAgentID
    }
    
    return r.http.Delete(ctx, fmt.Sprintf("/agents/%s", agentID), nil)
}

// buildQueryPath constructs URL with query parameters
func (r *AgentsResource) buildQueryPath(basePath string, opts *ListOptions) string {
    params := url.Values{}
    
    if opts.Limit > 0 {
        params.Add("limit", strconv.Itoa(opts.Limit))
    }
    if opts.Offset > 0 {
        params.Add("offset", strconv.Itoa(opts.Offset))
    }
    if opts.Filter != "" {
        params.Add("filter", opts.Filter)
    }
    
    if len(params) > 0 {
        return fmt.Sprintf("%s?%s", basePath, params.Encode())
    }
    return basePath
}
```

### Type Definitions

```go
// Core domain types
type Agent struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Status      AgentStatus       `json:"status"`
    Version     string            `json:"version"`
    LastSeen    time.Time         `json:"last_seen"`
    HealthScore int               `json:"health_score"`
    Location    *AgentLocation    `json:"location,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type AgentStatus string

const (
    AgentStatusOnline  AgentStatus = "online"
    AgentStatusOffline AgentStatus = "offline"
    AgentStatusError   AgentStatus = "error"
)

type Policy struct {
    ID                  string                 `json:"id"`
    Name                string                 `json:"name"`
    ComplianceFramework string                 `json:"compliance_framework,omitempty"`
    EnforcementMode     EnforcementMode        `json:"enforcement_mode"`
    Rules               map[string]interface{} `json:"rules"`
    Status              PolicyStatus           `json:"status"`
    CreatedAt           time.Time              `json:"created_at"`
    UpdatedAt           time.Time              `json:"updated_at"`
}

type EnforcementMode string

const (
    EnforcementModeObserve EnforcementMode = "observe"
    EnforcementModeEnforce EnforcementMode = "enforce"
)

// Request/Response types
type CreateAgentRequest struct {
    Name     string            `json:"name" validate:"required"`
    Location *AgentLocation    `json:"location,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

func (r *CreateAgentRequest) Validate() error {
    if r.Name == "" {
        return ErrMissingAgentName
    }
    if len(r.Name) > 100 {
        return ErrAgentNameTooLong
    }
    return nil
}

type SecurityStatus struct {
    HealthScore       int     `json:"health_score"`
    AgentsOnline      string  `json:"agents_online"`
    CriticalAlerts    int     `json:"critical_alerts"`
    PolicyCompliance  float64 `json:"policy_compliance"`
    LastUpdated       time.Time `json:"last_updated"`
}

type SetupOptions struct {
    ComplianceFramework string          `json:"compliance_framework"`
    EnforcementMode     EnforcementMode `json:"enforcement_mode"`
    PostQuantum         bool            `json:"post_quantum"`
}

type SetupResult struct {
    PolicyName string   `json:"policy_name"`
    PolicyID   string   `json:"policy_id"`
    NextSteps  []string `json:"next_steps"`
}
```

### Error Handling

```go
// Error types with proper wrapping and context
type VeliKeyError struct {
    Message    string
    Code       string
    StatusCode int
    Response   map[string]interface{}
    Err        error
}

func (e *VeliKeyError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
    }
    return e.Message
}

func (e *VeliKeyError) Unwrap() error {
    return e.Err
}

// Predefined errors
var (
    ErrInvalidAPIKey      = &VeliKeyError{Message: "invalid API key", Code: "invalid_api_key"}
    ErrInvalidAgentID     = &VeliKeyError{Message: "invalid agent ID", Code: "invalid_agent_id"}
    ErrMissingAgentName   = &VeliKeyError{Message: "agent name is required", Code: "missing_agent_name"}
    ErrAgentNameTooLong   = &VeliKeyError{Message: "agent name too long (max 100 characters)", Code: "agent_name_too_long"}
    ErrRateLimitExceeded  = &VeliKeyError{Message: "rate limit exceeded", Code: "rate_limit_exceeded"}
)

// Error constructors
func NewAuthenticationError(message string) error {
    return &VeliKeyError{
        Message:    message,
        Code:       "authentication_error",
        StatusCode: 401,
    }
}

func NewValidationError(field, message string) error {
    return &VeliKeyError{
        Message: fmt.Sprintf("validation failed for field '%s': %s", field, message),
        Code:    "validation_error",
        StatusCode: 400,
        Response: map[string]interface{}{
            "field": field,
            "message": message,
        },
    }
}

// HTTP error handling
func handleHTTPError(resp *http.Response) error {
    var apiError struct {
        Error struct {
            Code    string                 `json:"code"`
            Message string                 `json:"message"`
            Details map[string]interface{} `json:"details,omitempty"`
        } `json:"error"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
        return &VeliKeyError{
            Message:    "failed to parse error response",
            StatusCode: resp.StatusCode,
        }
    }

    return &VeliKeyError{
        Message:    apiError.Error.Message,
        Code:       apiError.Error.Code,
        StatusCode: resp.StatusCode,
        Response:   apiError.Error.Details,
    }
}
```

### Builder Pattern

```go
// PolicyBuilder provides a fluent interface for building policies
type PolicyBuilder struct {
    http   HTTPClient
    policy policyConfig
}

type policyConfig struct {
    Name                string                 `json:"name"`
    ComplianceFramework string                 `json:"compliance_framework"`
    EnforcementMode     EnforcementMode        `json:"enforcement_mode"`
    Rules               map[string]interface{} `json:"rules"`
    PostQuantum         bool                   `json:"post_quantum"`
}

// NewPolicyBuilder creates a new policy builder
func (c *Client) NewPolicyBuilder() *PolicyBuilder {
    return &PolicyBuilder{
        http: c.http,
        policy: policyConfig{
            Rules: make(map[string]interface{}),
        },
    }
}

// Name sets the policy name
func (b *PolicyBuilder) Name(name string) *PolicyBuilder {
    b.policy.Name = name
    return b
}

// ComplianceStandard sets the compliance framework
func (b *PolicyBuilder) ComplianceStandard(framework string) *PolicyBuilder {
    b.policy.ComplianceFramework = framework
    return b
}

// EnforcementMode sets the enforcement mode
func (b *PolicyBuilder) EnforcementMode(mode EnforcementMode) *PolicyBuilder {
    b.policy.EnforcementMode = mode
    return b
}

// PostQuantumReady enables post-quantum cryptography
func (b *PolicyBuilder) PostQuantumReady() *PolicyBuilder {
    b.policy.PostQuantum = true
    // Configure post-quantum specific rules
    if b.policy.Rules["tls"] == nil {
        b.policy.Rules["tls"] = make(map[string]interface{})
    }
    tlsRules := b.policy.Rules["tls"].(map[string]interface{})
    tlsRules["pq_ready"] = []string{
        "TLS_KYBER768_P256_SHA256",
        "TLS_KYBER1024_P384_SHA384",
    }
    return b
}

// AegisConfig adds Aegis-specific configuration
func (b *PolicyBuilder) AegisConfig(config AegisConfig) *PolicyBuilder {
    b.policy.Rules["aegis"] = config
    return b
}

// Build returns the policy configuration
func (b *PolicyBuilder) Build() policyConfig {
    return b.policy
}

// Create builds and creates the policy
func (b *PolicyBuilder) Create(ctx context.Context) (*Policy, error) {
    if b.policy.Name == "" {
        return nil, ErrMissingPolicyName
    }

    var policy Policy
    err := b.http.Post(ctx, "/policies", b.policy, &policy)
    return &policy, err
}

// Validate checks the policy configuration
func (b *PolicyBuilder) Validate() error {
    if b.policy.Name == "" {
        return ErrMissingPolicyName
    }
    
    if b.policy.EnforcementMode == "" {
        b.policy.EnforcementMode = EnforcementModeObserve
    }
    
    return nil
}
```

## Testing Guidelines

### Unit Testing

```go
// pkg/resources/agents_test.go
package resources

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
    mock.Mock
}

func (m *MockHTTPClient) Get(ctx context.Context, path string, result interface{}) error {
    args := m.Called(ctx, path, result)
    return args.Error(0)
}

func (m *MockHTTPClient) Post(ctx context.Context, path string, body, result interface{}) error {
    args := m.Called(ctx, path, body, result)
    return args.Error(0)
}

func TestAgentsResource_List(t *testing.T) {
    mockHTTP := &MockHTTPClient{}
    resource := NewAgentsResource(mockHTTP)

    expectedAgents := []*Agent{
        {
            ID:       "agent_1",
            Name:     "Test Agent",
            Status:   AgentStatusOnline,
            LastSeen: time.Now(),
        },
    }

    mockHTTP.On("Get", mock.Anything, "/agents", mock.AnythingOfType("*struct { Agents []*Agent }")).
        Run(func(args mock.Arguments) {
            result := args.Get(2).(*struct{ Agents []*Agent })
            result.Agents = expectedAgents
        }).
        Return(nil)

    agents, err := resource.List(context.Background(), nil)

    require.NoError(t, err)
    assert.Len(t, agents, 1)
    assert.Equal(t, "agent_1", agents[0].ID)
    assert.Equal(t, "Test Agent", agents[0].Name)
    mockHTTP.AssertExpectations(t)
}

func TestAgentsResource_Get(t *testing.T) {
    mockHTTP := &MockHTTPClient{}
    resource := NewAgentsResource(mockHTTP)

    expectedAgent := &Agent{
        ID:     "agent_1",
        Name:   "Test Agent",
        Status: AgentStatusOnline,
    }

    mockHTTP.On("Get", mock.Anything, "/agents/agent_1", mock.AnythingOfType("*Agent")).
        Run(func(args mock.Arguments) {
            result := args.Get(2).(*Agent)
            *result = *expectedAgent
        }).
        Return(nil)

    agent, err := resource.Get(context.Background(), "agent_1")

    require.NoError(t, err)
    assert.Equal(t, expectedAgent.ID, agent.ID)
    assert.Equal(t, expectedAgent.Name, agent.Name)
    mockHTTP.AssertExpectations(t)
}

func TestAgentsResource_Get_InvalidID(t *testing.T) {
    mockHTTP := &MockHTTPClient{}
    resource := NewAgentsResource(mockHTTP)

    _, err := resource.Get(context.Background(), "")

    assert.Error(t, err)
    assert.Equal(t, ErrInvalidAgentID, err)
    mockHTTP.AssertNotCalled(t, "Get")
}

// Benchmark test
func BenchmarkAgentsResource_List(b *testing.B) {
    mockHTTP := &MockHTTPClient{}
    resource := NewAgentsResource(mockHTTP)

    agents := make([]*Agent, 100)
    for i := 0; i < 100; i++ {
        agents[i] = &Agent{
            ID:     fmt.Sprintf("agent_%d", i),
            Name:   fmt.Sprintf("Agent %d", i),
            Status: AgentStatusOnline,
        }
    }

    mockHTTP.On("Get", mock.Anything, "/agents", mock.AnythingOfType("*struct { Agents []*Agent }")).
        Run(func(args mock.Arguments) {
            result := args.Get(2).(*struct{ Agents []*Agent })
            result.Agents = agents
        }).
        Return(nil)

    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := resource.List(ctx, nil)
        require.NoError(b, err)
    }
}
```

### Integration Testing

```go
// test/integration/client_test.go
//go:build integration
// +build integration

package integration

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/sgreysond/velikey-go-sdk"
)

var testClient *velikey.Client

func TestMain(m *testing.M) {
    apiKey := os.Getenv("VELIKEY_TEST_API_KEY")
    if apiKey == "" {
        fmt.Println("Skipping integration tests - VELIKEY_TEST_API_KEY not set")
        os.Exit(0)
    }

    testClient = velikey.NewClient(velikey.Config{
        APIKey: apiKey,
        Timeout: 30 * time.Second,
    })

    code := m.Run()
    os.Exit(code)
}

func TestClient_GetSecurityStatus(t *testing.T) {
    ctx := context.Background()

    status, err := testClient.GetSecurityStatus(ctx)
    require.NoError(t, err)
    require.NotNil(t, status)

    assert.GreaterOrEqual(t, status.HealthScore, 0)
    assert.LessOrEqual(t, status.HealthScore, 100)
    assert.Regexp(t, `\d+/\d+`, status.AgentsOnline)
    assert.GreaterOrEqual(t, status.CriticalAlerts, 0)
}

func TestClient_PolicyCRUD(t *testing.T) {
    ctx := context.Background()

    // Create policy
    builder := testClient.NewPolicyBuilder().
        Name("Integration Test Policy").
        ComplianceStandard("soc2").
        EnforcementMode(velikey.EnforcementModeObserve)

    policy, err := builder.Create(ctx)
    require.NoError(t, err)
    require.NotNil(t, policy)
    assert.Equal(t, "Integration Test Policy", policy.Name)
    assert.Equal(t, "soc2", policy.ComplianceFramework)

    defer func() {
        // Cleanup
        err := testClient.Policies.Delete(ctx, policy.ID)
        if err != nil {
            t.Logf("Failed to cleanup policy %s: %v", policy.ID, err)
        }
    }()

    // Get policy
    retrieved, err := testClient.Policies.Get(ctx, policy.ID)
    require.NoError(t, err)
    assert.Equal(t, policy.ID, retrieved.ID)
    assert.Equal(t, policy.Name, retrieved.Name)

    // Update policy
    updated, err := testClient.Policies.Update(ctx, policy.ID, &velikey.UpdatePolicyRequest{
        EnforcementMode: velikey.EnforcementModeEnforce,
    })
    require.NoError(t, err)
    assert.Equal(t, velikey.EnforcementModeEnforce, updated.EnforcementMode)

    // List policies (should include our test policy)
    policies, err := testClient.Policies.List(ctx, nil)
    require.NoError(t, err)
    assert.True(t, len(policies) > 0)

    found := false
    for _, p := range policies {
        if p.ID == policy.ID {
            found = true
            break
        }
    }
    assert.True(t, found, "Created policy should be in list")
}
```

### Example Tests

```go
// examples/basic/main_test.go
package main

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/sgreysond/velikey-go-sdk"
)

func TestBasicExample(t *testing.T) {
    // This tests that our basic example compiles and runs without panicking
    client := velikey.NewClient(velikey.Config{
        APIKey:  "test-key",
        BaseURL: "https://api-test.velikey.com",
        Timeout: 5 * time.Second,
    })

    assert.NotNil(t, client)
    assert.NotNil(t, client.Agents)
    assert.NotNil(t, client.Policies)

    // Test that we can create a policy builder
    builder := client.NewPolicyBuilder()
    assert.NotNil(t, builder)

    config := builder.
        Name("Test Policy").
        ComplianceStandard("soc2").
        EnforcementMode(velikey.EnforcementModeObserve).
        Build()

    assert.Equal(t, "Test Policy", config.Name)
    assert.Equal(t, "soc2", config.ComplianceFramework)
    assert.Equal(t, velikey.EnforcementModeObserve, config.EnforcementMode)
}
```

## Documentation

### Go Documentation Standards

```go
// Package wielkeygo provides a Go client for the VeliKey Aegis API.
//
// The VeliKey Aegis platform provides quantum-safe TLS policy management
// and enforcement capabilities. This SDK allows Go applications to
// interact with the Aegis control plane to manage security policies,
// monitor agent status, and ensure compliance.
//
// Basic Usage:
//
//  client := velikey.NewClient(velikey.Config{
//      APIKey: "your-api-key",
//  })
//  
//  status, err := client.GetSecurityStatus(context.Background())
//  if err != nil {
//      log.Fatal(err)
//  }
//  
//  fmt.Printf("Health Score: %d/100\n", status.HealthScore)
//
// The SDK provides several resource types for different operations:
//
// - Agents: Manage and monitor VeliKey agents
// - Policies: Create and deploy security policies  
// - Monitoring: Access metrics and alerts
// - Compliance: Validate compliance frameworks
// - Diagnostics: System health and troubleshooting
//
// All operations are context-aware and support cancellation and timeouts.
package velikey

// NewClient creates a new VeliKey API client.
//
// The client requires an API key for authentication. Optional configuration
// includes custom base URL, HTTP client, timeout, and retry settings.
//
// Example:
//  client := NewClient(Config{
//      APIKey:     "sk_prod_abc123...",
//      Timeout:    30 * time.Second,
//      RetryCount: 3,
//  })
//
// The client is safe for concurrent use across multiple goroutines.
func NewClient(config Config) *Client {
    // Implementation...
}

// GetSecurityStatus retrieves the overall security posture and health metrics.
//
// This provides a high-level view of the current security state including:
// - Overall health score (0-100)
// - Agent status summary
// - Critical alert count  
// - Policy compliance percentage
//
// Example:
//  status, err := client.GetSecurityStatus(ctx)
//  if err != nil {
//      return fmt.Errorf("failed to get status: %w", err)
//  }
//  
//  if status.CriticalAlerts > 0 {
//      log.Printf("WARNING: %d critical alerts active", status.CriticalAlerts)
//  }
func (c *Client) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
    // Implementation...
}
```

## Performance Considerations

### Concurrency and Context

```go
// Example of proper context usage and concurrency
func (c *Client) BulkAgentStatus(ctx context.Context, agentIDs []string) (map[string]*Agent, error) {
    if len(agentIDs) == 0 {
        return make(map[string]*Agent), nil
    }

    // Use semaphore to limit concurrent requests
    sem := make(chan struct{}, 10) // Max 10 concurrent requests
    results := make(chan agentResult, len(agentIDs))
    
    // Launch goroutines for concurrent fetching
    for _, id := range agentIDs {
        go func(agentID string) {
            sem <- struct{}{} // Acquire semaphore
            defer func() { <-sem }() // Release semaphore
            
            agent, err := c.Agents.Get(ctx, agentID)
            results <- agentResult{id: agentID, agent: agent, err: err}
        }(id)
    }
    
    // Collect results
    agentMap := make(map[string]*Agent)
    var errs []error
    
    for i := 0; i < len(agentIDs); i++ {
        select {
        case result := <-results:
            if result.err != nil {
                errs = append(errs, fmt.Errorf("agent %s: %w", result.id, result.err))
            } else {
                agentMap[result.id] = result.agent
            }
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    if len(errs) > 0 {
        return agentMap, fmt.Errorf("partial failure: %v", errs)
    }
    
    return agentMap, nil
}

type agentResult struct {
    id    string
    agent *Agent
    err   error
}
```

### Memory Management

```go
// HTTP client with proper connection pooling
func NewHTTPClient(config Config) HTTPClient {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    
    client := &http.Client{
        Transport: transport,
        Timeout:   config.Timeout,
    }
    
    return &httpClient{
        client: client,
        config: config,
    }
}

// Streaming interface for large datasets
func (r *AgentsResource) Stream(ctx context.Context, callback func(*Agent) error) error {
    offset := 0
    limit := 100 // Process in batches
    
    for {
        agents, err := r.List(ctx, &ListOptions{
            Offset: offset,
            Limit:  limit,
        })
        if err != nil {
            return fmt.Errorf("failed to fetch agents: %w", err)
        }
        
        if len(agents) == 0 {
            break // No more agents
        }
        
        for _, agent := range agents {
            if err := callback(agent); err != nil {
                return fmt.Errorf("callback failed: %w", err)
            }
        }
        
        offset += len(agents)
        
        // Check if we got fewer results than requested (end of data)
        if len(agents) < limit {
            break
        }
    }
    
    return nil
}
```

## Git Commit Guidelines

Use conventional commit messages:

```
feat(client): add support for bulk operations
fix(resources): handle pagination edge cases
docs(readme): add Kubernetes operator example
test(integration): add comprehensive policy tests
refactor(http): improve error handling and retries
perf(client): optimize concurrent request handling
```

## Release Process

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with new features and fixes
3. **Tag release**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. **Create GitHub release** with detailed notes
5. **Verify Go module proxy** picks up the release

## Getting Help

- **Go Package Documentation**: [pkg.go.dev/github.com/sgreysond/velikey-go-sdk](https://pkg.go.dev/github.com/sgreysond/velikey-go-sdk)
- **GitHub Discussions**: For questions and ideas
- **GitHub Issues**: For bug reports and feature requests
- **Community Forum**: [community.velikey.com](https://community.velikey.com)
- **Email**: [go-sdk@velikey.com](mailto:go-sdk@velikey.com)

## License

By contributing to VeliKey Go SDK, you agree that your contributions will be licensed under the MIT License.
