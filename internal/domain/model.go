package domain

import "fmt"

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

type Action struct {
	Type   ActionType        `json:"type"`
	Target string            `json:"target"`
	Params map[string]string `json:"params,omitempty"`
}

type Rollback struct {
	Strategy string            `json:"strategy"`
	Params   map[string]string `json:"params,omitempty"`
}

type Proposal struct {
	ID               string         `json:"id"`
	IncidentID       string         `json:"incident_id"`
	Authority        AuthorityLevel `json:"authority"`
	Risk             RiskClass      `json:"risk"`
	Hypothesis       Hypothesis     `json:"hypothesis"`
	Evidence         []EvidenceRef  `json:"evidence"`
	Action           Action         `json:"action"`
	Preconditions    []string       `json:"preconditions"`
	Verification     []string       `json:"verification"`
	Rollback         *Rollback      `json:"rollback,omitempty"`
	HumanApproved    bool           `json:"human_approved"`
	OperatorOverride bool           `json:"operator_override"`
	PolicyVersion    string         `json:"policy_version"`
}
