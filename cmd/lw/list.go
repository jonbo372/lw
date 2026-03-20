package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/session"
)

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// sessionStatus classifies the health of a session/worktree entry.
type sessionStatus int

const (
	statusMain     sessionStatus = iota // the main repo worktree
	statusActive                        // session JSON + worktree both exist
	statusOrphaned                      // worktree exists but no session JSON
	statusDead                          // session JSON exists but worktree is gone
)

// listEntry represents a single row in the list output.
type listEntry struct {
	Name            string
	Branch          string
	WorktreePath    string
	Status          sessionStatus
	ClaudeSessionID string
}

// statusLabel returns a color-coded status string.
func statusLabel(s sessionStatus) string {
	switch s {
	case statusMain:
		return colorDim + "main" + colorReset
	case statusActive:
		return colorGreen + "active" + colorReset
	case statusOrphaned:
		return colorYellow + "orphaned" + colorReset
	case statusDead:
		return colorRed + "dead" + colorReset
	default:
		return "unknown"
	}
}

// statusLabelPlain returns a plain (no color) status string for testing.
func statusLabelPlain(s sessionStatus) string {
	switch s {
	case statusMain:
		return "main"
	case statusActive:
		return "active"
	case statusOrphaned:
		return "orphaned"
	case statusDead:
		return "dead"
	default:
		return "unknown"
	}
}

// buildListEntries cross-references sessions and worktrees, returning
// a sorted list of entries with the main worktree first.
func buildListEntries(sessions map[string]*session.Session, worktrees []git.WorktreeEntry, mainWorktreePath string) []listEntry {
	// Track which worktree paths have a session
	worktreePathSet := make(map[string]bool)
	for _, wt := range worktrees {
		worktreePathSet[wt.Path] = true
	}

	// Track which sessions we've already matched to a worktree
	matchedSessions := make(map[string]bool)

	var entries []listEntry

	// 1. Main worktree entry first
	for _, wt := range worktrees {
		if wt.Path == mainWorktreePath {
			entries = append(entries, listEntry{
				Name:         filepath.Base(wt.Path),
				Branch:       wt.Branch,
				WorktreePath: wt.Path,
				Status:       statusMain,
			})
			break
		}
	}

	// 2. Active sessions (session JSON + worktree exists)
	for name, sess := range sessions {
		if worktreePathSet[sess.WorktreeDir] {
			matchedSessions[name] = true
			entries = append(entries, listEntry{
				Name:            name,
				Branch:          sess.Branch,
				WorktreePath:    sess.WorktreeDir,
				Status:          statusActive,
				ClaudeSessionID: sess.ClaudeSessionID,
			})
		}
	}

	// 3. Dead sessions (session JSON exists but worktree is gone)
	for name, sess := range sessions {
		if matchedSessions[name] {
			continue
		}
		entries = append(entries, listEntry{
			Name:            name,
			Branch:          sess.Branch,
			WorktreePath:    sess.WorktreeDir,
			Status:          statusDead,
			ClaudeSessionID: sess.ClaudeSessionID,
		})
	}

	// 4. Orphaned worktrees (worktree exists but no session JSON)
	// Build a set of all worktree dirs that have sessions
	sessionWorktreeDirs := make(map[string]bool)
	for _, sess := range sessions {
		sessionWorktreeDirs[sess.WorktreeDir] = true
	}
	for _, wt := range worktrees {
		if wt.Path == mainWorktreePath {
			continue // skip main, already shown
		}
		if wt.Bare {
			continue
		}
		if sessionWorktreeDirs[wt.Path] {
			continue // has a session
		}
		entries = append(entries, listEntry{
			Name:         filepath.Base(wt.Path),
			Branch:       wt.Branch,
			WorktreePath: wt.Path,
			Status:       statusOrphaned,
		})
	}

	return entries
}

// maxLen returns the maximum of a and b.
func maxLen(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderTable writes a formatted table to the given writer.
func renderTable(w io.Writer, entries []listEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No sessions or worktrees found.")
		return
	}

	// Calculate column widths (based on plain text, not ANSI codes)
	nameW, branchW, pathW, statusW, claudeW := 4, 6, 8, 6, 10 // header minimums
	for _, e := range entries {
		nameW = maxLen(nameW, len(e.Name))
		branchW = maxLen(branchW, len(e.Branch))
		pathW = maxLen(pathW, len(e.WorktreePath))
		statusW = maxLen(statusW, len(statusLabelPlain(e.Status)))
		claudeW = maxLen(claudeW, len(e.ClaudeSessionID))
	}

	// Header
	fmt.Fprintf(w, "%s%-*s  %-*s  %-*s  %-*s  %-*s%s\n",
		colorBold,
		nameW, "NAME",
		branchW, "BRANCH",
		pathW, "WORKTREE",
		statusW, "STATUS",
		claudeW, "CLAUDE ID",
		colorReset,
	)

	// Separator
	fmt.Fprintln(w, strings.Repeat("-", nameW+branchW+pathW+statusW+claudeW+8))

	// Rows
	for _, e := range entries {
		label := statusLabel(e.Status)
		plainLabel := statusLabelPlain(e.Status)
		// Pad the status column accounting for ANSI escape codes
		extraChars := len(label) - len(plainLabel)
		fmt.Fprintf(w, "%-*s  %-*s  %-*s  %-*s  %s\n",
			nameW, e.Name,
			branchW, e.Branch,
			pathW, e.WorktreePath,
			statusW+extraChars, label,
			e.ClaudeSessionID,
		)
	}
}

// listDeps holds the external dependencies for the list command,
// allowing tests to inject fakes.
type listDeps struct {
	sessionsDir    string
	repoName       string
	listSessions   func(sessionsDir, repoName string) (map[string]*session.Session, error)
	listWorktrees  func() ([]git.WorktreeEntry, error)
	mainWorktree   func() (string, error)
	output         io.Writer
}

// cmdListWithDeps is the testable core of the list command.
func cmdListWithDeps(deps listDeps) error {
	sessions, err := deps.listSessions(deps.sessionsDir, deps.repoName)
	if err != nil {
		return fmt.Errorf("reading sessions: %w", err)
	}

	worktrees, err := deps.listWorktrees()
	if err != nil {
		return fmt.Errorf("listing worktrees: %w", err)
	}

	mainPath, err := deps.mainWorktree()
	if err != nil {
		return fmt.Errorf("finding main worktree: %w", err)
	}

	entries := buildListEntries(sessions, worktrees, mainPath)
	renderTable(deps.output, entries)
	return nil
}

// cmdList is the top-level entry point called by Cobra.
func cmdList() error {
	gitRoot, err := git.MainRoot()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}
	repoName := filepath.Base(gitRoot)

	deps := listDeps{
		sessionsDir:   session.DefaultSessionsDir(),
		repoName:      repoName,
		listSessions:  session.ListAll,
		listWorktrees: git.WorktreeList,
		mainWorktree:  git.MainRoot,
		output:        os.Stdout,
	}
	return cmdListWithDeps(deps)
}
