package Php

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/halleck45/ast-metrics/src/CommandExecutor"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Driver"
	"github.com/halleck45/ast-metrics/src/File"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

// This allows to embed PHP sources in GO binary
//
//go:embed phpsources
var phpSources embed.FS

// PhpRunner is the runner for PHP
type PhpRunner struct {
	progressbar               *pterm.SpinnerPrinter
	configuration             *Configuration.Configuration
	foundFiles                File.FileList
	workspaceOfSourceAnalyzer CommandExecutor.EmbeddedWorkspace
	commandExecutor           CommandExecutor.CommandExecutor
}

// IsRequired returns true if at least one PHP file is found
func (r PhpRunner) IsRequired() bool {
	// If at least one PHP file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

// SetProgressbar sets the progressbar
func (r *PhpRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

// SetConfiguration sets the configuration
func (r *PhpRunner) SetConfiguration(configuration *Configuration.Configuration) {
	(*r).configuration = configuration
}

// Ensure ensures PHP is ready to run. It pulls the Docker image if needed, and runs the container if needed.
// It also try to obtain the PHP version, in order to check if PHP is ready.
func (r *PhpRunner) Ensure() error {

	// Create workspace for PHP sources. Sources are embedded in the binary
	r.workspaceOfSourceAnalyzer = CommandExecutor.NewEmbeddedWorkspace("PHP", phpSources)
	err := r.workspaceOfSourceAnalyzer.Ensure()
	if err != nil {
		return err
	}

	// Create command executor, allowing to run commands in the container or natively
	r.commandExecutor = CommandExecutor.NewCommandExecutor(*r.configuration, "php:8.1-cli-alpine", r.workspaceOfSourceAnalyzer, r.progressbar)

	// Ensure outdir exists
	if _, err := os.Stat(r.getLocalOutDirectory()); os.IsNotExist(err) {
		if err := os.Mkdir(r.getLocalOutDirectory(), 0755); err != nil {
			log.Error(err)
			return err
		}
	}

	// Ensure container is pulled and running (if needed)
	r.commandExecutor.Ensure("ast-php")

	// Get PHP version
	var phpVersion string
	r.progressbar.UpdateText("Checking PHP version")
	commandToExecute := r.getPHPBinaryPath() + " -r 'echo PHP_VERSION;'"
	phpVersion, err = r.commandExecutor.ExecuteAndReturnsOutput("ast-php", commandToExecute, "phpversion.txt")
	if err != nil {
		log.Error("Cannot get PHP version")
		r.progressbar.Fail("Error while checking PHP version")
		return err
	}

	// Inform user
	r.progressbar.Info("🐘 PHP " + phpVersion + " is ready")
	r.progressbar.Stop()

	return nil
}

// DumpAST dumps the AST of PHP files in protobuf format
// It uses a independant PHP program (dump.php) to dump the AST of PHP files
func (r PhpRunner) DumpAST() {

	maxParallelCommands := os.Getenv("MAX_PARALLEL_COMMANDS")
	if maxParallelCommands == "" {
		// if maxParallelCommands is empty, set default value
		maxParallelCommands = "100"
	}
	maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
	if err != nil {
		r.progressbar.Fail("Error while parsing MAX_PARALLEL_COMMANDS env variable")
		return
	}

	// Wait for end of all goroutines
	var wg sync.WaitGroup
	var nbFiles int = len(r.getFileList().Files)

	nbParsingFiles := 0
	sem := make(chan struct{}, maxParallelCommandsInt)

	for directory, files := range r.getFileList().FilesByDirectory {
		// We iterate over the list of files to analyze, and we run a goroutine for each file

		var directoryToAnalyze string = directory //  Please keep this intermediate vareiable. Avoid 'directory captured by func literal' error
		for _, file := range files {
			wg.Add(1)
			nbParsingFiles++
			sem <- struct{}{}
			go func(file string) {
				defer wg.Done()
				r.executePHPCommandForFile(r.getLocalOutDirectory(), directoryToAnalyze, file)

				// details is the number of files processed / total number of files
				details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(nbFiles)
				r.progressbar.UpdateText("🐘 Parsing PHP files (" + details + ")")
				<-sem
			}(file)
		}
	}

	// Wait for all goroutines to finish
	for i := 0; i < maxParallelCommandsInt; i++ {
		sem <- struct{}{}
	}

	wg.Wait()
	r.progressbar.Info("🐘 PHP code dumped")
}

// Finish cleans up the workspace
func (r PhpRunner) Finish() error {
	r.workspaceOfSourceAnalyzer.Cleanup()
	return nil
}

// Provides the hash of a file, in order to avoid to parse it twice
func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// This subroutine executes the PHP command to dump the AST of a file
func (r PhpRunner) executePHPCommandForFile(tmpDir string, currentlyAnalysedDirectory string, file string) {

	hash, err := getFileHash(file)
	if err != nil {
		log.Printf("Cannot get hash for file %s : %v\n", file, err)
		return
	}

	relativeFilePath, err := r.commandExecutor.GetRelativePath(file, currentlyAnalysedDirectory)
	if err != nil {
		log.Printf("Cannot get relative path for file %s : %v\n", file, err)
		return
	}

	outputFilePath := hash + ".bin"
	if r.commandExecutor.FileExists(outputFilePath) {
		// if file already exists, skip
		return
	}

	// Execute command
	command := r.getPHPBinaryPath() + " " +
		r.commandExecutor.GetEmbeddedWorkspacePath("phpsources/dump.php") + " " +
		relativeFilePath
	r.commandExecutor.ExecuteAndReturnsOutput("ast-php", command, outputFilePath)
}

// getPHPBinaryPath returns the path to the PHP binary, depending on the driver
func (r PhpRunner) getPHPBinaryPath() string {

	if r.configuration.Driver == Driver.Docker {
		return "php"
	}

	phpBinaryPath := os.Getenv("PHP_BINARY_PATH")
	if phpBinaryPath == "" {
		phpBinaryPath = "php"
	}

	return phpBinaryPath
}

// getLocalOutDirectory returns the path to the local output directory
func (r *PhpRunner) getLocalOutDirectory() string {
	return Storage.Path() + "/output"
}

// getFileList returns the list of PHP files to analyze, and caches it in memory
func (r *PhpRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".php")

	return r.foundFiles
}
