package policy

import (
	"fmt"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

type Decision struct {
	Allowed bool     `json:"allowed"`
	Reasons []string `json:"reasons"`
}

type Guard struct {
	actionMaxRisk map[domain.ActionType]domain.RiskClass
}

func NewGuard() *Guard {
	return &Guard{actionMaxRisk: map[domain.ActionType]domain.RiskClass{
		domain.ActionRestartWorkload:    domain.RiskR1,
		domain.ActionRollbackDeployment: domain.RiskR2,
		domain.ActionScaleWorkload:      domain.RiskR2,
		domain.ActionDrainNode:          domain.RiskR3,
		domain.ActionFailoverDatabase:   domain.RiskR3,
	}}
}

func (g *Guard) Evaluate(p domain.Proposal) Decision {
	d := Decision{Allowed: true}
	deny := func(reason string) {
		d.Allowed = false
		d.Reasons = append(d.Reasons, reason)
	}

	if p.ID == "" || p.IncidentID == "" {
		deny("proposal and incident IDs are required")
	}
	if p.OperatorOverride {
		deny("operator override is active; automation must relinquish control")
	}
	if p.Authority == domain.AuthorityDeny {
		deny("authority is DENY")
	}
	if p.Authority != domain.AuthorityDelegated && p.Authority != domain.AuthorityApproval {
		deny("proposal lacks execution authority")
	}
	if len(p.Evidence) == 0 {
		deny("no evidence provenance supplied")
	}
	if len(p.Preconditions) == 0 {
		deny("mutable action requires explicit preconditions")
	}
	if len(p.Verification) == 0 {
		deny("mutable action requires independent verification criteria")
	}
	if p.Hypothesis.Confidence < 0 || p.Hypothesis.Confidence > 1 {
		deny("hypothesis confidence must be within [0,1]")
	}

	maxRisk, known := g.actionMaxRisk[p.Action.Type]
	if !known {
		deny("action type is not present in the semantic remediation allowlist")
	} else if p.Risk > maxRisk {
		deny(fmt.Sprintf("declared risk %s exceeds catalog maximum %s for action %s", p.Risk, maxRisk, p.Action.Type))
	}

	if p.Risk >= domain.RiskR2 && p.Rollback == nil {
		deny("R2+ remediation requires a rollback plan before execution")
	}
	if p.Risk >= domain.RiskR3 && !(p.Authority == domain.AuthorityApproval && p.HumanApproved) {
		deny("R3+ remediation requires L5 authority and explicit human approval")
	}
	if p.Risk == domain.RiskR2 && p.Authority == domain.AuthorityDelegated {
		deny("R2 remediation is not autonomously delegated in the MVP policy")
	}
	if p.Risk >= domain.RiskR4 {
		deny("R4/R5 actions are hard-denied by the MVP policy")
	}

	if d.Allowed {
		d.Reasons = append(d.Reasons, "proposal satisfies the current deterministic remediation policy")
	}
	return d
}
