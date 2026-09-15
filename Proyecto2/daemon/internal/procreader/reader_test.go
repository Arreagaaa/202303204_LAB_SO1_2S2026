package procreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadParsesSnapshot(t *testing.T) {
	t.Parallel()

	content := "MEM_TOTAL_KB:8192\nMEM_FREE_KB:2048\nMEM_USED_KB:6144\nPID,NAME,CMD_OR_CONTAINER_ID,VSZ_KB,RSS_KB,MEM_PCT,CPU_PCT,CLASS\n101,worker,/bin/sh,2048,1024,12.50,4.25,low\n202,grafana,/usr/share/grafana,4096,2048,30.00,10.00,grafana\n"
	path := filepath.Join(t.TempDir(), "snapshot.proc")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	snapshot, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if snapshot.Memory.TotalKB != 8192 || snapshot.Memory.FreeKB != 2048 || snapshot.Memory.UsedKB != 6144 {
		t.Fatalf("unexpected memory snapshot: %#v", snapshot.Memory)
	}
	if len(snapshot.Processes) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(snapshot.Processes))
	}
	if snapshot.Processes[0].PID != 101 || snapshot.Processes[0].Class != "low" {
		t.Fatalf("unexpected first process: %#v", snapshot.Processes[0])
	}
	if snapshot.Processes[1].PID != 202 || snapshot.Processes[1].Class != "grafana" {
		t.Fatalf("unexpected second process: %#v", snapshot.Processes[1])
	}
}
