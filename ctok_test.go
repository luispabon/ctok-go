package ctok

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestCountAPI(t *testing.T) {
	for _, family := range []Family{"", "3.0", "4.7", "5.0", "4.8 "} {
		if counter, err := New(family); counter != nil || !errors.Is(err, ErrUnsupportedFamily) {
			t.Fatalf("family%q: counter%v err%v", family, counter, err)
		}
	}
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"", "\uf8ff", "\x01", "\ue000\x01\u0085"} {
		if counter.CountContent(text) != 0 || counter.Count(text) != 6 {
			t.Fatalf("stripped%q count differs", text)
		}
	}
	for _, text := range []string{"hello", " a", "123", "日本", "\n", "a\n\n", "Hello, world!", "\U000f0000"} {
		if counter.Count(text) != counter.CountContent(text)+6 {
			t.Fatalf("frame for%q differs", text)
		}
	}
	if counter.Count(" a") != counter.Count("a")+1 {
		t.Fatal("V48 absorbed leading space")
	}
	if counter.Count("a\ue000b") != counter.Count("ab") || counter.CountContent("\U000f0000") != 4 {
		t.Fatal("private-use behavior differs")
	}
	if counter.CountContent("hello")+counter.CountContent("world") == counter.CountContent("helloworld") {
		t.Fatal("non-additive witness no longer differs")
	}
}

func TestCountIssueDelta(t *testing.T) {
	payload := strings.Repeat("STEINER CONTEXT PROBE REFERENCE DATA 0123456789 abcdefghijklmnopqrstuvwxyz. ", 200)[:10000]
	sum := sha256.Sum256([]byte(payload))
	if hex.EncodeToString(sum[:]) != "31d2500f97354cdfad22b5e6acb757c0569965048f7a17fb427cbaf28817c53f" {
		t.Fatal("issue payload changed")
	}
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	if delta := counter.CountContent("Reply only OK."+payload) - counter.CountContent("Reply only OK."); delta != 6446 {
		t.Fatalf("delta%d want6446", delta)
	}
}

func TestCountInvalidBytes(t *testing.T) {
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"\xff", "a\xff\xfez", "\xf0\x80a\xff", "\xed\xa0\x80", "\xc0\xaf\x00\xff", "valid😀\xff\xfe"} {
		want := strings.ToValidUTF8(text, "\ufffd")
		if counter.Count(text) != counter.Count(want) {
			t.Fatalf("invalid input%x differs from replacement%q", text, want)
		}
	}
}

func TestCountConcurrent(t *testing.T) {
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	texts := []string{"hello world", "日本語\n", "\ufdd0a\n\n", "A a \u0301", ""}
	want := make([]int, len(texts))
	for i, text := range texts {
		want[i] = counter.Count(text)
	}
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			for range 100 {
				for i, text := range texts {
					if got := counter.Count(text); got != want[i] {
						t.Errorf("concurrent count%d want%d", got, want[i])
					}
				}
			}
		})
	}
	group.Wait()
}

func TestNewLoadError(t *testing.T) {
	loadErr := errors.New("load failed")
	calls := 0
	load := func() (*vocabulary, error) {
		calls++
		return nil, loadErr
	}
	if counter, err := newCounter("3.0", load); counter != nil || !errors.Is(err, ErrUnsupportedFamily) || calls != 0 {
		t.Fatalf("unsupported family: counter %v err %v calls %d", counter, err, calls)
	}
	counter, err := newCounter(V48, load)
	if counter != nil || !errors.Is(err, loadErr) || !strings.HasPrefix(err.Error(), "ctok: load vocabulary: ") || calls != 1 {
		t.Fatalf("load error: counter %v err %v calls %d", counter, err, calls)
	}
}

func TestNewDecodeError(t *testing.T) {
	var decodeErr error
	load := func() (*vocabulary, error) {
		_, decodeErr = decodeVocabulary(nil)
		return nil, decodeErr
	}
	counter, err := newCounter(V48, load)
	if counter != nil || decodeErr == nil || !errors.Is(err, decodeErr) || err.Error() != "ctok: load vocabulary: invalid vocabulary header or schema" {
		t.Fatalf("decode error: counter %v err %v decodeErr %v", counter, err, decodeErr)
	}
}
