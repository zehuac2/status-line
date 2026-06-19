//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type target struct {
	goos   string
	goarch string
	suffix string
}

var targets = []target{
	{"darwin", "arm64", "darwin-arm64"},
	{"linux", "amd64", "linux-amd64"},
}

func main() {
	outDir := "dist"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fatalf("mkdir dist: %v", err)
	}

	for _, t := range targets {
		out := filepath.Join(outDir, "status-line-"+t.suffix)
		fmt.Printf("building %s...\n", out)

		cmd := exec.Command("go", "build", "-o", out, ".")
		cmd.Env = append(os.Environ(),
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fatalf("build %s: %v", out, err)
		}
		fmt.Printf("  -> %s\n", out)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "build: "+format+"\n", args...)
	os.Exit(1)
}
