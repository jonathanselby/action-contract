package exampleinitiators

import (
	"context"
	"errors"
	"reflect"
	"testing"

	actioncontract "github.com/jonathanselby/action-contract"
	"github.com/jonathanselby/action-contract/internal/exampleapp"
)

func TestMultipleInitiatorsSameExecution(t *testing.T) {
	principal := Principal{ActorID: "verified-actor", AuthorityID: "approved-role"}
	trace := Trace{CorrelationID: "corr-42", CausationID: "cause-17"}
	claims := UntrustedClaims{ActorID: "spoofed-admin", AuthorityID: "root"}
	var called int
	var seen []actioncontract.Action[exampleapp.ApproveCase]
	fn := func(ctx context.Context, a actioncontract.Action[exampleapp.ApproveCase]) (actioncontract.Result[exampleapp.Approval], error) {
		called++
		seen = append(seen, a)
		return exampleapp.Execute(ctx, a)
	}
	invocations := []struct {
		name string
		run  func() (actioncontract.Result[exampleapp.Approval], error)
	}{
		{"human", func() (actioncontract.Result[exampleapp.Approval], error) {
			return Human(context.Background(), principal, trace, HumanForm{CaseID: "42", UntrustedClaims: claims}, fn)
		}},
		{"api", func() (actioncontract.Result[exampleapp.Approval], error) {
			return API(context.Background(), principal, trace, APIRequest{ResourceID: "42", UntrustedClaims: claims}, fn)
		}},
		{"mcp", func() (actioncontract.Result[exampleapp.Approval], error) {
			return MCP(context.Background(), principal, trace, MCPArguments{Case: "42", UntrustedClaims: claims}, fn)
		}},
		{"event", func() (actioncontract.Result[exampleapp.Approval], error) {
			return Event(context.Background(), principal, trace, EventMessage{Subject: "42", UntrustedClaims: claims}, fn)
		}},
	}
	var first actioncontract.Result[exampleapp.Approval]
	for i, tc := range invocations {
		got, err := tc.run()
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if i == 0 {
			first = got
		} else if !reflect.DeepEqual(first, got) {
			t.Fatalf("%s produced different result: %+v vs %+v", tc.name, got, first)
		}
		if got.Outcome != actioncontract.Success || got.Meta.ActorID != principal.ActorID || got.Meta.AuthorityID != principal.AuthorityID || got.Meta.CorrelationID != trace.CorrelationID || got.Meta.CausationID != trace.CausationID || len(got.Evidence) != 1 || len(got.Events) != 1 {
			t.Fatalf("%s lost trusted metadata or references: %+v", tc.name, got)
		}
	}
	if called != 4 || len(seen) != 4 {
		t.Fatalf("same executor must be invoked four times: %d", called)
	}
	for _, a := range seen {
		if !reflect.DeepEqual(a, seen[0]) || a.Meta.ActorID == claims.ActorID || a.Meta.AuthorityID == claims.AuthorityID {
			t.Fatalf("adapter diverged or accepted spoofed claims: %+v", a)
		}
	}
}

func TestForgedAuthorityCannotElevate(t *testing.T) {
	spoof := UntrustedClaims{ActorID: "administrator", AuthorityID: "all-access"}
	var called int
	fn := func(ctx context.Context, a actioncontract.Action[exampleapp.ApproveCase]) (actioncontract.Result[exampleapp.Approval], error) {
		called++
		return exampleapp.Execute(ctx, a)
	}
	invokers := []func() (actioncontract.Result[exampleapp.Approval], error){
		func() (actioncontract.Result[exampleapp.Approval], error) {
			return Human(context.Background(), Principal{}, Trace{}, HumanForm{CaseID: "42", UntrustedClaims: spoof}, fn)
		},
		func() (actioncontract.Result[exampleapp.Approval], error) {
			return API(context.Background(), Principal{}, Trace{}, APIRequest{ResourceID: "42", UntrustedClaims: spoof}, fn)
		},
		func() (actioncontract.Result[exampleapp.Approval], error) {
			return MCP(context.Background(), Principal{}, Trace{}, MCPArguments{Case: "42", UntrustedClaims: spoof}, fn)
		},
		func() (actioncontract.Result[exampleapp.Approval], error) {
			return Event(context.Background(), Principal{}, Trace{}, EventMessage{Subject: "42", UntrustedClaims: spoof}, fn)
		},
	}
	for i, invoke := range invokers {
		got, err := invoke()
		if err != nil || got.Outcome != actioncontract.Refused {
			t.Fatalf("initiator %d elevated claims: %+v %v", i, got, err)
		}
	}
	if called != 0 {
		t.Fatalf("forged authority reached application: %d calls", called)
	}
}

func TestExampleDomainRefusalAndCancellation(t *testing.T) {
	principal := Principal{ActorID: "verified", AuthorityID: "approver"}
	got, err := Human(context.Background(), principal, Trace{}, HumanForm{CaseID: "restricted"}, exampleapp.Execute)
	if err != nil || got.Outcome != actioncontract.Refused {
		t.Fatalf("domain refusal broken: %+v %v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err = MCP(ctx, principal, Trace{}, MCPArguments{Case: "42"}, exampleapp.Execute)
	if got.Outcome != actioncontract.Failed || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation failed: %+v %v", got, err)
	}
}
