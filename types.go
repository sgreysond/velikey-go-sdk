package velikey

import (
	"time"
)

// Agent represents a VeliKey agent
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Status       string            `json:"status"`
	Location     string            `json:"location"`
	Capabilities []string          `json:"capabilities"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	Uptime       string            `json:"uptime"`
	Metadata     map[string]string `json:"metadata"`
}

// Policy represents a security policy
type Policy struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	ComplianceFramework string               `json:"compliance_framework"`
	Rules             map[string]interface{} `json:"rules"`
	EnforcementMode   string                 `json:"enforcement_mode"`
	IsActive          bool                   `json:"is_active"`
	Version           int                    `json:"version"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         string                 `json:"created_by,omitempty"`
}

// PolicyTemplate represents a pre-built policy template
type PolicyTemplate struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	ComplianceFramework string                 `json:"compliance_framework"`
	Algorithms          map[string]interface{} `json:"algorithms"`
	RecommendedFor      []string               `json:"recommended_for"`
}

// HealthScore represents customer health metrics
type HealthScore struct {
	OverallScore    int               `json:"overall_score"`
	CategoryScores  map[string]int    `json:"category_scores"`
	RiskFactors     []string          `json:"risk_factors"`
	Recommendations []string          `json:"recommendations"`
	Trend           string            `json:"trend"`
	CalculatedAt    time.Time         `json:"calculated_at"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Category    string            `json:"category"`
	Source      string            `json:"source"`
	CreatedAt   time.Time         `json:"created_at"`
	Resolved    bool              `json:"resolved"`
	Metadata    map[string]string `json:"metadata"`
}

// UsageMetrics represents customer usage data
type UsageMetrics struct {
	AgentsDeployed       int       `json:"agents_deployed"`
	PoliciesActive       int       `json:"policies_active"`
	ConnectionsProcessed int64     `json:"connections_processed"`
	BytesAnalyzed        int64     `json:"bytes_analyzed"`
	UptimePercentage     float64   `json:"uptime_percentage"`
	AvgLatencyMs         float64   `json:"avg_latency_ms"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// DiagnosticSuite represents diagnostic test results
type DiagnosticSuite struct {
	ID          string             `json:"id"`
	CustomerID  string             `json:"customer_id"`
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt *time.Time         `json:"completed_at"`
	Status      string             `json:"status"`
	Results     []DiagnosticResult `json:"results"`
	Summary     DiagnosticSummary  `json:"summary"`
}

// DiagnosticResult represents a single diagnostic test result
type DiagnosticResult struct {
	TestName       string            `json:"test_name"`
	Category       string            `json:"category"`
	Status         string            `json:"status"`
	Message        string            `json:"message"`
	Details        string            `json:"details,omitempty"`
	FixSuggestions []FixSuggestion   `json:"fix_suggestions"`
	DurationMs     int64             `json:"duration_ms"`
}

// DiagnosticSummary represents overall diagnostic results
type DiagnosticSummary struct {
	TotalTests      int      `json:"total_tests"`
	PassedTests     int      `json:"passed_tests"`
	FailedTests     int      `json:"failed_tests"`
	WarningTests    int      `json:"warning_tests"`
	OverallHealth   string   `json:"overall_health"`
	CriticalIssues  []string `json:"critical_issues"`
	NextSteps       []string `json:"next_steps"`
}

// FixSuggestion represents an automated fix suggestion
type FixSuggestion struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	Command          string `json:"command,omitempty"`
	DocumentationURL string `json:"documentation_url,omitempty"`
	AutoFixable      bool   `json:"auto_fixable"`
}

// SecurityStatus represents comprehensive security posture
type SecurityStatus struct {
	AgentsOnline    string    `json:"agents_online"`
	PoliciesActive  int       `json:"policies_active"`
	HealthScore     int       `json:"health_score"`
	CriticalAlerts  int       `json:"critical_alerts"`
	Recommendations []string  `json:"recommendations"`
	LastUpdated     time.Time `json:"last_updated"`
}

// CustomerInfo represents customer account information
type CustomerInfo struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Company      string     `json:"company"`
	Plan         string     `json:"plan"`
	Status       string     `json:"status"`
	TrialEndsAt  *time.Time `json:"trial_ends_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SetupOptions for quick setup
type SetupOptions struct {
	ComplianceFramework string `json:"compliance_framework"`
	EnforcementMode     string `json:"enforcement_mode"`
	PostQuantum         bool   `json:"post_quantum"`
}

// SetupResult from quick setup
type SetupResult struct {
	PolicyID               string            `json:"policy_id"`
	PolicyName             string            `json:"policy_name"`
	DeploymentInstructions map[string]string `json:"deployment_instructions"`
	NextSteps              []string          `json:"next_steps"`
}

// ComplianceValidation represents compliance check results
type ComplianceValidation struct {
	Compliant       bool     `json:"compliant"`
	Score           int      `json:"score"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// OptimizationSuggestions represents system optimization recommendations
type OptimizationSuggestions struct {
	Performance []string `json:"performance"`
	Security    []string `json:"security"`
	Cost        []string `json:"cost"`
}

// PolicyUpdate represents a policy update operation
type PolicyUpdate struct {
	PolicyID    string                 `json:"policy_id"`
	Changes     map[string]interface{} `json:"changes"`
	Description string                 `json:"description,omitempty"`
}

// BulkUpdateResult represents bulk policy update results
type BulkUpdateResult struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Results    []Policy `json:"results"`
}



// UpdateOptions for agent updates
type UpdateOptions struct {
	Version  string `json:"version"`
	Strategy string `json:"strategy"` // "rolling", "blue-green", "canary"
}

// HealthResponse represents API health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Enums

// ComplianceFramework represents supported compliance frameworks
type ComplianceFramework string

const (
	SOC2TypeII ComplianceFramework = "soc2"
	PCIDSS40   ComplianceFramework = "pci-dss"
	HIPAA      ComplianceFramework = "hipaa"
	GDPR       ComplianceFramework = "gdpr"
	Custom     ComplianceFramework = "custom"
)

// Error types

// APIError represents a general API error
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// AuthenticationError represents an authentication failure
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// ValidationError represents a validation failure
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// Metrics represents system metrics
type Metrics struct {
	CPU         float64           `json:"cpu"`
	Memory      float64           `json:"memory"`
	Connections int               `json:"connections"`
	Throughput  float64           `json:"throughput"`
	Latency     map[string]float64 `json:"latency"`
}

// DiagnosticsResult represents the result of a diagnostic run
type DiagnosticsResult struct {
	Status  string                 `json:"status"`
	Checks  map[string]bool        `json:"checks"`
	Details map[string]interface{} `json:"details"`
	Errors  []string               `json:"errors"`
}

// PolicyMode represents policy enforcement modes
type PolicyMode string

const (
	Observe PolicyMode = "observe"
	Enforce PolicyMode = "enforce"
	Canary  PolicyMode = "canary"
)

// AgentStatus represents agent status
type AgentStatus string

const (
	Online   AgentStatus = "online"
	Offline  AgentStatus = "offline"
	Updating AgentStatus = "updating"
	Error    AgentStatus = "error"
	Degraded AgentStatus = "degraded"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	Info      AlertSeverity = "info"
	Warning   AlertSeverity = "warning"
	AlertError AlertSeverity = "error"
	Critical  AlertSeverity = "critical"
	Emergency AlertSeverity = "emergency"
)
