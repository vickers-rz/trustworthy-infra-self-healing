package domain

import (
	"strings"
	"testing"
	"time"
)

func TestEvidenceBundleResolveReturnsProposalSelectionInOrder(t *testing.T) {
	bundle := EvidenceBundle{
		Items: []EvidenceRef{
			resolveEvidence("ev-a"),
			resolveEvidence("ev-b"),
		},
	}
	selected, err := bundle.Resolve([]string{"ev-b", "ev-a"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(selected) != 2 || selected[0].ID != "ev-b" || selected[1].ID != "ev-a" {
		t.Fatalf("unexpected selection: %#v", selected)
	}
}

func TestEvidenceBundleResolveRejectsUnknownAndDuplicateProposalReferences(t *testing.T) {
	bundle := EvidenceBundle{Items: []EvidenceRef{resolveEvidence("ev-a")}}
	if _, err := bundle.Resolve([]string{"missing"}); err == nil || !strings.Contains(err.Error(), "unknown evidence id") {
		t.Fatalf("expected unknown-id failure, got %v", err)
	}
	if _, err := bundle.Resolve([]string{"ev-a", "ev-a"}); err == nil || !strings.Contains(err.Error(), "duplicate proposal evidence id") {
		t.Fatalf("expected duplicate-reference failure, got %v", err)
	}
}

func resolveEvidence(id string) EvidenceRef {
	observed := time.Date(2026, 8, 10, 5, 0, 0, 0, time.UTC)
	freshUntil := observed.Add(time.Hour)
	return EvidenceRef{
		ID:          id,
		Kind:        EvidenceMetric,
		URI:         "evidence://" + id,
		Source:      "test",
		Collector:   "test/v1",
		Summary:     "test",
		ObservedAt:  observed,
		CollectedAt: observed,
		FreshUntil:  &freshUntil,
		Trust:       EvidenceTrustHigh,
	}
}
