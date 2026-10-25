//go:build perfguard && !race

package ctok

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	perfRuns         = 5
	perfMaxRun       = time.Second
	perfMinMiBPerSec = 2.0
)

// TestPerfGuard is run only by the CI perf job, with -tags perfguard and -count=1.
func TestPerfGuard(t *testing.T) {
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	mixed := strings.Repeat("Hello, 世界! café 🙂 naïve\u0301 x_y-42\n", (1<<20)/40+1)
	expanding := strings.Repeat("\u0344", 1<<19)
	best := perfBestRun(t, counter, "mixed", mixed)
	if rate := float64(len(mixed)) / (1 << 20) / best.Seconds(); rate < perfMinMiBPerSec {
		t.Errorf("mixed throughput %.2f MiB/s, want at least %.1f", rate, perfMinMiBPerSec)
	}
	perfBestRun(t, counter, "expanding", expanding)
}

// perfBestRun warms counter on text, then returns the fastest of perfRuns timed
// counts. Each timed run must finish within perfMaxRun. GC runs before each timer.
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
		elapsed := time.Since(start)
		if elapsed > perfMaxRun {
			t.Errorf("%s: run took %v, want at most %v", name, elapsed, perfMaxRun)
		}
		best = min(best, elapsed)
	}
	return best
}
