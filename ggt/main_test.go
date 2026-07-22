package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func TestUpdateIndexAddSingleFile(t *testing.T) {
	testRepoPath := setupGitRepo(t)
	if err := os.WriteFile(filepath.Join(testRepoPath, "file.txt"), []byte("hello world"), 0666); err != nil {
		t.Fatalf("Creating the test file failed: %s", err)
	}

	cmd := exec.Command(ggtBinaryPath, "update-index", "--add", "file.txt")
	cmd.Dir = testRepoPath
	_, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %s", err)
	}

	cmd = exec.Command("git", "ls-files")
	cmd.Dir = testRepoPath
	result, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %s", err)
	}

	if !strings.Contains(string(result), "file.txt") {
		t.Fatalf("Test failed")
	}
}
