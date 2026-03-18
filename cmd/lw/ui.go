package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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
