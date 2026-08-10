package evidence

import (
	"strings"
	"testing"
	"time"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

var policyTime = time.Date(2026, 8, 10, 6, 0, 0, 0, time.UTC)

func TestFreshOperationalEvidenceAllowsMutableDecision(t *testing.T) {
	items := []domain.EvidenceRef{operationalEvidence("ev-fresh", policyTime.Add(time.Minute))}
	reasons, err := NewFreshnessPolicy().Evaluate(items, policyTime, domain.RiskR2)
	if err != nil {
		t.Fatalf("fresh evidence rejected: %v", err)
	}
	if !containsReason(reasons, "1 fresh operational") {
		t.Fatalf("missing freshness summary: %#v", reasons)
	}
}

func TestStaleOperationalEvidenceDenied(t *testing.T) {
	items := []domain.EvidenceRef{operationalEvidence("ev-stale", policyTime.Add(-time.Second))}
	_, err := NewFreshnessPolicy().Evaluate(items, policyTime, domain.RiskR1)
	if err == nil || !strings.Contains(err.Error(), "stale at decision time") {
		t.Fatalf("expected stale denial, got %v", err)
	}
}

func TestUnknownOperationalFreshnessDenied(t *testing.T) {
	item := operationalEvidence("ev-unknown", policyTime.Add(time.Minute))
	item.FreshUntil = nil
	_, err := NewFreshnessPolicy().Evaluate([]domain.EvidenceRef{item}, policyTime, domain.RiskR1)
	if err == nil || !strings.Contains(err.Error(), "unknown freshness") {
		t.Fatalf("expected unknown freshness denial, got %v", err)
	}
}

func TestRunbookMayBeContextButCannotAloneAuthorizeMutation(t *testing.T) {
	item := operationalEvidence("ev-runbook", policyTime.Add(time.Minute))
	item.Kind = domain.EvidenceRunbook
	item.FreshUntil = nil
	reasons, err := NewFreshnessPolicy().Evaluate([]domain.EvidenceRef{item}, policyTime, domain.RiskR1)
	if err == nil || !strings.Contains(err.Error(), "at least one fresh operational") {
		t.Fatalf("expected reference-only denial, got reasons=%#v err=%v", reasons, err)
	}
}

func TestReferenceEvidenceCanAccompanyFreshOperationalEvidence(t *testing.T) {
	live := operationalEvidence("ev-live", policyTime.Add(time.Minute))
	runbook := operationalEvidence("ev-runbook", policyTime.Add(time.Minute))
	runbook.Kind = domain.EvidenceRunbook
	runbook.FreshUntil = nil

	if _, err := NewFreshnessPolicy().Evaluate([]domain.EvidenceRef{live, runbook}, policyTime, domain.RiskR2); err != nil {
		t.Fatalf("fresh live + reference evidence should pass freshness policy: %v", err)
	}
}

func TestDecisionTimeRequired(t *testing.T) {
	_, err := NewFreshnessPolicy().Evaluate([]domain.EvidenceRef{operationalEvidence("ev", policyTime.Add(time.Minute))}, time.Time{}, domain.RiskR1)
	if err == nil || !strings.Contains(err.Error(), "decision time") {
		t.Fatalf("expected decision time failure, got %v", err)
	}
}

func operationalEvidence(id string, freshUntil time.Time) domain.EvidenceRef {
	observed := policyTime.Add(-30 * time.Second)
	return domain.EvidenceRef{
		ID:          id,
		Kind:        domain.EvidenceMetric,
		URI:         "prometheus://test/" + id,
		Source:      "prometheus",
		Collector:   "test-adapter/v1",
		Summary:     "test evidence",
		ObservedAt:  observed,
		CollectedAt: observed.Add(time.Second),
		FreshUntil:  &freshUntil,
		Trust:       domain.EvidenceTrustHigh,
	}
}

func containsReason(reasons []string, needle string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, needle) {
			return true
		}
	}
	return false
}
