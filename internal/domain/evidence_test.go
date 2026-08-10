package domain

import (
	"strings"
	"testing"
	"time"
)

func TestEvidenceRefValidateAndFreshness(t *testing.T) {
	observed := time.Date(2026, 8, 10, 5, 0, 0, 0, time.UTC)
	collected := observed.Add(5 * time.Second)
	freshUntil := observed.Add(2 * time.Minute)
	item := EvidenceRef{
		ID:           "ev-metric-001",
		Kind:         EvidenceMetric,
		URI:          "prometheus://payment-api/error-rate?window=5m",
		Source:       "prometheus",
		Collector:    "infraheal-prometheus-adapter/v0",
		Subject:      "k8s://default/deployment/payment-api",
		Summary:      "5xx error rate rose above the baseline",
		ObservedAt:   observed,
		CollectedAt:  collected,
		FreshUntil:   &freshUntil,
		Trust:        EvidenceTrustHigh,
		DigestSHA256: strings.Repeat("a", 64),
	}

	if err := item.Validate(); err != nil {
		t.Fatalf("valid evidence rejected: %v", err)
	}
	if got := item.Freshness(observed.Add(time.Minute)); got != FreshnessFresh {
		t.Fatalf("expected fresh evidence, got %s", got)
	}
	if got := item.Freshness(observed.Add(3 * time.Minute)); got != FreshnessStale {
		t.Fatalf("expected stale evidence, got %s", got)
	}
}

func TestEvidenceFreshnessUnknownWhenNoExpiryIsDeclared(t *testing.T) {
	item := validEvidence("ev-runbook", EvidenceRunbook)
	item.FreshUntil = nil
	if got := item.Freshness(time.Now().UTC()); got != FreshnessUnknown {
		t.Fatalf("expected unknown freshness, got %s", got)
	}
}

func TestEvidenceRefRejectsMissingCollectorProvenance(t *testing.T) {
	item := validEvidence("ev-1", EvidenceMetric)
	item.Collector = ""
	if err := item.Validate(); err == nil || !strings.Contains(err.Error(), "collector provenance") {
		t.Fatalf("expected collector provenance failure, got %v", err)
	}
}

func TestEvidenceRefRejectsInvalidDigestAndWindow(t *testing.T) {
	item := validEvidence("ev-1", EvidenceMetric)
	item.DigestSHA256 = "not-a-sha256"
	if err := item.Validate(); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("expected SHA-256 failure, got %v", err)
	}

	item = validEvidence("ev-2", EvidenceMetric)
	end := item.ObservedAt.Add(-time.Minute)
	start := item.ObservedAt
	item.WindowStart = &start
	item.WindowEnd = &end
	if err := item.Validate(); err == nil || !strings.Contains(err.Error(), "window_end precedes") {
		t.Fatalf("expected invalid window failure, got %v", err)
	}
}

func TestEvidenceBundleValidatesRelationsAndExplicitMissingEvidence(t *testing.T) {
	bundle := EvidenceBundle{
		ID:         "bundle-001",
		IncidentID: "inc-001",
		CreatedAt:  time.Date(2026, 8, 10, 5, 2, 0, 0, time.UTC),
		Items: []EvidenceRef{
			validEvidence("ev-support", EvidenceMetric),
			validEvidence("ev-contradict", EvidenceKubernetesState),
		},
		Relations: []EvidenceRelation{
			{EvidenceID: "ev-support", ClaimID: "hyp-deploy-regression", Type: EvidenceSupports},
			{EvidenceID: "ev-contradict", ClaimID: "hyp-deploy-regression", Type: EvidenceContradicts},
		},
		Missing: []MissingEvidence{
			{Kind: EvidenceTrace, Subject: "payment-api", Reason: "tracing backend unavailable during incident window"},
		},
	}

	if err := bundle.Validate(); err != nil {
		t.Fatalf("valid evidence bundle rejected: %v", err)
	}
}

func TestEvidenceBundleMayRepresentOnlyExplicitlyMissingEvidence(t *testing.T) {
	bundle := EvidenceBundle{
		ID:         "bundle-missing",
		IncidentID: "inc-missing",
		CreatedAt:  time.Date(2026, 8, 10, 5, 2, 0, 0, time.UTC),
		Missing: []MissingEvidence{
			{Kind: EvidenceMetric, Subject: "payment-api", Reason: "metrics backend unreachable"},
		},
	}
	if err := bundle.Validate(); err != nil {
		t.Fatalf("explicitly incomplete bundle should remain representable: %v", err)
	}
}

func TestEvidenceBundleRejectsDuplicateIDsAndDanglingRelations(t *testing.T) {
	duplicate := validEvidence("ev-dup", EvidenceMetric)
	bundle := EvidenceBundle{
		ID:         "bundle-dup",
		IncidentID: "inc-dup",
		CreatedAt:  time.Now().UTC(),
		Items:      []EvidenceRef{duplicate, duplicate},
	}
	if err := bundle.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate evidence id") {
		t.Fatalf("expected duplicate ID failure, got %v", err)
	}

	bundle = EvidenceBundle{
		ID:         "bundle-relation",
		IncidentID: "inc-relation",
		CreatedAt:  time.Now().UTC(),
		Items:      []EvidenceRef{validEvidence("ev-known", EvidenceMetric)},
		Relations: []EvidenceRelation{
			{EvidenceID: "ev-missing", ClaimID: "hyp-1", Type: EvidenceSupports},
		},
	}
	if err := bundle.Validate(); err == nil || !strings.Contains(err.Error(), "unknown evidence id") {
		t.Fatalf("expected dangling relation failure, got %v", err)
	}
}

func TestEvidenceBundleFreshnessSummaryKeepsUnknownDistinctFromStale(t *testing.T) {
	now := time.Date(2026, 8, 10, 6, 0, 0, 0, time.UTC)
	fresh := validEvidence("fresh", EvidenceMetric)
	freshExpiry := now.Add(time.Minute)
	fresh.FreshUntil = &freshExpiry

	stale := validEvidence("stale", EvidenceMetric)
	staleExpiry := now.Add(-time.Minute)
	stale.ObservedAt = now.Add(-2 * time.Minute)
	stale.FreshUntil = &staleExpiry

	unknown := validEvidence("unknown", EvidenceRunbook)
	unknown.FreshUntil = nil

	bundle := EvidenceBundle{Items: []EvidenceRef{fresh, stale, unknown}}
	summary := bundle.FreshnessSummary(now)
	if summary.Fresh != 1 || summary.Stale != 1 || summary.Unknown != 1 {
		t.Fatalf("unexpected freshness summary: %#v", summary)
	}
}

func validEvidence(id string, kind EvidenceKind) EvidenceRef {
	observed := time.Date(2026, 8, 10, 5, 0, 0, 0, time.UTC)
	freshUntil := observed.Add(2 * time.Hour)
	return EvidenceRef{
		ID:          id,
		Kind:        kind,
		URI:         "evidence://" + id,
		Source:      "test-source",
		Collector:   "test-collector/v1",
		Subject:     "test-subject",
		Summary:     "test evidence",
		ObservedAt:  observed,
		CollectedAt: observed.Add(time.Second),
		FreshUntil:  &freshUntil,
		Trust:       EvidenceTrustHigh,
	}
}
