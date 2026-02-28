package velikey

import (
	"context"
	"os"
	"testing"
)

func TestClientIntegrationSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := os.Getenv("TEST_AXIS_BASE_URL")
	if baseURL == "" {
		t.Skip("TEST_AXIS_BASE_URL not set")
	}

	cfg := Config{
		BaseURL:       baseURL,
		SessionCookie: os.Getenv("TEST_AXIS_SESSION_COOKIE"),
		APIKey:        os.Getenv("TEST_AXIS_API_KEY"),
		BearerToken:   os.Getenv("TEST_AXIS_BEARER_TOKEN"),
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatalf("expected non-nil client")
	}

	_, _ = client.GetHealth(context.Background())
}
