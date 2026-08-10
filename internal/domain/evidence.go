package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type EvidenceKind string

const (
	EvidenceMetric             EvidenceKind = "metric"
	EvidenceLog                EvidenceKind = "log"
	EvidenceTrace              EvidenceKind = "trace"
	EvidenceKubernetesState    EvidenceKind = "kubernetes_state"
	EvidenceKubernetesEvent    EvidenceKind = "kubernetes_event"
	EvidenceChange             EvidenceKind = "change"
	EvidenceRunbook            EvidenceKind = "runbook"
	EvidenceIncident           EvidenceKind = "incident"
	EvidenceTopology           EvidenceKind = "topology"
	EvidenceOperatorAnnotation EvidenceKind = "operator_annotation"
)

type EvidenceTrust string

const (
	EvidenceTrustUnknown   EvidenceTrust = "unknown"
	EvidenceTrustUntrusted EvidenceTrust = "untrusted"
	EvidenceTrustLow       EvidenceTrust = "low"
	EvidenceTrustMedium    EvidenceTrust = "medium"
	EvidenceTrustHigh      EvidenceTrust = "high"
)

type FreshnessState string

const (
	FreshnessUnknown FreshnessState = "unknown"
	FreshnessFresh   FreshnessState = "fresh"
	FreshnessStale   FreshnessState = "stale"
)

// EvidenceRef is a provenance-bearing reference to an observation or document.
// Summary is untrusted data from the evidence source; it is never interpreted
// as an instruction or an authorization primitive.
type EvidenceRef struct {
	ID        string       `json:"id"`
	Kind      EvidenceKind `json:"kind"`
	URI       string       `json:"uri"`
	Source    string       `json:"source"`
	Collector string       `json:"collector"`
	Subject   string       `json:"subject,omitempty"`
	Summary   string       `json:"summary"`

	ObservedAt  time.Time  `json:"observed_at"`
	CollectedAt time.Time  `json:"collected_at"`
	WindowStart *time.Time `json:"window_start,omitempty"`
	WindowEnd   *time.Time `json:"window_end,omitempty"`
	FreshUntil  *time.Time `json:"fresh_until,omitempty"`

	Trust        EvidenceTrust `json:"trust"`
	DigestSHA256 string        `json:"sha256,omitempty"`
}

func (e EvidenceRef) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("evidence id is required")
	}
	if !e.Kind.Valid() {
		return fmt.Errorf("unsupported evidence kind %q", e.Kind)
	}
	if strings.TrimSpace(e.URI) == "" {
		return fmt.Errorf("evidence %s requires a URI", e.ID)
	}
	if strings.TrimSpace(e.Source) == "" {
		return fmt.Errorf("evidence %s requires a source", e.ID)
	}
	if strings.TrimSpace(e.Collector) == "" {
		return fmt.Errorf("evidence %s requires collector provenance", e.ID)
	}
	if strings.TrimSpace(e.Summary) == "" {
		return fmt.Errorf("evidence %s requires a summary", e.ID)
	}
	if e.ObservedAt.IsZero() {
		return fmt.Errorf("evidence %s requires observed_at", e.ID)
	}
	if e.CollectedAt.IsZero() {
		return fmt.Errorf("evidence %s requires collected_at", e.ID)
	}
	if !e.Trust.Valid() {
		return fmt.Errorf("unsupported evidence trust level %q", e.Trust)
	}

	if (e.WindowStart == nil) != (e.WindowEnd == nil) {
		return fmt.Errorf("evidence %s window_start and window_end must be supplied together", e.ID)
	}
	if e.WindowStart != nil && e.WindowEnd.Before(*e.WindowStart) {
		return fmt.Errorf("evidence %s window_end precedes window_start", e.ID)
	}
	if e.FreshUntil != nil && e.FreshUntil.Before(e.ObservedAt) {
		return fmt.Errorf("evidence %s fresh_until precedes observed_at", e.ID)
	}
	if e.DigestSHA256 != "" {
		decoded, err := hex.DecodeString(e.DigestSHA256)
		if err != nil || len(decoded) != sha256.Size {
			return fmt.Errorf("evidence %s sha256 must be 64 hexadecimal characters", e.ID)
		}
	}
	return nil
}

func (e EvidenceRef) Freshness(at time.Time) FreshnessState {
	if e.FreshUntil == nil || e.FreshUntil.IsZero() {
		return FreshnessUnknown
	}
	if at.After(*e.FreshUntil) {
		return FreshnessStale
	}
	return FreshnessFresh
}

func (k EvidenceKind) Valid() bool {
	switch k {
	case EvidenceMetric,
		EvidenceLog,
		EvidenceTrace,
		EvidenceKubernetesState,
		EvidenceKubernetesEvent,
		EvidenceChange,
		EvidenceRunbook,
		EvidenceIncident,
		EvidenceTopology,
		EvidenceOperatorAnnotation:
		return true
	default:
		return false
	}
}

func (t EvidenceTrust) Valid() bool {
	switch t {
	case EvidenceTrustUnknown, EvidenceTrustUntrusted, EvidenceTrustLow, EvidenceTrustMedium, EvidenceTrustHigh:
		return true
	default:
		return false
	}
}

type EvidenceRelationType string

const (
	EvidenceSupports    EvidenceRelationType = "supports"
	EvidenceContradicts EvidenceRelationType = "contradicts"
	EvidenceContextFor  EvidenceRelationType = "context_for"
)

func (r EvidenceRelationType) Valid() bool {
	switch r {
	case EvidenceSupports, EvidenceContradicts, EvidenceContextFor:
		return true
	default:
		return false
	}
}

// EvidenceRelation records how one evidence item relates to a hypothesis or
// claim. The relation is explicit so retrieval cannot silently hide
// contradictory evidence behind a single model-generated narrative.
type EvidenceRelation struct {
	EvidenceID string               `json:"evidence_id"`
	ClaimID    string               `json:"claim_id"`
	Type       EvidenceRelationType `json:"type"`
}

type MissingEvidence struct {
	Kind    EvidenceKind `json:"kind"`
	Subject string       `json:"subject"`
	Reason  string       `json:"reason"`
}

// EvidenceBundle is a reconstructable incident evidence pack. A bundle may be
// intentionally incomplete, but missing evidence must then be represented
// explicitly rather than silently omitted.
type EvidenceBundle struct {
	ID         string             `json:"id"`
	IncidentID string             `json:"incident_id"`
	CreatedAt  time.Time          `json:"created_at"`
	Items      []EvidenceRef      `json:"items"`
	Relations  []EvidenceRelation `json:"relations,omitempty"`
	Missing    []MissingEvidence  `json:"missing,omitempty"`
}

func (b EvidenceBundle) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return fmt.Errorf("evidence bundle id is required")
	}
	if strings.TrimSpace(b.IncidentID) == "" {
		return fmt.Errorf("evidence bundle incident_id is required")
	}
	if b.CreatedAt.IsZero() {
		return fmt.Errorf("evidence bundle created_at is required")
	}
	if len(b.Items) == 0 && len(b.Missing) == 0 {
		return fmt.Errorf("evidence bundle must contain evidence or explicitly declared missing evidence")
	}

	ids := make(map[string]struct{}, len(b.Items))
	for i := range b.Items {
		item := b.Items[i]
		if err := item.Validate(); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
		if _, exists := ids[item.ID]; exists {
			return fmt.Errorf("duplicate evidence id %q", item.ID)
		}
		ids[item.ID] = struct{}{}
	}

	for i, relation := range b.Relations {
		if _, exists := ids[relation.EvidenceID]; !exists {
			return fmt.Errorf("relations[%d] references unknown evidence id %q", i, relation.EvidenceID)
		}
		if strings.TrimSpace(relation.ClaimID) == "" {
			return fmt.Errorf("relations[%d] requires claim_id", i)
		}
		if !relation.Type.Valid() {
			return fmt.Errorf("relations[%d] has unsupported relation type %q", i, relation.Type)
		}
	}

	for i, missing := range b.Missing {
		if !missing.Kind.Valid() {
			return fmt.Errorf("missing[%d] has unsupported evidence kind %q", i, missing.Kind)
		}
		if strings.TrimSpace(missing.Subject) == "" {
			return fmt.Errorf("missing[%d] requires subject", i)
		}
		if strings.TrimSpace(missing.Reason) == "" {
			return fmt.Errorf("missing[%d] requires reason", i)
		}
	}
	return nil
}

type FreshnessSummary struct {
	Fresh   int `json:"fresh"`
	Stale   int `json:"stale"`
	Unknown int `json:"unknown"`
}

func (b EvidenceBundle) FreshnessSummary(at time.Time) FreshnessSummary {
	var summary FreshnessSummary
	for _, item := range b.Items {
		switch item.Freshness(at) {
		case FreshnessFresh:
			summary.Fresh++
		case FreshnessStale:
			summary.Stale++
		default:
			summary.Unknown++
		}
	}
	return summary
}
