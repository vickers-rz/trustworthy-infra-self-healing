package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/executor"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/policy"
)

func main() {
	estimatedRisk := domain.RiskR1
	observedAt := time.Now().UTC()
	freshUntil := observedAt.Add(2 * time.Minute)
	proposal := domain.Proposal{
		ID:            "rem_demo_001",
		IncidentID:    "inc_demo_001",
		EstimatedRisk: &estimatedRisk,
		Hypothesis:    domain.Hypothesis{Type: "unhealthy_workload", Confidence: 0.91},
		Evidence: []domain.EvidenceRef{
			{
				ID:          "ev_demo_health",
				Kind:        domain.EvidenceMetric,
				URI:         "prometheus://demo-api/health",
				Source:      "prometheus",
				Collector:   "infraheal-demo-adapter/v1",
				Subject:     "k8s://default/deployment/demo-api",
				Summary:     "health signal degraded",
				ObservedAt:  observedAt,
				CollectedAt: observedAt.Add(time.Second),
				FreshUntil:  &freshUntil,
				Trust:       domain.EvidenceTrustHigh,
			},
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
