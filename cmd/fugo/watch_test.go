package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsGoFileChange(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile("main.go", []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	w, err := newWatcher()
	if err != nil {
		t.Fatalf("newWatcher: %v", err)
	}
	defer func() { _ = w.Close() }()

	done := make(chan struct{})
	go func() {
		waitForChange(w)
		close(done)
	}()

	// Give the watcher goroutine time to be blocked on select before writing,
	// then edit the watched file.
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("waitForChange did not return after a .go file was written")
	}
}

func TestWatcherIgnoresNonGoFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile("main.go", []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	w, err := newWatcher()
	if err != nil {
		t.Fatalf("newWatcher: %v", err)
	}
	defer func() { _ = w.Close() }()

	done := make(chan struct{})
	go func() {
		waitForChange(w)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile("notes.txt", []byte("hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
		t.Fatal("waitForChange returned for a non-.go file change")
	case <-time.After(500 * time.Millisecond):
	}

	// Confirm it still works once a real .go change happens.
	if err := os.WriteFile("main.go", []byte("package main\n\n// changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("waitForChange did not return after a .go file was written")
	}
}

func TestIsWatchExcluded(t *testing.T) {
	t.Parallel()

	for _, dir := range []string{".git", "bin", distDir, "logs", "vendor"} {
		if !isWatchExcluded(filepath.Join("project", dir)) {
			t.Errorf("isWatchExcluded(%q) = false, want true", dir)
		}
	}

	if isWatchExcluded(filepath.Join("project", "ui")) {
		t.Error("isWatchExcluded(ui) = true, want false")
	}
}
