package evidence

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrahealv1alpha1 "github.com/vickers-rz/trustworthy-infra-self-healing/api/v1alpha1"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

func TestBuildHealingPolicyStatusEvidenceIsDeterministic(t *testing.T) {
	policy, status, observed := evidenceFixture()
	first, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	second, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}

	if first.ID != second.ID || first.DigestSHA256 != second.DigestSHA256 {
		t.Fatalf("same snapshot produced different identities: first=%#v second=%#v", first, second)
	}
	if !strings.HasPrefix(first.ID, "ev_k8s_") || len(first.DigestSHA256) != 64 {
		t.Fatalf("unexpected evidence identity: %#v", first)
	}
	if first.Kind != domain.EvidenceKubernetesState || first.Source != "kubernetes-apiserver" || first.Collector != kubernetesObserverCollector {
		t.Fatalf("unexpected provenance: %#v", first)
	}
	if first.Subject != "k8s://default/deployment/demo-api" {
		t.Fatalf("unexpected subject: %s", first.Subject)
	}
	if first.FreshUntil == nil || !first.FreshUntil.Equal(observed.Add(15*time.Second)) {
		t.Fatalf("unexpected freshness deadline: %v", first.FreshUntil)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf("built evidence did not satisfy domain validation: %v", err)
	}
}

func TestBuildHealingPolicyStatusEvidenceChangesWhenObservedStateChanges(t *testing.T) {
	policy, status, observed := evidenceFixture()
	first, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatal(err)
	}
	status.AvailableReplicas = 0
	status.ReadyReplicas = 0
	for i := range status.Conditions {
		if status.Conditions[i].Type == infrahealv1alpha1.ConditionObservedHealthy {
			status.Conditions[i].Status = metav1.ConditionFalse
			status.Conditions[i].Reason = "Unavailable"
		}
	}
	second, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatal(err)
	}
	if first.DigestSHA256 == second.DigestSHA256 || first.ID == second.ID {
		t.Fatalf("state change must change evidence identity")
	}
}

func TestBuildHealingPolicyStatusEvidenceUsesExplicitTargetNamespace(t *testing.T) {
	policy, status, observed := evidenceFixture()
	policy.Namespace = "ops"
	policy.Spec.Target.Namespace = "apps"
	item, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatal(err)
	}
	if item.Subject != "k8s://apps/deployment/demo-api" {
		t.Fatalf("unexpected explicit target subject: %s", item.Subject)
	}
}

func TestBuildHealingPolicyStatusEvidenceDefaultsObservationTTL(t *testing.T) {
	policy, status, observed := evidenceFixture()
	policy.Spec.ObserveIntervalSeconds = 0
	item, err := BuildHealingPolicyStatusEvidence(policy, status, observed)
	if err != nil {
		t.Fatal(err)
	}
	if item.FreshUntil == nil || !item.FreshUntil.Equal(observed.Add(defaultObservationTTL)) {
		t.Fatalf("unexpected default freshness: %v", item.FreshUntil)
	}
}

func TestBuildHealingPolicyStatusEvidenceRejectsMissingObservationTime(t *testing.T) {
	policy, status, observed := evidenceFixture()
	status.LastObservedTime = nil
	if _, err := BuildHealingPolicyStatusEvidence(policy, status, observed); err == nil || !strings.Contains(err.Error(), "lastObservedTime") {
		t.Fatalf("expected missing observation-time error, got %v", err)
	}
}

func evidenceFixture() (*infrahealv1alpha1.HealingPolicy, infrahealv1alpha1.HealingPolicyStatus, time.Time) {
	observed := time.Date(2026, 8, 10, 6, 30, 0, 0, time.UTC)
	mt := metav1.NewTime(observed)
	policy := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-policy", Namespace: "default"},
		Spec: infrahealv1alpha1.HealingPolicySpec{
			Target:                 infrahealv1alpha1.TargetReference{Name: "demo-api"},
			Mode:                   infrahealv1alpha1.ObservationModeObserve,
			ObserveIntervalSeconds: 15,
		},
	}
	status := infrahealv1alpha1.HealingPolicyStatus{
		ObservedGeneration: 3,
		TargetFound:        true,
		DesiredReplicas:    2,
		AvailableReplicas:  2,
		ReadyReplicas:      2,
		UpdatedReplicas:    2,
		LastObservedTime:   &mt,
		Conditions: []metav1.Condition{
			{Type: infrahealv1alpha1.ConditionTargetResolved, Status: metav1.ConditionTrue, Reason: "DeploymentResolved", Message: "resolved", LastTransitionTime: mt},
			{Type: infrahealv1alpha1.ConditionObservedHealthy, Status: metav1.ConditionTrue, Reason: "Available", Message: "healthy", LastTransitionTime: mt},
		},
	}
	return policy, status, observed
}
