package actioncontract

import (
	"context"
	"testing"
)

var benchmarkResult Result[testOutput]

// Benchmarks measure call-path cost, not a promised timing guarantee.
func BenchmarkDirectExecute(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = goodExecutor(ctx, validAction)
	}
}

func BenchmarkRun(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = Run(ctx, trusted, validAction, goodExecutor)
	}
}
