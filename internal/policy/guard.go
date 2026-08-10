package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
	infraevidence "github.com/vickers-rz/trustworthy-infra-self-healing/internal/evidence"
	infrarisk "github.com/vickers-rz/trustworthy-infra-self-healing/internal/risk"
)

type Decision struct {
	Allowed       bool             `json:"allowed"`
	EffectiveRisk domain.RiskClass `json:"effective_risk"`
	Reasons       []string         `json:"reasons"`
}

type RiskClassifier interface {
	Classify(domain.Proposal, domain.ExecutionContext) (domain.RiskClass, []string, error)
}

type EvidenceFreshnessPolicy interface {
	Evaluate([]domain.EvidenceRef, time.Time, domain.RiskClass) ([]string, error)
}

type Guard struct {
	classifier      RiskClassifier
	freshnessPolicy EvidenceFreshnessPolicy
}

func NewGuard() *Guard {
	return &Guard{
		classifier:      infrarisk.NewClassifier(),
		freshnessPolicy: infraevidence.NewFreshnessPolicy(),
	}
}

func NewGuardWithPolicies(classifier RiskClassifier, freshnessPolicy EvidenceFreshnessPolicy) *Guard {
	return &Guard{classifier: classifier, freshnessPolicy: freshnessPolicy}
}

func (g *Guard) Evaluate(p domain.Proposal, ctx domain.ExecutionContext, bundle domain.EvidenceBundle) Decision {
	d := Decision{Allowed: true, EffectiveRisk: domain.RiskR5}
	deny := func(reason string) {
		d.Allowed = false
		d.Reasons = append(d.Reasons, reason)
	}

	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.IncidentID) == "" {
		deny("proposal and incident IDs are required")
	}
	if strings.TrimSpace(p.Hypothesis.ID) == "" || strings.TrimSpace(p.Hypothesis.Type) == "" {
		deny("proposal hypothesis requires id and type")
	}
	if strings.TrimSpace(ctx.ActorID) == "" {
		deny("trusted execution context requires actor identity")
	}
	if strings.TrimSpace(ctx.Environment) == "" {
		deny("trusted execution context requires environment")
	}
	if strings.TrimSpace(ctx.PolicyVersion) == "" {
		deny("trusted execution context requires policy version")
	}
	if ctx.DecisionTime.IsZero() {
		deny("trusted execution context requires decision time")
	}
	if ctx.OperatorOverride {
		deny("operator override is active; automation must relinquish control")
	}
	if ctx.Authority == domain.AuthorityDeny {
		deny("authority is DENY")
	}
	if ctx.Authority != domain.AuthorityDelegated && ctx.Authority != domain.AuthorityApproval {
		deny("trusted context lacks execution authority")
	}
	if ctx.Authority == domain.AuthorityApproval && !ctx.HasHumanApproval() {
		deny("L5 approval authority requires explicit human approval provenance")
	}

	var selectedEvidence []domain.EvidenceRef
	if err := bundle.Validate(); err != nil {
		deny(fmt.Sprintf("trusted evidence bundle is invalid: %v", err))
	} else {
		if strings.TrimSpace(p.EvidenceBundleID) == "" {
			deny("proposal requires evidence_bundle_id")
		} else if p.EvidenceBundleID != bundle.ID {
			deny(fmt.Sprintf("proposal evidence bundle %q does not match trusted bundle %q", p.EvidenceBundleID, bundle.ID))
		}
		if p.IncidentID != "" && bundle.IncidentID != p.IncidentID {
			deny(fmt.Sprintf("trusted evidence bundle incident %q does not match proposal incident %q", bundle.IncidentID, p.IncidentID))
		}
		resolved, err := bundle.Resolve(p.EvidenceIDs)
		if err != nil {
			deny(fmt.Sprintf("proposal evidence selection is invalid: %v", err))
		} else {
			selectedEvidence = resolved
		}
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

	effectiveRisk, riskReasons, err := g.classifier.Classify(p, ctx)
	if err != nil {
		deny(fmt.Sprintf("deterministic risk classification failed: %v", err))
	} else {
		d.EffectiveRisk = effectiveRisk
		d.Reasons = append(d.Reasons, riskReasons...)

		if len(selectedEvidence) > 0 && !ctx.DecisionTime.IsZero() {
			freshnessReasons, freshnessErr := g.freshnessPolicy.Evaluate(selectedEvidence, ctx.DecisionTime, effectiveRisk)
			d.Reasons = append(d.Reasons, freshnessReasons...)
			if freshnessErr != nil {
				deny(fmt.Sprintf("evidence freshness policy denied remediation: %v", freshnessErr))
			}
		}

		if p.EstimatedRisk != nil {
			if *p.EstimatedRisk < effectiveRisk {
				deny(fmt.Sprintf("proposal estimated risk %s under-reports deterministic effective risk %s", p.EstimatedRisk.String(), effectiveRisk.String()))
			} else if *p.EstimatedRisk > effectiveRisk {
				d.Reasons = append(d.Reasons, fmt.Sprintf("proposal estimated risk %s is more conservative than effective risk %s", p.EstimatedRisk.String(), effectiveRisk.String()))
			}
		}

		if effectiveRisk >= domain.RiskR2 {
			if p.Rollback == nil {
				deny("R2+ remediation requires a typed rollback action before execution")
			} else if err := p.Rollback.Validate(); err != nil {
				deny(fmt.Sprintf("rollback action is invalid: %v", err))
			}
		}
		if effectiveRisk >= domain.RiskR3 && ctx.Authority != domain.AuthorityApproval {
			deny("R3+ remediation requires L5 authority")
		}
		if effectiveRisk == domain.RiskR2 && ctx.Authority == domain.AuthorityDelegated {
			deny("R2 remediation is not autonomously delegated in the MVP policy")
		}
		if effectiveRisk >= domain.RiskR4 {
			deny("R4/R5 actions are hard-denied by the MVP policy")
		}
	}

	if d.Allowed {
		d.Reasons = append(d.Reasons, "proposal satisfies trusted evidence, deterministic risk classification and remediation policy")
	}
	return d
}
