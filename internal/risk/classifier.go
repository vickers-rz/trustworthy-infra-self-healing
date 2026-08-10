package risk

import (
	"fmt"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

// Classifier derives the effective operational risk from trusted action
// semantics. Model/planner risk estimates are advisory and never used as the
// authorization source of truth.
type Classifier struct{}

func NewClassifier() Classifier { return Classifier{} }

func (Classifier) Classify(p domain.Proposal, _ domain.ExecutionContext) (domain.RiskClass, []string, error) {
	if err := p.Action.Validate(); err != nil {
		return domain.RiskR5, nil, fmt.Errorf("invalid semantic action: %w", err)
	}

	switch p.Action.Type {
	case domain.ActionRestartWorkload:
		return domain.RiskR1, []string{"restart_workload catalogued as R1"}, nil
	case domain.ActionRollbackDeployment:
		return domain.RiskR2, []string{"rollback_deployment catalogued as R2"}, nil
	case domain.ActionScaleWorkload:
		return domain.RiskR2, []string{"scale_workload catalogued as R2"}, nil
	case domain.ActionDrainNode:
		return domain.RiskR3, []string{"drain_node catalogued as R3"}, nil
	case domain.ActionFailoverDatabase:
		return domain.RiskR3, []string{"failover_database catalogued as R3"}, nil
	default:
		return domain.RiskR5, nil, fmt.Errorf("action type %q is not in the deterministic risk catalogue", p.Action.Type)
	}
}
