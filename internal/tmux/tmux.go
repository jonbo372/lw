package tmux

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), err
}

// Active returns true when running inside a tmux session.
func Active() bool {
	return os.Getenv("TMUX") != ""
}

// MaxWindowNameLen is the maximum length for a tmux window name.
const MaxWindowNameLen = 50

// CreateOrSwitch opens a new tmux window or switches to an existing one.
// Returns the window name used, or "" if not in tmux.
func CreateOrSwitch(name, dir string) string {
	if !Active() {
		return ""
	}
	if len(name) > MaxWindowNameLen {
		name = name[:MaxWindowNameLen]
	}

	out, _ := run("list-windows", "-F", "#{window_name}")
	for _, line := range strings.Split(out, "\n") {
		if line == name {
			run("select-window", "-t", name)
			return name
		}
	}

	run("new-window", "-c", dir, "-n", name)
	return name
}

// FindWindow returns the index and name of the first window whose name contains prefix.
func FindWindow(prefix string) (index, name string) {
	if !Active() {
		return "", ""
	}
	out, _ := run("list-windows", "-F", "#{window_index} #{window_name}")
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, prefix) {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
		}
	}
	return "", ""
}

// KillWindow closes the tmux window at the given index.
func KillWindow(index string) {
	run("kill-window", "-t", index)
}

// SendKeys sends text to a tmux window followed by Enter.
// Does nothing if not in a tmux session.
func SendKeys(window, text string) {
	if !Active() {
		return
	}
	run("send-keys", "-t", window, text, "Enter")
}
