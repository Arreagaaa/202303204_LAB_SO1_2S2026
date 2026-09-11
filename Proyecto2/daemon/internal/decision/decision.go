package decision

import (
	"sort"

	"proyecto2/daemon/internal/model"
)

type Policy struct {
	MinimumLow           int
	MinimumHigh          int
	MemoryPercentLimit   float64
	CPUPercentLimit      float64
}

func DefaultPolicy() Policy {
	return Policy{
		MinimumLow: 3,
		MinimumHigh: 2,
		MemoryPercentLimit: 25,
		CPUPercentLimit: 80,
	}
}

func SelectKills(snapshot model.Snapshot, policy Policy) []model.Process {
	lowCount, highCount := countWorkloads(snapshot.Processes)
	candidates := make([]model.Process, 0)
	for _, process := range snapshot.Processes {
		if process.Class == model.ClassGrafana || !overLimit(process, policy) {
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

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].RSSKB != candidates[j].RSSKB {
			return candidates[i].RSSKB > candidates[j].RSSKB
		}
		if candidates[i].VSZKB != candidates[j].VSZKB {
			return candidates[i].VSZKB > candidates[j].VSZKB
		}
		if candidates[i].CPUPct != candidates[j].CPUPct {
			return candidates[i].CPUPct > candidates[j].CPUPct
		}
		return candidates[i].MemoryPct > candidates[j].MemoryPct
	})
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
