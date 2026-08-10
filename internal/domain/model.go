package domain

import (
	"fmt"
	"strings"
)

type AuthorityLevel string

const (
	AuthorityObserve   AuthorityLevel = "L0_OBSERVE"
	AuthorityDiagnose  AuthorityLevel = "L1_DIAGNOSE"
	AuthorityPropose   AuthorityLevel = "L2_PROPOSE"
	AuthorityPrepare   AuthorityLevel = "L3_PREPARE"
	AuthorityDelegated AuthorityLevel = "L4_DELEGATED"
	AuthorityApproval  AuthorityLevel = "L5_APPROVAL"
	AuthorityDeny      AuthorityLevel = "DENY"
)

type RiskClass int

const (
	RiskR0 RiskClass = iota
	RiskR1
	RiskR2
	RiskR3
	RiskR4
	RiskR5
)

func (r RiskClass) String() string { return fmt.Sprintf("R%d", int(r)) }

type ActionType string

const (
	ActionRestartWorkload    ActionType = "restart_workload"
	ActionRollbackDeployment ActionType = "rollback_deployment"
	ActionScaleWorkload      ActionType = "scale_workload"
	ActionDrainNode          ActionType = "drain_node"
	ActionFailoverDatabase   ActionType = "failover_database"
)

type EvidenceRef struct {
	URI     string  `json:"uri"`
	Summary string  `json:"summary"`
	Weight  float64 `json:"weight,omitempty"`
}

type Hypothesis struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
}

type RestartWorkloadAction struct {
	Namespace string `json:"namespace"`
	Workload  string `json:"workload"`
}

type RollbackDeploymentAction struct {
	Namespace    string `json:"namespace"`
	Deployment   string `json:"deployment"`
	FromRevision int64  `json:"from_revision"`
	ToRevision   int64  `json:"to_revision"`
}

type ScaleWorkloadAction struct {
	Namespace string `json:"namespace"`
	Workload  string `json:"workload"`
	Replicas  int32  `json:"replicas"`
}

type DrainNodeAction struct {
	Node string `json:"node"`
}

type FailoverDatabaseAction struct {
	Cluster       string `json:"cluster"`
	TargetReplica string `json:"target_replica"`
}

// Action is a closed semantic-action union. Exactly one typed payload must be
// present and it must match Type. There is deliberately no arbitrary shell,
// kubectl, SSH, or free-form command payload.
type Action struct {
	Type ActionType `json:"type"`

	RestartWorkload    *RestartWorkloadAction    `json:"restart_workload,omitempty"`
	RollbackDeployment *RollbackDeploymentAction `json:"rollback_deployment,omitempty"`
	ScaleWorkload      *ScaleWorkloadAction      `json:"scale_workload,omitempty"`
	DrainNode          *DrainNodeAction          `json:"drain_node,omitempty"`
	FailoverDatabase   *FailoverDatabaseAction   `json:"failover_database,omitempty"`
}

func (a Action) Validate() error {
	if a.payloadCount() != 1 {
		return fmt.Errorf("semantic action must contain exactly one typed payload")
	}

	require := func(value, field string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}

	switch a.Type {
	case ActionRestartWorkload:
		if a.RestartWorkload == nil {
			return fmt.Errorf("restart_workload payload is required")
		}
		if err := require(a.RestartWorkload.Namespace, "restart_workload.namespace"); err != nil {
			return err
		}
		return require(a.RestartWorkload.Workload, "restart_workload.workload")
	case ActionRollbackDeployment:
		if a.RollbackDeployment == nil {
			return fmt.Errorf("rollback_deployment payload is required")
		}
		if err := require(a.RollbackDeployment.Namespace, "rollback_deployment.namespace"); err != nil {
			return err
		}
		if err := require(a.RollbackDeployment.Deployment, "rollback_deployment.deployment"); err != nil {
			return err
		}
		if a.RollbackDeployment.FromRevision <= 0 || a.RollbackDeployment.ToRevision <= 0 {
			return fmt.Errorf("rollback revisions must be positive")
		}
		if a.RollbackDeployment.FromRevision == a.RollbackDeployment.ToRevision {
			return fmt.Errorf("rollback revisions must differ")
		}
		return nil
	case ActionScaleWorkload:
		if a.ScaleWorkload == nil {
			return fmt.Errorf("scale_workload payload is required")
		}
		if err := require(a.ScaleWorkload.Namespace, "scale_workload.namespace"); err != nil {
			return err
		}
		if err := require(a.ScaleWorkload.Workload, "scale_workload.workload"); err != nil {
			return err
		}
		if a.ScaleWorkload.Replicas < 0 {
			return fmt.Errorf("scale_workload.replicas cannot be negative")
		}
		return nil
	case ActionDrainNode:
		if a.DrainNode == nil {
			return fmt.Errorf("drain_node payload is required")
		}
		return require(a.DrainNode.Node, "drain_node.node")
	case ActionFailoverDatabase:
		if a.FailoverDatabase == nil {
			return fmt.Errorf("failover_database payload is required")
		}
		if err := require(a.FailoverDatabase.Cluster, "failover_database.cluster"); err != nil {
			return err
		}
		return require(a.FailoverDatabase.TargetReplica, "failover_database.target_replica")
	default:
		return fmt.Errorf("unknown semantic action type %q", a.Type)
	}
}

func (a Action) payloadCount() int {
	count := 0
	if a.RestartWorkload != nil {
		count++
	}
	if a.RollbackDeployment != nil {
		count++
	}
	if a.ScaleWorkload != nil {
		count++
	}
	if a.DrainNode != nil {
		count++
	}
	if a.FailoverDatabase != nil {
		count++
	}
	return count
}

type Proposal struct {
	ID         string     `json:"id"`
	IncidentID string     `json:"incident_id"`
	Hypothesis Hypothesis `json:"hypothesis"`
	Evidence   []EvidenceRef `json:"evidence"`

	// EstimatedRisk is advisory model/planner output only. It never grants
	// authority. The deterministic risk engine computes EffectiveRisk.
	EstimatedRisk *RiskClass `json:"estimated_risk,omitempty"`

	Action        Action   `json:"action"`
	Preconditions []string `json:"preconditions"`
	Verification  []string `json:"verification"`
	Rollback      *Action  `json:"rollback,omitempty"`
}

type HumanApproval struct {
	ApprovalID string `json:"approval_id"`
	ApprovedBy string `json:"approved_by"`
}

// ExecutionContext is trusted runtime state supplied by the control plane,
// identity/approval systems and operator controls. It is intentionally not part
// of model-generated Proposal data.
type ExecutionContext struct {
	Authority        AuthorityLevel  `json:"authority"`
	HumanApproval    *HumanApproval  `json:"human_approval,omitempty"`
	OperatorOverride bool            `json:"operator_override"`
	PolicyVersion    string          `json:"policy_version"`
	Environment      string          `json:"environment"`
	ActorID          string          `json:"actor_id"`
}

func (c ExecutionContext) HasHumanApproval() bool {
	return c.HumanApproval != nil &&
		strings.TrimSpace(c.HumanApproval.ApprovalID) != "" &&
		strings.TrimSpace(c.HumanApproval.ApprovedBy) != ""
}
