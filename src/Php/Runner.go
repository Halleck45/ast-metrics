package Php

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/mount"
	"github.com/halleck45/ast-metrics/src/CommandExecutor"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Docker"
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

type PhpRunner struct {
	progressbar               *pterm.SpinnerPrinter
	configuration             *Configuration.Configuration
	foundFiles                File.FileList
	workspaceOfSourceAnalyzer CommandExecutor.EmbeddedWorkspace
}

func (r PhpRunner) IsRequired() bool {
	// If at least one PHP file is found, we need to run PHP engine
	return len(r.getFileList().Files) > 0
}

func (r *PhpRunner) getFileList() File.FileList {

	if r.foundFiles.Files != nil {
		return r.foundFiles
	}

	finder := File.Finder{Configuration: *r.configuration}
	r.foundFiles = finder.Search(".php")

	return r.foundFiles
}

func (r *PhpRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
	(*r).progressbar = progressbar
}

func (r *PhpRunner) SetConfiguration(configuration *Configuration.Configuration) {
	(*r).configuration = configuration
}

func (r *PhpRunner) getContainerOutDirectory() string {
	return "/root/output"
}
func (r *PhpRunner) getLocalOutDirectory() string {
	return Storage.Path() + "/output"
}

func (r *PhpRunner) Ensure() error {

	r.workspaceOfSourceAnalyzer = CommandExecutor.EmbeddedWorkspace{Name: "PHP", PathToLocalSources: phpSources}
	err := r.workspaceOfSourceAnalyzer.Ensure()

	if err != nil {
		log.Error(err)
		return err
	}

	// Ensure outdir exists
	if _, err := os.Stat(r.getLocalOutDirectory()); os.IsNotExist(err) {
		if err := os.Mkdir(r.getLocalOutDirectory(), 0755); err != nil {
			log.Error(err)
			return err
		}
	}

	var phpVersion string
	if r.configuration.Driver == Driver.Docker {
		// Pull
		imageName := "php:8.1-cli-alpine"
		var wg sync.WaitGroup
		wg.Add(1)
		r.progressbar.UpdateText("🐘 Pulling docker " + imageName + " image")
		go Docker.PullImage(&wg, r.progressbar, imageName)
		wg.Wait()

		// Run container
		// do not mount /tmp : permissions issues
		debugMountsAsString := ""
		mounts := []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   r.workspaceOfSourceAnalyzer.GetPath() + "/phpsources",
				Target:   "/tmp/engine",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   r.getLocalOutDirectory(),
				Target:   r.getContainerOutDirectory(),
				ReadOnly: false,
			},
		}
		debugMountsAsString += " -v " + r.workspaceOfSourceAnalyzer.GetPath() + "/phpsources:/tmp/engine:ro"
		debugMountsAsString += " -v " + r.getLocalOutDirectory() + ":" + r.getContainerOutDirectory() + ":rw"

		// for each path to analyze, add a mount
		for index, path := range r.configuration.SourcesToAnalyzePath {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   path,
				Target:   "/tmp/sources" + strconv.Itoa(index),
				ReadOnly: true,
			})
			debugMountsAsString += " -v " + path + ":/tmp/sources" + strconv.Itoa(index) + ":ro"
		}

		// Create and start container. We want a deamonized, container, with an infinite loop. Loop stops when /tmp/engine is deleted
		loopString := []string{"sh", "-c", "until [ ! -f /tmp/engine/dump.php ]; do echo wait; sleep 1; done"}
		Docker.RunImage(imageName, "ast-php", mounts, loopString)

		// Give to the user the CLI command to run the same container with the same options
		if log.GetLevel() == log.DebugLevel {
			log.Debug("🐘 Run the following command to run the same container with the same options :")
			log.Debug("docker run --rm -it " + debugMountsAsString + " " + imageName + " " + strings.Join(loopString, " "))
		}
	}

	// Execute command
	r.progressbar.UpdateText("Checking PHP version")

	if r.configuration.Driver == Driver.Docker {
		command := []string{"sh", "-c", "php -r 'echo PHP_VERSION;' > " + r.getContainerOutDirectory() + "/php_version"}
		Docker.ExecuteInRunningContainer("ast-php", command)
	} else {
		phpBinaryPath := getPHPBinaryPath()
		cmd := exec.Command("sh", "-c", phpBinaryPath+" -r 'echo PHP_VERSION;' > "+r.getLocalOutDirectory()+"/php_version")
		if err := cmd.Run(); err != nil {
			log.Error("Cannot execute command: \n", cmd.String(), err)
			log.Error(err)
			return err
		}
	}

	// get content of local file
	phpVersionBytes, err := os.ReadFile(r.getLocalOutDirectory() + "/php_version")
	if err != nil {
		log.Error("Cannot read file " + r.getLocalOutDirectory() + "/php_version")
		r.progressbar.Fail("Error while checking PHP version")
		return err
	}
	phpVersion = string(phpVersionBytes)

	r.progressbar.Info("🐘 PHP " + phpVersion + " is ready")
	r.progressbar.Stop()

	return nil
}

func (r PhpRunner) DumpAST() {

	maxParallelCommands := os.Getenv("MAX_PARALLEL_COMMANDS")
	// if maxParallelCommands is empty, set default value
	if maxParallelCommands == "" {
		maxParallelCommands = "100"
	}
	// to int
	maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
	if err != nil {
		r.progressbar.Fail("Error while parsing MAX_PARALLEL_COMMANDS env variable")
		return
	}
	workDir := r.getLocalOutDirectory()

	// Wait for end of all goroutines
	var wg sync.WaitGroup
	var nbFiles int = len(r.getFileList().Files)

	nbParsingFiles := 0
	sem := make(chan struct{}, maxParallelCommandsInt)

	for directory, files := range r.getFileList().FilesByDirectory {

		var directoryToAnalyze string = directory //  directory captured by func literal
		for _, file := range files {
			wg.Add(1)
			nbParsingFiles++
			sem <- struct{}{}
			go func(file string) {
				defer wg.Done()
				r.executePHPCommandForFile(workDir, directoryToAnalyze, file)

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

func (r PhpRunner) Finish() error {
	r.workspaceOfSourceAnalyzer.Cleanup()
	return nil
}

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

func (r PhpRunner) executePHPCommandForFile(tmpDir string, currentlyAnalysedDirectory string, file string) {

	hash, err := getFileHash(file)
	if err != nil {
		log.Printf("Cannot get hash for file %s : %v\n", file, err)
		return
	}
	outputFilePath := filepath.Join(tmpDir, hash+".bin")

	if log.GetLevel() == log.DebugLevel {
		log.Debug("Dumping file " + file + " to " + outputFilePath)
	}

	// if file already exists, skip
	if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
		return
	}
	if r.configuration.Driver == Driver.Docker {
		// Get the index of the directory in the list of directories to analyze
		// Each directory is mounted in a different directory in the container
		// /tmp/sources0, /tmp/sources1, etc
		var directoryIndex int = -1
		for index, directory := range r.configuration.SourcesToAnalyzePath {
			if directory == currentlyAnalysedDirectory {
				directoryIndex = index
				break
			}
		}

		if directoryIndex == -1 {
			log.Error("Cannot find directory " + currentlyAnalysedDirectory + " in list of directories to analyze")

			if log.GetLevel() == log.DebugLevel {
				log.Debug("Directories map looks incorrect : " + strings.Join(r.configuration.SourcesToAnalyzePath, ", ") + "\n")
			}
			return
		}

		mountDirectory := "/tmp/sources" + strconv.Itoa(directoryIndex)

		relativePath := strings.Replace(file, currentlyAnalysedDirectory, "", 1)
		relativePath = strings.TrimLeft(relativePath, "/")

		containerOutputFilePath := r.getContainerOutDirectory() + "/" + hash + ".bin"
		command := "(php /tmp/engine/dump.php " + mountDirectory + "/" + relativePath + " > " + containerOutputFilePath + ") || rm " + containerOutputFilePath
		if log.GetLevel() == log.DebugLevel {
			log.Debug("Executing command in container : " + command + "\n")
		}

		Docker.ExecuteInRunningContainer("ast-php", []string{"sh", "-c", command})

	} else {
		phpBinaryPath := getPHPBinaryPath()
		tempDir := r.workspaceOfSourceAnalyzer.GetPath()
		if log.GetLevel() == log.DebugLevel {
			log.Debug("Executing command : " + phpBinaryPath + " " + tempDir + "/phpsources/dump.php " + file + " > " + outputFilePath)
		}
		cmd := exec.Command("sh", "-c", phpBinaryPath+" "+tempDir+"/phpsources/dump.php "+file+" > "+outputFilePath)
		if err := cmd.Run(); err != nil {
			log.Error("[SKIP] "+file+" - Cannot execute command %s :\n", cmd.String(), "\n", err)

			// remove file
			if err := os.Remove(outputFilePath); err != nil {
				log.Error("Cannot remove file "+outputFilePath+" : %v\n", err)
			}
			return
		}
	}
}

func getPHPBinaryPath() string {
	var useDocker bool = true // @todo add an option
	if useDocker {
		return "php"
	}

	phpBinaryPath := os.Getenv("PHP_BINARY_PATH")
	if phpBinaryPath == "" {
		phpBinaryPath = "php"
	}

	return phpBinaryPath
}
