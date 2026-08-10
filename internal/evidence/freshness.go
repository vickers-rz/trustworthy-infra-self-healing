package evidence

import (
	"fmt"
	"time"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

// FreshnessPolicy decides whether the evidence selected by a proposal is
// temporally sufficient for a mutation decision. The decision time is supplied
// by trusted runtime context so the policy is deterministic and replayable.
type FreshnessPolicy struct{}

func NewFreshnessPolicy() FreshnessPolicy { return FreshnessPolicy{} }

func (FreshnessPolicy) Evaluate(items []domain.EvidenceRef, decisionTime time.Time, risk domain.RiskClass) ([]string, error) {
	if decisionTime.IsZero() {
		return nil, fmt.Errorf("trusted decision time is required for evidence freshness evaluation")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no selected evidence")
	}

	reasons := make([]string, 0, len(items)+1)
	freshOperational := 0
	for _, item := range items {
		if !requiresOperationalFreshness(item.Kind) {
			reasons = append(reasons, fmt.Sprintf("reference evidence %s (%s) does not establish current operational state", item.ID, item.Kind))
			continue
		}

		switch item.Freshness(decisionTime) {
		case domain.FreshnessFresh:
			freshOperational++
			reasons = append(reasons, fmt.Sprintf("operational evidence %s is fresh at decision time", item.ID))
		case domain.FreshnessStale:
			return reasons, fmt.Errorf("operational evidence %s (%s) is stale at decision time", item.ID, item.Kind)
		case domain.FreshnessUnknown:
			return reasons, fmt.Errorf("operational evidence %s (%s) has unknown freshness", item.ID, item.Kind)
		default:
			return reasons, fmt.Errorf("operational evidence %s has unsupported freshness state", item.ID)
		}
	}

	if risk >= domain.RiskR1 && freshOperational == 0 {
		return reasons, fmt.Errorf("mutable remediation requires at least one fresh operational evidence item")
	}
	reasons = append(reasons, fmt.Sprintf("selected evidence contains %d fresh operational item(s)", freshOperational))
	return reasons, nil
}

func requiresOperationalFreshness(kind domain.EvidenceKind) bool {
	switch kind {
	case domain.EvidenceMetric,
		domain.EvidenceLog,
		domain.EvidenceTrace,
		domain.EvidenceKubernetesState,
		domain.EvidenceKubernetesEvent,
		domain.EvidenceChange,
		domain.EvidenceTopology,
		domain.EvidenceOperatorAnnotation:
		return true
	case domain.EvidenceRunbook, domain.EvidenceIncident:
		return false
	default:
		// Unknown kinds are rejected by EvidenceRef.Validate before this policy.
		return true
	}
}
