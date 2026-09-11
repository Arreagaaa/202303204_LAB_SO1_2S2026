package cronmanager

import (
	"fmt"
	"os/exec"
	"strings"
)

const marker = "# proyecto2-so1"

type Manager struct {
	Executable string
	ScriptPath string
}

func (manager Manager) Install() error {
	current, err := list()
	if err != nil {
		return err
	}
	current = removeProjectEntries(current)
	entry := fmt.Sprintf("*/2 * * * * %s %s %s", manager.Executable, manager.ScriptPath, marker)
	return write(strings.TrimSpace(current+"\n"+entry) + "\n")
}

func (manager Manager) Remove() error {
	current, err := list()
	if err != nil {
		return err
	}
	return write(removeProjectEntries(current))
}

func list() (string, error) {
	output, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("listar crontab: %w", err)
	}
	return string(output), nil
}

func write(content string) error {
	command := exec.Command("crontab", "-")
	command.Stdin = strings.NewReader(content)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("actualizar crontab: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func removeProjectEntries(content string) string {
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, marker) {
			kept = append(kept, line)
		}
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}
