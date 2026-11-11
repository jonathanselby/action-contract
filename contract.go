// Package actioncontract provides a typed, transport-independent execution
// boundary. Callers own authentication, authorisation and domain behaviour.
package actioncontract

import (
	"context"
	"errors"
	"strings"
)

// Meta identifies an action and carries caller-resolved actor and trace
// references. References are not proof of identity or authority.
type Meta struct {
	Type          string
	Version       string
	ActorID       string
	AuthorityID   string
	CorrelationID string
	CausationID   string
}

// Action holds application-specific input without untyped payload encoding.
type Action[I any] struct {
	Meta  Meta
	Input I
}

// Outcome is the explicit application execution classification.
type Outcome string

const (
	Success       Outcome = "success"
	Refused       Outcome = "refused"
	Failed        Outcome = "failed"
	Indeterminate Outcome = "indeterminate"
)

// EvidenceRef references application-owned evidence without storing it.
type EvidenceRef struct {
	ID   string
	Kind string
}

// EventRef references an application-produced event without publishing it.
type EventRef struct {
	ID   string
	Type string
}

// Result holds the outcome, typed output and optional evidence/event references.
type Result[O any] struct {
	Meta     Meta
	Outcome  Outcome
	Output   O
	Evidence []EvidenceRef
	Events   []EventRef
}

// Definition is selected by trusted application code, never by an initiator.
// Governed requires non-blank actor and authority references; it does not
// authenticate or authorise them.
type Definition struct {
	Type     string
	Version  string
	Governed bool
}

var (
	ErrInvalidDefinition = errors.New("action contract: invalid definition")
	ErrNilContext        = errors.New("action contract: nil context")
	ErrNilExecutor       = errors.New("action contract: nil executor")
	ErrInvalidOutcome    = errors.New("action contract: invalid executor outcome")
)

// Run validates the trusted definition and shared metadata, then calls execute
// once with the original context and typed action. No transport, retries,
// authorisation logic or hidden state is introduced.
//
// Invalid action metadata is a deterministic refusal without execution.
// Invalid local configuration is failed with an error without execution.
// Contradictory/unknown executor outcomes become indeterminate, with a
// contract error and no claimed business output or event/evidence references.
// The application must classify uncertain external side effects itself.
func Run[I, O any](
	ctx context.Context,
	definition Definition,
	action Action[I],
	execute func(context.Context, Action[I]) (Result[O], error),
) (Result[O], error) {
	result := Result[O]{Meta: action.Meta}

	switch {
	case strings.TrimSpace(definition.Type) == "" || strings.TrimSpace(definition.Version) == "":
		result.Outcome = Failed
		return result, ErrInvalidDefinition
	case ctx == nil:
		result.Outcome = Failed
		return result, ErrNilContext
	case execute == nil:
		result.Outcome = Failed
		return result, ErrNilExecutor
	case strings.TrimSpace(action.Meta.Type) == "" || strings.TrimSpace(action.Meta.Version) == "" ||
		action.Meta.Type != definition.Type || action.Meta.Version != definition.Version:
		result.Outcome = Refused
		return result, nil
	case definition.Governed &&
		(strings.TrimSpace(action.Meta.ActorID) == "" || strings.TrimSpace(action.Meta.AuthorityID) == ""):
		result.Outcome = Refused
		return result, nil
	}

	result, err := execute(ctx, action)
	if !validOutcome(result.Outcome) ||
		((result.Outcome == Success || result.Outcome == Refused) && err != nil) {
		// Never present an uncertain/contradictory executor outcome as a
		// successful action. Do not leak untrusted business artefacts.
		return Result[O]{Meta: action.Meta, Outcome: Indeterminate}, errors.Join(ErrInvalidOutcome, err)
	}
	result.Meta = action.Meta
	return result, err
}

func validOutcome(outcome Outcome) bool {
	switch outcome {
	case Success, Refused, Failed, Indeterminate:
		return true
	default:
		return false
	}
}
