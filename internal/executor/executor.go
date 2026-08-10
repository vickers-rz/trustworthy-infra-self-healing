package executor

import (
	"context"
	"fmt"

	"github.com/vickers-rz/trustworthy-infra-self-healing/internal/domain"
)

type Result struct {
	ProposalID string
	Action     domain.ActionType
	DryRun     bool
	Message    string
}

type Executor interface {
	Execute(ctx context.Context, proposal domain.Proposal, dryRun bool) (Result, error)
}

// MockExecutor demonstrates the semantic-action boundary. It never executes
// model-generated shell text and is intentionally non-mutating.
type MockExecutor struct{}

func (MockExecutor) Execute(_ context.Context, p domain.Proposal, dryRun bool) (Result, error) {
	if err := p.Action.Validate(); err != nil {
		return Result{}, fmt.Errorf("invalid semantic action: %w", err)
	}
	return Result{
		ProposalID: p.ID,
		Action:     p.Action.Type,
		DryRun:     dryRun,
		Message:    "mock executor accepted a validated typed semantic action; no infrastructure was mutated",
	}, nil
}
