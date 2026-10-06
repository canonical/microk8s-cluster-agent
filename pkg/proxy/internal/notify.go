package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

func newWatcher(file string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to setup new watcher: %w", err)
	}
	if err := watcher.Add(file); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("could not watch for changes in %s: %w", file, err)
	}
	return watcher, nil
}

// notifyOnChange owns watcher and closes it when ctx is done. The proxy creates
// a new watcher on every restart, so a watcher left open here leaks one inotify
// instance per restart until fs.inotify.max_user_instances is exhausted.
func notifyOnChange(ctx context.Context, watcher *fsnotify.Watcher, ch chan<- struct{}) error {
	defer watcher.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-watcher.Events:
			select {
			case <-ctx.Done():
				return nil
			case ch <- struct{}{}:
			}
		case err := <-watcher.Errors:
			log.Printf("WARNING: error while watching file: %q\n", err)
		}
	}
}
