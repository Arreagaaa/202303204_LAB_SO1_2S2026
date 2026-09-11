package procreader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"proyecto2/daemon/internal/model"
)

func Read(path string) (model.Snapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer file.Close()

	var snapshot model.Snapshot
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, ":") && !strings.Contains(line, ",") {
			if err := readMemoryLine(&snapshot.Memory, line); err != nil {
				return model.Snapshot{}, fmt.Errorf("línea %d: %w", lineNumber, err)
			}
			continue
		}
		if strings.EqualFold(line, "PID,NAME,CMD_OR_CONTAINER_ID,VSZ_KB,RSS_KB,MEM_PCT,CPU_PCT,CLASS") {
			continue
		}
		process, err := readProcessLine(line)
		if err != nil {
			return model.Snapshot{}, fmt.Errorf("línea %d: %w", lineNumber, err)
		}
		snapshot.Processes = append(snapshot.Processes, process)
	}
	if err := scanner.Err(); err != nil {
		return model.Snapshot{}, fmt.Errorf("leer %s: %w", path, err)
	}
	return snapshot, nil
}

func readMemoryLine(memory *model.Memory, line string) error {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("métrica de memoria inválida")
	}
	value, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("valor de memoria inválido: %w", err)
	}
	switch strings.TrimSpace(parts[0]) {
	case "MEM_TOTAL_KB":
		memory.TotalKB = value
	case "MEM_FREE_KB":
		memory.FreeKB = value
	case "MEM_USED_KB":
		memory.UsedKB = value
	default:
		return fmt.Errorf("métrica de memoria desconocida %q", parts[0])
	}
	return nil
}

func readProcessLine(line string) (model.Process, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 8 {
		return model.Process{}, fmt.Errorf("se esperaban 8 campos y llegaron %d", len(fields))
	}
	pid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil {
		return model.Process{}, fmt.Errorf("PID inválido: %w", err)
	}
	vsz, err := parseUint(fields[3])
	if err != nil {
		return model.Process{}, fmt.Errorf("VSZ inválido: %w", err)
	}
	rss, err := parseUint(fields[4])
	if err != nil {
		return model.Process{}, fmt.Errorf("RSS inválido: %w", err)
	}
	memoryPct, err := parseFloat(fields[5])
	if err != nil {
		return model.Process{}, fmt.Errorf("porcentaje de memoria inválido: %w", err)
	}
	cpuPct, err := parseFloat(fields[6])
	if err != nil {
		return model.Process{}, fmt.Errorf("porcentaje de CPU inválido: %w", err)
	}
	class := model.WorkloadClass(strings.ToLower(strings.TrimSpace(fields[7])))
	if class == "" {
		class = model.ClassUnknown
	}
	return model.Process{
		PID: pid, Name: strings.TrimSpace(fields[1]), Command: strings.TrimSpace(fields[2]),
		VSZKB: vsz, RSSKB: rss, MemoryPct: memoryPct, CPUPct: cpuPct, Class: class,
	}, nil
}

func parseUint(value string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(value), 10, 64)
}

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}
