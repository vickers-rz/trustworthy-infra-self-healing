package controller

import (
	"context"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrahealv1alpha1 "github.com/vickers-rz/trustworthy-infra-self-healing/api/v1alpha1"
)

func TestReconcilePublishesDeterministicKubernetesEvidenceIdentity(t *testing.T) {
	scheme := testScheme(t)
	replicas := int32(1)
	policy := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "evidence", Namespace: "default", Generation: 1},
		Spec: infrahealv1alpha1.HealingPolicySpec{
			Target:                 infrahealv1alpha1.TargetReference{Name: "evidence"},
			Mode:                   infrahealv1alpha1.ObservationModeObserve,
			ObserveIntervalSeconds: 30,
		},
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "evidence", Namespace: "default", Generation: 2},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 2,
			AvailableReplicas:  1,
			ReadyReplicas:      1,
			UpdatedReplicas:    1,
		},
	}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infrahealv1alpha1.HealingPolicy{}).
		WithObjects(policy, deployment).
		Build()

	observed := time.Date(2026, 8, 10, 7, 0, 0, 0, time.UTC)
	r := &HealingPolicyReconciler{Client: c, Scheme: scheme, Now: func() time.Time { return observed }}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "evidence"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var got infrahealv1alpha1.HealingPolicy
	if err := c.Get(context.Background(), req.NamespacedName, &got); err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if !strings.HasPrefix(got.Status.LastEvidenceID, "ev_k8s_") {
		t.Fatalf("unexpected evidence id: %q", got.Status.LastEvidenceID)
	}
	if len(got.Status.LastEvidenceDigestSHA256) != 64 {
		t.Fatalf("expected 64-character evidence digest, got %q", got.Status.LastEvidenceDigestSHA256)
	}
	if got.Status.LastEvidenceID != "ev_k8s_"+got.Status.LastEvidenceDigestSHA256 {
		t.Fatalf("evidence id and digest are not bound: id=%q digest=%q", got.Status.LastEvidenceID, got.Status.LastEvidenceDigestSHA256)
	}
}

func TestMissingTargetAlsoPublishesObservationEvidence(t *testing.T) {
	scheme := testScheme(t)
	policy := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "missing-evidence", Namespace: "default", Generation: 1},
		Spec:       infrahealv1alpha1.HealingPolicySpec{Target: infrahealv1alpha1.TargetReference{Name: "missing"}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&infrahealv1alpha1.HealingPolicy{}).WithObjects(policy).Build()
	r := &HealingPolicyReconciler{Client: c, Scheme: scheme, Now: func() time.Time {
		return time.Date(2026, 8, 10, 7, 5, 0, 0, time.UTC)
	}}
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing-evidence"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile missing target: %v", err)
	}

	var got infrahealv1alpha1.HealingPolicy
	if err := c.Get(context.Background(), req.NamespacedName, &got); err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if got.Status.TargetFound {
		t.Fatal("target should remain missing")
	}
	if got.Status.LastEvidenceID == "" || got.Status.LastEvidenceDigestSHA256 == "" {
		t.Fatalf("missing-target observation should still have evidence identity: %#v", got.Status)
	}
}
