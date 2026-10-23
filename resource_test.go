package ctok

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestResourceColdProcess(t *testing.T) {
	if os.Getenv("CTOK_COLD_CHILD") == "1" {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start := time.Now()
		counter, err := New(V48)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		runtime.GC()
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(counter)
		retained := uint64(0)
		if after.HeapAlloc > before.HeapAlloc {
			retained = after.HeapAlloc - before.HeapAlloc
		}
		if err := json.NewEncoder(os.Stdout).Encode(struct {
			Nanos    int64
			Retained uint64
		}{elapsed.Nanoseconds(), retained}); err != nil {
			t.Fatal(err)
		}
		if elapsed > 150*time.Millisecond || retained > 48<<20 {
			t.Fatalf("cold load%s retained%d exceeds gate", elapsed, retained)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestResourceColdProcess$")
	command.Env = append(os.Environ(), "CTOK_COLD_CHILD=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cold subprocess: %v\n%s", err, output)
	}
	var report struct {
		Nanos    int64
		Retained uint64
	}
	if err := json.Unmarshal([]byte(strings.SplitN(string(output), "\n", 2)[0]), &report); err != nil {
		t.Fatalf("cold report: %v\n%s", err, output)
	}
	if report.Nanos <= 0 || report.Retained == 0 {
		t.Fatal("missing cold measurements")
	}
	t.Logf("cold load=%s retained=%d bytes", time.Duration(report.Nanos), report.Retained)
}

// TestResourceAllocations measures all transient bytes for one call per size.
// Run without -race for timing evidence; instrumentation changes runtime cost.
func TestResourceAllocations(t *testing.T) {
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, atom string }{
		{"ascii", "func Count(s string) int { return len(s) }\n"},
		{"mixed", "Hello 世界 Καλημέρα مرحبا नमस्ते 👩\u200d💻\n"},
		{"short_words", "A a "}, {"dense_upper", "A,"}, {"dense_lower", "a."},
		{"punct_space", " , "}, {"upper_space", "A, "}, {"stray", "\u0345A,"}, {"digits", " १ ! "},
		{"combining", "\u0315\u0300"}, {"newlines", "\n"},
		{"exclusion", "\u0344"}, {"tibetan", "\u0f73"}, {"tibetan75", "\u0f75"}, {"tibetan81", "\u0f81"},
		{"singleton40", "\u0340"}, {"singleton41", "\u0341"}, {"singleton43", "\u0343"},
		{"hebrew2c", "\ufb2c"}, {"hebrew2d", "\ufb2d"}, {"musical", "\U0001d160"},
		{"devanagari", "\u0958"}, {"hangul", "\u1100\u1161\u11a8"}, {"mixed_expansion", "\u0344"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var allocated [2]int64
			for i, size := range []int{1 << 20, 4 << 20} {
				text := benchmarkInput(tt.atom, size)
				if tt.name == "mixed_expansion" {
					text = strings.Repeat("\u0344", (size-3)/2) + "\u0f73" + strings.Repeat("a", (size-3)%2)
				}
				// A timed one-call measurement includes all transient allocations.
				// GC is completed before the measurement, not during input creation.
				runtime.GC()
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				start := time.Now()
				got := counter.CountContent(text)
				elapsed := time.Since(start)
				runtime.ReadMemStats(&after)
				runtime.KeepAlive(text)
				if got < 0 {
					t.Fatal("negative content count")
				}
				allocated[i] = int64(after.TotalAlloc - before.TotalAlloc)
				budget := int64(24*len(text) + (1 << 20))
				if allocated[i] > budget {
					t.Fatalf("size%d allocated%d exceeds%d", size, allocated[i], budget)
				}
				t.Logf("size%d bytes%d elapsed%s", size, allocated[i], elapsed)
			}
			if allocated[1]-allocated[0] > 24*(3<<20) {
				t.Fatalf("allocation slope exceeds24: %d to%d", allocated[0], allocated[1])
			}
		})
	}
}
