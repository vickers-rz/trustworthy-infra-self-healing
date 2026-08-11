package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrahealv1alpha1 "github.com/vickers-rz/trustworthy-infra-self-healing/api/v1alpha1"
	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

const (
	kubernetesObserverCollector = "infraheal-healingpolicy-observer/v1"
	defaultObservationTTL       = 30 * time.Second
)

type healingPolicySnapshot struct {
	PolicyNamespace string `json:"policy_namespace"`
	PolicyName      string `json:"policy_name"`
	TargetNamespace string `json:"target_namespace"`
	TargetName      string `json:"target_name"`

	ObservedGeneration int64 `json:"observed_generation"`
	TargetFound        bool  `json:"target_found"`
	DesiredReplicas    int32 `json:"desired_replicas"`
	AvailableReplicas  int32 `json:"available_replicas"`
	ReadyReplicas      int32 `json:"ready_replicas"`
	UpdatedReplicas    int32 `json:"updated_replicas"`

	TargetResolved  metav1.ConditionStatus `json:"target_resolved"`
	ObservedHealthy metav1.ConditionStatus `json:"observed_healthy"`
	ObservedAt      string                 `json:"observed_at"`
}

// BuildHealingPolicyStatusEvidence converts an API-server-backed observation
// status into a deterministic, provenance-bearing EvidenceRef. The hash is over
// canonical struct JSON with a fixed field order, not over human-formatted YAML.
func BuildHealingPolicyStatusEvidence(policy *infrahealv1alpha1.HealingPolicy, status infrahealv1alpha1.HealingPolicyStatus, collectedAt time.Time) (domain.EvidenceRef, error) {
	if policy == nil {
		return domain.EvidenceRef{}, fmt.Errorf("HealingPolicy is required")
	}
	if strings.TrimSpace(policy.Name) == "" || strings.TrimSpace(policy.Namespace) == "" {
		return domain.EvidenceRef{}, fmt.Errorf("HealingPolicy namespace and name are required")
	}
	if strings.TrimSpace(policy.Spec.Target.Name) == "" {
		return domain.EvidenceRef{}, fmt.Errorf("HealingPolicy target name is required")
	}
	if status.LastObservedTime == nil || status.LastObservedTime.IsZero() {
		return domain.EvidenceRef{}, fmt.Errorf("HealingPolicy status lastObservedTime is required")
	}
	if collectedAt.IsZero() {
		return domain.EvidenceRef{}, fmt.Errorf("collection time is required")
	}

	observedAt := status.LastObservedTime.Time.UTC()
	collectedAt = collectedAt.UTC()
	targetNamespace := strings.TrimSpace(policy.Spec.Target.Namespace)
	if targetNamespace == "" {
		targetNamespace = policy.Namespace
	}

	snapshot := healingPolicySnapshot{
		PolicyNamespace:   policy.Namespace,
		PolicyName:        policy.Name,
		TargetNamespace:   targetNamespace,
		TargetName:        policy.Spec.Target.Name,
		ObservedGeneration: status.ObservedGeneration,
		TargetFound:        status.TargetFound,
		DesiredReplicas:    status.DesiredReplicas,
		AvailableReplicas:  status.AvailableReplicas,
		ReadyReplicas:      status.ReadyReplicas,
		UpdatedReplicas:    status.UpdatedReplicas,
		TargetResolved:     statusCondition(status.Conditions, infrahealv1alpha1.ConditionTargetResolved),
		ObservedHealthy:    statusCondition(status.Conditions, infrahealv1alpha1.ConditionObservedHealthy),
		ObservedAt:         observedAt.Format(time.RFC3339Nano),
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return domain.EvidenceRef{}, fmt.Errorf("marshal canonical HealingPolicy snapshot: %w", err)
	}
	digestBytes := sha256.Sum256(encoded)
	digest := hex.EncodeToString(digestBytes[:])

	ttl := observationTTL(policy.Spec.ObserveIntervalSeconds)
	freshUntil := observedAt.Add(ttl)
	targetSubject := fmt.Sprintf("k8s://%s/deployment/%s", targetNamespace, policy.Spec.Target.Name)
	uri := fmt.Sprintf("k8s://%s/healingpolicy/%s/status?observedGeneration=%d", policy.Namespace, policy.Name, status.ObservedGeneration)

	item := domain.EvidenceRef{
		ID:          "ev_k8s_" + digest,
		Kind:        domain.EvidenceKubernetesState,
		URI:         uri,
		Source:      "kubernetes-apiserver",
		Collector:   kubernetesObserverCollector,
		Subject:     targetSubject,
		Summary:     fmt.Sprintf("target_found=%t desired=%d available=%d ready=%d updated=%d target_resolved=%s observed_healthy=%s", status.TargetFound, status.DesiredReplicas, status.AvailableReplicas, status.ReadyReplicas, status.UpdatedReplicas, snapshot.TargetResolved, snapshot.ObservedHealthy),
		ObservedAt:  observedAt,
		CollectedAt: collectedAt,
		FreshUntil:  &freshUntil,
		Trust:       domain.EvidenceTrustHigh,
		DigestSHA256: digest,
	}
	if err := item.Validate(); err != nil {
		return domain.EvidenceRef{}, fmt.Errorf("validate Kubernetes observation evidence: %w", err)
	}
	return item, nil
}

func observationTTL(seconds int32) time.Duration {
	if seconds <= 0 {
		return defaultObservationTTL
	}
	return time.Duration(seconds) * time.Second
}

func statusCondition(conditions []metav1.Condition, conditionType string) metav1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}
	return metav1.ConditionUnknown
}
