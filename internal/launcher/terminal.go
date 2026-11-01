package launcher

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func StartGUI() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find application executable: %w", err)
	}
	if err = exec.Command(executable, "--ui", "gui", "--no-prompt").Start(); err != nil {
		return fmt.Errorf("start dashboard: %w", err)
	}
	return nil
}

func StartTerminal() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find application executable: %w", err)
	}
	appArgs := []string{executable, "--ui", "tui", "--no-prompt"}
	commands := terminalCommands(appArgs)
	for _, candidate := range commands {
		path, lookupErr := exec.LookPath(candidate[0])
		if lookupErr != nil {
			continue
		}
		if err := exec.Command(path, candidate[1:]...).Start(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no supported terminal emulator was found; run this command manually: %s --ui tui", executable)
}

func terminalCommands(appArgs []string) [][]string {
	switch runtime.GOOS {
	case "windows":
		if _, err := exec.LookPath("wt.exe"); err == nil {
			return [][]string{{"wt.exe", "new-tab", "--title", "AttackPath Visualizer", appArgs[0], appArgs[1], appArgs[2], appArgs[3]}}
		}
		return [][]string{{"cmd.exe", "/k", appArgs[0], appArgs[1], appArgs[2], appArgs[3]}}
	case "darwin":
		return [][]string{{"open", "-a", "Terminal", "--args", appArgs[0], appArgs[1], appArgs[2], appArgs[3]}}
	default:
		result := make([][]string, 0, 8)
		if terminal := strings.Fields(os.Getenv("TERMINAL")); len(terminal) > 0 {
			result = append(result, append(append([]string(nil), terminal...), append([]string{"-e"}, appArgs...)...))
		}
		for _, candidate := range []struct {
			name string
			args []string
		}{{"x-terminal-emulator", []string{"-e"}}, {"gnome-terminal", []string{"--"}}, {"konsole", []string{"-e"}}, {"xfce4-terminal", []string{"-e"}}, {"xterm", []string{"-e"}}} {
			result = append(result, append(append([]string{candidate.name}, candidate.args...), appArgs...))
		}
		return result
	}
}

func Prompt(input io.Reader, output io.Writer, saved string) (string, error) {
	defaultMode := preferred(saved)
	defaultChoice := "1"
	if defaultMode == "tui" {
		defaultChoice = "2"
	}
	if _, err := fmt.Fprintf(output, "How do you want to open AttackPath Visualizer?\n1 Dashboard (graphical window)\n2 Terminal UI\nChoice [%s]: ", defaultChoice); err != nil {
		return "", err
	}
	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read launcher choice: %w", err)
	}
	choice := strings.TrimSpace(line)
	mode := defaultMode
	if choice == "1" {
		mode = "gui"
	} else if choice == "2" {
		mode = "tui"
	} else if choice != "" {
		return "", errors.New("choose 1 for Dashboard or 2 for Terminal UI")
	}
	if _, err = fmt.Fprint(output, "Remember this choice? [y/N]: "); err != nil {
		return "", err
	}
	remember, readErr := reader.ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", fmt.Errorf("read preference choice: %w", readErr)
	}
	if strings.EqualFold(strings.TrimSpace(remember), "y") || strings.EqualFold(strings.TrimSpace(remember), "yes") {
		if err := SavePreference(mode); err != nil {
			return "", err
		}
	} else if err := SavePreference(""); err != nil {
		return "", err
	}
	return mode, nil
}
