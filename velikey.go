// Package velikey provides a Go SDK for VeliKey Aegis quantum-safe crypto policy management.
//
// Example usage:
//
//	client := velikey.NewClient(velikey.Config{
//		APIKey: os.Getenv("VELIKEY_API_KEY"),
//	})
//
//	// Quick setup
//	setup, err := client.QuickSetup(context.Background(), velikey.SetupOptions{
//		ComplianceFramework: "soc2",
//		EnforcementMode:     "observe",
//		PostQuantum:         true,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Monitor agents
//	agents, err := client.Agents.List(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Found %d agents\n", len(agents))
package velikey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// Client is the main VeliKey API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
	limiter    *rate.Limiter

	// Resource managers
	Agents      *AgentsService
	Policies    *PoliciesService
	Monitoring  *MonitoringService
	Compliance  *ComplianceService
	Diagnostics *DiagnosticsService
	Billing     *BillingService
}

// Config holds client configuration
type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	Timeout    time.Duration
	RateLimit  rate.Limit // requests per second
}

// NewClient creates a new VeliKey client
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.velikey.com"
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
		if config.Timeout == 0 {
			config.HTTPClient.Timeout = 30 * time.Second
		}
	}
	if config.UserAgent == "" {
		config.UserAgent = "velikey-go-sdk/0.1.0"
	}
	if config.RateLimit == 0 {
		config.RateLimit = 10 // 10 requests per second default
	}

	client := &Client{
		httpClient: config.HTTPClient,
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		userAgent:  config.UserAgent,
		limiter:    rate.NewLimiter(config.RateLimit, 1),
	}

	// Initialize service managers
	client.Agents = &AgentsService{client: client}
	client.Policies = &PoliciesService{client: client}
	client.Monitoring = &MonitoringService{client: client}
	client.Compliance = &ComplianceService{client: client}
	client.Diagnostics = &DiagnosticsService{client: client}
	client.Billing = &BillingService{client: client}

	return client
}

// request makes an authenticated HTTP request
func (c *Client) request(ctx context.Context, method, endpoint string, params map[string]string, body interface{}) (*http.Response, error) {
	// Rate limiting
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Build URL
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		
		switch resp.StatusCode {
		case 401:
			return nil, &AuthenticationError{Message: "Invalid API key or expired token"}
		case 400:
			return nil, &ValidationError{Message: string(bodyBytes)}
		case 404:
			return nil, &NotFoundError{Message: "Resource not found"}
		case 429:
			return nil, &RateLimitError{Message: "Rate limit exceeded"}
		default:
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(bodyBytes),
			}
		}
	}

	return resp, nil
}

// QuickSetup performs initial setup for new customers
func (c *Client) QuickSetup(ctx context.Context, options SetupOptions) (*SetupResult, error) {
	resp, err := c.request(ctx, "POST", "/api/setup/quick", nil, options)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SetupResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSecurityStatus returns comprehensive security status
func (c *Client) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
	resp, err := c.request(ctx, "GET", "/api/security/status", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status SecurityStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// GetHealth checks API health
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.request(ctx, "GET", "/health", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &health, nil
}

// ValidateCompliance checks compliance against a framework
func (c *Client) ValidateCompliance(ctx context.Context, framework string) (*ComplianceValidation, error) {
	body := map[string]string{"framework": framework}
	resp, err := c.request(ctx, "POST", "/api/compliance/validate", nil, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var validation ComplianceValidation
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &validation, nil
}

// GetOptimizationSuggestions returns performance and cost optimization suggestions
func (c *Client) GetOptimizationSuggestions(ctx context.Context) (*OptimizationSuggestions, error) {
	resp, err := c.request(ctx, "GET", "/api/optimization/suggestions", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var suggestions OptimizationSuggestions
	if err := json.NewDecoder(resp.Body).Decode(&suggestions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &suggestions, nil
}

// BulkPolicyUpdate updates multiple policies atomically
func (c *Client) BulkPolicyUpdate(ctx context.Context, updates []PolicyUpdate) (*BulkUpdateResult, error) {
	body := map[string]interface{}{"updates": updates}
	resp, err := c.request(ctx, "POST", "/api/policies/bulk-update", nil, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BulkUpdateResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// PolicyBuilder provides a fluent interface for building policies
type PolicyBuilder struct {
	client          *Client
	rules           map[string]interface{}
	enforcementMode string
	name            string
	description     string
}

// NewPolicyBuilder creates a new policy builder
func (c *Client) NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		client: c,
		rules: map[string]interface{}{
			"compliance_standard": "Custom",
			"aegis":               map[string]interface{}{},
			"somnus":              map[string]interface{}{},
			"logos":               map[string]interface{}{},
		},
		enforcementMode: "observe",
	}
}

// ComplianceStandard sets the compliance standard
func (pb *PolicyBuilder) ComplianceStandard(standard string) *PolicyBuilder {
	pb.rules["compliance_standard"] = standard
	return pb
}

// AegisConfig sets Aegis TLS configuration
func (pb *PolicyBuilder) AegisConfig(config map[string]interface{}) *PolicyBuilder {
	aegis := pb.rules["aegis"].(map[string]interface{})
	for k, v := range config {
		aegis[k] = v
	}
	return pb
}

// EnforcementMode sets the policy enforcement mode
func (pb *PolicyBuilder) EnforcementMode(mode string) *PolicyBuilder {
	pb.enforcementMode = mode
	return pb
}

// PostQuantumReady enables post-quantum cryptography
func (pb *PolicyBuilder) PostQuantumReady() *PolicyBuilder {
	aegis := pb.rules["aegis"].(map[string]interface{})
	somnus := pb.rules["somnus"].(map[string]interface{})
	logos := pb.rules["logos"].(map[string]interface{})

	aegis["pq_ready"] = []string{"TLS_KYBER768_P256_SHA256"}
	somnus["pq_ready"] = []string{"Kyber-768 + AES-KWP"}
	logos["pq_ready"] = []string{"Kyber-768 + AES-KWP (DEK/Field Key Wrap)"}

	return pb
}

// Name sets the policy name
func (pb *PolicyBuilder) Name(name string) *PolicyBuilder {
	pb.name = name
	return pb
}

// Description sets the policy description
func (pb *PolicyBuilder) Description(description string) *PolicyBuilder {
	pb.description = description
	return pb
}

// Create builds and creates the policy
func (pb *PolicyBuilder) Create(ctx context.Context) (*Policy, error) {
	if pb.name == "" {
		return nil, &ValidationError{Message: "policy name is required"}
	}

	return pb.client.Policies.Create(ctx, CreatePolicyRequest{
		Name:            pb.name,
		Description:     pb.description,
		Rules:           pb.rules,
		EnforcementMode: pb.enforcementMode,
	})
}

// Build returns the policy configuration without creating it
func (pb *PolicyBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"name":             pb.name,
		"description":      pb.description,
		"rules":            pb.rules,
		"enforcement_mode": pb.enforcementMode,
	}
}

// Utility functions

// CreateFromTemplate creates a policy from a compliance template
func CreateFromTemplate(template ComplianceFramework, name string, options ...PolicyOption) map[string]interface{} {
	templates := map[ComplianceFramework]map[string]interface{}{
		SOC2TypeII: {
			"compliance_standard": "SOC2 Type II",
			"aegis": map[string]interface{}{
				"pq_ready":             []string{"TLS_KYBER768_P256_SHA256"},
				"preferred":            []string{"TLS_AES_256_GCM_SHA384"},
				"fallback_acceptable":  []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
				"prohibited":           []string{"TLS 1.0", "TLS 1.1", "SSL V2", "SSL V3"},
			},
			"somnus": map[string]interface{}{
				"pq_ready":            []string{"Kyber-768 + AES-KWP"},
				"preferred":           []string{"XChaCha20-Poly1305", "AES-GCM-SIV-256"},
				"fallback_acceptable": []string{"AES-256-CBC + HMAC-SHA256/512"},
				"prohibited":          []string{"AES-ECB", "DES", "3DES", "RC4"},
			},
		},
		PCIDSS40: {
			"compliance_standard": "PCI DSS 4.0",
			"aegis": map[string]interface{}{
				"pq_ready":            []string{"TLS_KYBER768_P256_SHA256"},
				"preferred":           []string{"TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"},
				"fallback_acceptable": []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
				"prohibited":          []string{"SSLv2", "SSLv3", "TLS 1.0", "TLS 1.1", "RC4"},
			},
		},
	}

	config := templates[template]
	if config == nil {
		config = map[string]interface{}{
			"compliance_standard": "Custom",
			"aegis":               map[string]interface{}{},
		}
	}

	// Apply options
	for _, option := range options {
		option(config)
	}

	return config
}

// PolicyOption is a function that modifies policy configuration
type PolicyOption func(map[string]interface{})

// WithPostQuantum enables post-quantum cryptography
func WithPostQuantum() PolicyOption {
	return func(config map[string]interface{}) {
		if aegis, ok := config["aegis"].(map[string]interface{}); ok {
			if pqReady, exists := aegis["pq_ready"]; exists {
				if preferred, ok := aegis["preferred"].([]string); ok {
					if pqSlice, ok := pqReady.([]string); ok {
						aegis["preferred"] = append(pqSlice, preferred...)
					}
				}
			}
		}
	}
}

// WithEnforcementMode sets the enforcement mode
func WithEnforcementMode(mode string) PolicyOption {
	return func(config map[string]interface{}) {
		config["enforcement_mode"] = mode
	}
}

// Kubernetes operator helpers

// AgentConfigBuilder helps build agent configurations for Kubernetes
type AgentConfigBuilder struct {
	config map[string]interface{}
}

// NewAgentConfigBuilder creates a new agent configuration builder
func NewAgentConfigBuilder() *AgentConfigBuilder {
	return &AgentConfigBuilder{
		config: map[string]interface{}{
			"deployment_method": "helm",
			"namespace":         "aegis-system",
			"replicas":          1,
			"resources": map[string]interface{}{
				"cpu":    "100m",
				"memory": "128Mi",
			},
			"networking": map[string]interface{}{
				"listen_port": 8444,
				"health_port": 9080,
			},
		},
	}
}

// Namespace sets the Kubernetes namespace
func (acb *AgentConfigBuilder) Namespace(namespace string) *AgentConfigBuilder {
	acb.config["namespace"] = namespace
	return acb
}

// Replicas sets the number of agent replicas
func (acb *AgentConfigBuilder) Replicas(count int) *AgentConfigBuilder {
	acb.config["replicas"] = count
	return acb
}

// Resources sets CPU and memory limits
func (acb *AgentConfigBuilder) Resources(cpu, memory string) *AgentConfigBuilder {
	acb.config["resources"] = map[string]interface{}{
		"cpu":    cpu,
		"memory": memory,
	}
	return acb
}

// BackendURL sets the backend URL for proxying
func (acb *AgentConfigBuilder) BackendURL(url string) *AgentConfigBuilder {
	networking := acb.config["networking"].(map[string]interface{})
	networking["backend_url"] = url
	return acb
}

// Build returns the agent configuration
func (acb *AgentConfigBuilder) Build() map[string]interface{} {
	return acb.config
}

// Monitoring and alerting helpers

// AlertHandler defines the interface for handling security alerts
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert SecurityAlert) error
}

// AlertHandlerFunc is an adapter to allow functions to be used as AlertHandlers
type AlertHandlerFunc func(ctx context.Context, alert SecurityAlert) error

// HandleAlert calls the function
func (f AlertHandlerFunc) HandleAlert(ctx context.Context, alert SecurityAlert) error {
	return f(ctx, alert)
}

// StartAlertMonitoring starts monitoring for security alerts
func (c *Client) StartAlertMonitoring(ctx context.Context, handler AlertHandler) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			alerts, err := c.Monitoring.GetActiveAlerts(ctx)
			if err != nil {
				continue // Log error but don't stop monitoring
			}

			for _, alert := range alerts {
				// Only handle new alerts (created in last minute)
				if time.Since(alert.CreatedAt) < time.Minute {
					if err := handler.HandleAlert(ctx, alert); err != nil {
						// Log error but continue processing other alerts
						continue
					}
				}
			}
		}
	}
}

// Compliance automation helpers

// ComplianceChecker provides automated compliance validation
type ComplianceChecker struct {
	client     *Client
	frameworks []string
	thresholds map[string]int
}

// NewComplianceChecker creates a new compliance checker
func (c *Client) NewComplianceChecker(frameworks []string) *ComplianceChecker {
	return &ComplianceChecker{
		client:     c,
		frameworks: frameworks,
		thresholds: map[string]int{
			"soc2":    90, // 90% compliance required
			"pci-dss": 95, // 95% compliance required
			"hipaa":   90,
			"gdpr":    85,
		},
	}
}

// ValidateAll validates compliance against all configured frameworks
func (cc *ComplianceChecker) ValidateAll(ctx context.Context) (map[string]*ComplianceValidation, error) {
	results := make(map[string]*ComplianceValidation)

	for _, framework := range cc.frameworks {
		validation, err := cc.client.ValidateCompliance(ctx, framework)
		if err != nil {
			return nil, fmt.Errorf("compliance validation failed for %s: %w", framework, err)
		}
		results[framework] = validation
	}

	return results, nil
}

// CheckThresholds returns frameworks that don't meet compliance thresholds
func (cc *ComplianceChecker) CheckThresholds(ctx context.Context) ([]string, error) {
	validations, err := cc.ValidateAll(ctx)
	if err != nil {
		return nil, err
	}

	var failing []string
	for framework, validation := range validations {
		threshold := cc.thresholds[framework]
		if threshold > 0 && validation.Score < threshold {
			failing = append(failing, framework)
		}
	}

	return failing, nil
}
