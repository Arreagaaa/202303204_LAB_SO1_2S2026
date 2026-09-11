package containerctl

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Container struct {
	ID      string
	Name    string
	PID     int
	Class   string
	Grafana bool
}

func List() ([]Container, error) {
	format := "{{.ID}}|{{.Names}}|{{.Label \"so1.pr2.class\"}}"
	output, err := exec.Command("docker", "ps", "--format", format).Output()
	if err != nil {
		return nil, fmt.Errorf("listar contenedores: %w", err)
	}

	var containers []Container
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) != 3 {
			return nil, fmt.Errorf("fila de Docker inválida: %q", line)
		}
		pid, err := inspectPID(fields[0])
		if err != nil {
			return nil, err
		}
		containers = append(containers, Container{
			ID: fields[0], Name: fields[1], PID: pid,
			Class: fields[2], Grafana: fields[1] == "so1-pr2-grafana",
		})
	}
	return containers, nil
}

func Stop(containerID string) error {
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("ID de contenedor vacío")
	}
	if output, err := exec.Command("docker", "stop", containerID).CombinedOutput(); err != nil {
		return fmt.Errorf("detener %s: %w: %s", containerID, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func inspectPID(containerID string) (int, error) {
	output, err := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", containerID).Output()
	if err != nil {
		return 0, fmt.Errorf("inspeccionar %s: %w", containerID, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("PID inválido para %s: %w", containerID, err)
	}
	return pid, nil
}
