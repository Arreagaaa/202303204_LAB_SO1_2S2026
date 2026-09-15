package killwatch

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const tracePipePath = "/sys/kernel/debug/tracing/trace_pipe"
const tracePrefix = "SO1_PR2_KILL"

var traceLinePattern = regexp.MustCompile(`SO1_PR2_KILL sender=([0-9]+) target=([0-9]+) signal=([0-9]+)`)

type Event struct {
	TimestampNS uint64
	SenderPID   uint32
	TargetPID   uint32
	Signal      uint32
}

type Watcher struct {
	file    *os.File
	stream  chan Event
	pending []Event
	mu      sync.Mutex
	total   uint64
	closed  chan struct{}
	once    sync.Once
}

func Start(objectPath string) (*Watcher, error) {
	if objectPath != "" {
		if _, err := os.Stat(objectPath); err != nil {
			return nil, fmt.Errorf("verificar objeto eBPF: %w", err)
		}
	}
	file, err := os.Open(tracePipePath)
	if err != nil {
		return nil, fmt.Errorf("abrir trace_pipe: %w", err)
	}
	watcher := &Watcher{
		file:   file,
		stream: make(chan Event, 128),
		closed: make(chan struct{}),
	}
	go watcher.run()
	return watcher, nil
}

func (watcher *Watcher) run() {
	defer close(watcher.stream)
	reader := bufio.NewScanner(watcher.file)
	reader.Buffer(make([]byte, 0, 1024), 64*1024)
	for {
		if !reader.Scan() {
			if err := reader.Err(); err != nil && !errors.Is(err, os.ErrClosed) {
				return
			}
			return
		}
		event, ok := parseTraceLine(reader.Text())
		if !ok {
			continue
		}
		atomic.AddUint64(&watcher.total, 1)
		select {
		case watcher.stream <- event:
		case <-watcher.closed:
			return
		}
	}
}

func (watcher *Watcher) Close() error {
	if watcher == nil {
		return nil
	}
	watcher.once.Do(func() {
		close(watcher.closed)
	})
	if watcher.file != nil {
		_ = watcher.file.Close()
	}
	return nil
}

func (watcher *Watcher) Total() uint64 {
	if watcher == nil {
		return 0
	}
	return atomic.LoadUint64(&watcher.total)
}

func (watcher *Watcher) WaitFor(targetPID uint32, timeout time.Duration) (Event, bool) {
	if watcher == nil {
		return Event{}, false
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		if event, ok := watcher.popPending(targetPID); ok {
			return event, true
		}
		select {
		case event, ok := <-watcher.stream:
			if !ok {
				return Event{}, false
			}
			watcher.mu.Lock()
			watcher.pending = append(watcher.pending, event)
			watcher.mu.Unlock()
		case <-deadline.C:
			return Event{}, false
		}
	}
}

func (watcher *Watcher) popPending(targetPID uint32) (Event, bool) {
	watcher.mu.Lock()
	defer watcher.mu.Unlock()
	for index, event := range watcher.pending {
		if event.TargetPID != targetPID {
			continue
		}
		watcher.pending = append(watcher.pending[:index], watcher.pending[index+1:]...)
		return event, true
	}
	return Event{}, false
}

func parseTraceLine(line string) (Event, bool) {
	if !strings.Contains(line, tracePrefix) {
		return Event{}, false
	}
	matches := traceLinePattern.FindStringSubmatch(line)
	if len(matches) != 4 {
		return Event{}, false
	}
	sender, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return Event{}, false
	}
	target, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return Event{}, false
	}
	signal, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return Event{}, false
	}
	return Event{
		TimestampNS: uint64(time.Now().UnixNano()),
		SenderPID:   uint32(sender),
		TargetPID:   uint32(target),
		Signal:      uint32(signal),
	}, true
}
