package velikey

import "time"

// Agent represents an agent record returned by Axis APIs.
type Agent struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenantId,omitempty"`
	AgentID       string     `json:"agentId,omitempty"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	Version       string     `json:"version,omitempty"`
	Capabilities  []string   `json:"capabilities,omitempty"`
	LastHeartbeat *time.Time `json:"lastHeartbeat,omitempty"`
	EnrolledAt    *time.Time `json:"enrolledAt,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`
}

// Policy represents a policy returned by Axis APIs.
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Scope       string                 `json:"scope,omitempty"`
	ScopeValue  string                 `json:"scopeValue,omitempty"`
	PolicyType  string                 `json:"policyType,omitempty"`
	Rules       map[string]interface{} `json:"rules,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Analysis    map[string]interface{} `json:"analysis,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time             `json:"updatedAt,omitempty"`
}

// Alert represents a tenant-scoped alert.
type Alert struct {
	ID          string                 `json:"id"`
	Severity    string                 `json:"severity"`
	Category    string                 `json:"category"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Resolved    bool                   `json:"resolved"`
	Source      map[string]interface{} `json:"source,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   *time.Time             `json:"createdAt,omitempty"`
	ResolvedAt  *time.Time             `json:"resolvedAt,omitempty"`
}

// UsagePoint captures usage and estimated cost for a period.
type UsagePoint struct {
	Period        string  `json:"period"`
	DayOfPeriod   int     `json:"dayOfPeriod"`
	DaysInPeriod  int     `json:"daysInPeriod"`
	EncryptionGB  float64 `json:"encryptionGB"`
	TelemetryGB   float64 `json:"telemetryGB"`
	Environments  int     `json:"environments"`
	Agents        int     `json:"agents"`
	EstimatedCost float64 `json:"estimatedCost"`
}

// UsageResponse is the response payload from /api/usage.
type UsageResponse struct {
	Current    UsagePoint   `json:"current"`
	Usage      UsagePoint   `json:"usage"`
	Historical []UsagePoint `json:"historical"`
}

// UsageMetrics is kept as a compatibility alias.
type UsageMetrics = UsagePoint

// HealthResponse represents /api/health output.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// SecurityStatus is a normalized summary built from alert stats.
type SecurityStatus struct {
	AgentsOnline    string    `json:"agents_online"`
	PoliciesActive  int       `json:"policies_active"`
	HealthScore     int       `json:"health_score"`
	CriticalAlerts  int       `json:"critical_alerts"`
	Recommendations []string  `json:"recommendations"`
	LastUpdated     time.Time `json:"last_updated"`
}

// MaintenanceWindow defines a rollout maintenance window.
type MaintenanceWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// PlanRolloutRequest is sent to /api/rollouts/plan.
type PlanRolloutRequest struct {
	PolicyID             string              `json:"policyId"`
	CanaryPercent        *int                `json:"canaryPercent,omitempty"`
	StabilizationWindowS *int                `json:"stabilizationWindowS,omitempty"`
	MaintenanceWindows   []MaintenanceWindow `json:"maintenanceWindows,omitempty"`
	Explain              bool                `json:"explain,omitempty"`
}

// ApplyRolloutRequest is sent to /api/rollouts/apply.
type ApplyRolloutRequest struct {
	PlanID             string              `json:"planId"`
	DryRun             bool                `json:"dryRun"`
	IdempotencyKey     string              `json:"idempotencyKey,omitempty"`
	Confirmation       string              `json:"confirmation,omitempty"`
	Confirm            bool                `json:"confirm,omitempty"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenanceWindows,omitempty"`
}

// RollbackRolloutRequest is sent to /api/rollouts/rollback.
type RollbackRolloutRequest struct {
	RollbackToken string `json:"rollbackToken"`
	Confirmation  string `json:"confirmation,omitempty"`
	Confirm       bool   `json:"confirm,omitempty"`
}

// RolloutReceiptRef references the signed receipt generated by Axis.
type RolloutReceiptRef struct {
	ID string `json:"id"`
}

// RolloutOperationData captures common fields returned from plan/apply/rollback endpoints.
type RolloutOperationData struct {
	PlanID        string `json:"plan_id,omitempty"`
	RolloutID     string `json:"rollout_id,omitempty"`
	RollbackID    string `json:"rollback_id,omitempty"`
	RollbackToken string `json:"rollback_token,omitempty"`
}

// RolloutOperationResponse is a normalized rollout operation response.
type RolloutOperationResponse struct {
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Data           RolloutOperationData   `json:"data,omitempty"`
	RolloutReceipt *RolloutReceiptRef     `json:"rolloutReceipt,omitempty"`
	Raw            map[string]interface{} `json:"-"`
}

// RolloutReceipt represents an item returned by /api/rollout-receipts.
type RolloutReceipt struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resourceId"`
	CreatedAt *time.Time             `json:"createdAt,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TelemetryIngestRequest is accepted by /api/telemetry/ingest.
type TelemetryIngestRequest struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  string                 `json:"timestamp,omitempty"`
}

// TelemetryIngestResponse captures ingest acknowledgement.
type TelemetryIngestResponse struct {
	Success   bool   `json:"success"`
	Accepted  bool   `json:"accepted"`
	Queued    bool   `json:"queued"`
	Timestamp string `json:"timestamp"`
}

// SetupOptions is retained for API compatibility only.
type SetupOptions struct {
	ComplianceFramework string `json:"compliance_framework"`
	EnforcementMode     string `json:"enforcement_mode"`
	PostQuantum         bool   `json:"post_quantum"`
}

// SetupResult is retained for API compatibility only.
type SetupResult struct {
	PolicyID               string            `json:"policy_id"`
	PolicyName             string            `json:"policy_name"`
	DeploymentInstructions map[string]string `json:"deployment_instructions"`
	NextSteps              []string          `json:"next_steps"`
}

// ComplianceValidation is retained for API compatibility only.
type ComplianceValidation struct {
	Compliant       bool     `json:"compliant"`
	Score           int      `json:"score"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// OptimizationSuggestions is retained for API compatibility only.
type OptimizationSuggestions struct {
	Performance []string `json:"performance"`
	Security    []string `json:"security"`
	Cost        []string `json:"cost"`
}

// PolicyUpdate is retained for API compatibility only.
type PolicyUpdate struct {
	PolicyID    string                 `json:"policy_id"`
	Changes     map[string]interface{} `json:"changes"`
	Description string                 `json:"description,omitempty"`
}

// BulkUpdateResult is retained for API compatibility only.
type BulkUpdateResult struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Results    []Policy `json:"results"`
}

// ComplianceFramework represents compliance policy templates.
type ComplianceFramework string

const (
	SOC2TypeII ComplianceFramework = "soc2"
	PCIDSS40   ComplianceFramework = "pci-dss"
	HIPAA      ComplianceFramework = "hipaa"
	GDPR       ComplianceFramework = "gdpr"
	Custom     ComplianceFramework = "custom"
)

// PolicyMode represents rollout enforcement modes.
type PolicyMode string

const (
	Observe PolicyMode = "observe"
	Enforce PolicyMode = "enforce"
	Canary  PolicyMode = "canary"
)

// APIError represents a generic API failure.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// AuthenticationError represents an authentication failure.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NotFoundError represents a resource-not-found failure.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// RateLimitError represents a rate limit failure.
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// UnsupportedOperationError indicates an SDK call that does not map to current Axis APIs.
type UnsupportedOperationError struct {
	Method  string
	Message string
}

func (e *UnsupportedOperationError) Error() string {
	return e.Message
}

// AlertSeverity represents alert severity levels.
type AlertSeverity string

const (
	Info       AlertSeverity = "info"
	Warning    AlertSeverity = "warning"
	AlertError AlertSeverity = "error"
	Critical   AlertSeverity = "critical"
	Emergency  AlertSeverity = "emergency"
)
