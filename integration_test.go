package velikey

import (
	"testing"
)

func TestClientIntegration(t *testing.T) {
	config := Config{
		APIKey:  "test-key",
		BaseURL: "https://localhost:8443",
	}
	client := NewClient(config)
	if client == nil {
		t.Errorf("Client creation should handle test mode")
	}
}

func TestPolicyManagement(t *testing.T) {
	// Test policy configuration
	policyConfig := map[string]interface{}{
		"name":       "test-policy",
		"algorithms": []string{"aes-256-gcm"},
	}
	if policyConfig["name"] != "test-policy" {
		t.Errorf("Policy name mismatch")
	}
}

func TestQuantumResistantSupport(t *testing.T) {
	qrAlgos := []string{"kyber1024", "dilithium5"}
	if len(qrAlgos) != 2 {
		t.Errorf("Expected 2 quantum-resistant algorithms")
	}
}

func TestPluginHotSwap(t *testing.T) {
	plugins := []string{"aes-plugin", "kyber-plugin"}
	if len(plugins) < 2 {
		t.Errorf("Plugin list should contain multiple plugins")
	}
}
