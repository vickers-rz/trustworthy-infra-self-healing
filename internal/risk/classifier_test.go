package risk

import (
	"testing"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

func TestClassifierIgnoresPlannerEstimatedRiskForAuthorization(t *testing.T) {
	estimated := domain.RiskR1
	p := domain.Proposal{
		ID:            "rem-risk",
		IncidentID:    "inc-risk",
		EstimatedRisk: &estimated,
		Action: domain.Action{
			Type: domain.ActionRollbackDeployment,
			RollbackDeployment: &domain.RollbackDeploymentAction{
				Namespace:    "default",
				Deployment:   "api",
				FromRevision: 2,
				ToRevision:   1,
			},
		},
	}

	risk, _, err := NewClassifier().Classify(p, domain.ExecutionContext{})
	if err != nil {
		t.Fatalf("unexpected classification error: %v", err)
	}
	if risk != domain.RiskR2 {
		t.Fatalf("expected action-derived R2, got %s", risk)
	}
}

func TestClassifierFailsClosedForMalformedTypedAction(t *testing.T) {
	p := domain.Proposal{
		ID:         "rem-invalid",
		IncidentID: "inc-invalid",
		Action: domain.Action{
			Type: domain.ActionScaleWorkload,
			ScaleWorkload: &domain.ScaleWorkloadAction{
				Namespace: "default",
				Workload:  "api",
				Replicas:  -1,
			},
		},
	}

	if _, _, err := NewClassifier().Classify(p, domain.ExecutionContext{}); err == nil {
		t.Fatal("expected malformed action to fail classification")
	}
}
