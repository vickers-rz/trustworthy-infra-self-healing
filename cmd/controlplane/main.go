package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/executor"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/policy"
)

func main() {
	estimatedRisk := domain.RiskR1
	proposal := domain.Proposal{
		ID:            "rem_demo_001",
		IncidentID:    "inc_demo_001",
		EstimatedRisk: &estimatedRisk,
		Hypothesis:    domain.Hypothesis{Type: "unhealthy_workload", Confidence: 0.91},
		Evidence: []domain.EvidenceRef{
			{URI: "metric://prometheus/demo", Summary: "health signal degraded"},
		},
		Action: domain.Action{
			Type: domain.ActionRestartWorkload,
			RestartWorkload: &domain.RestartWorkloadAction{
				Namespace: "default",
				Workload:  "demo-api",
			},
		},
		Preconditions: []string{"target is stateless", "healthy replica remains available"},
		Verification:  []string{"error_rate_5m < 1%", "ready_replicas >= desired_replicas"},
	}

	trustedContext := domain.ExecutionContext{
		Authority:     domain.AuthorityDelegated,
		PolicyVersion: "mvp-v2",
		Environment:   "staging",
		ActorID:       "controlplane:demo",
	}

	guard := policy.NewGuard()
	decision := guard.Evaluate(proposal, trustedContext)
	encoded, _ := json.MarshalIndent(decision, "", "  ")
	fmt.Println(string(encoded))
	if !decision.Allowed {
		return
	}

	result, err := (executor.MockExecutor{}).Execute(context.Background(), proposal, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("dry-run: %s\n", result.Message)
}
