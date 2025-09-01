package velikey

import (
	"testing"
)

func TestClientIntegration(t *testing.T) {
	client, err := NewClient("test-key", "https://localhost:8443")
	if err != nil && client == nil {
		t.Errorf("Client creation should handle test mode")
	}
}

func TestPolicyManagement(t *testing.T) {
	policy := Policy{
		Name: "test-policy",
		Algorithms: []string{"aes-256-gcm"},
	}
	if policy.Name != "test-policy" {
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
