package controller

import (
	"context"
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	infrahealv1alpha1 "github.com/vickers-rz/trustworthy-infra-self-healing/api/v1alpha1"
)

func TestReconcileObservesHealthyDeploymentWithoutMutatingIt(t *testing.T) {
	scheme := testScheme(t)
	replicas := int32(2)
	policy := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default", Generation: 3},
		Spec: infrahealv1alpha1.HealingPolicySpec{
			Target: infrahealv1alpha1.TargetReference{Name: "api"},
			Mode: infrahealv1alpha1.ObservationModeObserve,
			ObserveIntervalSeconds: 15,
		},
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default", Generation: 7},
		Spec: appsv1.DeploymentSpec{Replicas: &replicas},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 7,
			AvailableReplicas:  2,
			ReadyReplicas:      2,
			UpdatedReplicas:    2,
		},
	}
	originalSpec := *deployment.Spec.DeepCopy()

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infrahealv1alpha1.HealingPolicy{}).
		WithObjects(policy, deployment).
		Build()

	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	r := &HealingPolicyReconciler{Client: c, Scheme: scheme, Now: func() time.Time { return now }}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "api"}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.RequeueAfter != 15*time.Second {
		t.Fatalf("expected 15s requeue, got %s", result.RequeueAfter)
	}

	var gotPolicy infrahealv1alpha1.HealingPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "api"}, &gotPolicy); err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if !gotPolicy.Status.TargetFound || gotPolicy.Status.AvailableReplicas != 2 || gotPolicy.Status.DesiredReplicas != 2 {
		t.Fatalf("unexpected status: %#v", gotPolicy.Status)
	}
	if status := conditionStatus(gotPolicy.Status.Conditions, infrahealv1alpha1.ConditionObservedHealthy); status != metav1.ConditionTrue {
		t.Fatalf("expected healthy condition true, got %s", status)
	}
	if status := conditionStatus(gotPolicy.Status.Conditions, infrahealv1alpha1.ConditionObserveOnly); status != metav1.ConditionTrue {
		t.Fatalf("expected observe-only safety condition true, got %s", status)
	}

	var gotDeployment appsv1.Deployment
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "api"}, &gotDeployment); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	if !reflect.DeepEqual(originalSpec, gotDeployment.Spec) {
		t.Fatalf("controller mutated target Deployment spec: before=%#v after=%#v", originalSpec, gotDeployment.Spec)
	}
}

func TestReconcileReportsMissingDeployment(t *testing.T) {
	scheme := testScheme(t)
	policy := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "missing", Namespace: "default", Generation: 1},
		Spec: infrahealv1alpha1.HealingPolicySpec{Target: infrahealv1alpha1.TargetReference{Name: "does-not-exist"}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&infrahealv1alpha1.HealingPolicy{}).WithObjects(policy).Build()
	r := &HealingPolicyReconciler{Client: c, Scheme: scheme, Now: func() time.Time { return time.Unix(0, 0).UTC() }}

	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing"}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.RequeueAfter != defaultObserveInterval {
		t.Fatalf("expected default requeue %s, got %s", defaultObserveInterval, result.RequeueAfter)
	}

	var got infrahealv1alpha1.HealingPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "missing"}, &got); err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if got.Status.TargetFound {
		t.Fatalf("missing target must not be marked found")
	}
	if status := conditionStatus(got.Status.Conditions, infrahealv1alpha1.ConditionTargetResolved); status != metav1.ConditionFalse {
		t.Fatalf("expected target-resolved false, got %s", status)
	}
	if status := conditionStatus(got.Status.Conditions, infrahealv1alpha1.ConditionObservedHealthy); status != metav1.ConditionUnknown {
		t.Fatalf("expected health unknown, got %s", status)
	}
}

func TestDeploymentEventMapsToMatchingPolicy(t *testing.T) {
	scheme := testScheme(t)
	matching := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "match", Namespace: "ops"},
		Spec: infrahealv1alpha1.HealingPolicySpec{Target: infrahealv1alpha1.TargetReference{Namespace: "apps", Name: "api"}},
	}
	nonMatching := &infrahealv1alpha1.HealingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "ops"},
		Spec: infrahealv1alpha1.HealingPolicySpec{Target: infrahealv1alpha1.TargetReference{Namespace: "apps", Name: "worker"}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(matching, nonMatching).Build()
	r := &HealingPolicyReconciler{Client: c, Scheme: scheme}

	requests := r.mapDeploymentToPolicies(context.Background(), &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "apps", Name: "api"}})
	if len(requests) != 1 || requests[0].Namespace != "ops" || requests[0].Name != "match" {
		t.Fatalf("unexpected mapped requests: %#v", requests)
	}
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := appsv1.AddToScheme(s); err != nil {
		t.Fatalf("add apps scheme: %v", err)
	}
	if err := infrahealv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add infraheal scheme: %v", err)
	}
	return s
}

func conditionStatus(conditions []metav1.Condition, kind string) metav1.ConditionStatus {
	for _, c := range conditions {
		if c.Type == kind {
			return c.Status
		}
	}
	return metav1.ConditionUnknown
}
