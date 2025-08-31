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

### Custom Alerting & Incident Response

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
    "time"
    
    "github.com/velikey/velikey-go-sdk"
)

// IncidentManager handles security incidents with automated response
type IncidentManager struct {
    client          *velikey.Client
    slackWebhook    string
    pagerDutyKey    string
    escalationRules map[string]EscalationRule
    responseActions map[string]ResponseAction
}

type EscalationRule struct {
    Severity        string
    TimeToEscalate  time.Duration
    EscalationLevel int
    Recipients      []string
}

type ResponseAction struct {
    Trigger     string
    Action      func(ctx context.Context, alert velikey.SecurityAlert) error
    Description string
    Automatic   bool
}

// AlertContext provides additional context for alert handling
type AlertContext struct {
    Alert          velikey.SecurityAlert
    Agent          *velikey.Agent
    Policy         *velikey.Policy
    HistoricalData map[string]interface{}
}

func NewIncidentManager(client *velikey.Client) *IncidentManager {
    im := &IncidentManager{
        client:          client,
        slackWebhook:    os.Getenv("SLACK_WEBHOOK_URL"),
        pagerDutyKey:    os.Getenv("PAGERDUTY_INTEGRATION_KEY"),
        escalationRules: make(map[string]EscalationRule),
        responseActions: make(map[string]ResponseAction),
    }
    
    // Configure escalation rules
    im.escalationRules["critical"] = EscalationRule{
        Severity:        "critical",
        TimeToEscalate:  5 * time.Minute,
        EscalationLevel: 1,
        Recipients:      []string{"security-team", "on-call-engineer"},
    }
    
    im.escalationRules["high"] = EscalationRule{
        Severity:        "high", 
        TimeToEscalate:  15 * time.Minute,
        EscalationLevel: 2,
        Recipients:      []string{"security-team"},
    }
    
    // Configure automated response actions
    im.responseActions["policy_violation"] = ResponseAction{
        Trigger:     "policy_violation",
        Action:      im.handlePolicyViolation,
        Description: "Automatic policy violation response",
        Automatic:   true,
    }
    
    im.responseActions["agent_compromise"] = ResponseAction{
        Trigger:     "agent_compromise",
        Action:      im.handleAgentCompromise,
        Description: "Automatic agent isolation",
        Automatic:   true,
    }
    
    im.responseActions["compliance_failure"] = ResponseAction{
        Trigger:     "compliance_failure",
        Action:      im.handleComplianceFailure,
        Description: "Compliance failure notification",
        Automatic:   false,
    }
    
    return im
}

// StartMonitoring begins continuous security monitoring
func (im *IncidentManager) StartMonitoring(ctx context.Context) error {
    // Subscribe to real-time alerts
    alertChan, err := im.client.Monitoring.StreamAlerts(ctx)
    if err != nil {
        return fmt.Errorf("failed to start alert stream: %w", err)
    }
    
    // Start alert processing goroutine
    go im.processAlerts(ctx, alertChan)
    
    // Start periodic health checks
    go im.runPeriodicHealthChecks(ctx)
    
    // Start compliance monitoring
    go im.runComplianceMonitoring(ctx)
    
    log.Println("Incident management system started")
    return nil
}

// processAlerts handles incoming security alerts
func (im *IncidentManager) processAlerts(ctx context.Context, alertChan <-chan velikey.SecurityAlert) {
    for {
        select {
        case alert := <-alertChan:
            go im.handleAlert(ctx, alert)
        case <-ctx.Done():
            log.Println("Alert processing stopped")
            return
        }
    }
}

// handleAlert processes individual alerts with context enrichment
func (im *IncidentManager) handleAlert(ctx context.Context, alert velikey.SecurityAlert) {
    // Enrich alert with context
    alertCtx, err := im.enrichAlertContext(ctx, alert)
    if err != nil {
        log.Printf("Failed to enrich alert context: %v", err)
        alertCtx = &AlertContext{Alert: alert}
    }
    
    // Execute automated responses if configured
    if action, exists := im.responseActions[alert.Category]; exists && action.Automatic {
        log.Printf("Executing automated response for %s: %s", alert.Category, action.Description)
        if err := action.Action(ctx, alert); err != nil {
            log.Printf("Automated response failed: %v", err)
        }
    }
    
    // Send notifications based on severity
    switch alert.Severity {
    case "critical", "emergency":
        im.sendCriticalAlert(alertCtx)
        im.createIncident(ctx, alertCtx)
        
    case "high":
        im.sendHighPriorityAlert(alertCtx)
        
    case "medium":
        im.sendMediumPriorityAlert(alertCtx)
        
    case "low", "info":
        im.logAlert(alertCtx)
    }
    
    // Start escalation timer for unacknowledged critical alerts
    if alert.Severity == "critical" {
        go im.startEscalationTimer(ctx, alert)
    }
}

// enrichAlertContext adds additional context to alerts
func (im *IncidentManager) enrichAlertContext(ctx context.Context, alert velikey.SecurityAlert) (*AlertContext, error) {
    alertCtx := &AlertContext{
        Alert:          alert,
        HistoricalData: make(map[string]interface{}),
    }
    
    // Get agent information if available
    if alert.AgentID != "" {
        agent, err := im.client.Agents.Get(ctx, alert.AgentID)
        if err != nil {
            log.Printf("Failed to get agent info: %v", err)
        } else {
            alertCtx.Agent = agent
            
            // Get historical data for this agent
            metrics, err := im.client.Monitoring.GetAgentMetrics(ctx, alert.AgentID, 24*time.Hour)
            if err == nil {
                alertCtx.HistoricalData["metrics"] = metrics
            }
        }
    }
    
    // Get policy information if relevant
    if alert.PolicyID != "" {
        policy, err := im.client.Policies.Get(ctx, alert.PolicyID)
        if err != nil {
            log.Printf("Failed to get policy info: %v", err)
        } else {
            alertCtx.Policy = policy
        }
    }
    
    return alertCtx, nil
}

// handlePolicyViolation implements automated policy violation response
func (im *IncidentManager) handlePolicyViolation(ctx context.Context, alert velikey.SecurityAlert) error {
    log.Printf("Handling policy violation: %s", alert.Title)
    
    // Get agent that violated policy
    if alert.AgentID == "" {
        return fmt.Errorf("no agent ID in policy violation alert")
    }
    
    agent, err := im.client.Agents.Get(ctx, alert.AgentID)
    if err != nil {
        return fmt.Errorf("failed to get agent: %w", err)
    }
    
    // Temporarily isolate agent if violation is severe
    if alert.Severity == "critical" {
        log.Printf("Isolating agent %s due to critical policy violation", agent.Name)
        
        err = im.client.Agents.Isolate(ctx, alert.AgentID, "Automated isolation due to policy violation")
        if err != nil {
            return fmt.Errorf("failed to isolate agent: %w", err)
        }
        
        // Schedule review in 1 hour
        time.AfterFunc(time.Hour, func() {
            im.scheduleAgentReview(alert.AgentID, "Policy violation isolation review")
        })
    }
    
    // Force policy refresh on agent
    err = im.client.Agents.RefreshPolicies(ctx, alert.AgentID)
    if err != nil {
        log.Printf("Failed to refresh policies on agent %s: %v", agent.Name, err)
    }
    
    // Create detailed incident report
    incident := map[string]interface{}{
        "type":        "policy_violation",
        "agent_id":    alert.AgentID,
        "agent_name":  agent.Name,
        "policy_id":   alert.PolicyID,
        "severity":    alert.Severity,
        "description": alert.Description,
        "timestamp":   time.Now(),
        "automated_response": map[string]interface{}{
            "isolation_applied": alert.Severity == "critical",
            "policies_refreshed": true,
        },
    }
    
    return im.storeIncidentReport(incident)
}

// handleAgentCompromise implements automated response to compromised agents
func (im *IncidentManager) handleAgentCompromise(ctx context.Context, alert velikey.SecurityAlert) error {
    log.Printf("Handling potential agent compromise: %s", alert.Title)
    
    if alert.AgentID == "" {
        return fmt.Errorf("no agent ID in compromise alert")
    }
    
    // Immediately isolate compromised agent
    err := im.client.Agents.Isolate(ctx, alert.AgentID, "Automated isolation due to suspected compromise")
    if err != nil {
        return fmt.Errorf("failed to isolate agent: %w", err)
    }
    
    // Revoke agent certificates
    err = im.client.Agents.RevokeCertificates(ctx, alert.AgentID)
    if err != nil {
        log.Printf("Failed to revoke certificates for agent %s: %v", alert.AgentID, err)
    }
    
    // Collect forensic data
    forensicData, err := im.client.Agents.CollectForensicData(ctx, alert.AgentID)
    if err != nil {
        log.Printf("Failed to collect forensic data: %v", err)
    } else {
        im.storeForensicEvidence(alert.AgentID, forensicData)
    }
    
    // Notify security team immediately
    message := fmt.Sprintf("🚨 CRITICAL: Agent %s has been automatically isolated due to suspected compromise. Forensic data collection initiated.", alert.AgentID)
    im.sendSlackAlert(message, "critical")
    
    return nil
}

// runPeriodicHealthChecks performs regular system health monitoring
func (im *IncidentManager) runPeriodicHealthChecks(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            im.performHealthCheck(ctx)
        case <-ctx.Done():
            return
        }
    }
}

// performHealthCheck checks overall system health
func (im *IncidentManager) performHealthCheck(ctx context.Context) {
    status, err := im.client.GetSecurityStatus(ctx)
    if err != nil {
        log.Printf("Health check failed: %v", err)
        return
    }
    
    // Check for degraded health score
    if status.HealthScore < 80 {
        im.sendSlackAlert(fmt.Sprintf("⚠️ System health score degraded: %d/100", status.HealthScore), "warning")
    }
    
    // Check for offline agents
    agents, err := im.client.Agents.List(ctx, &velikey.ListOptions{Filter: "status=offline"})
    if err != nil {
        log.Printf("Failed to get offline agents: %v", err)
        return
    }
    
    if len(agents) > 5 { // More than 5 agents offline
        agentNames := make([]string, 0, len(agents))
        for _, agent := range agents {
            agentNames = append(agentNames, agent.Name)
        }
        
        message := fmt.Sprintf("⚠️ %d agents are offline: %s", len(agents), strings.Join(agentNames, ", "))
        im.sendSlackAlert(message, "warning")
    }
}

// runComplianceMonitoring performs periodic compliance checks
func (im *IncidentManager) runComplianceMonitoring(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    frameworks := []string{"soc2", "pci-dss", "hipaa"}
    
    for {
        select {
        case <-ticker.C:
            for _, framework := range frameworks {
                compliance, err := im.client.Compliance.ValidateFramework(ctx, framework)
                if err != nil {
                    log.Printf("Compliance check failed for %s: %v", framework, err)
                    continue
                }
                
                if !compliance.Compliant {
                    message := fmt.Sprintf("🚨 %s compliance check failed. Issues: %d", 
                        strings.ToUpper(framework), len(compliance.Issues))
                    im.sendSlackAlert(message, "high")
                }
            }
        case <-ctx.Done():
            return
        }
    }
}

// sendCriticalAlert handles critical alert notifications
func (im *IncidentManager) sendCriticalAlert(alertCtx *AlertContext) {
    // Send to Slack
    message := fmt.Sprintf("🚨 CRITICAL ALERT: %s\n%s", alertCtx.Alert.Title, alertCtx.Alert.Description)
    if alertCtx.Agent != nil {
        message += fmt.Sprintf("\nAgent: %s (%s)", alertCtx.Agent.Name, alertCtx.Agent.ID)
    }
    im.sendSlackAlert(message, "critical")
    
    // Page on-call engineer
    im.sendPagerDutyAlert(alertCtx.Alert)
    
    // Send email to security team
    im.sendEmailAlert(alertCtx.Alert, []string{"security-team@company.com"})
}

// Utility functions for external integrations
func (im *IncidentManager) sendSlackAlert(message, severity string) {
    if im.slackWebhook == "" {
        return
    }
    
    color := "good"
    switch severity {
    case "critical":
        color = "danger"
    case "warning", "high":
        color = "warning"
    }
    
    payload := map[string]interface{}{
        "text": message,
        "attachments": []map[string]interface{}{
            {
                "color": color,
                "fields": []map[string]interface{}{
                    {"title": "Severity", "value": severity, "short": true},
                    {"title": "Timestamp", "value": time.Now().Format(time.RFC3339), "short": true},
                },
            },
        },
    }
    
    // Send webhook (implementation details omitted for brevity)
    im.sendWebhook(im.slackWebhook, payload)
}

// Example usage in main application
func main() {
    ctx := context.Background()
    
    client := velikey.NewClient(velikey.Config{
        APIKey: os.Getenv("VELIKEY_API_KEY"),
    })
    
    // Create and start incident manager
    im := NewIncidentManager(client)
    
    if err := im.StartMonitoring(ctx); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }
    
    // Keep the application running
    select {}
}
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
