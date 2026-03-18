package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), err
}

func runInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), err
}

// MainRoot returns the root of the main repository, even when called from a worktree.
func MainRoot() (string, error) {
	commonDir, err := run("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(commonDir, "/.git"), nil
}

// RefExists checks whether a git ref exists.
func RefExists(ref string) bool {
	_, err := run("show-ref", "--verify", "--quiet", ref)
	return err == nil
}

// BranchExists checks whether a local branch exists.
func BranchExists(branch string) bool {
	return RefExists("refs/heads/" + branch)
}

// CurrentBranch returns the current branch name.
func CurrentBranch() (string, error) {
	out, err := run("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not determine current branch (detached HEAD?)")
	}
	return out, nil
}

// CurrentBranchInDir returns the current branch name for a worktree directory.
func CurrentBranchInDir(dir string) (string, error) {
	return runInDir(dir, "symbolic-ref", "--short", "HEAD")
}

// FetchBranch fetches a branch from origin into a local branch of the same name.
func FetchBranch(branch string) error {
	_, err := run("fetch", "origin", branch+":"+branch)
	return err
}

// FetchOriginBranch fetches a branch from origin without creating a local tracking branch.
func FetchOriginBranch(branch string) error {
	_, err := run("fetch", "origin", branch)
	return err
}

// FetchOrigin runs git fetch origin.
func FetchOrigin() {
	run("fetch", "origin")
}

// CreateBranch creates a new branch at the given start point.
func CreateBranch(name, startPoint string) error {
	_, err := run("branch", name, startPoint)
	return err
}

// DeleteBranch force-deletes a local branch.
func DeleteBranch(branch string) error {
	_, err := run("branch", "-D", branch)
	return err
}

// DetectDefaultBranch tries to determine the default branch from origin.
func DetectDefaultBranch() (string, error) {
	out, err := run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil && out != "" {
		return strings.TrimPrefix(out, "refs/remotes/origin/"), nil
	}

	FetchOrigin()
	for _, candidate := range []string{"main", "master", "trunk", "develop"} {
		if RefExists("refs/remotes/origin/" + candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not determine the default branch — set it with: git remote set-head origin --auto")
}

// EnsureLocalBranch makes sure the branch exists locally, fetching from origin if needed.
func EnsureLocalBranch(branch string) error {
	if BranchExists(branch) {
		return nil
	}
	return FetchBranch(branch)
}

// WorktreeAdd creates a new worktree at the given path for the given branch.
func WorktreeAdd(path, branch string) error {
	_, err := run("worktree", "add", path, branch)
	return err
}

// WorktreeRemove force-removes a worktree and prunes.
func WorktreeRemove(path string) error {
	_, err := run("worktree", "remove", "--force", path)
	run("worktree", "prune")
	return err
}

// WorktreeBranchInUse checks if a branch is still referenced by any worktree.
func WorktreeBranchInUse(branch string) bool {
	out, _ := run("worktree", "list", "--porcelain")
	return strings.Contains(out, "refs/heads/"+branch)
}

// IsDirty returns true if the worktree has uncommitted changes.
func IsDirty(dir string) bool {
	out, _ := runInDir(dir, "status", "--porcelain")
	return out != ""
}

// ShortStatus returns the short status output for a directory.
func ShortStatus(dir string) string {
	out, _ := runInDir(dir, "status", "--short")
	return out
}

// HasUnpushedCommits returns true if local HEAD differs from origin's branch tip.
func HasUnpushedCommits(dir, branch string) bool {
	local, _ := runInDir(dir, "rev-parse", "HEAD")
	remote, _ := runInDir(dir, "rev-parse", "origin/"+branch)
	return local != "" && remote != "" && local != remote
}

// DirExists checks if a directory exists on the filesystem.
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
