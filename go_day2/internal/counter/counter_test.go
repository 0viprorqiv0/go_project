package counter

import "testing"

func TestEventCounter(t *testing.T) {
	counter := New()

	if got := counter.Add("192.0.2.10"); got != 1 {
		t.Fatalf("first Add()=%d, want 1", got)
	}

	if got := counter.Add("192.0.2.10"); got != 2 {
		t.Fatalf("second Add()=%d, want 2", got)
	}

	got, exists := counter.Count("192.0.2.10")
	if !exists || got != 2 {
		t.Fatalf("Count()=(%d, %v), want (2, true)", got, exists)
	}
}

func TestEventCounterMissingIP(t *testing.T) {
	counter := New()

	got, exists := counter.Count("203.0.113.7")
	if exists || got != 0 {
		t.Fatalf("Count()=(%d, %v), want (0, false)", got, exists)
	}
}

func TestEventCounterEmptyIP(t *testing.T) {
	counter := New()

	if got := counter.Add(""); got != 0 {
		t.Fatalf("Add(empty)=%d, want 0", got)
	}

	if _, exists := counter.Count(""); exists {
		t.Fatal("empty IP should not be stored")
	}
}
