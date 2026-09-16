package decision

import (
	"sort"

	"proyecto2/daemon/internal/model"
)

type Policy struct {
	MinimumLow         int
	MinimumHigh        int
	MemoryPercentLimit float64
	CPUPercentLimit    float64
}

func DefaultPolicy() Policy {
	return Policy{
		MinimumLow:         3,
		MinimumHigh:        2,
		MemoryPercentLimit: 3,
		CPUPercentLimit:    50,
	}
}

func SelectKills(snapshot model.Snapshot, policy Policy) []model.Process {
	processes := append([]model.Process(nil), snapshot.Processes...)
	sort.SliceStable(processes, func(i, j int) bool {
		if processes[i].RSSKB != processes[j].RSSKB {
			return processes[i].RSSKB > processes[j].RSSKB
		}
		if processes[i].VSZKB != processes[j].VSZKB {
			return processes[i].VSZKB > processes[j].VSZKB
		}
		if processes[i].CPUPct != processes[j].CPUPct {
			return processes[i].CPUPct > processes[j].CPUPct
		}
		return processes[i].MemoryPct > processes[j].MemoryPct
	})

	lowCount, highCount := countWorkloads(processes)
	candidates := make([]model.Process, 0)
	for _, process := range processes {
		if process.Class == model.ClassGrafana || process.Class == model.ClassUnknown || !overLimit(process, policy) {
			continue
		}
		if process.Class == model.ClassLow {
			if lowCount <= policy.MinimumLow {
				continue
			}
			lowCount--
		}
		if process.Class == model.ClassHigh {
			if highCount <= policy.MinimumHigh {
				continue
			}
			highCount--
		}
		candidates = append(candidates, process)
	}

	return candidates
}

func countWorkloads(processes []model.Process) (low, high int) {
	for _, process := range processes {
		switch process.Class {
		case model.ClassLow:
			low++
		case model.ClassHigh:
			high++
		}
	}
	return low, high
}

func overLimit(process model.Process, policy Policy) bool {
	return process.MemoryPct >= policy.MemoryPercentLimit || process.CPUPct >= policy.CPUPercentLimit
}
