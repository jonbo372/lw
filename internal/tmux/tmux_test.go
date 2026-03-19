package tmux

import (
	"testing"
)

func TestSendKeys_NotInTmux(t *testing.T) {
	// When TMUX env is not set, SendKeys should be a no-op
	t.Setenv("TMUX", "")

	// Should not panic or error
	SendKeys("some-window", "echo hello")
}
