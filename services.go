package velikey

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// AgentsService provides agent APIs.
type AgentsService struct {
	client *Client
}

// List returns all tenant-visible agents.
func (s *AgentsService) List(ctx context.Context) ([]Agent, error) {
	return s.ListByAgentID(ctx, "")
}

// ListByAgentID returns agents filtered by agentId when provided.
func (s *AgentsService) ListByAgentID(ctx context.Context, agentID string) ([]Agent, error) {
	params := map[string]string{}
	if strings.TrimSpace(agentID) != "" {
		params["agentId"] = strings.TrimSpace(agentID)
	}

	var response struct {
		Agents []Agent `json:"agents"`
	}
	if err := s.client.getJSON(ctx, "/api/agents", params, &response); err != nil {
		return nil, err
	}
	return response.Agents, nil
}

// Get returns one agent by agentId.
func (s *AgentsService) Get(ctx context.Context, agentID string) (*Agent, error) {
	trimmed := strings.TrimSpace(agentID)
	if trimmed == "" {
		return nil, &ValidationError{Message: "agentID is required"}
	}

	agents, err := s.ListByAgentID(ctx, trimmed)
	if err != nil {
		return nil, err
	}
	for _, agent := range agents {
		if agent.AgentID == trimmed {
			copy := agent
			return &copy, nil
		}
	}
	if len(agents) == 1 {
		copy := agents[0]
		return &copy, nil
	}

	return nil, &NotFoundError{Message: "agent not found"}
}

// PoliciesService provides policy APIs.
type PoliciesService struct {
	client *Client
}

// ListPoliciesOptions controls policy list filters.
type ListPoliciesOptions struct {
	Scope    string
	IsActive *bool
}

// List returns all policies.
func (s *PoliciesService) List(ctx context.Context) ([]Policy, error) {
	return s.ListWithOptions(ctx, ListPoliciesOptions{})
}

// ListWithOptions returns policies with optional filters.
func (s *PoliciesService) ListWithOptions(ctx context.Context, opts ListPoliciesOptions) ([]Policy, error) {
	params := map[string]string{}
	if strings.TrimSpace(opts.Scope) != "" {
		params["scope"] = strings.TrimSpace(opts.Scope)
	}
	if opts.IsActive != nil {
		params["isActive"] = strconv.FormatBool(*opts.IsActive)
	}

	var response struct {
		Policies []Policy `json:"policies"`
	}
	if err := s.client.getJSON(ctx, "/api/policies", params, &response); err != nil {
		return nil, err
	}
	return response.Policies, nil
}

// AgentPoliciesResponse captures /api/agents/{id}/policies response.
type AgentPoliciesResponse struct {
	AgentID   string   `json:"agentId"`
	Policies  []Policy `json:"policies"`
	FetchedAt string   `json:"fetchedAt"`
}

// ListForAgent returns policies visible to a specific agent.
func (s *PoliciesService) ListForAgent(ctx context.Context, agentID string) (*AgentPoliciesResponse, error) {
	trimmed := strings.TrimSpace(agentID)
	if trimmed == "" {
		return nil, &ValidationError{Message: "agentID is required"}
	}

	var response AgentPoliciesResponse
	endpoint := fmt.Sprintf("/api/agents/%s/policies", trimmed)
	if err := s.client.getJSON(ctx, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// AlertsService provides alert APIs.
type AlertsService struct {
	client *Client
}

// ListAlertsOptions controls alert listing filters.
type ListAlertsOptions struct {
	Severity string
	Category string
	Resolved *bool
	Limit    int
}

// List returns alerts without filters.
func (s *AlertsService) List(ctx context.Context) ([]Alert, error) {
	return s.ListWithOptions(ctx, ListAlertsOptions{})
}

// ListWithOptions returns alerts filtered by options.
func (s *AlertsService) ListWithOptions(ctx context.Context, opts ListAlertsOptions) ([]Alert, error) {
	params := map[string]string{}
	if strings.TrimSpace(opts.Severity) != "" {
		params["severity"] = strings.TrimSpace(opts.Severity)
	}
	if strings.TrimSpace(opts.Category) != "" {
		params["category"] = strings.TrimSpace(opts.Category)
	}
	if opts.Resolved != nil {
		params["resolved"] = strconv.FormatBool(*opts.Resolved)
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}

	var response struct {
		Alerts []Alert `json:"alerts"`
	}
	if err := s.client.getJSON(ctx, "/api/alerts", params, &response); err != nil {
		return nil, err
	}
	return response.Alerts, nil
}

// CreateAlertRequest contains required fields for POST /api/alerts.
type CreateAlertRequest struct {
	Severity    string                 `json:"severity"`
	Category    string                 `json:"category"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      map[string]interface{} `json:"source"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Create creates a new alert.
func (s *AlertsService) Create(ctx context.Context, req CreateAlertRequest) (*Alert, error) {
	if strings.TrimSpace(req.Severity) == "" || strings.TrimSpace(req.Category) == "" ||
		strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" {
		return nil, &ValidationError{Message: "severity, category, title, and description are required"}
	}
	if req.Source == nil {
		req.Source = map[string]interface{}{"component": "velikey-go-sdk"}
	}

	var response struct {
		Alert Alert `json:"alert"`
	}
	if err := s.client.postJSON(ctx, "/api/alerts", nil, req, &response); err != nil {
		return nil, err
	}
	return &response.Alert, nil
}

// UsageService provides usage/metering APIs.
type UsageService struct {
	client *Client
}

// Get returns usage for a period. Allowed values include current, 3months, 6months, year.
func (s *UsageService) Get(ctx context.Context, period string) (*UsageResponse, error) {
	params := map[string]string{}
	if strings.TrimSpace(period) != "" {
		params["period"] = strings.TrimSpace(period)
	}

	var response UsageResponse
	if err := s.client.getJSON(ctx, "/api/usage", params, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// RolloutsService provides rollout plan/apply/rollback APIs.
type RolloutsService struct {
	client *Client
}

// Plan creates a rollout plan (dry-run explanation by default on server side).
func (s *RolloutsService) Plan(ctx context.Context, req PlanRolloutRequest) (*RolloutOperationResponse, error) {
	if strings.TrimSpace(req.PolicyID) == "" {
		return nil, &ValidationError{Message: "policyId is required"}
	}

	var response RolloutOperationResponse
	if err := s.client.postJSON(ctx, "/api/rollouts/plan", nil, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Apply applies a rollout plan. Confirmation is auto-populated for non-dry-run requests.
func (s *RolloutsService) Apply(ctx context.Context, req ApplyRolloutRequest) (*RolloutOperationResponse, error) {
	if strings.TrimSpace(req.PlanID) == "" {
		return nil, &ValidationError{Message: "planId is required"}
	}
	if !req.DryRun {
		if !req.Confirm {
			req.Confirm = true
		}
		if strings.TrimSpace(req.Confirmation) == "" {
			req.Confirmation = "APPLY"
		}
	}

	var response RolloutOperationResponse
	if err := s.client.postJSON(ctx, "/api/rollouts/apply", nil, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// Rollback rolls back an applied rollout. Confirmation is auto-populated.
func (s *RolloutsService) Rollback(ctx context.Context, req RollbackRolloutRequest) (*RolloutOperationResponse, error) {
	if strings.TrimSpace(req.RollbackToken) == "" {
		return nil, &ValidationError{Message: "rollbackToken is required"}
	}
	if !req.Confirm {
		req.Confirm = true
	}
	if strings.TrimSpace(req.Confirmation) == "" {
		req.Confirmation = "ROLLBACK"
	}

	var response RolloutOperationResponse
	if err := s.client.postJSON(ctx, "/api/rollouts/rollback", nil, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// RolloutReceiptsService provides receipt listing APIs.
type RolloutReceiptsService struct {
	client *Client
}

// List returns recent rollout receipts.
func (s *RolloutReceiptsService) List(ctx context.Context, limit int) ([]RolloutReceipt, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	var response struct {
		Items []RolloutReceipt `json:"items"`
	}
	if err := s.client.getJSON(ctx, "/api/rollout-receipts", params, &response); err != nil {
		return nil, err
	}
	return response.Items, nil
}

// TelemetryService provides telemetry ingest APIs.
type TelemetryService struct {
	client *Client
}

// Ingest submits one telemetry event.
func (s *TelemetryService) Ingest(ctx context.Context, req TelemetryIngestRequest) (*TelemetryIngestResponse, error) {
	if strings.TrimSpace(req.Event) == "" {
		return nil, &ValidationError{Message: "event is required"}
	}

	var response TelemetryIngestResponse
	if err := s.client.postJSON(ctx, "/api/telemetry/ingest", nil, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// MonitoringService is a compatibility wrapper over alerts/metrics.
type MonitoringService struct {
	client *Client
}

// GetActiveAlerts returns unresolved alerts.
func (s *MonitoringService) GetActiveAlerts(ctx context.Context) ([]Alert, error) {
	resolved := false
	return s.client.Alerts.ListWithOptions(ctx, ListAlertsOptions{Resolved: &resolved})
}

// Metrics represents monitoring metrics.
type Metrics struct {
	CPU         float64            `json:"cpu"`
	Memory      float64            `json:"memory"`
	Connections int                `json:"connections"`
	Throughput  float64            `json:"throughput"`
	Latency     map[string]float64 `json:"latency"`
}

// GetMetrics is retained for compatibility and returns unsupported-operation error.
func (s *MonitoringService) GetMetrics(_ context.Context) (*Metrics, error) {
	return nil, unsupportedOperation("Monitoring.GetMetrics", "tenant dashboard analytics endpoints")
}

// ComplianceService compatibility wrapper.
type ComplianceService struct {
	client *Client
}

// Validate validates a compliance framework (unsupported in current public route set).
func (s *ComplianceService) Validate(ctx context.Context, framework string) (*ComplianceValidation, error) {
	return s.client.ValidateCompliance(ctx, framework)
}

// DiagnosticsService compatibility wrapper.
type DiagnosticsService struct {
	client *Client
}

// DiagnosticsResult is retained for compatibility.
type DiagnosticsResult struct {
	Status  string                 `json:"status"`
	Checks  map[string]bool        `json:"checks"`
	Details map[string]interface{} `json:"details"`
	Errors  []string               `json:"errors"`
}

// RunDiagnostics is retained for compatibility and returns unsupported-operation error.
func (s *DiagnosticsService) RunDiagnostics(_ context.Context) (*DiagnosticsResult, error) {
	return nil, unsupportedOperation("Diagnostics.RunDiagnostics", "route-specific readiness checks")
}

// BillingService compatibility wrapper.
type BillingService struct {
	client *Client
}

// GetUsage returns the current usage point from /api/usage.
func (s *BillingService) GetUsage(ctx context.Context) (*UsageMetrics, error) {
	usage, err := s.client.Usage.Get(ctx, "current")
	if err != nil {
		return nil, err
	}
	point := usage.Current
	return &point, nil
}
