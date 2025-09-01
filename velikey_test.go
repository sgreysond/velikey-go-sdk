package velikey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		client := NewClient(Config{
			APIKey: "test-api-key",
		})

		assert.Equal(t, "test-api-key", client.apiKey)
		assert.Equal(t, "https://api.velikey.com", client.baseURL)
		assert.Equal(t, "velikey-go-sdk/0.1.0", client.userAgent)
		assert.NotNil(t, client.Agents)
		assert.NotNil(t, client.Policies)
		assert.NotNil(t, client.Monitoring)
	})

	t.Run("custom configuration", func(t *testing.T) {
		client := NewClient(Config{
			APIKey:  "test-key",
			BaseURL: "https://custom.api.com",
			Timeout: 10 * time.Second,
		})

		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
	})
}

func TestClient_request(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication header
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-api-key", auth)

		// Verify user agent
		userAgent := r.Header.Get("User-Agent")
		assert.Contains(t, userAgent, "velikey-go-sdk")

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	resp, err := client.request(ctx, "GET", "/test", nil, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestClient_GetHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "ok",
			"timestamp": "2024-01-01T12:00:00Z",
			"version": "1.0.0"
		}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	health, err := client.GetHealth(ctx)

	require.NoError(t, err)
	assert.Equal(t, "ok", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
}

func TestClient_QuickSetup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/setup/quick", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"policy_id": "policy-123",
			"policy_name": "SOC2 Policy",
			"deployment_instructions": {
				"helm": "helm install aegis velikey/aegis"
			},
			"next_steps": [
				"Deploy agents using Helm",
				"Verify agent connectivity"
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	result, err := client.QuickSetup(ctx, SetupOptions{
		ComplianceFramework: "soc2",
		EnforcementMode:     "observe",
		PostQuantum:         true,
	})

	require.NoError(t, err)
	assert.Equal(t, "policy-123", result.PolicyID)
	assert.Equal(t, "SOC2 Policy", result.PolicyName)
	assert.Len(t, result.NextSteps, 2)
}

func TestPolicyBuilder(t *testing.T) {
	mockAxiosInstance := &http.Client{}
	client := &Client{httpClient: mockAxiosInstance}

	t.Run("build configuration", func(t *testing.T) {
		builder := client.NewPolicyBuilder()
		config := builder.
			ComplianceStandard("SOC2 Type II").
			PostQuantumReady().
			EnforcementMode("enforce").
			Name("Test Policy").
			Description("Test policy description").
			Build()

		assert.Equal(t, "Test Policy", config["name"])
		assert.Equal(t, "Test policy description", config["description"])
		assert.Equal(t, "enforce", config["enforcement_mode"])

		rules := config["rules"].(map[string]interface{})
		assert.Equal(t, "SOC2 Type II", rules["compliance_standard"])

		aegis := rules["aegis"].(map[string]interface{})
		pqReady := aegis["pq_ready"].([]string)
		assert.Contains(t, pqReady, "TLS_KYBER768_P256_SHA256")
	})

	t.Run("validation errors", func(t *testing.T) {
		builder := client.NewPolicyBuilder()

		// Should fail without name
		_, err := builder.Create(context.Background())
		assert.Error(t, err)
		assert.IsType(t, &ValidationError{}, err)
	})
}

func TestCreateFromTemplate(t *testing.T) {
	t.Run("SOC2 template", func(t *testing.T) {
		config := CreateFromTemplate(SOC2TypeII, "Test SOC2 Policy")

		assert.Equal(t, "SOC2 Type II", config["compliance_standard"])

		aegis := config["aegis"].(map[string]interface{})
		assert.NotEmpty(t, aegis["pq_ready"])
		assert.NotEmpty(t, aegis["preferred"])
		assert.NotEmpty(t, aegis["prohibited"])
	})

	t.Run("with options", func(t *testing.T) {
		config := CreateFromTemplate(
			SOC2TypeII,
			"Test Policy",
			WithPostQuantum(),
			WithEnforcementMode("enforce"),
		)

		assert.Equal(t, "enforce", config["enforcement_mode"])
	})
}

func TestComplianceChecker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/compliance/validate":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"compliant": true,
				"score": 95,
				"issues": [],
				"recommendations": []
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	checker := client.NewComplianceChecker([]string{"soc2", "pci-dss"})

	ctx := context.Background()
	results, err := checker.ValidateAll(ctx)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results["soc2"].Compliant)
	assert.Equal(t, 95, results["soc2"].Score)
}

func TestAgentConfigBuilder(t *testing.T) {
	builder := NewAgentConfigBuilder()
	config := builder.
		Namespace("custom-namespace").
		Replicas(3).
		Resources("200m", "512Mi").
		BackendURL("https://backend.example.com").
		Build()

	assert.Equal(t, "custom-namespace", config["namespace"])
	assert.Equal(t, 3, config["replicas"])

	resources := config["resources"].(map[string]interface{})
	assert.Equal(t, "200m", resources["cpu"])
	assert.Equal(t, "512Mi", resources["memory"])

	networking := config["networking"].(map[string]interface{})
	assert.Equal(t, "https://backend.example.com", networking["backend_url"])
}

// Benchmark tests
func BenchmarkClient_request(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.request(ctx, "GET", "/test", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkPolicyBuilder(b *testing.B) {
	client := NewClient(Config{APIKey: "test"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.NewPolicyBuilder().
			ComplianceStandard("SOC2 Type II").
			PostQuantumReady().
			EnforcementMode("enforce").
			Build()
	}
}

// Integration tests (require test server)
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	// Only run if integration test server is available
	testAPIKey := os.Getenv("TEST_VELIKEY_API_KEY")
	if testAPIKey == "" {
		t.Skip("TEST_VELIKEY_API_KEY not set, skipping integration tests")
	}

	client := NewClient(Config{
		APIKey:  testAPIKey,
		BaseURL: "https://api-test.velikey.com",
	})

	ctx := context.Background()

	t.Run("health check", func(t *testing.T) {
		health, err := client.GetHealth(ctx)
		require.NoError(t, err)
		assert.Equal(t, "ok", health.Status)
	})

	t.Run("security status", func(t *testing.T) {
		status, err := client.GetSecurityStatus(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, status.HealthScore, 0)
		assert.LessOrEqual(t, status.HealthScore, 100)
	})
}
