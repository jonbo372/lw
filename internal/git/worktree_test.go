package git

import (
	"testing"
)

func TestParseWorktreeList_Empty(t *testing.T) {
	entries := ParseWorktreeList("")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseWorktreeList_SingleEntry(t *testing.T) {
	input := "/home/user/repo  abc1234 [main]"
	entries := ParseWorktreeList(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "/home/user/repo" {
		t.Errorf("expected path /home/user/repo, got %s", entries[0].Path)
	}
	if entries[0].Branch != "main" {
		t.Errorf("expected branch main, got %s", entries[0].Branch)
	}
	if entries[0].Bare {
		t.Error("expected bare=false")
	}
}

func TestParseWorktreeList_MultipleEntries(t *testing.T) {
	input := `/home/user/repo          abc1234 [main]
/home/user/.superclaude/repo/VOI-42  def5678 [jonbo372/VOI-42/add-auth]
/home/user/.superclaude/repo/fuzzy_cobra  aaa1111 [fuzzy_cobra]
/home/user/.superclaude/repo/review-feature  bbb2222 [feature/cool-thing]`

	entries := ParseWorktreeList(input)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	if entries[1].Path != "/home/user/.superclaude/repo/VOI-42" {
		t.Errorf("unexpected path: %s", entries[1].Path)
	}
	if entries[1].Branch != "jonbo372/VOI-42/add-auth" {
		t.Errorf("unexpected branch: %s", entries[1].Branch)
	}
}

func TestParseWorktreeList_BareEntry(t *testing.T) {
	input := "/home/user/repo  0000000 (bare)"
	entries := ParseWorktreeList(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].Bare {
		t.Error("expected bare=true")
	}
	if entries[0].Branch != "" {
		t.Errorf("expected empty branch for bare entry, got %s", entries[0].Branch)
	}
}

func TestParseWorktreeList_DetachedHead(t *testing.T) {
	input := "/home/user/repo  abc1234 (detached HEAD)"
	entries := ParseWorktreeList(input)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Branch != "" {
		t.Errorf("expected empty branch for detached HEAD, got %s", entries[0].Branch)
	}
}

func TestMatchWorktrees_ExactPathMatch(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/.superclaude/repo/fuzzy_cobra", Branch: "fuzzy_cobra"},
	}
	matches := MatchWorktrees(entries, "/home/user/.superclaude/repo/fuzzy_cobra")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestMatchWorktrees_ByDirName(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/fuzzy_cobra", Branch: "fuzzy_cobra"},
		{Path: "/home/user/.superclaude/repo/VOI-42", Branch: "jonbo372/VOI-42/add-auth"},
	}
	matches := MatchWorktrees(entries, "fuzzy_cobra")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Path != "/home/user/.superclaude/repo/fuzzy_cobra" {
		t.Errorf("unexpected path: %s", matches[0].Path)
	}
}

func TestMatchWorktrees_ByBranch(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/VOI-42", Branch: "jonbo372/VOI-42/add-auth"},
	}
	matches := MatchWorktrees(entries, "jonbo372/VOI-42/add-auth")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestMatchWorktrees_SubstringMatch(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/VOI-42", Branch: "jonbo372/VOI-42/add-auth"},
		{Path: "/home/user/.superclaude/repo/VOI-43", Branch: "jonbo372/VOI-43/fix-bug"},
	}
	// "VOI-42" should match both path and branch for entry 2
	matches := MatchWorktrees(entries, "VOI-42")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Path != "/home/user/.superclaude/repo/VOI-42" {
		t.Errorf("unexpected match: %s", matches[0].Path)
	}
}

func TestMatchWorktrees_MultipleMatches(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/.superclaude/repo/VOI-42", Branch: "jonbo372/VOI-42/add-auth"},
		{Path: "/home/user/.superclaude/repo/VOI-43", Branch: "jonbo372/VOI-43/fix-bug"},
	}
	// "VOI-4" matches both
	matches := MatchWorktrees(entries, "VOI-4")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestMatchWorktrees_NoMatch(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
	}
	matches := MatchWorktrees(entries, "nonexistent")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestMatchWorktrees_ReviewWorktree(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/.superclaude/repo/review-feature-cool", Branch: "feature/cool-thing"},
	}
	matches := MatchWorktrees(entries, "review-feature-cool")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestMatchWorktrees_SkipsBareEntries(t *testing.T) {
	entries := []WorktreeEntry{
		{Path: "/home/user/repo", Branch: "", Bare: true},
		{Path: "/home/user/.superclaude/repo/work", Branch: "work"},
	}
	matches := MatchWorktrees(entries, "repo")
	// Should only match the non-bare entry via substring
	// The bare one has "repo" in its path too, but bare entries should be skipped
	for _, m := range matches {
		if m.Bare {
			t.Error("bare entries should not be matched")
		}
	}
}
