package velikey

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Gateway represents a productized data-plane deployment unit.
// See aegis/docs/api/gateways.openapi.yaml for the full schema.
type Gateway struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenantId"`
	Name          string     `json:"name"`
	Mode          string     `json:"mode"`
	Template      *string    `json:"template,omitempty"`
	Status        string     `json:"status"`
	AgentID       *string    `json:"agentId,omitempty"`
	AgentVersion  *string    `json:"agentVersion,omitempty"`
	ChartVersion  *string    `json:"chartVersion,omitempty"`
	CertExpiresAt *time.Time `json:"certExpiresAt,omitempty"`
	BackendURL    *string    `json:"backendUrl,omitempty"`
	LastRolloutID *string    `json:"lastRolloutId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// InstallPlan is the response from POST /api/gateway/install-plans.
// The bootstrap_token is single-use and expires in 15 minutes — treat
// as a secret.
type InstallPlan struct {
	PlanID         string `json:"planId"`
	ExpiresAt      string `json:"expiresAt"`
	BootstrapToken string `json:"bootstrapToken"`
	InstallScript  string `json:"installScript"`
	GatewayID      string `json:"gatewayId"`
	TenantID       string `json:"tenantId"`
}

// InstallPlanRequest is the request body for POST /api/gateway/install-plans.
type InstallPlanRequest struct {
	Name       string  `json:"name"`
	Mode       string  `json:"mode"`               // INGRESS | EGRESS | BOTH
	Template   *string `json:"template,omitempty"` // SOC2 | PCI | HIPAA | GDPR | CUSTOM
	BackendURL *string `json:"backendUrl,omitempty"`
	HostHint   *string `json:"hostHint,omitempty"`
}

// RotateRequest is the body for POST /api/gateways/:id/rotate. The
// confirm field MUST equal "ROTATE" or the server returns 400.
type RotateRequest struct {
	Target         string `json:"target"` // cert | key | plugin-trust-anchor | all
	Confirm        string `json:"confirm"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// RotationResult is the 202 Accepted response from rotate.
type RotationResult struct {
	RotationID     string `json:"rotationId"`
	GatewayID      string `json:"gatewayId"`
	Target         string `json:"target"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotencyKey"`
	Replayed       bool   `json:"replayed"`
	Message        string `json:"message,omitempty"`
}

// GatewayList is the paginated list response.
type GatewayList struct {
	Gateways   []Gateway `json:"gateways"`
	NextCursor *string   `json:"nextCursor,omitempty"`
}

// ListGatewaysOptions narrows a list call.
type ListGatewaysOptions struct {
	Limit  int
	Cursor string
	Status string
}

// GatewaysService exposes /api/gateway/* and /api/gateways/* endpoints.
type GatewaysService struct {
	client *Client
}

// InstallPlan mints a single-use install plan + bootstrap token + script.
func (s *GatewaysService) InstallPlan(
	ctx context.Context,
	req InstallPlanRequest,
) (*InstallPlan, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, &ValidationError{Message: "name is required"}
	}
	if strings.TrimSpace(req.Mode) == "" {
		return nil, &ValidationError{Message: "mode is required"}
	}
	var resp InstallPlan
	if err := s.client.postJSON(ctx, "/api/gateway/install-plans", nil, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// List returns gateways for the caller's tenant.
func (s *GatewaysService) List(
	ctx context.Context,
	opts ListGatewaysOptions,
) (*GatewayList, error) {
	params := map[string]string{}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if strings.TrimSpace(opts.Cursor) != "" {
		params["cursor"] = opts.Cursor
	}
	if strings.TrimSpace(opts.Status) != "" {
		params["status"] = opts.Status
	}
	var resp GatewayList
	if err := s.client.getJSON(ctx, "/api/gateways", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get fetches a single gateway by id.
func (s *GatewaysService) Get(ctx context.Context, id string) (*Gateway, error) {
	if strings.TrimSpace(id) == "" {
		return nil, &ValidationError{Message: "gateway id is required"}
	}
	var g Gateway
	if err := s.client.getJSON(ctx, "/api/gateways/"+id, nil, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// Rotate triggers cert/key/plugin-trust-anchor rotation. Returns 202
// Accepted with a rotationId — reconcile happens on the next agent
// heartbeat (Phase 2 controller orchestrates the actual reload).
func (s *GatewaysService) Rotate(
	ctx context.Context,
	id string,
	req RotateRequest,
) (*RotationResult, error) {
	if strings.TrimSpace(id) == "" {
		return nil, &ValidationError{Message: "gateway id is required"}
	}
	if req.Confirm != "ROTATE" {
		return nil, &ValidationError{
			Message: "confirm must equal \"ROTATE\"",
		}
	}
	var resp RotationResult
	if err := s.client.postJSON(
		ctx,
		fmt.Sprintf("/api/gateways/%s/rotate", id),
		nil,
		req,
		&resp,
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Decommission removes a gateway. Idempotent.
func (s *GatewaysService) Decommission(ctx context.Context, id string) (*Gateway, error) {
	if strings.TrimSpace(id) == "" {
		return nil, &ValidationError{Message: "gateway id is required"}
	}
	// Server requires ?confirm=DECOMMISSION to guard against accidental
	// destructive ops.
	endpoint := fmt.Sprintf("/api/gateways/%s?confirm=DECOMMISSION", id)
	payload, err := s.client.request(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, nil
	}
	var g Gateway
	if jerr := json.Unmarshal(payload, &g); jerr != nil {
		return nil, fmt.Errorf("decode gateway: %w", jerr)
	}
	return &g, nil
}
