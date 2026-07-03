package main

import (
	"os/exec"
	"strings"
)

func getGitBranch(dir string) (string, bool) {
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err != nil {
		return "", false
	}

	out, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
	if err == nil {
		if branch := strings.TrimSpace(string(out)); branch != "" {
			return branch, true
		}
	}

	out, err = exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		if branch := strings.TrimSpace(string(out)); branch != "" {
			return branch, true
		}
	}

	return "", false
}
