package velikey

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientDefaults(t *testing.T) {
	client := NewClient(Config{APIKey: "test-key"})

	assert.Equal(t, "https://axis.velikey.com", client.baseURL)
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, defaultUserAgent, client.userAgent)
	assert.Equal(t, defaultMaxRetries, client.maxRetries)
	assert.NotNil(t, client.Agents)
	assert.NotNil(t, client.Policies)
	assert.NotNil(t, client.Rollouts)
}

func TestRequestUsesBearerAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{APIKey: "test-api-key", BaseURL: server.URL})
	_, err := client.request(context.Background(), http.MethodGet, "/api/test", nil, nil)
	require.NoError(t, err)
}

func TestRequestUsesSessionCookie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "next-auth.session-token=session-token-value", r.Header.Get("Cookie"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL:         server.URL,
		SessionToken:    "session-token-value",
		MaxRetries:      1,
		RetryMinBackoff: 1 * time.Millisecond,
		RetryMaxBackoff: 1 * time.Millisecond,
	})

	_, err := client.request(context.Background(), http.MethodGet, "/api/test", nil, nil)
	require.NoError(t, err)
}

func TestRequestRetriesOnTransientErrors(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"temporary upstream failure"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL:         server.URL,
		APIKey:          "test-key",
		MaxRetries:      3,
		RetryMinBackoff: 1 * time.Millisecond,
		RetryMaxBackoff: 2 * time.Millisecond,
	})

	_, err := client.request(context.Background(), http.MethodGet, "/api/retry", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestGetHealthUsesAxisRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","timestamp":"2026-02-28T00:00:00Z","version":"1.0.0"}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	health, err := client.GetHealth(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
}

func TestAgentsListParsesEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agents", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"agents":[{"id":"1","agentId":"agent-1","name":"Agent One","status":"active"}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIKey: "vk_test"})
	agents, err := client.Agents.List(context.Background())
	require.NoError(t, err)
	require.Len(t, agents, 1)
	assert.Equal(t, "agent-1", agents[0].AgentID)
}

func TestPoliciesListForAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agents/agent-123/policies", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"agentId":"agent-123","policies":[{"id":"p1","name":"Transit Policy","isActive":true}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, BearerToken: "bootstrap-token"})
	resp, err := client.Policies.ListForAgent(context.Background(), "agent-123")
	require.NoError(t, err)
	assert.Equal(t, "agent-123", resp.AgentID)
	require.Len(t, resp.Policies, 1)
	assert.Equal(t, "p1", resp.Policies[0].ID)
}

func TestUsageGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/usage", r.URL.Path)
		assert.Equal(t, "current", r.URL.Query().Get("period"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"current":{"period":"February 2026","dayOfPeriod":28,"daysInPeriod":28,"encryptionGB":12,"telemetryGB":1,"environments":2,"agents":3,"estimatedCost":1001},"usage":{"period":"February 2026","dayOfPeriod":28,"daysInPeriod":28,"encryptionGB":12,"telemetryGB":1,"environments":2,"agents":3,"estimatedCost":1001},"historical":[]}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, SessionCookie: "next-auth.session-token=test"})
	usage, err := client.Usage.Get(context.Background(), "current")
	require.NoError(t, err)
	assert.Equal(t, 3, usage.Current.Agents)
}

func TestRolloutsApplyAutoConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/rollouts/apply", r.URL.Path)
		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, true, payload["confirm"])
		assert.Equal(t, "APPLY", payload["confirmation"])
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"rollout_id":"rollout-1","rollback_token":"rb-token"}}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, SessionCookie: "next-auth.session-token=test"})
	resp, err := client.Rollouts.Apply(context.Background(), ApplyRolloutRequest{PlanID: "plan-1", DryRun: false})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "rollout-1", resp.Data.RolloutID)
}

func TestUnsupportedOperations(t *testing.T) {
	client := NewClient(Config{APIKey: "test"})

	_, err := client.ValidateCompliance(context.Background(), "soc2")
	var unsupported *UnsupportedOperationError
	assert.True(t, errors.As(err, &unsupported))
	assert.Equal(t, "ValidateCompliance", unsupported.Method)
}
