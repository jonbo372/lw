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

// MaxNameLen is the maximum length for a tmux session name.
const MaxNameLen = 50

// sanitizeName replaces characters that are special in tmux target syntax.
// Colons and periods are used as separators (session:window.pane), so they
// must not appear in session names.
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, ".", "-")
	return name
}

// CreateOrSwitch creates a new tmux session or switches to an existing one.
// Returns the session name used, or "" if not in tmux.
func CreateOrSwitch(name, dir string) string {
	if !Active() {
		return ""
	}
	name = sanitizeName(name)
	if len(name) > MaxNameLen {
		name = name[:MaxNameLen]
	}

	// Check if session already exists (exact match via = prefix)
	if exec.Command("tmux", "has-session", "-t", "="+name).Run() == nil {
		run("switch-client", "-t", name)
		return name
	}

	run("new-session", "-d", "-s", name, "-c", dir)
	run("switch-client", "-t", name)
	return name
}

// FindSession returns the name of the first session whose name contains prefix.
func FindSession(prefix string) string {
	if !Active() {
		return ""
	}
	prefix = sanitizeName(prefix)
	out, _ := run("list-sessions", "-F", "#{session_name}")
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, prefix) {
			return line
		}
	}
	return ""
}

// KillSession closes the tmux session with the given name.
func KillSession(name string) {
	run("kill-session", "-t", "="+name)
}

// SendKeys sends text to a tmux session followed by Enter.
// Does nothing if not in a tmux session.
func SendKeys(target, text string) {
	if !Active() {
		return
	}
	run("send-keys", "-t", target, text, "Enter")
}
