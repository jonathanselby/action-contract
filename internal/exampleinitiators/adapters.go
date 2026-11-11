// Package exampleinitiators demonstrates transport-free human, API, MCP and
// event translation into one application-owned action contract.
package exampleinitiators

import (
	"context"

	actioncontract "github.com/jonathanselby/action-contract"
	"github.com/jonathanselby/action-contract/internal/exampleapp"
)

// Principal is supplied by the host's trusted security layer, never obtained
// from input claims. Supplying one is not itself authentication or permission.
type Principal struct {
	ActorID     string
	AuthorityID string
}

// UntrustedClaims model spoofable transport values for the security example.
// They cannot set action authority.
type UntrustedClaims struct {
	ActorID     string
	AuthorityID string
}

type HumanForm struct {
	CaseID string
	UntrustedClaims
}

type APIRequest struct {
	ResourceID string
	UntrustedClaims
}

type MCPArguments struct {
	Case string
	UntrustedClaims
}

type EventMessage struct {
	Subject string
	UntrustedClaims
}

type Trace struct {
	CorrelationID string
	CausationID   string
}

type Executor = func(context.Context, actioncontract.Action[exampleapp.ApproveCase]) (actioncontract.Result[exampleapp.Approval], error)

// All example initiators use the identical definition and execution helper.
var definition = actioncontract.Definition{Type: exampleapp.Type, Version: exampleapp.Version, Governed: true}

func Human(ctx context.Context, p Principal, t Trace, form HumanForm, execute Executor) (actioncontract.Result[exampleapp.Approval], error) {
	return run(ctx, p, t, form.CaseID, execute)
}

func API(ctx context.Context, p Principal, t Trace, request APIRequest, execute Executor) (actioncontract.Result[exampleapp.Approval], error) {
	return run(ctx, p, t, request.ResourceID, execute)
}

func MCP(ctx context.Context, p Principal, t Trace, args MCPArguments, execute Executor) (actioncontract.Result[exampleapp.Approval], error) {
	return run(ctx, p, t, args.Case, execute)
}

func Event(ctx context.Context, p Principal, t Trace, message EventMessage, execute Executor) (actioncontract.Result[exampleapp.Approval], error) {
	return run(ctx, p, t, message.Subject, execute)
}

func run(ctx context.Context, principal Principal, trace Trace, id string, execute Executor) (actioncontract.Result[exampleapp.Approval], error) {
	return actioncontract.Run(ctx, definition, actioncontract.Action[exampleapp.ApproveCase]{
		Meta: actioncontract.Meta{
			Type:          definition.Type,
			Version:       definition.Version,
			ActorID:       principal.ActorID,
			AuthorityID:   principal.AuthorityID,
			CorrelationID: trace.CorrelationID,
			CausationID:   trace.CausationID,
		},
		Input: exampleapp.ApproveCase{CaseID: id},
	}, execute)
}
