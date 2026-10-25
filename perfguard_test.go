//go:build perfguard && !race

package ctok

import (
	"runtime"
	"testing"
	"time"
)

const (
	perfSize         = 1 << 20
	perfRuns         = 5
	perfMaxRun       = time.Second
	perfMinMiBPerSec = 2.0
)

// TestPerfGuard is run only by the CI perf job, with -tags perfguard and
// -count=1.
func TestPerfGuard(t *testing.T) {
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	mixed := benchmarkInput("Hello 世界 Καλημέρα مرحبا नमस्ते 👩\u200d💻\n", perfSize)
	expanding := benchmarkInput("\u0344", perfSize)
	best := perfBestRun(t, counter, "mixed", mixed)
	if rate := float64(len(mixed)) / perfSize / best.Seconds(); rate < perfMinMiBPerSec {
		t.Errorf("mixed throughput %.2f MiB/s, want at least %.1f", rate, perfMinMiBPerSec)
	}
	perfBestRun(t, counter, "expanding", expanding)
}

// perfBestRun warms counter on text, then times perfRuns counts with GC run
// before each timer. It reports an error only after the loop, when the fastest
// run exceeds perfMaxRun.
func perfBestRun(t *testing.T, counter *Counter, name, text string) time.Duration {
	t.Helper()
	if counter.CountContent(text) == 0 {
		t.Fatalf("%s: warm-up count is zero", name)
	}
	best := time.Duration(1<<63 - 1)
	for range perfRuns {
		runtime.GC()
		start := time.Now()
		if counter.CountContent(text) == 0 {
			t.Fatalf("%s: count is zero", name)
		}
		best = min(best, time.Since(start))
	}
	if best > perfMaxRun {
		t.Errorf("%s: best of %d runs took %v, want at most %v", name, perfRuns, best, perfMaxRun)
	}
	return best
}
