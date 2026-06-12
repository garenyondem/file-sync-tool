package watch

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type debouncer struct {
	mu       sync.Mutex
	events   map[string]bool
	timer    *time.Timer
	interval time.Duration
	fn       func(paths []string)
	done     chan struct{}
}

func newDebouncer(interval time.Duration, fn func(paths []string)) *debouncer {
	return &debouncer{
		events:   make(map[string]bool),
		interval: interval,
		fn:       fn,
		done:     make(chan struct{}),
	}
}

func (d *debouncer) add(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.events[path] = true

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.interval, d.fire)
}

func (d *debouncer) fire() {
	d.mu.Lock()
	if len(d.events) == 0 {
		d.mu.Unlock()
		return
	}
	paths := make([]string, 0, len(d.events))
	for p := range d.events {
		paths = append(paths, p)
	}
	d.events = make(map[string]bool)
	d.mu.Unlock()

	d.fn(paths)
}

func (d *debouncer) stop() {
	close(d.done)
	if d.timer != nil {
		d.timer.Stop()
	}
}

func Start(root string, debounce time.Duration, onChange func(paths []string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	deb := newDebouncer(debounce, onChange)
	defer deb.stop()

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				if event.Op&(fsnotify.Create|fsnotify.Rename) != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						watcher.Add(event.Name)
					}
				}
				deb.add(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("watch error: %v", err)

		case <-deb.done:
			return nil
		}
	}
}
