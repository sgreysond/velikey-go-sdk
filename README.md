# VeliKey Go SDK

Go SDK for Axis/Aegis control-plane APIs.

## Install

```bash
go get github.com/velikey/velikey-go-sdk
```

## Runtime/Toolchain Requirement

Security closure for current Go stdlib advisories requires operator/runtime hosts to run:

- Go runtime `>= 1.25.7`

Validate locally before production rollout:

```bash
GO_SDK_MIN_RUNTIME_VERSION=1.25.7 bash scripts/check-go-runtime.sh
```

If the default `go` binary is older than the required baseline, point the gate at a patched toolchain explicitly:

```bash
GO_BINARY=/path/to/go1.25.7 GO_SDK_MIN_RUNTIME_VERSION=1.25.7 bash scripts/check-go-runtime.sh
```

## Authentication Modes

You can authenticate with one or more modes depending on endpoint contract:

- `APIKey`: for API-key-compatible routes (for example `/api/agents`).
- `BearerToken`: for bearer-token routes (for example agent policy fetch routes).
- `SessionCookie` (or `SessionToken`): for user-session routes (rollouts, usage, alerts).

```go
client := velikey.NewClient(velikey.Config{
    BaseURL:       "https://axis.velikey.com",
    APIKey:        os.Getenv("VELIKEY_API_KEY"),
    BearerToken:   os.Getenv("VELIKEY_BEARER_TOKEN"),
    SessionCookie: os.Getenv("AXIS_SESSION_COOKIE"),
})
```

## Minimal Health Check

```go
ctx := context.Background()
health, err := client.GetHealth(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("status=%s version=%s\n", health.Status, health.Version)
```

## Production Flow: Canary Rollout + Rollback

Prerequisites:

- Valid Axis session cookie with rollout permissions.
- Existing `policyId` in the target tenant.
- Non-production validation should use `DryRun=true` first.

```go
canary := 10
plan, err := client.Rollouts.Plan(ctx, velikey.PlanRolloutRequest{
    PolicyID:      policyID,
    CanaryPercent: &canary,
    Explain:       true,
})
if err != nil {
    log.Fatal(err)
}

apply, err := client.Rollouts.Apply(ctx, velikey.ApplyRolloutRequest{
    PlanID: plan.Data.PlanID,
    DryRun: true,
})
if err != nil {
    log.Fatal(err)
}

// For live rollouts only (manual operator confirmation and change window required):
// live, _ := client.Rollouts.Apply(ctx, velikey.ApplyRolloutRequest{PlanID: plan.Data.PlanID, DryRun: false})
// rollback, _ := client.Rollouts.Rollback(ctx, velikey.RollbackRolloutRequest{RollbackToken: live.Data.RollbackToken})
```

## Metering Readback

```go
usage, err := client.Usage.Get(ctx, "current")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("telemetry_gb=%.4f estimated_cost=%.2f\n", usage.Current.TelemetryGB, usage.Current.EstimatedCost)
```

## Alerts and Telemetry

```go
_, _ = client.Telemetry.Ingest(ctx, velikey.TelemetryIngestRequest{
    Event: "sdk.go.validation",
    Properties: map[string]interface{}{
        "lane": "go",
    },
})

activeAlerts, _ := client.Alerts.ListWithOptions(ctx, velikey.ListAlertsOptions{Resolved: boolPtr(false)})
fmt.Printf("active_alerts=%d\n", len(activeAlerts))
```

## Retries and Timeouts

The client retries transient failures (`429`, `5xx`, and network failures) with bounded exponential backoff.

```go
client := velikey.NewClient(velikey.Config{
    Timeout:         20 * time.Second,
    MaxRetries:      3,
    RetryMinBackoff: 250 * time.Millisecond,
    RetryMaxBackoff: 2 * time.Second,
})
```

## Unsupported Legacy Methods

Legacy methods that do not map to current Axis public routes return `UnsupportedOperationError`:

- `QuickSetup`
- `ValidateCompliance`
- `GetOptimizationSuggestions`
- `BulkPolicyUpdate`
- `PolicyBuilder.Create`

## Development

```bash
go mod tidy
go build ./...
go test ./...
GO_SDK_MIN_RUNTIME_VERSION=1.25.7 bash scripts/check-go-runtime.sh
GO_BINARY=/path/to/go1.25.7 GO_SDK_MIN_RUNTIME_VERSION=1.25.7 bash scripts/check-go-runtime.sh
```

```go
func boolPtr(v bool) *bool { return &v }
```
