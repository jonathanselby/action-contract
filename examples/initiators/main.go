package main

import (
	"context"
	"fmt"

	"github.com/jonathanselby/action-contract/internal/exampleapp"
	"github.com/jonathanselby/action-contract/internal/exampleinitiators"
)

func main() {
	ctx := context.Background()
	principal := exampleinitiators.Principal{ActorID: "verified-user", AuthorityID: "case-approver"}
	trace := exampleinitiators.Trace{CorrelationID: "correlation-123", CausationID: "event-456"}

	results := []struct {
		name   string
		status string
	}{}

	human, _ := exampleinitiators.Human(ctx, principal, trace, exampleinitiators.HumanForm{CaseID: "123"}, exampleapp.Execute)
	api, _ := exampleinitiators.API(ctx, principal, trace, exampleinitiators.APIRequest{ResourceID: "123"}, exampleapp.Execute)
	mcp, _ := exampleinitiators.MCP(ctx, principal, trace, exampleinitiators.MCPArguments{Case: "123"}, exampleapp.Execute)
	event, _ := exampleinitiators.Event(ctx, principal, trace, exampleinitiators.EventMessage{Subject: "123"}, exampleapp.Execute)
	results = append(results,
		struct{ name, status string }{"human", string(human.Outcome)},
		struct{ name, status string }{"api", string(api.Outcome)},
		struct{ name, status string }{"mcp", string(mcp.Outcome)},
		struct{ name, status string }{"event", string(event.Outcome)},
	)
	for _, r := range results {
		fmt.Printf("%s: %s\n", r.name, r.status)
	}
}
