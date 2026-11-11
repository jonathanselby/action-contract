// Package exampleapp is a tiny demonstration domain, not production business
// logic or a substitute for application authentication/authorisation.
package exampleapp

import (
	"context"
	"strings"

	actioncontract "github.com/jonathanselby/action-contract"
)

const (
	Type    = "case.approve"
	Version = "1"
)

// ApproveCase is the same typed input for every example initiator.
type ApproveCase struct {
	CaseID string
}

type Approval struct {
	CaseID string
	Status string
}

// Execute is the single application function used by all example adapters.
// Real applications enforce domain-level permissions here before side effects.
func Execute(ctx context.Context, action actioncontract.Action[ApproveCase]) (actioncontract.Result[Approval], error) {
	if err := ctx.Err(); err != nil {
		// No effect has started: the failure is known.
		return actioncontract.Result[Approval]{Outcome: actioncontract.Failed}, err
	}
	id := strings.TrimSpace(action.Input.CaseID)
	if id == "" || id == "restricted" {
		return actioncontract.Result[Approval]{Outcome: actioncontract.Refused}, nil
	}
	return actioncontract.Result[Approval]{
		Outcome:  actioncontract.Success,
		Output:   Approval{CaseID: id, Status: "approved"},
		Evidence: []actioncontract.EvidenceRef{{ID: "approval/" + id, Kind: "approval"}},
		Events:   []actioncontract.EventRef{{ID: "event/" + id, Type: "case.approved"}},
	}, nil
}
