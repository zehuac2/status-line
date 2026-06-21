package main

import (
	"os/exec"
	"strings"
)

type gitStatus struct {
	branch string
	dirty  bool
}

func getGitStatus(dir string) (gitStatus, bool) {
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err != nil {
		return gitStatus{}, false
	}

	var branch string
	out, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(out))
	} else {
		out, err = exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
		if err == nil {
			branch = strings.TrimSpace(string(out))
		}
	}

	if branch == "" {
		return gitStatus{}, false
	}

	out, _ = exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	dirty := strings.TrimSpace(string(out)) != ""

	return gitStatus{branch: branch, dirty: dirty}, true
}
