package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// verboseMode controls whether verbose() produces output.
var verboseMode bool

// setVerbose sets the verbose mode.
func setVerbose(v bool) {
	verboseMode = v
}

// isVerbose returns whether verbose mode is enabled.
func isVerbose() bool {
	return verboseMode
}

// verbose prints a message to stderr only when --verbose is enabled.
func verbose(msg string, args ...any) {
	if verboseMode {
		fmt.Fprintf(os.Stderr, "  "+msg+"\n", args...)
	}
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+msg+"\n", args...)
	os.Exit(1)
}

func info(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "▸ "+msg+"\n", args...)
}

func confirm(msg string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
	}
	return false
}
