package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"proyecto2/daemon/internal/decision"
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

	kills := decision.SelectKills(snapshot, policy)
	fmt.Printf("RAM total=%dKB libre=%dKB usada=%dKB procesos=%d candidatos=%d\n",
		snapshot.Memory.TotalKB, snapshot.Memory.FreeKB, snapshot.Memory.UsedKB,
		len(snapshot.Processes), len(kills))
	for _, process := range kills {
		fmt.Printf("candidato pid=%d nombre=%s clase=%s rss=%dKB cpu=%.2f%%\n",
			process.PID, process.Name, process.Class, process.RSSKB, process.CPUPct)
	}
	return nil
}
