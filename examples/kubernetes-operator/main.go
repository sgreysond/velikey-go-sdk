// Package main demonstrates contract-aligned VeliKey SDK usage in operator automation.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	velikey "github.com/velikey/velikey-go-sdk"
)

func main() {
	ctx := context.Background()

	client := velikey.NewClient(velikey.Config{
		BaseURL:       getenv("AXIS_BASE_URL", "https://axis.velikey.com"),
		APIKey:        os.Getenv("VELIKEY_API_KEY"),
		BearerToken:   os.Getenv("VELIKEY_BEARER_TOKEN"),
		SessionCookie: os.Getenv("AXIS_SESSION_COOKIE"),
	})

	health, err := client.GetHealth(ctx)
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}
	fmt.Printf("health status=%s version=%s\n", health.Status, health.Version)

	agents, err := client.Agents.List(ctx)
	if err != nil {
		log.Printf("agent list unavailable (expected when auth mode lacks agent scope): %v", err)
	} else {
		fmt.Printf("agents discovered=%d\n", len(agents))
	}

	policyID := stringsTrim(os.Getenv("VELIKEY_POLICY_ID"))
	if policyID == "" {
		fmt.Println("VELIKEY_POLICY_ID not set; skipping rollout plan/apply demonstration")
		return
	}

	canaryPercent := 10
	plan, err := client.Rollouts.Plan(ctx, velikey.PlanRolloutRequest{
		PolicyID:      policyID,
		CanaryPercent: &canaryPercent,
		Explain:       true,
	})
	if err != nil {
		log.Fatalf("rollout plan failed: %v", err)
	}
	if !plan.Success || stringsTrim(plan.Data.PlanID) == "" {
		log.Fatalf("rollout plan did not return a usable plan_id")
	}
	fmt.Printf("planned rollout plan_id=%s\n", plan.Data.PlanID)

	apply, err := client.Rollouts.Apply(ctx, velikey.ApplyRolloutRequest{
		PlanID: plan.Data.PlanID,
		DryRun: true,
	})
	if err != nil {
		log.Fatalf("rollout apply dry-run failed: %v", err)
	}
	fmt.Printf("dry-run apply success=%t rollout_id=%s\n", apply.Success, apply.Data.RolloutID)
}

func getenv(key, fallback string) string {
	value := stringsTrim(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func stringsTrim(value string) string {
	if value == "" {
		return ""
	}
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t' || value[end-1] == '\n' || value[end-1] == '\r') {
		end--
	}
	return value[start:end]
}
