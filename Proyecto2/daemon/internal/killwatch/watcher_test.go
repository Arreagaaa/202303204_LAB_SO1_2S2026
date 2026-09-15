package killwatch

import (
	"testing"
	"time"
)

func TestParseTraceLine(t *testing.T) {
	t.Parallel()

	event, ok := parseTraceLine("task-1234 [001] ..s1 123.456789: SO1_PR2_KILL sender=45 target=67 signal=9")
	if !ok {
		t.Fatal("parseTraceLine() did not match a valid trace line")
	}
	if event.SenderPID != 45 || event.TargetPID != 67 || event.Signal != 9 {
		t.Fatalf("parseTraceLine() = %#v", event)
	}
}

func TestPopPendingFindsTargetPID(t *testing.T) {
	t.Parallel()

	watcher := &Watcher{pending: []Event{{TargetPID: 11}, {TargetPID: 22}}}
	event, ok := watcher.popPending(22)
	if !ok || event.TargetPID != 22 {
		t.Fatalf("popPending() = %#v, %v", event, ok)
	}
	if len(watcher.pending) != 1 || watcher.pending[0].TargetPID != 11 {
		t.Fatalf("pending queue was not updated correctly: %#v", watcher.pending)
	}
}

func TestWaitForReturnsFalseOnTimeoutWithNoStream(t *testing.T) {
	t.Parallel()

	watcher := &Watcher{stream: make(chan Event), closed: make(chan struct{})}
	if _, ok := watcher.WaitFor(999, 10*time.Millisecond); ok {
		t.Fatal("WaitFor() should time out when no event is available")
	}
}
