package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	infrahealv1alpha1 "github.com/vickers-rz/trustworthy-infra-self-healing/api/v1alpha1"
	infraevidence "github.com/vickers-rz/trustworthy-infra-self-healing/internal/evidence"
)

const (
	defaultObserveInterval   = 30 * time.Second
	healingPolicyTargetIndex = "infraheal.io/targetDeployment"
)

// HealingPolicyReconciler is deliberately observation-only. It may read
// Deployments and update HealingPolicy status, but it has no mutation path to
// the target workload.
//
// +kubebuilder:rbac:groups=infraheal.io,resources=healingpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=infraheal.io,resources=healingpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
type HealingPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Now    func() time.Time
}

func (r *HealingPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var policy infrahealv1alpha1.HealingPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	interval := observationInterval(policy.Spec.ObserveIntervalSeconds)
	if policy.Spec.Mode != "" && policy.Spec.Mode != infrahealv1alpha1.ObservationModeObserve {
		return ctrl.Result{}, fmt.Errorf("unsupported HealingPolicy mode %q", policy.Spec.Mode)
	}

	targetNamespace := policy.Spec.Target.Namespace
	if targetNamespace == "" {
		targetNamespace = policy.Namespace
	}

	var deployment appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{Namespace: targetNamespace, Name: policy.Spec.Target.Name}, &deployment)
	if apierrors.IsNotFound(err) {
		status := r.missingTargetStatus(&policy, targetNamespace)
		if evidenceErr := r.attachEvidenceIdentity(&policy, &status); evidenceErr != nil {
			return ctrl.Result{}, evidenceErr
		}
		if updateErr := r.updateStatus(ctx, &policy, status); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{RequeueAfter: interval}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	status := r.observedDeploymentStatus(&policy, &deployment)
	if err := r.attachEvidenceIdentity(&policy, &status); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.updateStatus(ctx, &policy, status); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *HealingPolicyReconciler) observedDeploymentStatus(policy *infrahealv1alpha1.HealingPolicy, deployment *appsv1.Deployment) infrahealv1alpha1.HealingPolicyStatus {
	now := metav1.NewTime(r.now())
	desired := int32(1)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}

	status := infrahealv1alpha1.HealingPolicyStatus{
		ObservedGeneration: policy.Generation,
		TargetFound:        true,
		DesiredReplicas:   desired,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		LastObservedTime:  &now,
		Conditions:        append([]metav1.Condition(nil), policy.Status.Conditions...),
	}

	setCondition(&status.Conditions,
		infrahealv1alpha1.ConditionTargetResolved,
		metav1.ConditionTrue,
		"DeploymentResolved",
		"target Deployment was resolved",
		policy.Generation,
		now,
	)
	setCondition(&status.Conditions,
		infrahealv1alpha1.ConditionObserveOnly,
		metav1.ConditionTrue,
		"SafetyBoundary",
		"controller is read-only for the target workload",
		policy.Generation,
		now,
	)

	healthy := desired > 0 && deployment.Status.AvailableReplicas >= desired && deployment.Status.ObservedGeneration >= deployment.Generation
	if healthy {
		setCondition(&status.Conditions,
			infrahealv1alpha1.ConditionObservedHealthy,
			metav1.ConditionTrue,
			"Available",
			"observed Deployment satisfies the MVP availability check",
			policy.Generation,
			now,
		)
	} else {
		setCondition(&status.Conditions,
			infrahealv1alpha1.ConditionObservedHealthy,
			metav1.ConditionFalse,
			"Unavailable",
			"observed Deployment does not satisfy the MVP availability check",
			policy.Generation,
			now,
		)
	}
	return status
}

func (r *HealingPolicyReconciler) missingTargetStatus(policy *infrahealv1alpha1.HealingPolicy, namespace string) infrahealv1alpha1.HealingPolicyStatus {
	now := metav1.NewTime(r.now())
	status := infrahealv1alpha1.HealingPolicyStatus{
		ObservedGeneration: policy.Generation,
		TargetFound:        false,
		LastObservedTime:   &now,
		Conditions:         append([]metav1.Condition(nil), policy.Status.Conditions...),
	}
	setCondition(&status.Conditions,
		infrahealv1alpha1.ConditionTargetResolved,
		metav1.ConditionFalse,
		"DeploymentNotFound",
		fmt.Sprintf("target Deployment %s/%s was not found", namespace, policy.Spec.Target.Name),
		policy.Generation,
		now,
	)
	setCondition(&status.Conditions,
		infrahealv1alpha1.ConditionObserveOnly,
		metav1.ConditionTrue,
		"SafetyBoundary",
		"controller is read-only for the target workload",
		policy.Generation,
		now,
	)
	setCondition(&status.Conditions,
		infrahealv1alpha1.ConditionObservedHealthy,
		metav1.ConditionUnknown,
		"TargetUnavailable",
		"health cannot be assessed until the target exists",
		policy.Generation,
		now,
	)
	return status
}

func (r *HealingPolicyReconciler) attachEvidenceIdentity(policy *infrahealv1alpha1.HealingPolicy, status *infrahealv1alpha1.HealingPolicyStatus) error {
	if status == nil || status.LastObservedTime == nil {
		return fmt.Errorf("cannot build Kubernetes evidence without observation time")
	}
	item, err := infraevidence.BuildHealingPolicyStatusEvidence(policy, *status, status.LastObservedTime.Time)
	if err != nil {
		return fmt.Errorf("build HealingPolicy evidence: %w", err)
	}
	status.LastEvidenceID = item.ID
	status.LastEvidenceDigestSHA256 = item.DigestSHA256
	return nil
}

func (r *HealingPolicyReconciler) updateStatus(ctx context.Context, policy *infrahealv1alpha1.HealingPolicy, next infrahealv1alpha1.HealingPolicyStatus) error {
	policy.Status = next
	return r.Status().Update(ctx, policy)
}

func (r *HealingPolicyReconciler) now() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}
	return time.Now().UTC()
}

func observationInterval(seconds int32) time.Duration {
	if seconds <= 0 {
		return defaultObserveInterval
	}
	return time.Duration(seconds) * time.Second
}

func setCondition(conditions *[]metav1.Condition, t string, status metav1.ConditionStatus, reason, message string, observedGeneration int64, now metav1.Time) {
	apimeta.SetStatusCondition(conditions, metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: now,
	})
}

func healingPolicyTargetKey(policy *infrahealv1alpha1.HealingPolicy) string {
	name := strings.TrimSpace(policy.Spec.Target.Name)
	if name == "" {
		return ""
	}

	namespace := strings.TrimSpace(policy.Spec.Target.Namespace)
	if namespace == "" {
		namespace = strings.TrimSpace(policy.Namespace)
	}
	if namespace == "" {
		return ""
	}

	return namespace + "/" + name
}

func indexHealingPolicyTarget(obj client.Object) []string {
	policy, ok := obj.(*infrahealv1alpha1.HealingPolicy)
	if !ok {
		return nil
	}
	key := healingPolicyTargetKey(policy)
	if key == "" {
		return nil
	}
	return []string{key}
}

func (r *HealingPolicyReconciler) mapDeploymentToPolicies(ctx context.Context, obj client.Object) []ctrl.Request {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil
	}

	targetKey := deployment.Namespace + "/" + deployment.Name
	var policies infrahealv1alpha1.HealingPolicyList
	if err := r.List(ctx, &policies, client.MatchingFields{healingPolicyTargetIndex: targetKey}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "unable to map Deployment event through HealingPolicy target index", "target", targetKey)
		return nil
	}

	requests := make([]ctrl.Request, 0, len(policies.Items))
	for i := range policies.Items {
		policy := &policies.Items[i]
		requests = append(requests, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: policy.Namespace, Name: policy.Name}})
	}
	return requests
}

func (r *HealingPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &infrahealv1alpha1.HealingPolicy{}, healingPolicyTargetIndex, indexHealingPolicyTarget); err != nil {
		return fmt.Errorf("index HealingPolicy targets: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrahealv1alpha1.HealingPolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(r.mapDeploymentToPolicies)).
		Complete(r)
}
