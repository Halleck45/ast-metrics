package Watcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/halleck45/ast-metrics/src/Command"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Storage"
	log "github.com/sirupsen/logrus"
)

type CommandWatcher struct {
	// The path to the sources to analyze
	SourcesToAnalyzePath []string

	// Configuration
	Configuration *Configuration.Configuration
}

func NewCommandWatcher(configuration *Configuration.Configuration) *CommandWatcher {
	return &CommandWatcher{
		Configuration:        configuration,
		SourcesToAnalyzePath: configuration.SourcesToAnalyzePath,
	}
}

func (c *CommandWatcher) Start(command *Command.AnalyzeCommand) error {
	if !c.Configuration.Watching {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	command.FileWatcher = watcher

	if err != nil {
		log.Fatal(err)
	}

	// get subdirectories in all sources
	subdirectories := []string{}
	for _, path := range c.SourcesToAnalyzePath {
		// explore the directory
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				subdirectories = append(subdirectories, path)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Ignore swp files
				if filepath.Ext(event.Name) == ".swp" {
					continue
				}
				// when last character is ~, it's a backup file
				if event.Name[len(event.Name)-1] == '~' {
					continue
				}

				// get file concerned by the event, and remove it from cache
				Storage.DeleteCache(event.Name)

				// Re-execute analyze command
				command.Execute()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	// watch all subdirectories
	for _, dir := range subdirectories {
		err = watcher.Add(dir)
		if err != nil {
			return err
		}
	}

	return nil
}
