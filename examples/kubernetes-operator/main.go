// Package main demonstrates using VeliKey SDK in a Kubernetes operator
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/velikey/velikey-go-sdk"
)

func main() {
	fmt.Println("🛡️ VeliKey Kubernetes Operator Example")
	fmt.Println(strings.Repeat("=", 60))

	// Initialize VeliKey client
	client := velikey.NewClient(velikey.Config{
		APIKey: os.Getenv("VELIKEY_API_KEY"),
	})

	ctx := context.Background()

	// 1. Validate API connectivity
	fmt.Println("\n1. 🔗 API Connectivity Check")
	health, err := client.GetHealth(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to VeliKey API: %v", err)
	}
	fmt.Printf("✅ API Status: %s\n", health.Status)

	// 2. Deploy security policies for Kubernetes workloads
	fmt.Println("\n2. 📋 Kubernetes Security Policy Deployment")

	// Create policy for microservices
	policyBuilder := client.NewPolicyBuilder().
		Name("Kubernetes Microservices Security").
		Description("Security policy for containerized microservices").
		ComplianceStandard("SOC2 Type II").
		PostQuantumReady().
		EnforcementMode("enforce") // Production enforcement

	policy, err := policyBuilder.Create(ctx)
	if err != nil {
		log.Fatalf("Failed to create policy: %v", err)
	}
	fmt.Printf("✅ Created policy: %s (%s)\n", policy.Name, policy.ID)

	// 3. Configure agents for Kubernetes deployment
	fmt.Println("\n3. 🤖 Agent Configuration for Kubernetes")

	agentConfig := velikey.NewAgentConfigBuilder().
		Namespace("aegis-system").
		Replicas(3).                // High availability
		Resources("200m", "256Mi"). // Production resources
		BackendURL("http://app-service.default.svc.cluster.local:8080").
		Build()

	fmt.Printf("📦 Agent Configuration:\n")
	fmt.Printf("  Namespace: %s\n", agentConfig["namespace"])
	fmt.Printf("  Replicas: %v\n", agentConfig["replicas"])
	fmt.Printf("  Resources: %v\n", agentConfig["resources"])

	// 4. Monitor cluster security posture
	fmt.Println("\n4. 🔐 Cluster Security Monitoring")

	// Start continuous monitoring
	go startSecurityMonitoring(ctx, client)

	// 5. Compliance validation for Kubernetes
	fmt.Println("\n5. ✅ Kubernetes Compliance Validation")

	complianceChecker := client.NewComplianceChecker([]string{"soc2", "pci-dss"})
	failingFrameworks, err := complianceChecker.CheckThresholds(ctx)
	if err != nil {
		log.Printf("Compliance check failed: %v", err)
	} else if len(failingFrameworks) > 0 {
		fmt.Printf("⚠️ Compliance issues in: %v\n", failingFrameworks)
	} else {
		fmt.Println("✅ All compliance frameworks passing")
	}

	// 6. Agent lifecycle management
	fmt.Println("\n6. 🔄 Agent Lifecycle Management")

	agents, err := client.Agents.List(ctx)
	if err != nil {
		log.Printf("Failed to list agents: %v", err)
	} else {
		fmt.Printf("📊 Managing %d agents:\n", len(agents))
		for _, agent := range agents {
			status := "❌"
			if agent.Status == "online" {
				status = "✅"
			}
			fmt.Printf("  %s %s (%s) - %s\n", status, agent.Name, agent.Version, agent.Location)

			// Check if agent needs updates
			if needsUpdate(agent) {
				fmt.Printf("    🔄 Scheduling update for %s\n", agent.Name)
				// In a real implementation, this would trigger an update
				// err := client.Agents.Update(ctx, agent.ID, velikey.UpdateOptions{
				//     Version: "latest",
				//     Strategy: "rolling",
				// })
				log.Printf("Would update agent %s to latest version", agent.Name)
			}
		}
	}

	// 7. Performance optimization for Kubernetes
	fmt.Println("\n7. 📊 Performance Optimization")

	suggestions, err := client.GetOptimizationSuggestions(ctx)
	if err != nil {
		log.Printf("Failed to get optimization suggestions: %v", err)
	} else {
		if len(suggestions.Performance) > 0 {
			fmt.Println("💡 Performance Suggestions:")
			for _, suggestion := range suggestions.Performance {
				fmt.Printf("  • %s\n", suggestion)
			}
		}

		if len(suggestions.Cost) > 0 {
			fmt.Println("💰 Cost Optimization:")
			for _, suggestion := range suggestions.Cost {
				fmt.Printf("  • %s\n", suggestion)
			}
		}
	}

	// 8. Custom Resource Definition (CRD) integration example
	fmt.Println("\n8. ⚙️ Kubernetes CRD Integration")
	fmt.Println("Example AegisPolicy CRD that could be managed by this operator:")
	fmt.Printf(`
apiVersion: security.velikey.com/v1
kind: AegisPolicy
metadata:
  name: microservices-security
  namespace: aegis-system
spec:
  complianceFramework: soc2
  enforcementMode: enforce
  postQuantum: true
  targetWorkloads:
    - namespace: default
      labels:
        app: microservice
status:
  policyId: %s
  deployed: true
  agentsTargeted: %d
`, policy.ID, len(agents))

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Kubernetes operator example complete!")
	fmt.Println("📚 Documentation: https://docs.velikey.com/sdk/go")
	fmt.Println("🔧 Operator Framework: https://docs.velikey.com/kubernetes/operator")

	// Keep monitoring running
	select {}
}

func startSecurityMonitoring(ctx context.Context, client *velikey.Client) {
	fmt.Println("🔍 Starting continuous security monitoring...")

	// Monitor for security alerts
	alertHandler := velikey.AlertHandlerFunc(func(ctx context.Context, alert velikey.SecurityAlert) error {
		emoji := map[string]string{
			"info":      "ℹ️",
			"warning":   "⚠️",
			"error":     "❌",
			"critical":  "🔥",
			"emergency": "🚨",
		}[alert.Severity]

		fmt.Printf("\n🚨 Security Alert: %s %s\n", emoji, alert.Title)
		fmt.Printf("   Description: %s\n", alert.Description)
		fmt.Printf("   Source: %s\n", alert.Source)

		// In a real operator, you might:
		// - Create Kubernetes events
		// - Update CRD status
		// - Trigger remediation workflows
		// - Send notifications to ops teams

		return nil
	})

	if err := client.StartAlertMonitoring(ctx, alertHandler); err != nil {
		log.Printf("Alert monitoring stopped: %v", err)
	}
}

func needsUpdate(agent velikey.Agent) bool {
	// Simple version check - in production, you'd have more sophisticated logic
	return agent.Version != "latest" &&
		time.Since(agent.LastHeartbeat) < 5*time.Minute // Only update healthy agents
}

// Kubernetes operator utilities

func generateHelmValues(agentConfig map[string]interface{}, policy *velikey.Policy) map[string]interface{} {
	return map[string]interface{}{
		"agent": map[string]interface{}{
			"enabled":   true,
			"replicas":  agentConfig["replicas"],
			"resources": agentConfig["resources"],
			"image": map[string]interface{}{
				"repository": "velikey/aegis-agent",
				"tag":        "latest",
			},
		},
		"controlPlane": map[string]interface{}{
			"enabled": true,
			"service": map[string]interface{}{
				"type": "ClusterIP",
			},
		},
		"policies": map[string]interface{}{
			"default": policy.ID,
		},
		"monitoring": map[string]interface{}{
			"enabled": true,
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
	}
}

func generateKubernetesManifests(namespace string, agentConfig map[string]interface{}) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    name: %s
    app.kubernetes.io/name: aegis
    app.kubernetes.io/managed-by: velikey-operator
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aegis-agent
  namespace: %s
  labels:
    app: aegis-agent
spec:
  replicas: %v
  selector:
    matchLabels:
      app: aegis-agent
  template:
    metadata:
      labels:
        app: aegis-agent
    spec:
      containers:
      - name: agent
        image: velikey/aegis-agent:latest
        resources:
          requests:
            cpu: %s
            memory: %s
        ports:
        - containerPort: 8444
          name: tls-proxy
        - containerPort: 9080
          name: health
        env:
        - name: AEGIS_LISTEN
          value: "0.0.0.0:8444"
        - name: AEGIS_HEALTH_ADDR
          value: "0.0.0.0:9080"
`,
		namespace, namespace, namespace,
		agentConfig["replicas"],
		agentConfig["resources"].(map[string]interface{})["cpu"],
		agentConfig["resources"].(map[string]interface{})["memory"],
	)
}
