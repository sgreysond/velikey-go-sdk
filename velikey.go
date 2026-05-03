// Package velikey provides the official Go SDK for Axis/Aegis control-plane APIs.
package velikey

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultBaseURL         = "https://axis.velikey.com"
	defaultUserAgent       = "velikey-go-sdk/0.2.0"
	defaultTimeout         = 30 * time.Second
	defaultRateLimit       = 10
	defaultMaxRetries      = 2
	defaultRetryMinBackoff = 250 * time.Millisecond
	defaultRetryMaxBackoff = 2 * time.Second
)

// Client is the primary Axis API client.
type Client struct {
	httpClient            *http.Client
	baseURL               string
	apiKey                string
	bearerToken           string
	sessionCookie         string
	sessionToken          string
	useSecureSessionToken bool
	userAgent             string
	limiter               *rate.Limiter
	maxRetries            int
	retryMinBackoff       time.Duration
	retryMaxBackoff       time.Duration

	// Resource services.
	Agents          *AgentsService
	Policies        *PoliciesService
	Alerts          *AlertsService
	Usage           *UsageService
	Rollouts        *RolloutsService
	RolloutReceipts *RolloutReceiptsService
	Telemetry       *TelemetryService
	Gateways        *GatewaysService

	// Compatibility service aliases.
	Monitoring  *MonitoringService
	Compliance  *ComplianceService
	Diagnostics *DiagnosticsService
	Billing     *BillingService
}

// Config controls client behavior and authentication.
type Config struct {
	// APIKey is sent as Authorization: Bearer <key> for API-key-compatible endpoints.
	APIKey string

	// BearerToken is sent as Authorization: Bearer <token>.
	BearerToken string

	// SessionCookie is a fully formed Cookie header value.
	// Example: "next-auth.session-token=...; __Secure-next-auth.session-token=..."
	SessionCookie string

	// SessionToken sets a single NextAuth session cookie automatically.
	SessionToken string

	// UseSecureSessionCookie switches SessionToken cookie name from
	// next-auth.session-token to __Secure-next-auth.session-token.
	UseSecureSessionCookie bool

	BaseURL         string
	HTTPClient      *http.Client
	UserAgent       string
	Timeout         time.Duration
	RateLimit       rate.Limit
	MaxRetries      int
	RetryMinBackoff time.Duration
	RetryMaxBackoff time.Duration
}

// NewClient creates a new VeliKey SDK client with sane defaults.
func NewClient(config Config) *Client {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		httpClient.Timeout = timeout
	}

	userAgent := strings.TrimSpace(config.UserAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	rps := config.RateLimit
	if rps <= 0 {
		rps = defaultRateLimit
	}

	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}

	retryMin := config.RetryMinBackoff
	if retryMin <= 0 {
		retryMin = defaultRetryMinBackoff
	}
	retryMax := config.RetryMaxBackoff
	if retryMax <= 0 {
		retryMax = defaultRetryMaxBackoff
	}
	if retryMax < retryMin {
		retryMax = retryMin
	}

	client := &Client{
		httpClient:            httpClient,
		baseURL:               baseURL,
		apiKey:                strings.TrimSpace(config.APIKey),
		bearerToken:           strings.TrimSpace(config.BearerToken),
		sessionCookie:         strings.TrimSpace(config.SessionCookie),
		sessionToken:          strings.TrimSpace(config.SessionToken),
		useSecureSessionToken: config.UseSecureSessionCookie,
		userAgent:             userAgent,
		limiter:               rate.NewLimiter(rps, 1),
		maxRetries:            maxRetries,
		retryMinBackoff:       retryMin,
		retryMaxBackoff:       retryMax,
	}

	client.Agents = &AgentsService{client: client}
	client.Policies = &PoliciesService{client: client}
	client.Alerts = &AlertsService{client: client}
	client.Usage = &UsageService{client: client}
	client.Rollouts = &RolloutsService{client: client}
	client.RolloutReceipts = &RolloutReceiptsService{client: client}
	client.Telemetry = &TelemetryService{client: client}
	client.Gateways = &GatewaysService{client: client}

	client.Monitoring = &MonitoringService{client: client}
	client.Compliance = &ComplianceService{client: client}
	client.Diagnostics = &DiagnosticsService{client: client}
	client.Billing = &BillingService{client: client}

	return client
}

func (c *Client) buildURL(endpoint string, params map[string]string) (string, error) {
	e := strings.TrimSpace(endpoint)
	if e == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	if !strings.HasPrefix(e, "/") {
		e = "/" + e
	}

	u, err := url.Parse(c.baseURL + e)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if len(params) > 0 {
		query := u.Query()
		for key, value := range params {
			if strings.TrimSpace(value) == "" {
				continue
			}
			query.Set(key, value)
		}
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

func (c *Client) applyAuth(req *http.Request) {
	token := strings.TrimSpace(c.bearerToken)
	if token == "" {
		token = strings.TrimSpace(c.apiKey)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	cookie := strings.TrimSpace(c.sessionCookie)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
		return
	}

	if strings.TrimSpace(c.sessionToken) == "" {
		return
	}
	cookieName := "next-auth.session-token"
	if c.useSecureSessionToken {
		cookieName = "__Secure-next-auth.session-token"
	}
	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", cookieName, c.sessionToken))
}

func (c *Client) request(
	ctx context.Context,
	method, endpoint string,
	params map[string]string,
	body interface{},
) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	urlString, err := c.buildURL(endpoint, params)
	if err != nil {
		return nil, err
	}

	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithContext(ctx, c.backoffDelay(attempt)); err != nil {
				return nil, err
			}
		}

		var reqBody io.Reader
		if payload != nil {
			reqBody = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, urlString, reqBody)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		c.applyAuth(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries && isRetryableError(err) {
				continue
			}
			return nil, fmt.Errorf("request failed: %w", err)
		}

		responseBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read response body: %w", readErr)
		}

		if resp.StatusCode >= 400 {
			apiErr := mapStatusError(resp.StatusCode, responseBytes)
			lastErr = apiErr
			if attempt < c.maxRetries && isRetryableStatus(resp.StatusCode) {
				continue
			}
			return nil, apiErr
		}

		return responseBytes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, &APIError{StatusCode: 500, Message: "request failed with unknown error"}
}

func (c *Client) getJSON(
	ctx context.Context,
	endpoint string,
	params map[string]string,
	out interface{},
) error {
	payload, err := c.request(ctx, http.MethodGet, endpoint, params, nil)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) postJSON(
	ctx context.Context,
	endpoint string,
	params map[string]string,
	requestBody interface{},
	out interface{},
) error {
	payload, err := c.request(ctx, http.MethodPost, endpoint, params, requestBody)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func mapStatusError(statusCode int, body []byte) error {
	message := extractErrorMessage(statusCode, body)

	switch statusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{Message: message}
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusForbidden:
		return &ValidationError{Message: message}
	case http.StatusNotFound:
		return &NotFoundError{Message: message}
	case http.StatusTooManyRequests:
		return &RateLimitError{Message: message}
	default:
		return &APIError{StatusCode: statusCode, Message: message}
	}
}

func extractErrorMessage(statusCode int, body []byte) string {
	if len(body) == 0 {
		return fmt.Sprintf("request failed with HTTP %d", statusCode)
	}

	var structured struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &structured); err == nil {
		if strings.TrimSpace(structured.Error) != "" {
			return strings.TrimSpace(structured.Error)
		}
		if strings.TrimSpace(structured.Message) != "" {
			return strings.TrimSpace(structured.Message)
		}
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Sprintf("request failed with HTTP %d", statusCode)
	}
	return message
}

func isRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return true
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) backoffDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	base := float64(c.retryMinBackoff)
	max := float64(c.retryMaxBackoff)
	delay := base * math.Pow(2, float64(attempt-1))
	if delay > max {
		delay = max
	}

	// Add low jitter to avoid synchronized retries.
	jitter := 0.75 + (rand.Float64() * 0.5) // 75%..125%
	withJitter := time.Duration(delay * jitter)
	if withJitter < c.retryMinBackoff {
		return c.retryMinBackoff
	}
	if withJitter > c.retryMaxBackoff {
		return c.retryMaxBackoff
	}
	return withJitter
}

func unsupportedOperation(method, replacement string) error {
	message := fmt.Sprintf("%s is not supported by current Axis public routes", method)
	if replacement != "" {
		message = fmt.Sprintf("%s; use %s instead", message, replacement)
	}
	return &UnsupportedOperationError{Method: method, Message: message}
}

// QuickSetup is retained for compatibility and returns an explicit unsupported error.
func (c *Client) QuickSetup(_ context.Context, _ SetupOptions) (*SetupResult, error) {
	return nil, unsupportedOperation("QuickSetup", "Rollouts.Plan + Rollouts.Apply")
}

// GetSecurityStatus returns a best-effort health summary derived from /api/alerts/stats.
func (c *Client) GetSecurityStatus(ctx context.Context) (*SecurityStatus, error) {
	var stats struct {
		GeneratedAt string `json:"generatedAt"`
		BySeverity  []struct {
			Severity string `json:"severity"`
			Count    int    `json:"count"`
		} `json:"bySeverity"`
	}

	if err := c.getJSON(ctx, "/api/alerts/stats", nil, &stats); err != nil {
		return nil, err
	}

	criticalCount := 0
	for _, bucket := range stats.BySeverity {
		if bucket.Severity == string(Critical) || bucket.Severity == string(Emergency) {
			criticalCount += bucket.Count
		}
	}

	healthScore := 100 - (criticalCount * 10)
	if healthScore < 0 {
		healthScore = 0
	}

	status := &SecurityStatus{
		AgentsOnline:    "unknown",
		PoliciesActive:  0,
		HealthScore:     healthScore,
		CriticalAlerts:  criticalCount,
		Recommendations: []string{},
		LastUpdated:     time.Now().UTC(),
	}

	if parsed, err := time.Parse(time.RFC3339, stats.GeneratedAt); err == nil {
		status.LastUpdated = parsed
	}

	return status, nil
}

// GetHealth checks Axis API health.
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	if err := c.getJSON(ctx, "/api/health", nil, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

// ValidateCompliance is retained for compatibility and returns an explicit unsupported error.
func (c *Client) ValidateCompliance(_ context.Context, _ string) (*ComplianceValidation, error) {
	return nil, unsupportedOperation("ValidateCompliance", "Compliance bundle APIs")
}

// GetOptimizationSuggestions is retained for compatibility and returns an explicit unsupported error.
func (c *Client) GetOptimizationSuggestions(_ context.Context) (*OptimizationSuggestions, error) {
	return nil, unsupportedOperation("GetOptimizationSuggestions", "tenant-specific observability queries")
}

// BulkPolicyUpdate is retained for compatibility and returns an explicit unsupported error.
func (c *Client) BulkPolicyUpdate(_ context.Context, _ []PolicyUpdate) (*BulkUpdateResult, error) {
	return nil, unsupportedOperation("BulkPolicyUpdate", "policy management routes via Axis UI/API")
}

// PolicyBuilder provides a fluent interface for building policy payloads.
type PolicyBuilder struct {
	client          *Client
	rules           map[string]interface{}
	enforcementMode string
	name            string
	description     string
}

// NewPolicyBuilder creates a new policy builder.
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

// ComplianceStandard sets the compliance standard.
func (pb *PolicyBuilder) ComplianceStandard(standard string) *PolicyBuilder {
	pb.rules["compliance_standard"] = standard
	return pb
}

// AegisConfig sets Aegis configuration.
func (pb *PolicyBuilder) AegisConfig(config map[string]interface{}) *PolicyBuilder {
	aegis, _ := pb.rules["aegis"].(map[string]interface{})
	if aegis == nil {
		aegis = map[string]interface{}{}
		pb.rules["aegis"] = aegis
	}
	for key, value := range config {
		aegis[key] = value
	}
	return pb
}

// EnforcementMode sets the enforcement mode.
func (pb *PolicyBuilder) EnforcementMode(mode string) *PolicyBuilder {
	pb.enforcementMode = mode
	return pb
}

// PostQuantumReady enables post-quantum-ready defaults.
func (pb *PolicyBuilder) PostQuantumReady() *PolicyBuilder {
	aegis := pb.ensureRulesMap("aegis")
	somnus := pb.ensureRulesMap("somnus")
	logos := pb.ensureRulesMap("logos")

	aegis["pq_ready"] = []string{"TLS_KYBER768_P256_SHA256"}
	somnus["pq_ready"] = []string{"Kyber-768 + AES-KWP"}
	logos["pq_ready"] = []string{"Kyber-768 + AES-KWP (DEK/Field Key Wrap)"}

	return pb
}

func (pb *PolicyBuilder) ensureRulesMap(key string) map[string]interface{} {
	current, _ := pb.rules[key].(map[string]interface{})
	if current == nil {
		current = map[string]interface{}{}
		pb.rules[key] = current
	}
	return current
}

// Name sets the policy name.
func (pb *PolicyBuilder) Name(name string) *PolicyBuilder {
	pb.name = name
	return pb
}

// Description sets the policy description.
func (pb *PolicyBuilder) Description(description string) *PolicyBuilder {
	pb.description = description
	return pb
}

// Create is retained for compatibility and returns an explicit unsupported error.
func (pb *PolicyBuilder) Create(_ context.Context) (*Policy, error) {
	if strings.TrimSpace(pb.name) == "" {
		return nil, &ValidationError{Message: "policy name is required"}
	}
	return nil, unsupportedOperation("PolicyBuilder.Create", "Rollouts.Plan against an existing policyId")
}

// Build returns policy configuration without applying it.
func (pb *PolicyBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"name":             pb.name,
		"description":      pb.description,
		"rules":            pb.rules,
		"enforcement_mode": pb.enforcementMode,
	}
}

// CreateFromTemplate creates policy configuration from a compliance template.
func CreateFromTemplate(template ComplianceFramework, _ string, options ...PolicyOption) map[string]interface{} {
	templates := map[ComplianceFramework]map[string]interface{}{
		SOC2TypeII: {
			"compliance_standard": "SOC2 Type II",
			"aegis": map[string]interface{}{
				"pq_ready":            []string{"TLS_KYBER768_P256_SHA256"},
				"preferred":           []string{"TLS_AES_256_GCM_SHA384"},
				"fallback_acceptable": []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
				"prohibited":          []string{"TLS 1.0", "TLS 1.1", "SSL V2", "SSL V3"},
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

	for _, option := range options {
		option(config)
	}

	return config
}

// PolicyOption mutates a policy configuration.
type PolicyOption func(map[string]interface{})

// WithPostQuantum prepends post-quantum suites to preferred algorithms.
func WithPostQuantum() PolicyOption {
	return func(config map[string]interface{}) {
		aegis, _ := config["aegis"].(map[string]interface{})
		if aegis == nil {
			return
		}
		pqReady, _ := aegis["pq_ready"].([]string)
		preferred, _ := aegis["preferred"].([]string)
		if len(pqReady) == 0 {
			return
		}
		aegis["preferred"] = append(append([]string{}, pqReady...), preferred...)
	}
}

// WithEnforcementMode sets enforcement_mode in template config.
func WithEnforcementMode(mode string) PolicyOption {
	return func(config map[string]interface{}) {
		config["enforcement_mode"] = mode
	}
}

// AgentConfigBuilder builds agent config for deployment templates.
type AgentConfigBuilder struct {
	config map[string]interface{}
}

// NewAgentConfigBuilder creates a new agent config builder.
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

// Namespace sets the Kubernetes namespace.
func (acb *AgentConfigBuilder) Namespace(namespace string) *AgentConfigBuilder {
	acb.config["namespace"] = namespace
	return acb
}

// Replicas sets the number of replicas.
func (acb *AgentConfigBuilder) Replicas(count int) *AgentConfigBuilder {
	acb.config["replicas"] = count
	return acb
}

// Resources sets CPU and memory resources.
func (acb *AgentConfigBuilder) Resources(cpu, memory string) *AgentConfigBuilder {
	acb.config["resources"] = map[string]interface{}{
		"cpu":    cpu,
		"memory": memory,
	}
	return acb
}

// BackendURL sets backend URL for networking.
func (acb *AgentConfigBuilder) BackendURL(backendURL string) *AgentConfigBuilder {
	networking, _ := acb.config["networking"].(map[string]interface{})
	if networking == nil {
		networking = map[string]interface{}{}
		acb.config["networking"] = networking
	}
	networking["backend_url"] = backendURL
	return acb
}

// Build returns final agent configuration.
func (acb *AgentConfigBuilder) Build() map[string]interface{} {
	return acb.config
}

// AlertHandler handles each alert in StartAlertMonitoring.
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert Alert) error
}

// AlertHandlerFunc allows plain functions to be used as AlertHandler.
type AlertHandlerFunc func(ctx context.Context, alert Alert) error

// HandleAlert invokes the function.
func (f AlertHandlerFunc) HandleAlert(ctx context.Context, alert Alert) error {
	return f(ctx, alert)
}

// StartAlertMonitoring polls unresolved alerts and forwards them to handler.
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
				continue
			}
			for _, alert := range alerts {
				if err := handler.HandleAlert(ctx, alert); err != nil {
					continue
				}
			}
		}
	}
}

// ComplianceChecker performs repeated compliance checks.
type ComplianceChecker struct {
	client     *Client
	frameworks []string
	thresholds map[string]int
}

// NewComplianceChecker creates a checker instance.
func (c *Client) NewComplianceChecker(frameworks []string) *ComplianceChecker {
	return &ComplianceChecker{
		client:     c,
		frameworks: frameworks,
		thresholds: map[string]int{
			"soc2":    90,
			"pci-dss": 95,
			"hipaa":   90,
			"gdpr":    85,
		},
	}
}

// ValidateAll validates configured compliance frameworks.
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

// CheckThresholds returns frameworks below configured threshold.
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
