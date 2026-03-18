package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// Env holds the environment variables exported to hook scripts.
type Env struct {
	WorktreeDir string
	Branch      string
	Ticket      string
	RepoName    string
	Phase       string // "setup" or "teardown"
	TmuxWindow  string
}

// Run executes hook scripts from ~/.lw/<phase>/ and <gitRoot>/.lw/<phase>/ in sorted order.
// Global hooks run first, then repo-local hooks.
func Run(phase, gitRoot string, env Env) error {
	globalDir := filepath.Join(os.Getenv("HOME"), ".lw", phase)
	repoDir := filepath.Join(gitRoot, ".lw", phase)

	var scripts []string
	for _, dir := range []string{globalDir, repoDir} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			if fi.Mode()&0111 != 0 {
				scripts = append(scripts, filepath.Join(dir, e.Name()))
			}
		}
	}

	if len(scripts) == 0 {
		return nil
	}
	sort.Strings(scripts)

	hookEnv := os.Environ()
	hookEnv = append(hookEnv,
		"LW_WORKTREE_DIR="+env.WorktreeDir,
		"LW_BRANCH="+env.Branch,
		"LW_TICKET="+env.Ticket,
		"LW_REPO_NAME="+env.RepoName,
		"LW_ACTION="+env.Phase,
		"LW_TMUX_WINDOW="+env.TmuxWindow,
	)

	workDir := env.WorktreeDir
	if workDir == "" {
		workDir = "."
	}

	for _, script := range scripts {
		name := filepath.Base(script)
		fmt.Fprintf(os.Stderr, "▸ Running %s hook: %s…\n", phase, name)
		cmd := exec.Command(script)
		cmd.Dir = workDir
		cmd.Env = hookEnv
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hook %s failed: %w", name, err)
		}
	}
	return nil
}
