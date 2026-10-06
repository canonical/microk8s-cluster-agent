package internal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestNotifyOnChangeClosesWatcher(t *testing.T) {
	file := filepath.Join(t.TempDir(), "provider.yaml")
	if err := os.WriteFile(file, []byte("tcp: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	watcher, err := newWatcher(file)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- notifyOnChange(ctx, watcher, make(chan struct{})) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("notifyOnChange returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("notifyOnChange did not return after context cancel")
	}

	// A closed watcher rejects new watches; an open one (the leak) accepts them.
	if err := watcher.Add(file); !errors.Is(err, fsnotify.ErrClosed) {
		t.Fatalf("watcher still open after context cancel: Add returned %v, want fsnotify.ErrClosed", err)
	}
}
