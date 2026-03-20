package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVerbose_WhenEnabled(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	verboseMode = true
	defer func() { verboseMode = false }()

	verbose("testing %s", "output")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = old

	got := buf.String()
	if !strings.Contains(got, "testing output") {
		t.Errorf("expected verbose output to contain 'testing output', got: %q", got)
	}
}

func TestVerbose_WhenDisabled(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	verboseMode = false

	verbose("should not appear %s", "ever")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = old

	got := buf.String()
	if got != "" {
		t.Errorf("expected no output when verbose is disabled, got: %q", got)
	}
}

func TestVerbose_Prefix(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	verboseMode = true
	defer func() { verboseMode = false }()

	verbose("hello")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = old

	got := buf.String()
	// Should have some prefix and newline
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected verbose output to end with newline, got: %q", got)
	}
}

func TestSetVerbose(t *testing.T) {
	verboseMode = false
	setVerbose(true)
	if !verboseMode {
		t.Error("expected verboseMode to be true after setVerbose(true)")
	}
	setVerbose(false)
	if verboseMode {
		t.Error("expected verboseMode to be false after setVerbose(false)")
	}
}

func TestIsVerbose(t *testing.T) {
	verboseMode = false
	if isVerbose() {
		t.Error("expected isVerbose() to return false")
	}
	verboseMode = true
	defer func() { verboseMode = false }()
	if !isVerbose() {
		t.Error("expected isVerbose() to return true")
	}
}
