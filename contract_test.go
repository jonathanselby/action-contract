package actioncontract

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testInput struct{ ID string }
type testOutput struct{ ID string }

var trusted = Definition{Type: "case.approve", Version: "1", Governed: true}
var validAction = Action[testInput]{Meta: Meta{
	Type: "case.approve", Version: "1", ActorID: "verified-user", AuthorityID: "approved-role",
	CorrelationID: "cor-123", CausationID: "event-456",
}, Input: testInput{ID: "case-123"}}

func goodExecutor(_ context.Context, a Action[testInput]) (Result[testOutput], error) {
	return Result[testOutput]{Outcome: Success, Output: testOutput{ID: a.Input.ID}}, nil
}

func TestActionMetadataValidation(t *testing.T) {
	cases := []struct {
		name       string
		definition Definition
		meta       Meta
		want       Outcome
		wantErr    error
	}{
		{"valid", trusted, validAction.Meta, Success, nil},
		{"missing type", trusted, Meta{Version: "1"}, Refused, nil},
		{"missing version", trusted, Meta{Type: "case.approve"}, Refused, nil},
		{"wrong type", trusted, Meta{Type: "case.reject", Version: "1"}, Refused, nil},
		{"wrong version", trusted, Meta{Type: "case.approve", Version: "2"}, Refused, nil},
		{"whitespace version", trusted, Meta{Type: "case.approve", Version: " "}, Refused, nil},
		{"bad expected type", Definition{Type: " ", Version: "1"}, validAction.Meta, Failed, ErrInvalidDefinition},
		{"bad expected version", Definition{Type: "case.approve"}, validAction.Meta, Failed, ErrInvalidDefinition},
		{"ungoverned reference optional", Definition{Type: "case.approve", Version: "1"}, Meta{Type: "case.approve", Version: "1"}, Success, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got, err := Run(context.Background(), tc.definition, Action[testInput]{Meta: tc.meta}, func(ctx context.Context, action Action[testInput]) (Result[testOutput], error) {
				calls++
				return goodExecutor(ctx, action)
			})
			if got.Outcome != tc.want || !errors.Is(err, tc.wantErr) {
				t.Fatalf("got outcome=%q err=%v, want outcome=%q err=%v", got.Outcome, err, tc.want, tc.wantErr)
			}
			if wantCalls := 0; tc.want == Success {
				wantCalls = 1
				if calls != wantCalls {
					t.Fatalf("execute called %d times; want %d", calls, wantCalls)
				}
			} else if calls != 0 {
				t.Fatalf("invalid identity executed %d times", calls)
			}
		})
	}
}

func TestInvalidMetadataRefused(t *testing.T) {
	cases := []struct {
		name string
		meta Meta
	}{
		{"empty actor", Meta{Type: trusted.Type, Version: trusted.Version, AuthorityID: "role"}},
		{"blank actor", Meta{Type: trusted.Type, Version: trusted.Version, ActorID: " \t", AuthorityID: "role"}},
		{"empty authority", Meta{Type: trusted.Type, Version: trusted.Version, ActorID: "actor"}},
		{"blank authority", Meta{Type: trusted.Type, Version: trusted.Version, ActorID: "actor", AuthorityID: "  "}},
		{"missing identity", Meta{Type: "different", Version: trusted.Version, ActorID: "actor", AuthorityID: "role"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			got, err := Run(context.Background(), trusted, Action[testInput]{Meta: tc.meta}, func(context.Context, Action[testInput]) (Result[testOutput], error) {
				called = true
				return Result[testOutput]{Outcome: Success}, nil
			})
			if err != nil || got.Outcome != Refused || called || got.Meta != tc.meta {
				t.Fatalf("invalid metadata: got=%+v err=%v called=%v", got, err, called)
			}
		})
	}
}

func TestContextMetadataPreserved(t *testing.T) {
	original := context.WithValue(context.Background(), contextKey{}, "original")
	seen := Meta{}
	got, err := Run(original, trusted, validAction, func(ctx context.Context, action Action[testInput]) (Result[testOutput], error) {
		if ctx != original || ctx.Value(contextKey{}) != "original" {
			t.Fatal("context was replaced or lost")
		}
		seen = action.Meta
		return Result[testOutput]{Meta: Meta{Type: "forged", ActorID: "attacker", AuthorityID: "superuser"}, Outcome: Success}, nil
	})
	if err != nil || got.Outcome != Success || seen != validAction.Meta || got.Meta != validAction.Meta {
		t.Fatalf("metadata was not preserved: %+v, err=%v, executor saw=%+v", got, err, seen)
	}
}

type contextKey struct{}

func TestOutcomeClassification(t *testing.T) {
	errExternal := errors.New("external timeout")
	cases := []struct {
		name            string
		outcome         Outcome
		execErr         error
		want            Outcome
		wantContractErr bool
		keepArtifacts   bool
	}{
		{"success", Success, nil, Success, false, true},
		{"refused", Refused, nil, Refused, false, true},
		{"failed", Failed, errExternal, Failed, false, true},
		{"failed without error", Failed, nil, Failed, false, true},
		{"indeterminate", Indeterminate, errExternal, Indeterminate, false, true},
		{"indeterminate without error", Indeterminate, nil, Indeterminate, false, true},
		{"success with error", Success, errExternal, Indeterminate, true, false},
		{"refused with error", Refused, errExternal, Indeterminate, true, false},
		{"missing", "", nil, Indeterminate, true, false},
		{"unknown", "sometimes", nil, Indeterminate, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Run(context.Background(), trusted, validAction, func(context.Context, Action[testInput]) (Result[testOutput], error) {
				return Result[testOutput]{Outcome: tc.outcome, Output: testOutput{ID: "claimed"}, Evidence: []EvidenceRef{{ID: "e"}}, Events: []EventRef{{ID: "evt"}}}, tc.execErr
			})
			if got.Outcome != tc.want || errors.Is(err, ErrInvalidOutcome) != tc.wantContractErr {
				t.Fatalf("outcome=%q err=%v; wanted %q contractError=%v", got.Outcome, err, tc.want, tc.wantContractErr)
			}
			if got.Meta != validAction.Meta {
				t.Fatal("result metadata not preserved")
			}
			if tc.keepArtifacts {
				if got.Output.ID != "claimed" || len(got.Events) != 1 || len(got.Evidence) != 1 {
					t.Fatal("valid result artefacts discarded")
				}
			} else if got.Output != (testOutput{}) || len(got.Events) != 0 || len(got.Evidence) != 0 {
				t.Fatal("invalid result leaked success artefacts")
			}
			if tc.execErr != nil && !errors.Is(err, tc.execErr) {
				t.Fatalf("original diagnostic error lost: %v", err)
			}
		})
	}
}

func TestEvidenceReferences(t *testing.T) {
	want := []EvidenceRef{{ID: "receipt/123", Kind: "receipt"}, {ID: "record/456", Kind: "document"}}
	got, err := Run(context.Background(), trusted, validAction, func(context.Context, Action[testInput]) (Result[testOutput], error) {
		return Result[testOutput]{Outcome: Success, Evidence: want}, nil
	})
	if err != nil || !reflect.DeepEqual(got.Evidence, want) {
		t.Fatalf("evidence references lost: %+v %v", got, err)
	}
}

func TestEventReferences(t *testing.T) {
	want := []EventRef{{ID: "event/123", Type: "case.approved"}}
	got, err := Run(context.Background(), trusted, validAction, func(context.Context, Action[testInput]) (Result[testOutput], error) {
		return Result[testOutput]{Outcome: Success, Events: want}, nil
	})
	if err != nil || !reflect.DeepEqual(got.Events, want) {
		t.Fatalf("event references lost: %+v %v", got, err)
	}
}

func TestContextCancellation(t *testing.T) {
	t.Run("pre-cancel reaches executor", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		called := false
		got, err := Run(ctx, trusted, validAction, func(received context.Context, _ Action[testInput]) (Result[testOutput], error) {
			called = true
			if received != ctx || !errors.Is(received.Err(), context.Canceled) {
				t.Fatal("original cancelled context not forwarded")
			}
			return Result[testOutput]{Outcome: Failed}, received.Err()
		})
		if !called || got.Outcome != Failed || !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled context result: %+v, %v", got, err)
		}
	})
	t.Run("deadline survives", func(t *testing.T) {
		deadline := time.Now().Add(30 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		_, err := Run(ctx, trusted, validAction, func(received context.Context, _ Action[testInput]) (Result[testOutput], error) {
			gotDeadline, ok := received.Deadline()
			if received != ctx || !ok || !gotDeadline.Equal(deadline) {
				t.Fatal("deadline or context changed")
			}
			return Result[testOutput]{Outcome: Success}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ambiguous timeout after effect", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		got, err := Run(ctx, trusted, validAction, func(received context.Context, _ Action[testInput]) (Result[testOutput], error) {
			cancel() // Simulate cancellation after an uncertain remote effect.
			return Result[testOutput]{Outcome: Indeterminate}, received.Err()
		})
		if got.Outcome != Indeterminate || !errors.Is(err, context.Canceled) {
			t.Fatalf("unknown state incorrectly classified: %+v %v", got, err)
		}
	})
}

func TestInvalidConfigurationFailsClosed(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		def  Definition
		exec func(context.Context, Action[testInput]) (Result[testOutput], error)
		want error
	}{
		{"nil context", nil, trusted, goodExecutor, ErrNilContext},
		{"nil executor", context.Background(), trusted, nil, ErrNilExecutor},
		{"invalid definition", context.Background(), Definition{}, goodExecutor, ErrInvalidDefinition},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Run(tc.ctx, tc.def, validAction, tc.exec)
			if got.Outcome != Failed || !errors.Is(err, tc.want) {
				t.Fatalf("expected failed with %v: %+v %v", tc.want, got, err)
			}
		})
	}
}

func TestConcurrentExecution(t *testing.T) {
	const workers = 64
	var count atomic.Int64
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := Run(context.Background(), trusted, validAction, func(ctx context.Context, a Action[testInput]) (Result[testOutput], error) {
				count.Add(1)
				return goodExecutor(ctx, a)
			})
			if err != nil || result.Outcome != Success || result.Meta != validAction.Meta {
				t.Errorf("concurrent result invalid: %+v %v", result, err)
			}
		}()
	}
	wg.Wait()
	if count.Load() != workers {
		t.Fatalf("calls = %d, want %d", count.Load(), workers)
	}
}

func TestCoreDependencyBoundary(t *testing.T) {
	allowed := map[string]bool{"context": true, "errors": true, "strings": true}
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".go" || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}
		seen++
		f, err := parser.ParseFile(token.NewFileSet(), file.Name(), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil || !allowed[path] {
				t.Errorf("unexpected core dependency %q in %s", path, file.Name())
			}
		}
	}
	if seen == 0 {
		t.Fatal("no root Go source inspected")
	}
	module, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(module), "require ") || strings.Contains(string(module), "replace ") {
		t.Fatal("core module unexpectedly contains external dependencies")
	}
}

func TestVersionMatchesFile(t *testing.T) {
	b, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) != Version {
		t.Fatalf("VERSION mismatch: file=%q constant=%q", strings.TrimSpace(string(b)), Version)
	}
}
