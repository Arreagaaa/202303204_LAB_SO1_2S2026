package main

import (
	"testing"

	"proyecto2/daemon/internal/containerctl"
	"proyecto2/daemon/internal/model"
)

func TestEnrichSnapshotAppliesContainerLabels(t *testing.T) {
	t.Parallel()

	snapshot := model.Snapshot{
		Processes: []model.Process{
			{PID: 11, Name: "task", Command: "task", Class: model.ClassUnknown},
			{PID: 22, Name: "grafana", Command: "grafana", Class: model.ClassUnknown},
		},
	}

	enrichSnapshot(&snapshot, []containerctl.Container{
		{ID: "abc123", Name: "so1-pr2-low-1", PID: 11, Class: "low"},
		{ID: "def456", Name: "so1-pr2-grafana", PID: 22, Grafana: true},
	})

	if snapshot.Processes[0].Command != "abc123" || snapshot.Processes[0].Name != "so1-pr2-low-1" || snapshot.Processes[0].Class != model.ClassLow {
		t.Fatalf("unexpected low container enrichment: %#v", snapshot.Processes[0])
	}
	if snapshot.Processes[1].Class != model.ClassGrafana {
		t.Fatalf("unexpected grafana enrichment: %#v", snapshot.Processes[1])
	}
}
