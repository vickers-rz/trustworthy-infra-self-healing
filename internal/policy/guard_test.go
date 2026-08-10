package policy

import (
	"strings"
	"testing"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

func baseProposal() domain.Proposal {
	return domain.Proposal{
		ID:            "rem_test",
		IncidentID:    "inc_test",
		Authority:     domain.AuthorityDelegated,
		Risk:          domain.RiskR1,
		Hypothesis:    domain.Hypothesis{Type: "unhealthy_workload", Confidence: 0.9},
		Evidence:      []domain.EvidenceRef{{URI: "metric://test", Summary: "degraded"}},
		Action:        domain.Action{Type: domain.ActionRestartWorkload, Target: "api"},
		Preconditions: []string{"stateless"},
		Verification:  []string{"healthy"},
		PolicyVersion: "test-v1",
	}
}

func TestAllowsLowRiskDelegatedSemanticAction(t *testing.T) {
	d := NewGuard().Evaluate(baseProposal())
	if !d.Allowed {
		t.Fatalf("expected allow, got %#v", d)
	}
}

func TestOperatorOverrideAlwaysWins(t *testing.T) {
	p := baseProposal()
	p.OperatorOverride = true
	d := NewGuard().Evaluate(p)
	if d.Allowed || !contains(d.Reasons, "operator override") {
		t.Fatalf("expected override denial, got %#v", d)
	}
}

func TestUnknownActionFailsClosed(t *testing.T) {
	p := baseProposal()
	p.Action.Type = domain.ActionType("run_shell")
	d := NewGuard().Evaluate(p)
	if d.Allowed || !contains(d.Reasons, "allowlist") {
		t.Fatalf("expected allowlist denial, got %#v", d)
	}
}

func TestR2NeedsRollbackAndApprovalInMVP(t *testing.T) {
	p := baseProposal()
	p.Risk = domain.RiskR2
	p.Action.Type = domain.ActionRollbackDeployment
	d := NewGuard().Evaluate(p)
	if d.Allowed {
		t.Fatalf("expected R2 denial, got %#v", d)
	}
	if !contains(d.Reasons, "rollback") || !contains(d.Reasons, "not autonomously delegated") {
		t.Fatalf("expected rollback and delegation reasons, got %#v", d)
	}
}

func TestR3RequiresExplicitHumanApproval(t *testing.T) {
	p := baseProposal()
	p.Risk = domain.RiskR3
	p.Action.Type = domain.ActionDrainNode
	p.Rollback = &domain.Rollback{Strategy: "uncordon_node"}
	p.Authority = domain.AuthorityApproval
	d := NewGuard().Evaluate(p)
	if d.Allowed || !contains(d.Reasons, "explicit human approval") {
		t.Fatalf("expected approval denial, got %#v", d)
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
