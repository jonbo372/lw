package tmux

import (
	"testing"
)

func TestSendKeys_NotInTmux(t *testing.T) {
	// When TMUX env is not set, SendKeys should be a no-op
	t.Setenv("TMUX", "")

	// Should not panic or error
	SendKeys("some-session", "echo hello")
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"[repo] ENG-123: Add auth", "[repo] ENG-123- Add auth"},
		{"[repo] review: feature/auth", "[repo] review- feature/auth"},
		{"name.with.dots", "name-with-dots"},
		{"no-special-chars", "no-special-chars"},
		{"mixed:colons.and:dots", "mixed-colons-and-dots"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
