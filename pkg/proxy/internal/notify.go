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
		return nil, fmt.Errorf("could not watch for changes in %s: %w", file, err)
	}
	return watcher, nil
}

func notifyOnChange(ctx context.Context, watcher *fsnotify.Watcher, ch chan<- struct{}) error {
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
