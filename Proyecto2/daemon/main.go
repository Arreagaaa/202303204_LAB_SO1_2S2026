package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"proyecto2/daemon/internal/containerctl"
	"proyecto2/daemon/internal/decision"
	"proyecto2/daemon/internal/model"
	"proyecto2/daemon/internal/procreader"
)

func main() {
	procPath := flag.String("proc", "/proc/continfo_pr2_so1_202303204", "archivo de métricas del módulo")
	once := flag.Bool("once", false, "procesa una lectura y termina")
	interval := flag.Duration("interval", 20*time.Second, "tiempo entre lecturas")
	flag.Parse()

	policy := decision.DefaultPolicy()
	for {
		if err := processSnapshot(*procPath, policy); err != nil {
			log.Printf("lectura no disponible: %v", err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}

func processSnapshot(path string, policy decision.Policy) error {
	snapshot, err := procreader.Read(path)
	if err != nil {
		return err
	}

	if containers, err := containerctl.List(); err != nil {
		log.Printf("metadata de Docker no disponible: %v", err)
	} else {
		enrichSnapshot(&snapshot, containers)
	}

	kills := decision.SelectKills(snapshot, policy)
	fmt.Printf("RAM total=%dKB libre=%dKB usada=%dKB procesos=%d candidatos=%d\n",
		snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB,
		len(snapshot.Processes), len(kills))
	for _, process := range kills {
		fmt.Printf("candidato pid=%d nombre=%s cmd=%s clase=%s rss=%dKB cpu=%.2f%%\n",
			process.PID, process.Name, process.Command, process.Class, process.RSSKB, process.CPUPct)
	}
	return nil
}

func enrichSnapshot(snapshot *model.Snapshot, containers []containerctl.Container) {
	containersByPID := make(map[int]containerctl.Container, len(containers))
	for _, container := range containers {
		containersByPID[container.PID] = container
	}

	for index := range snapshot.Processes {
		container, ok := containersByPID[snapshot.Processes[index].PID]
		if !ok {
			continue
		}

		snapshot.Processes[index].Command = container.ID
		snapshot.Processes[index].Name = container.Name
		if container.Grafana {
			snapshot.Processes[index].Class = model.ClassGrafana
			continue
		}
		if container.Class == "" {
			snapshot.Processes[index].Class = model.ClassUnknown
			continue
		}
		snapshot.Processes[index].Class = model.WorkloadClass(container.Class)
	}
}
