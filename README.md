# VeliKey Aegis Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/velikey/velikey-go-sdk.svg)](https://pkg.go.dev/github.com/velikey/velikey-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/velikey/velikey-go-sdk)](https://goreportcard.com/report/github.com/velikey/velikey-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Quantum-safe crypto policy management for Go applications and infrastructure**

The VeliKey Go SDK provides high-performance, type-safe integration with VeliKey Aegis for infrastructure automation, Kubernetes operators, and system integrations.

## 🚀 Installation

```bash
go get github.com/velikey/velikey-go-sdk
```

## ⚡ Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/velikey/velikey-go-sdk"
)

func main() {
    // Initialize client
    client := velikey.NewClient(velikey.Config{
        APIKey: os.Getenv("VELIKEY_API_KEY"),
    })
    
    ctx := context.Background()
    
    // Quick setup
    setup, err := client.QuickSetup(ctx, velikey.SetupOptions{
        ComplianceFramework: "soc2",
        EnforcementMode:     "observe",
        PostQuantum:         true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✅ Created policy: %s\n", setup.PolicyName)
    
    // Monitor security status
    status, err := client.GetSecurityStatus(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("🛡️ Health Score: %d/100\n", status.HealthScore)
    fmt.Printf("🤖 Agents Online: %s\n", status.AgentsOnline)
}
```

## 🏗️ Infrastructure Automation

### Terraform Provider Integration

```go
// Custom Terraform provider using VeliKey SDK
package main

import (
    "context"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/velikey/velikey-go-sdk"
)

func resourceAegisPolicy() *schema.Resource {
    return &schema.Resource{
        Create: resourceAegisCreatePolicy,
        Read:   resourceAegisReadPolicy,
        Update: resourceAegisUpdatePolicy,
        Delete: resourceAegisDeletePolicy,
        
        Schema: map[string]*schema.Schema{
            "name": {
                Type:     schema.TypeString,
                Required: true,
            },
            "compliance_framework": {
                Type:     schema.TypeString,
                Required: true,
            },
            "enforcement_mode": {
                Type:     schema.TypeString,
                Optional: true,
                Default:  "observe",
            },
        },
    }
}

func resourceAegisCreatePolicy(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*velikey.Client)
    ctx := context.Background()
    
    policy, err := client.Policies.Create(ctx, velikey.CreatePolicyRequest{
        Name:            d.Get("name").(string),
        ComplianceFramework: d.Get("compliance_framework").(string),
        EnforcementMode: d.Get("enforcement_mode").(string),
    })
    if err != nil {
        return err
    }
    
    d.SetId(policy.ID)
    return nil
}
```

### Kubernetes Operator

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/velikey/velikey-go-sdk"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/source"
)

// AegisPolicyReconciler reconciles AegisPolicy objects
type AegisPolicyReconciler struct {
    client.Client
    VeliKey *velikey.Client
}

func (r *AegisPolicyReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
    // Fetch the AegisPolicy instance
    var aegisPolicy AegisPolicy
    if err := r.Get(ctx, req.NamespacedName, &aegisPolicy); err != nil {
        return reconcile.Result{}, client.IgnoreNotFound(err)
    }
    
    // Create or update policy in VeliKey
    policyBuilder := r.VeliKey.NewPolicyBuilder().
        Name(aegisPolicy.Spec.Name).
        ComplianceStandard(aegisPolicy.Spec.ComplianceFramework).
        EnforcementMode(aegisPolicy.Spec.EnforcementMode)
    
    if aegisPolicy.Spec.PostQuantum {
        policyBuilder = policyBuilder.PostQuantumReady()
    }
    
    policy, err := policyBuilder.Create(ctx)
    if err != nil {
        return reconcile.Result{}, err
    }
    
    // Update status
    aegisPolicy.Status.PolicyID = policy.ID
    aegisPolicy.Status.Deployed = true
    aegisPolicy.Status.LastUpdated = time.Now()
    
    if err := r.Status().Update(ctx, &aegisPolicy); err != nil {
        return reconcile.Result{}, err
    }
    
    return reconcile.Result{RequeueAfter: time.Minute * 5}, nil
}

// Custom Resource Definition
type AegisPolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    
    Spec   AegisPolicySpec   `json:"spec,omitempty"`
    Status AegisPolicyStatus `json:"status,omitempty"`
}

type AegisPolicySpec struct {
    Name                string   `json:"name"`
    ComplianceFramework string   `json:"complianceFramework"`
    EnforcementMode     string   `json:"enforcementMode"`
    PostQuantum         bool     `json:"postQuantum"`
    TargetWorkloads     []string `json:"targetWorkloads"`
}

type AegisPolicyStatus struct {
    PolicyID      string      `json:"policyId"`
    Deployed      bool        `json:"deployed"`
    AgentsTargeted int        `json:"agentsTargeted"`
    LastUpdated   time.Time   `json:"lastUpdated"`
}
```

## 📊 Monitoring & Observability

### Prometheus Integration

```go
package main

import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/velikey/velikey-go-sdk"
)

var (
    healthScoreGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "velikey_customer_health_score",
        Help: "Customer health score from VeliKey",
    })
    
    agentsOnlineGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "velikey_agents_online_total",
        Help: "Number of online VeliKey agents",
    })
    
    policyViolationsCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "velikey_policy_violations_total",
        Help: "Total number of policy violations detected",
    })
)

func startMetricsCollection(client *velikey.Client) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        ctx := context.Background()
        
        // Collect security status
        status, err := client.GetSecurityStatus(ctx)
        if err != nil {
            continue
        }
        
        healthScoreGauge.Set(float64(status.HealthScore))
        
        // Parse agents online (format: "5/10")
        var online, total int
        fmt.Sscanf(status.AgentsOnline, "%d/%d", &online, &total)
        agentsOnlineGauge.Set(float64(online))
        
        // Collect alerts for violations
        alerts, err := client.Monitoring.GetActiveAlerts(ctx)
        if err != nil {
            continue
        }
        
        for _, alert := range alerts {
            if alert.Category == "policy_violation" {
                policyViolationsCounter.Inc()
            }
        }
    }
}
```

### Custom Alerting

```go
// Custom alert handler for integration with existing systems
type CustomAlertHandler struct {
    SlackWebhook   string
    PagerDutyKey   string
    EmailSMTP      SMTPConfig
}

func (h *CustomAlertHandler) HandleAlert(ctx context.Context, alert velikey.SecurityAlert) error {
    switch alert.Severity {
    case "critical", "emergency":
        // Page on-call engineer
        return h.sendPagerDutyAlert(alert)
        
    case "error":
        // Send to Slack
        return h.sendSlackAlert(alert)
        
    case "warning":
        // Email security team
        return h.sendEmailAlert(alert)
        
    default:
        // Log only
        log.Printf("Security alert: %s - %s", alert.Title, alert.Description)
    }
    
    return nil
}

// Start monitoring with custom handler
handler := &CustomAlertHandler{
    SlackWebhook: os.Getenv("SLACK_WEBHOOK_URL"),
    PagerDutyKey: os.Getenv("PAGERDUTY_INTEGRATION_KEY"),
}

go client.StartAlertMonitoring(ctx, handler)
```

## 🔧 Advanced Features

### Rate Limiting & Retry Logic

```go
import "golang.org/x/time/rate"

client := velikey.NewClient(velikey.Config{
    APIKey:    os.Getenv("VELIKEY_API_KEY"),
    RateLimit: rate.Limit(10), // 10 requests per second
    Timeout:   30 * time.Second,
})

// Automatic retries with exponential backoff
ctx := context.Background()
agents, err := client.Agents.ListWithRetry(ctx, 3) // 3 retry attempts
```

### Concurrent Operations

```go
// Concurrent policy deployment
func deployPoliciesAsync(client *velikey.Client, policyIDs []string) error {
    ctx := context.Background()
    
    errCh := make(chan error, len(policyIDs))
    
    for _, policyID := range policyIDs {
        go func(id string) {
            err := client.Policies.Deploy(ctx, id, nil)
            errCh <- err
        }(policyID)
    }
    
    // Wait for all deployments
    for i := 0; i < len(policyIDs); i++ {
        if err := <-errCh; err != nil {
            return fmt.Errorf("policy deployment failed: %w", err)
        }
    }
    
    return nil
}
```

## 🧪 Testing

```go
// tests/client_test.go
package velikey_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/velikey/velikey-go-sdk"
)

func TestClient_GetSecurityStatus(t *testing.T) {
    client := velikey.NewClient(velikey.Config{
        APIKey:  "test-api-key",
        BaseURL: "https://api-test.velikey.com",
    })
    
    ctx := context.Background()
    status, err := client.GetSecurityStatus(ctx)
    
    assert.NoError(t, err)
    assert.NotNil(t, status)
    assert.GreaterOrEqual(t, status.HealthScore, 0)
    assert.LessOrEqual(t, status.HealthScore, 100)
}

func TestPolicyBuilder(t *testing.T) {
    client := velikey.NewClient(velikey.Config{APIKey: "test"})
    
    config := client.NewPolicyBuilder().
        ComplianceStandard("SOC2 Type II").
        PostQuantumReady().
        EnforcementMode("enforce").
        Build()
    
    assert.Equal(t, "SOC2 Type II", config["rules"].(map[string]interface{})["compliance_standard"])
    assert.Equal(t, "enforce", config["enforcement_mode"])
}
```

## 📖 Documentation

- **API Reference**: [pkg.go.dev/github.com/velikey/velikey-go-sdk](https://pkg.go.dev/github.com/velikey/velikey-go-sdk)
- **Examples**: [github.com/velikey/velikey-go-sdk/examples](https://github.com/velikey/velikey-go-sdk/tree/main/examples)
- **Kubernetes Guide**: [docs.velikey.com/sdk/go/kubernetes](https://docs.velikey.com/sdk/go/kubernetes)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.
