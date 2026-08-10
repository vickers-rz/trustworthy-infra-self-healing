package policy

import (
	"strings"
	"testing"
	"time"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

func riskPtr(r domain.RiskClass) *domain.RiskClass { return &r }

var testObservedAt = time.Date(2026, 8, 10, 5, 0, 0, 0, time.UTC)

func baseProposal() domain.Proposal {
	return domain.Proposal{
		ID:               "rem_test",
		IncidentID:       "inc_test",
		Hypothesis:       domain.Hypothesis{ID: "hyp_unhealthy", Type: "unhealthy_workload", Confidence: 0.9},
		EvidenceBundleID: "bundle_test",
		EvidenceIDs:      []string{"ev-policy-test"},
		EstimatedRisk:    riskPtr(domain.RiskR1),
		Action: domain.Action{
			Type: domain.ActionRestartWorkload,
			RestartWorkload: &domain.RestartWorkloadAction{
				Namespace: "default",
				Workload:  "api",
			},
		},
		Preconditions: []string{"stateless"},
		Verification:  []string{"healthy"},
	}
}

func baseContext() domain.ExecutionContext {
	return domain.ExecutionContext{
		Authority:     domain.AuthorityDelegated,
		PolicyVersion: "test-v3",
		Environment:   "staging",
		ActorID:       "controlplane:test",
		DecisionTime:  testObservedAt.Add(2 * time.Minute),
	}
}

func baseBundle() domain.EvidenceBundle {
	return domain.EvidenceBundle{
		ID:         "bundle_test",
		IncidentID: "inc_test",
		CreatedAt:  testObservedAt.Add(10 * time.Second),
		Items:      []domain.EvidenceRef{validPolicyEvidence()},
		Relations: []domain.EvidenceRelation{
			{EvidenceID: "ev-policy-test", ClaimID: "hyp_unhealthy", Type: domain.EvidenceSupports},
		},
	}
}

func TestAllowsLowRiskDelegatedSemanticActionWithTrustedEvidence(t *testing.T) {
	d := NewGuard().Evaluate(baseProposal(), baseContext(), baseBundle())
	if !d.Allowed {
		t.Fatalf("expected allow, got %#v", d)
	}
	if d.EffectiveRisk != domain.RiskR1 {
		t.Fatalf("expected R1 effective risk, got %s", d.EffectiveRisk)
	}
	if !contains(d.Reasons, "fresh operational") {
		t.Fatalf("expected freshness reasoning, got %#v", d.Reasons)
	}
}

func TestInvalidTrustedEvidenceBundleFailsClosed(t *testing.T) {
	bundle := baseBundle()
	bundle.Items[0].Collector = ""
	d := NewGuard().Evaluate(baseProposal(), baseContext(), bundle)
	if d.Allowed || !contains(d.Reasons, "trusted evidence bundle is invalid") || !contains(d.Reasons, "collector provenance") {
		t.Fatalf("expected bundle provenance denial, got %#v", d)
	}
}

func TestProposalCannotSubstituteAnotherEvidenceBundle(t *testing.T) {
	p := baseProposal()
	p.EvidenceBundleID = "model-invented-bundle"
	d := NewGuard().Evaluate(p, baseContext(), baseBundle())
	if d.Allowed || !contains(d.Reasons, "does not match trusted bundle") {
		t.Fatalf("expected evidence bundle mismatch denial, got %#v", d)
	}
}

func TestProposalCannotReferenceUnknownEvidenceID(t *testing.T) {
	p := baseProposal()
	p.EvidenceIDs = []string{"ev-does-not-exist"}
	d := NewGuard().Evaluate(p, baseContext(), baseBundle())
	if d.Allowed || !contains(d.Reasons, "unknown evidence id") {
		t.Fatalf("expected unknown evidence reference denial, got %#v", d)
	}
}

func TestStaleOperationalEvidenceFailsClosed(t *testing.T) {
	ctx := baseContext()
	ctx.DecisionTime = testObservedAt.Add(10 * time.Minute)
	d := NewGuard().Evaluate(baseProposal(), ctx, baseBundle())
	if d.Allowed || !contains(d.Reasons, "stale at decision time") {
		t.Fatalf("expected stale evidence denial, got %#v", d)
	}
}

func TestUnknownOperationalFreshnessFailsClosed(t *testing.T) {
	bundle := baseBundle()
	bundle.Items[0].FreshUntil = nil
	d := NewGuard().Evaluate(baseProposal(), baseContext(), bundle)
	if d.Allowed || !contains(d.Reasons, "unknown freshness") {
		t.Fatalf("expected unknown freshness denial, got %#v", d)
	}
}

func TestReferenceDocumentsCannotAloneAuthorizeMutation(t *testing.T) {
	bundle := baseBundle()
	bundle.Items[0].Kind = domain.EvidenceRunbook
	bundle.Items[0].FreshUntil = nil
	d := NewGuard().Evaluate(baseProposal(), baseContext(), bundle)
	if d.Allowed || !contains(d.Reasons, "requires at least one fresh operational evidence") {
		t.Fatalf("expected runbook-only denial, got %#v", d)
	}
}

func TestDecisionTimeIsRequiredAndReplayExplicit(t *testing.T) {
	ctx := baseContext()
	ctx.DecisionTime = time.Time{}
	d := NewGuard().Evaluate(baseProposal(), ctx, baseBundle())
	if d.Allowed || !contains(d.Reasons, "requires decision time") {
		t.Fatalf("expected decision-time denial, got %#v", d)
	}
}

func TestOperatorOverrideAlwaysWins(t *testing.T) {
	ctx := baseContext()
	ctx.OperatorOverride = true
	d := NewGuard().Evaluate(baseProposal(), ctx, baseBundle())
	if d.Allowed || !contains(d.Reasons, "operator override") {
		t.Fatalf("expected override denial, got %#v", d)
	}
}

func TestUnknownActionFailsClosed(t *testing.T) {
	p := baseProposal()
	p.Action.Type = domain.ActionType("run_shell")
	d := NewGuard().Evaluate(p, baseContext(), baseBundle())
	if d.Allowed || !contains(d.Reasons, "risk classification failed") {
		t.Fatalf("expected classification denial, got %#v", d)
	}
}

func TestUnderreportedRiskFailsClosed(t *testing.T) {
	p := rollbackProposal()
	p.EstimatedRisk = riskPtr(domain.RiskR1)
	d := NewGuard().Evaluate(p, approvedContext(), baseBundle())
	if d.Allowed || !contains(d.Reasons, "under-reports deterministic effective risk") {
		t.Fatalf("expected risk under-reporting denial, got %#v", d)
	}
	if d.EffectiveRisk != domain.RiskR2 {
		t.Fatalf("expected deterministic R2 risk, got %s", d.EffectiveRisk)
	}
}

func TestR2NeedsRollbackAndIsNotDelegatedInMVP(t *testing.T) {
	p := rollbackProposal()
	p.Rollback = nil
	d := NewGuard().Evaluate(p, baseContext(), baseBundle())
	if d.Allowed {
		t.Fatalf("expected R2 denial, got %#v", d)
	}
	if !contains(d.Reasons, "typed rollback") || !contains(d.Reasons, "not autonomously delegated") {
		t.Fatalf("expected rollback and delegation reasons, got %#v", d)
	}
}

func TestR3RequiresExplicitHumanApprovalProvenance(t *testing.T) {
	p := baseProposal()
	p.EstimatedRisk = riskPtr(domain.RiskR3)
	p.Action = domain.Action{
		Type:      domain.ActionDrainNode,
		DrainNode: &domain.DrainNodeAction{Node: "worker-1"},
	}
	p.Rollback = &domain.Action{
		Type:         domain.ActionUncordonNode,
		UncordonNode: &domain.UncordonNodeAction{Node: "worker-1"},
	}

	ctx := baseContext()
	ctx.Authority = domain.AuthorityApproval

	d := NewGuard().Evaluate(p, ctx, baseBundle())
	if d.Allowed || !contains(d.Reasons, "explicit human approval provenance") {
		t.Fatalf("expected approval-provenance denial, got %#v", d)
	}
}

func TestApprovedR2UsesTrustedContextAndEvidence(t *testing.T) {
	p := rollbackProposal()
	d := NewGuard().Evaluate(p, approvedContext(), baseBundle())
	if !d.Allowed {
		t.Fatalf("expected approved R2 action to pass, got %#v", d)
	}
}

func rollbackProposal() domain.Proposal {
	p := baseProposal()
	p.EstimatedRisk = riskPtr(domain.RiskR2)
	p.Action = domain.Action{
		Type: domain.ActionRollbackDeployment,
		RollbackDeployment: &domain.RollbackDeploymentAction{
			Namespace:    "default",
			Deployment:   "api",
			FromRevision: 193,
			ToRevision:   192,
		},
	}
	p.Rollback = &domain.Action{
		Type: domain.ActionRollbackDeployment,
		RollbackDeployment: &domain.RollbackDeploymentAction{
			Namespace:    "default",
			Deployment:   "api",
			FromRevision: 192,
			ToRevision:   193,
		},
	}
	return p
}

func approvedContext() domain.ExecutionContext {
	ctx := baseContext()
	ctx.Authority = domain.AuthorityApproval
	ctx.HumanApproval = &domain.HumanApproval{
		ApprovalID: "chg-1234",
		ApprovedBy: "operator@example.test",
	}
	return ctx
}

func validPolicyEvidence() domain.EvidenceRef {
	freshUntil := testObservedAt.Add(5 * time.Minute)
	return domain.EvidenceRef{
		ID:          "ev-policy-test",
		Kind:        domain.EvidenceMetric,
		URI:         "prometheus://test/api-health",
		Source:      "prometheus",
		Collector:   "policy-test-adapter/v1",
		Subject:     "k8s://default/deployment/api",
		Summary:     "health signal degraded",
		ObservedAt:  testObservedAt,
		CollectedAt: testObservedAt.Add(time.Second),
		FreshUntil:  &freshUntil,
		Trust:       domain.EvidenceTrustHigh,
	}
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if strings.Contains(item, needle) {
			return true
		}
	}
	return false
}
