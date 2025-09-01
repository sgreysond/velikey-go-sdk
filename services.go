package velikey

import (
	"context"
	"encoding/json"
)

// AgentsService handles agent-related operations
type AgentsService struct {
	client *Client
}

// List returns all agents
func (s *AgentsService) List(ctx context.Context) ([]Agent, error) {
	resp, err := s.client.request(ctx, "GET", "/api/agents", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var agents []Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}
	return agents, nil
}

// Get returns a specific agent
func (s *AgentsService) Get(ctx context.Context, id string) (*Agent, error) {
	resp, err := s.client.request(ctx, "GET", "/api/agents/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var agent Agent
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// PoliciesService handles policy-related operations
type PoliciesService struct {
	client *Client
}

// CreatePolicyRequest represents a policy creation request
type CreatePolicyRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Rules           map[string]interface{} `json:"rules"`
	EnforcementMode string                 `json:"enforcement_mode"`
}

// Create creates a new policy
func (s *PoliciesService) Create(ctx context.Context, req CreatePolicyRequest) (*Policy, error) {
	resp, err := s.client.request(ctx, "POST", "/api/policies", nil, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// List returns all policies
func (s *PoliciesService) List(ctx context.Context) ([]Policy, error) {
	resp, err := s.client.request(ctx, "GET", "/api/policies", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// Get returns a specific policy
func (s *PoliciesService) Get(ctx context.Context, id string) (*Policy, error) {
	resp, err := s.client.request(ctx, "GET", "/api/policies/"+id, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// MonitoringService handles monitoring operations
type MonitoringService struct {
	client *Client
}

// GetActiveAlerts returns active security alerts
func (s *MonitoringService) GetActiveAlerts(ctx context.Context) ([]SecurityAlert, error) {
	resp, err := s.client.request(ctx, "GET", "/api/monitoring/alerts", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var alerts []SecurityAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

// GetMetrics returns current metrics
func (s *MonitoringService) GetMetrics(ctx context.Context) (*Metrics, error) {
	resp, err := s.client.request(ctx, "GET", "/api/monitoring/metrics", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var metrics Metrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}
	return &metrics, nil
}

// ComplianceService handles compliance operations
type ComplianceService struct {
	client *Client
}

// Validate validates compliance against a framework
func (s *ComplianceService) Validate(ctx context.Context, framework string) (*ComplianceValidation, error) {
	resp, err := s.client.request(ctx, "POST", "/api/compliance/validate", nil, map[string]string{"framework": framework})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var validation ComplianceValidation
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, err
	}
	return &validation, nil
}

// DiagnosticsService handles diagnostic operations
type DiagnosticsService struct {
	client *Client
}

// RunDiagnostics runs system diagnostics
func (s *DiagnosticsService) RunDiagnostics(ctx context.Context) (*DiagnosticsResult, error) {
	resp, err := s.client.request(ctx, "POST", "/api/diagnostics/run", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result DiagnosticsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BillingService handles billing operations
type BillingService struct {
	client *Client
}

// GetUsage returns current usage metrics
func (s *BillingService) GetUsage(ctx context.Context) (*UsageMetrics, error) {
	resp, err := s.client.request(ctx, "GET", "/api/billing/usage", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var usage UsageMetrics
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, err
	}
	return &usage, nil
}
