package domain

import "testing"

func TestActionRejectsMultiplePayloads(t *testing.T) {
	a := Action{
		Type: ActionRestartWorkload,
		RestartWorkload: &RestartWorkloadAction{
			Namespace: "default",
			Workload:  "api",
		},
		ScaleWorkload: &ScaleWorkloadAction{
			Namespace: "default",
			Workload:  "api",
			Replicas:  3,
		},
	}
	if err := a.Validate(); err == nil {
		t.Fatal("expected multiple semantic payloads to be rejected")
	}
}

func TestActionRejectsMismatchedPayload(t *testing.T) {
	a := Action{
		Type: ActionScaleWorkload,
		RestartWorkload: &RestartWorkloadAction{
			Namespace: "default",
			Workload:  "api",
		},
	}
	if err := a.Validate(); err == nil {
		t.Fatal("expected action type/payload mismatch to be rejected")
	}
}

func TestRollbackDeploymentValidation(t *testing.T) {
	a := Action{
		Type: ActionRollbackDeployment,
		RollbackDeployment: &RollbackDeploymentAction{
			Namespace:    "default",
			Deployment:   "api",
			FromRevision: 2,
			ToRevision:   1,
		},
	}
	if err := a.Validate(); err != nil {
		t.Fatalf("expected valid typed rollback action, got %v", err)
	}
}
