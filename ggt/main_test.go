package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var ggtBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ggtbin")
	if err != nil {
		log.Fatal(err)
	}

	binaryName := "ggt"
	if runtime.GOOS == "windows" {
		binaryName = binaryName + ".exe"
	}

	ggtBinaryPath = filepath.Join(dir, binaryName)

	cmd := exec.Command("go", "build", "-o", ggtBinaryPath)

	_, err = cmd.Output()
	if err != nil {
		log.Fatalf("Command failed: %s", err)
	}

	runResultCode := m.Run()
	os.RemoveAll(dir)
	os.Exit(runResultCode)
}

func setupGitRepo(t *testing.T) string {
	dir, err := os.MkdirTemp("", "ggt-test-repo")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init", dir)
	cmd.Dir = dir
	_, err = cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %s", err)
	}
	return dir
}
