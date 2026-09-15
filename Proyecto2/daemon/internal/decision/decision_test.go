package decision

import (
	"testing"

	"proyecto2/daemon/internal/model"
)

func TestSelectKillsHonorsMinimumsAndOrdering(t *testing.T) {
	t.Parallel()

	snapshot := model.Snapshot{
		Processes: []model.Process{
			{PID: 1, Name: "low-a", Class: model.ClassLow, RSSKB: 256, VSZKB: 512, MemoryPct: 40, CPUPct: 10},
			{PID: 2, Name: "low-b", Class: model.ClassLow, RSSKB: 1024, VSZKB: 2048, MemoryPct: 60, CPUPct: 90},
			{PID: 3, Name: "low-c", Class: model.ClassLow, RSSKB: 2048, VSZKB: 4096, MemoryPct: 70, CPUPct: 30},
			{PID: 4, Name: "high-a", Class: model.ClassHigh, RSSKB: 128, VSZKB: 256, MemoryPct: 50, CPUPct: 20},
			{PID: 5, Name: "high-b", Class: model.ClassHigh, RSSKB: 3072, VSZKB: 6144, MemoryPct: 75, CPUPct: 40},
			{PID: 6, Name: "grafana", Class: model.ClassGrafana, RSSKB: 4096, VSZKB: 8192, MemoryPct: 99, CPUPct: 99},
			{PID: 7, Name: "unknown-a", Class: model.ClassUnknown, RSSKB: 5120, VSZKB: 1024, MemoryPct: 80, CPUPct: 95},
			{PID: 8, Name: "unknown-b", Class: model.ClassUnknown, RSSKB: 4096, VSZKB: 2048, MemoryPct: 70, CPUPct: 50},
			{PID: 9, Name: "unknown-c", Class: model.ClassUnknown, RSSKB: 4096, VSZKB: 1024, MemoryPct: 60, CPUPct: 90},
		},
	}

	kills := SelectKills(snapshot, DefaultPolicy())
	if len(kills) != 3 {
		t.Fatalf("expected 3 candidates, got %d: %#v", len(kills), kills)
	}
	if kills[0].PID != 7 || kills[1].PID != 8 || kills[2].PID != 9 {
		t.Fatalf("unexpected kill order: %#v", kills)
	}
	for _, process := range kills {
		if process.Class == model.ClassGrafana {
			t.Fatalf("grafana should never be selected: %#v", process)
		}
	}
}
