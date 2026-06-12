package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/user/file-sync-tool/scan"
	"github.com/user/file-sync-tool/sync"
	"github.com/user/file-sync-tool/types"
	"github.com/user/file-sync-tool/watch"
)

var defaultStateDir = filepath.Join(os.Getenv("HOME"), ".file-sync-tool")

func main() {
	watchMode := flag.Bool("watch", false, "Live sync via fsnotify")
	interval := flag.Duration("interval", 30*time.Second, "Poll interval between syncs (0 = run once)")
	dryRun := flag.Bool("dry-run", false, "Show changes without copying")
	deleteFlag := flag.Bool("delete", false, "Remove dest files absent in source")
	statePath := flag.String("state", filepath.Join(defaultStateDir, "state.json"), "Path to state file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: file-sync-tool [flags] <source> <destination>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	sourceDir := flag.Arg(0)
	destDir := flag.Arg(1)

	if err := validateDirs(sourceDir, destDir); err != nil {
		log.Fatal(err)
	}

	var state *types.SyncState
	st, err := sync.LoadState(*statePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("warning: could not load state: %v (starting fresh)", err)
		}
		state = &types.SyncState{
			Source: sourceDir,
			Dest:   destDir,
			Files:  make(map[string]types.FileRecord),
		}
	} else {
		state = st
	}

	if *watchMode {
		runWatch(sourceDir, destDir, *statePath, *dryRun, *deleteFlag, state)
	} else {
		runPoll(sourceDir, destDir, *statePath, *dryRun, *deleteFlag, state, *interval)
	}
}

func validateDirs(sourceDir, destDir string) error {
	srcInfo, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", sourceDir)
	}

	dstInfo, err := os.Stat(destDir)
	if err != nil {
		return fmt.Errorf("destination: %w", err)
	}
	if !dstInfo.IsDir() {
		return fmt.Errorf("destination is not a directory: %s", destDir)
	}

	absSrc, _ := filepath.Abs(sourceDir)
	absDst, _ := filepath.Abs(destDir)
	if absSrc == absDst {
		return fmt.Errorf("source and destination must be different directories")
	}

	return nil
}

func runPoll(sourceDir, destDir, statePath string, dryRun, delete bool, state *types.SyncState, interval time.Duration) {
	for {
		if err := syncOnce(sourceDir, destDir, statePath, dryRun, delete, state); err != nil {
			log.Printf("sync error: %v", err)
		}
		if interval == 0 {
			break
		}
		time.Sleep(interval)
	}
}

func syncOnce(sourceDir, destDir, statePath string, dryRun, delete bool, state *types.SyncState) error {
	records, err := scan.Tree(sourceDir)
	if err != nil {
		return fmt.Errorf("scan source: %w", err)
	}

	changes := sync.DetectChanges(records, state.Files, delete)

	if len(changes) == 0 {
		fmt.Println("no changes detected")
		return nil
	}

	fmt.Printf("detected %d change(s)\n", len(changes))

	if err := sync.ApplyChanges(changes, sourceDir, destDir, dryRun); err != nil {
		return fmt.Errorf("apply changes: %w", err)
	}

	if !dryRun {
		var deletedPaths []string
		for _, c := range changes {
			if c.Type == types.Deleted {
				deletedPaths = append(deletedPaths, c.RelPath)
			}
		}
		sync.ApplyChangesToState(state, records, deletedPaths)
		if err := sync.SaveState(state, statePath); err != nil {
			return fmt.Errorf("save state: %w", err)
		}
	}

	return nil
}

func runWatch(sourceDir, destDir, statePath string, dryRun, delete bool, state *types.SyncState) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("watching %s for changes...", sourceDir)
		if err := watch.Start(sourceDir, 100*time.Millisecond, func(paths []string) {
			log.Printf("change detected (%d file(s))", len(paths))
			if err := syncOnce(sourceDir, destDir, statePath, dryRun, delete, state); err != nil {
				log.Printf("sync error: %v", err)
			}
		}); err != nil {
			log.Printf("watcher error: %v", err)
		}
	}()

	<-sig
	log.Println("shutting down...")

	if !dryRun {
		if err := sync.SaveState(state, statePath); err != nil {
			log.Printf("error saving state: %v", err)
		}
	}
}
