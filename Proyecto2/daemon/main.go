package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"proyecto2/daemon/internal/containerctl"
	"proyecto2/daemon/internal/decision"
	"proyecto2/daemon/internal/killwatch"
	"proyecto2/daemon/internal/model"
	"proyecto2/daemon/internal/procreader"
	"proyecto2/daemon/internal/valkey"
)

func main() {
	procPath := flag.String("proc", "/proc/continfo_pr2_so1_202303204", "archivo de métricas del módulo")
	valkeyAddr := flag.String("valkey", "127.0.0.1:6379", "dirección del servicio Valkey")
	ebpfObject := flag.String("ebpf-object", "../ebpf-probe/kill_tracer.bpf.o", "objeto eBPF para confirmación de kills")
	once := flag.Bool("once", false, "procesa una lectura y termina")
	interval := flag.Duration("interval", 20*time.Second, "tiempo entre lecturas")
	killTimeout := flag.Duration("kill-timeout", 5*time.Second, "tiempo máximo para esperar la confirmación eBPF")
	flag.Parse()

	policy := decision.DefaultPolicy()
	client, err := valkey.Dial(*valkeyAddr)
	if err != nil {
		log.Printf("valkey no disponible en %s: %v", *valkeyAddr, err)
	}
	if client != nil {
		defer client.Close()
	}
	watcher, err := killwatch.Start(*ebpfObject)
	if err != nil {
		log.Printf("eBPF no disponible en %s: %v", *ebpfObject, err)
	} else if watcher != nil {
		defer watcher.Close()
	}
	for {
		if err := processSnapshot(*procPath, policy, client, watcher, *killTimeout); err != nil {
			log.Printf("lectura no disponible: %v", err)
		}
		if *once {
			return
		}
		time.Sleep(*interval)
	}
}

func processSnapshot(path string, policy decision.Policy, client *valkey.Client, watcher *killwatch.Watcher, killTimeout time.Duration) error {
	snapshot, err := procreader.Read(path)
	if err != nil {
		return err
	}

	containers, err := containerctl.List()
	if err != nil {
		log.Printf("metadata de Docker no disponible: %v", err)
	}
	containersByPID := mapContainersByPID(containers)
	if len(containers) > 0 {
		enrichSnapshot(&snapshot, containersByPID)
	}

	kills := decision.SelectKills(snapshot, policy)
	confirmedKills := stopSelectedContainers(containersByPID, kills, watcher, killTimeout)
	if client != nil {
		if err := persistSnapshot(client, snapshot, kills, confirmedKills, watcher); err != nil {
			log.Printf("valkey no disponible: %v", err)
		}
	}
	fmt.Printf("RAM total=%dKB libre=%dKB usada=%dKB procesos=%d candidatos=%d\n",
		snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB,
		len(snapshot.Processes), len(kills))
	for _, process := range kills {
		fmt.Printf("candidato pid=%d nombre=%s cmd=%s clase=%s rss=%dKB cpu=%.2f%%\n",
			process.PID, process.Name, process.Command, process.Class, process.RSSKB, process.CPUPct)
	}
	return nil
}

func persistSnapshot(client *valkey.Client, snapshot model.Snapshot, kills []model.Process, confirmedKills []killwatch.Event, watcher *killwatch.Watcher) error {
	if err := client.Set("system:ram:current", fmt.Sprintf("total=%dKB,free=%dKB,used=%dKB", snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB)); err != nil {
		return err
	}
	if err := client.RPush("system:ram:history", fmt.Sprintf("%d|%d|%d", snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB)); err != nil {
		return err
	}
	if err := client.SetInteger("containers:killed:count", uint64(len(confirmedKills))); err != nil {
		return err
	}
	if err := client.RPush("containers:killed:history", formatKillEvents(confirmedKills)); err != nil {
		return err
	}
	if err := client.RPush("containers:ranking:ram", formatTopProcesses(snapshot.Processes, func(left, right model.Process) bool {
		if left.RSSKB != right.RSSKB {
			return left.RSSKB > right.RSSKB
		}
		if left.VSZKB != right.VSZKB {
			return left.VSZKB > right.VSZKB
		}
		return left.CPUPct > right.CPUPct
	})); err != nil {
		return err
	}
	if err := client.RPush("containers:ranking:cpu", formatTopProcesses(snapshot.Processes, func(left, right model.Process) bool {
		if left.CPUPct != right.CPUPct {
			return left.CPUPct > right.CPUPct
		}
		if left.RSSKB != right.RSSKB {
			return left.RSSKB > right.RSSKB
		}
		return left.VSZKB > right.VSZKB
	})); err != nil {
		return err
	}
	if watcher != nil {
		if err := client.SetInteger("ebpf:events:count", watcher.Total()); err != nil {
			return err
		}
	} else {
		if err := client.SetInteger("ebpf:events:count", 0); err != nil {
			return err
		}
	}
	return nil
}

func formatKillEvents(events []killwatch.Event) string {
	if len(events) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(events))
	for _, event := range events {
		parts = append(parts, fmt.Sprintf("target=%d signal=%d sender=%d ts=%d", event.TargetPID, event.Signal, event.SenderPID, event.TimestampNS))
	}
	return joinWithComma(parts)
}

func mapContainersByPID(containers []containerctl.Container) map[int]containerctl.Container {
	containersByPID := make(map[int]containerctl.Container, len(containers))
	for _, container := range containers {
		containersByPID[container.PID] = container
	}
	return containersByPID
}

func stopSelectedContainers(containersByPID map[int]containerctl.Container, kills []model.Process, watcher *killwatch.Watcher, killTimeout time.Duration) []killwatch.Event {
	confirmed := make([]killwatch.Event, 0, len(kills))
	for _, process := range kills {
		container, ok := containersByPID[process.PID]
		if !ok || container.Grafana {
			continue
		}
		if err := containerctl.Stop(container.ID); err != nil {
			log.Printf("no se pudo detener %s: %v", container.ID, err)
			continue
		}
		if watcher == nil {
			continue
		}
		event, ok := watcher.WaitFor(uint32(container.PID), killTimeout)
		if !ok {
			log.Printf("sin confirmación eBPF para pid=%d contenedor=%s", container.PID, container.ID)
			continue
		}
		confirmed = append(confirmed, event)
		log.Printf("contenedor confirmado por eBPF pid=%d signal=%d sender=%d", event.TargetPID, event.Signal, event.SenderPID)
	}
	return confirmed
}

func formatTopProcesses(processes []model.Process, less func(model.Process, model.Process) bool) string {
	items := append([]model.Process(nil), processes...)
	sort.SliceStable(items, func(i, j int) bool {
		return less(items[i], items[j])
	})
	if len(items) > 5 {
		items = items[:5]
	}
	parts := make([]string, 0, len(items))
	for _, process := range items {
		parts = append(parts, fmt.Sprintf("%d:%s:%d:%d:%.2f", process.PID, process.Name, process.RSSKB, process.VSZKB, process.CPUPct))
	}
	return joinWithComma(parts)
}

func joinWithComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for index := 1; index < len(parts); index++ {
		result += "," + parts[index]
	}
	return result
}

func enrichSnapshot(snapshot *model.Snapshot, containersByPID map[int]containerctl.Container) {
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
