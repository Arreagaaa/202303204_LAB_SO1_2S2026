package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"proyecto2/daemon/internal/containerctl"
	"proyecto2/daemon/internal/cronmanager"
	"proyecto2/daemon/internal/decision"
	"proyecto2/daemon/internal/killwatch"
	"proyecto2/daemon/internal/model"
	"proyecto2/daemon/internal/procreader"
	"proyecto2/daemon/internal/valkey"
)

func main() {
	procPath := flag.String("proc", "/proc/continfo_pr2_so1_202303204", "archivo de métricas del módulo")
	valkeyAddr := flag.String("valkey", "127.0.0.1:6379", "dirección del servicio Valkey")
	once := flag.Bool("once", false, "procesa una lectura y termina")
	interval := flag.Duration("interval", 30*time.Second, "tiempo entre lecturas")
	killTimeout := flag.Duration("kill-timeout", 5*time.Second, "tiempo máximo para esperar la confirmación eBPF")
	flag.Parse()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("no se pudo obtener el directorio actual: %v", err)
	}
	cron := cronmanager.Manager{
		Executable: "/bin/bash",
		ScriptPath: workDir + "/cronjob/spawn_containers.sh",
	}
	log.Printf("cronjob apuntara a: %s", cron.ScriptPath)

	log.Println("cargando modulo de kernel...")
	if output, err := exec.Command("/bin/bash", workDir+"/kernel-module/load.sh").CombinedOutput(); err != nil {
		log.Printf("script del modulo: %v: %s", err, strings.TrimSpace(string(output)))
	} else {
		log.Printf("modulo de kernel listo: %s", strings.TrimSpace(string(output)))
	}

	log.Println("levantando Grafana y Valkey...")
	if output, err := exec.Command("docker", "compose", "-f", workDir+"/infra/docker-compose.yml", "up", "-d").CombinedOutput(); err != nil {
		log.Printf("docker compose: %v: %s", err, strings.TrimSpace(string(output)))
	}

	log.Println("instalando cronjob de contenedores...")
	if err := cron.Install(); err != nil {
		log.Printf("error instalando cronjob: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("deteniendo daemon, eliminando cronjob...")
		_ = cron.Remove()
		os.Exit(0)
	}()

	policy := decision.DefaultPolicy()
	client, err := valkey.Dial(*valkeyAddr)
	if err != nil {
		log.Printf("valkey no disponible en %s: %v", *valkeyAddr, err)
	}
	if client != nil {
		defer client.Close()
	}
	watcher, err := killwatch.Start()
	if err != nil {
		log.Printf("eBPF no disponible: %v", err)
	} else if watcher != nil {
		defer watcher.Close()
	}
	var totalKills uint64
	for {
		kills, err := processSnapshot(*procPath, policy, client, watcher, *killTimeout)
		if err != nil {
			log.Printf("lectura no disponible: %v", err)
		}
		totalKills += kills
		if client != nil {
			_ = client.SetInteger("containers:killed:count", totalKills)
			_ = client.SetInteger("ebpf:events:count", totalKills)
		}
		if *once {
			_ = cron.Remove()
			return
		}
		time.Sleep(*interval)
	}
}

func processSnapshot(path string, policy decision.Policy, client *valkey.Client, watcher *killwatch.Watcher, killTimeout time.Duration) (uint64, error) {
	snapshot, err := procreader.Read(path)
	if err != nil {
		return 0, err
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
		if err := persistSnapshot(client, snapshot, confirmedKills, watcher); err != nil {
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
	return uint64(len(confirmedKills)), nil
}

func persistSnapshot(client *valkey.Client, snapshot model.Snapshot, confirmedKills []killwatch.Event, watcher *killwatch.Watcher) error {
	if err := client.SetInteger("system:ram:total", snapshot.Memory.TotalKB); err != nil {
		return err
	}
	if err := client.SetInteger("system:ram:free", snapshot.Memory.FreeKB); err != nil {
		return err
	}
	if err := client.SetInteger("system:ram:used", snapshot.Memory.UsedKB); err != nil {
		return err
	}
	if err := client.Set("system:ram:current", fmt.Sprintf("total=%dKB,free=%dKB,used=%dKB", snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB)); err != nil {
		return err
	}
	ts := time.Now().Unix()
	if err := client.RPush("system:ram:history", fmt.Sprintf("%d,%d,%d,%d", ts, snapshot.Memory.UsedKB, snapshot.Memory.FreeKB, snapshot.Memory.TotalKB)); err != nil {
		return err
	}
	_ = client.LTrim("system:ram:history", -180, -1)
	for _, event := range confirmedKills {
		if err := client.RPush("containers:killed:history", fmt.Sprintf("%d,1", ts)); err != nil {
			return err
		}
		_ = event
	}
	_ = client.LTrim("containers:killed:history", -180, -1)
	if err := persistTopHash(client, "containers:top:ram", snapshot.Processes, func(p model.Process) uint64 { return p.RSSKB }); err != nil {
		return err
	}
	if err := persistTopHash(client, "containers:top:cpu", snapshot.Processes, func(p model.Process) uint64 { return uint64(p.CPUPct * 100) }); err != nil {
		return err
	}
	if watcher != nil {
		if err := client.SetInteger("ebpf:events:raw", watcher.Total()); err != nil {
			return err
		}
	}
	return nil
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
		if err := containerctl.Signal(container.PID, syscall.SIGTERM); err != nil {
			log.Printf("no se pudo enviar SIGTERM a pid=%d contenedor=%s: %v", container.PID, container.ID, err)
			continue
		}
		log.Printf("SIGTERM enviado por el daemon pid=%d nombre=%s", container.PID, container.Name)
		event, confirmedByEBPF := watcher.WaitFor(uint32(container.PID), killTimeout)
		if watcher != nil && !confirmedByEBPF {
			log.Printf("sin confirmacion eBPF para pid=%d contenedor=%s (no se registra en Valkey)", container.PID, container.ID)
			continue
		}
		if err := containerctl.Stop(container.ID); err != nil {
			log.Printf("no se pudo detener %s: %v", container.ID, err)
			continue
		}
		log.Printf("contenedor detenido pid=%d nombre=%s clase=%s", container.PID, container.Name, container.Class)
		if watcher == nil {
			confirmed = append(confirmed, killwatch.Event{
				TargetPID: uint32(container.PID),
				Signal:    15,
			})
			continue
		}
		confirmed = append(confirmed, event)
		log.Printf("contenedor confirmado por eBPF pid=%d signal=%d sender=%d", event.TargetPID, event.Signal, event.SenderPID)
	}
	return confirmed
}

func persistTopHash(client *valkey.Client, key string, processes []model.Process, value func(model.Process) uint64) error {
	containers := make([]model.Process, 0, len(processes))
	for _, process := range processes {
		switch process.Class {
		case model.ClassLow, model.ClassHigh, model.ClassIntruder:
			containers = append(containers, process)
		}
	}
	items := containers
	sort.SliceStable(items, func(i, j int) bool {
		return value(items[i]) > value(items[j])
	})
	if len(items) > 5 {
		items = items[:5]
	}
	if err := client.Del(key); err != nil {
		return err
	}
	for _, process := range items {
		id := process.Command
		if len(id) > 12 {
			id = id[:12]
		}
		field := fmt.Sprintf("pid=%d %s (%s)", process.PID, process.Name, id)
		if err := client.HSet(key, field, fmt.Sprintf("%d", value(process))); err != nil {
			return err
		}
	}
	return nil
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
