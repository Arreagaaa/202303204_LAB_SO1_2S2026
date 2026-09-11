package model

type Snapshot struct {
	Memory Memory
	Processes []Process
}

type Memory struct {
	TotalKB uint64
	FreeKB  uint64
	UsedKB  uint64
}

type Process struct {
	PID       int
	Name      string
	Command   string
	VSZKB     uint64
	RSSKB     uint64
	MemoryPct float64
	CPUPct    float64
	Class     WorkloadClass
}

type WorkloadClass string

const (
	ClassUnknown WorkloadClass = "unknown"
	ClassLow     WorkloadClass = "low"
	ClassHigh    WorkloadClass = "high"
	ClassGrafana WorkloadClass = "grafana"
)
