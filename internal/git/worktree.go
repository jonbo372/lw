package git

import (
	"path/filepath"
	"strings"
)

// WorktreeEntry represents a single entry from `git worktree list`.
type WorktreeEntry struct {
	Path   string
	Branch string
	Bare   bool
}

// ParseWorktreeList parses the output of `git worktree list` (default format)
// into structured entries.
//
// Each line looks like:
//
//	/path/to/worktree  abc1234 [branch-name]
//	/path/to/worktree  abc1234 (bare)
//	/path/to/worktree  abc1234 (detached HEAD)
func ParseWorktreeList(output string) []WorktreeEntry {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	var entries []WorktreeEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := WorktreeEntry{}

		// The format is: path  hash [branch] or path  hash (info)
		// Split on whitespace — path is first field.
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		entry.Path = fields[0]

		// Remaining after path and hash is the branch info
		rest := strings.Join(fields[2:], " ")

		if rest == "(bare)" {
			entry.Bare = true
		} else if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
			entry.Branch = rest[1 : len(rest)-1]
		}
		// (detached HEAD) and other parenthetical forms leave Branch empty

		entries = append(entries, entry)
	}
	return entries
}

// MatchWorktrees finds worktree entries matching the given identifier.
// It matches against: exact path, directory basename, branch name, or
// substring match in path or branch. Bare entries are skipped.
func MatchWorktrees(entries []WorktreeEntry, identifier string) []WorktreeEntry {
	var exactMatches []WorktreeEntry
	var substringMatches []WorktreeEntry

	idLower := strings.ToLower(identifier)

	for _, e := range entries {
		if e.Bare {
			continue
		}

		dirName := filepath.Base(e.Path)

		// Exact matches (case-insensitive): full path, dir name, or branch name
		if e.Path == identifier ||
			strings.EqualFold(dirName, identifier) ||
			strings.EqualFold(e.Branch, identifier) {
			exactMatches = append(exactMatches, e)
			continue
		}

		// Substring matches (case-insensitive): identifier appears in path or branch
		if strings.Contains(strings.ToLower(e.Path), idLower) ||
			strings.Contains(strings.ToLower(e.Branch), idLower) {
			substringMatches = append(substringMatches, e)
		}
	}

	// Prefer exact matches if any exist
	if len(exactMatches) > 0 {
		return exactMatches
	}
	return substringMatches
}

// WorktreeList runs `git worktree list` and returns parsed entries.
func WorktreeList() ([]WorktreeEntry, error) {
	out, err := run("worktree", "list")
	if err != nil {
		return nil, err
	}
	return ParseWorktreeList(out), nil
}
