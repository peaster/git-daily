package main

import (
	"os/exec"
	"strings"
)

// runGit executes a git command and returns its stdout.
// Returns "" on any error (matching the bash || true pattern).
func runGit(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n")
}
