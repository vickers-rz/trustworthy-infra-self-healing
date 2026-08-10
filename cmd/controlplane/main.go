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
	decisionTime := time.Now().UTC()
	observedAt := decisionTime.Add(-30 * time.Second)
	freshUntil := decisionTime.Add(90 * time.Second)

	bundle := domain.EvidenceBundle{
		ID:         "bundle_demo_001",
		IncidentID: "inc_demo_001",
		CreatedAt:  observedAt.Add(time.Second),
		Items: []domain.EvidenceRef{
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
		Relations: []domain.EvidenceRelation{
			{EvidenceID: "ev_demo_health", ClaimID: "hyp_demo_unhealthy", Type: domain.EvidenceSupports},
		},
	}

	proposal := domain.Proposal{
		ID:               "rem_demo_001",
		IncidentID:       "inc_demo_001",
		EstimatedRisk:    &estimatedRisk,
		Hypothesis:       domain.Hypothesis{ID: "hyp_demo_unhealthy", Type: "unhealthy_workload", Confidence: 0.91},
		EvidenceBundleID: bundle.ID,
		EvidenceIDs:      []string{"ev_demo_health"},
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
		PolicyVersion: "mvp-v3",
		Environment:   "staging",
		ActorID:       "controlplane:demo",
		DecisionTime:  decisionTime,
	}

	guard := policy.NewGuard()
	decision := guard.Evaluate(proposal, trustedContext, bundle)
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
