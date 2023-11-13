package CommandExecutor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/mount"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Docker"
	"github.com/halleck45/ast-metrics/src/Driver"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type CommandExecutor struct {
	EmbeddedWorkspace EmbeddedWorkspace
	Configuration     Configuration.Configuration
	dockerImageName   string
	progressbar       *pterm.SpinnerPrinter
}

func NewCommandExecutor(Configuration Configuration.Configuration, dockerImageName string, embeddedWorkspace EmbeddedWorkspace, progressbar *pterm.SpinnerPrinter) CommandExecutor {
	return CommandExecutor{
		Configuration:     Configuration,
		dockerImageName:   dockerImageName,
		progressbar:       progressbar,
		EmbeddedWorkspace: embeddedWorkspace,
	}
}

func (r CommandExecutor) Ensure(nameOfContainer string) error {

	if r.Configuration.Driver == Driver.Native {
		// nothing to do when using native driver
		return nil
	}

	// Pull Docker image
	var wg sync.WaitGroup
	wg.Add(1)
	r.progressbar.UpdateText("🐘 Pulling docker " + r.dockerImageName + " image")
	go Docker.PullImage(&wg, r.progressbar, r.dockerImageName)
	wg.Wait()

	// Run container
	// do not mount /tmp : permissions issues
	debugMountsAsString := ""
	mounts := []mount.Mount{}

	// Mount Embedded workspace
	mounts = append(mounts, mount.Mount{
		Type:     mount.TypeBind,
		Source:   r.EmbeddedWorkspace.GetPath(),
		Target:   "/tmp/engine",
		ReadOnly: true,
	})

	debugMountsAsString += " -v " + r.EmbeddedWorkspace.GetPath() + ":/tmp/engine:ro"

	// Mount output directory
	mounts = append(mounts, mount.Mount{
		Type:     mount.TypeBind,
		Source:   r.getLocalOutDirectory(),
		Target:   r.getContainerOutDirectory(),
		ReadOnly: false,
	})
	debugMountsAsString += " -v " + r.getLocalOutDirectory() + ":" + r.getContainerOutDirectory() + ":rw"

	// for each path to analyze, add a mount
	for index, path := range r.Configuration.SourcesToAnalyzePath {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   path,
			Target:   "/tmp/sources" + strconv.Itoa(index),
			ReadOnly: true,
		})
		debugMountsAsString += " -v " + path + ":/tmp/sources" + strconv.Itoa(index) + ":ro"
	}

	// Create and start container. We want a deamonized, container, with an infinite loop. Loop stops when /tmp/engine is deleted
	loopString := []string{"sh", "-c", "until [ ! -f /tmp/engine/.keepme ]; do echo wait; sleep 1; done"}
	Docker.RunImage(r.dockerImageName, nameOfContainer, mounts, loopString)

	// Give to the user the CLI command to run the same container with the same options
	if log.GetLevel() == log.DebugLevel {
		log.Debug("🐋 Run the following command to run the same container with the same options :")
		log.Debug("🐋 docker run --rm -it " + debugMountsAsString + " " + r.dockerImageName + " " + strings.Join(loopString, " "))
		fmt.Println("🐋 docker run --rm -it " + debugMountsAsString + " " + r.dockerImageName + " " + strings.Join(loopString, " "))
	}

	return nil
}

func (r CommandExecutor) ExecuteAndReturnsOutput(dockerContainerName string, commandToRun string, outputDestinationPath string) (string, error) {

	if r.Configuration.Driver == Driver.Docker {
		command := []string{"sh", "-c", commandToRun + " > " + r.getContainerOutDirectory() + string(os.PathSeparator) + outputDestinationPath}
		Docker.ExecuteInRunningContainer(dockerContainerName, command)
	} else {
		cmd := exec.Command("sh", "-c", commandToRun+" > "+r.getLocalOutDirectory()+string(os.PathSeparator)+outputDestinationPath)
		if err := cmd.Run(); err != nil {
			log.Error("Cannot execute command: \n", cmd.String(), err)
			log.Error(err)
			return "", err
		}
	}

	// get content of local file
	resultInBytes, err := os.ReadFile(r.getLocalOutDirectory() + string(os.PathSeparator) + outputDestinationPath)
	if err != nil {
		log.Error("Cannot read file " + r.getLocalOutDirectory() + string(os.PathSeparator) + outputDestinationPath)
		r.progressbar.Fail("Error while checking PHP version")
		return "", err
	}

	return string(resultInBytes), nil
}

func (r CommandExecutor) FileExists(path string) bool {
	absolutePath := r.getLocalOutDirectory() + string(os.PathSeparator) + path

	if _, err := os.Stat(absolutePath); !os.IsNotExist(err) {
		return true
	}

	return false
}

func (r CommandExecutor) getLocalOutDirectory() string {
	return Storage.Path() + "/output"
}

func (r CommandExecutor) GetRelativePath(file string, currentlyAnalysedDirectory string) (string, error) {
	if r.Configuration.Driver == Driver.Native {
		return file, nil
	}

	// Get the index of the directory in the list of directories to analyze
	// Each directory is mounted in a different directory in the container
	// /tmp/sources0, /tmp/sources1, etc
	var directoryIndex int = -1
	for index, directory := range r.Configuration.SourcesToAnalyzePath {
		if directory == currentlyAnalysedDirectory {
			directoryIndex = index
			break
		}
	}

	if directoryIndex == -1 {
		log.Error("Cannot find directory " + currentlyAnalysedDirectory + " in list of directories to analyze")

		if log.GetLevel() == log.DebugLevel {
			log.Debug("Directories map looks incorrect : " + strings.Join(r.Configuration.SourcesToAnalyzePath, ", ") + "\n")
		}
		return "", errors.New("Cannot find directory " + currentlyAnalysedDirectory + " in list of directories to analyze")
	}

	dirInDocker := "/tmp/sources" + strconv.Itoa(directoryIndex)
	relativePath := strings.Replace(file, currentlyAnalysedDirectory, dirInDocker, 1)
	relativePath = strings.TrimLeft(relativePath, "/")

	return relativePath, nil
}

func (r CommandExecutor) getContainerOutDirectory() string {
	return "/root/output"
}

func (r CommandExecutor) GetEmbeddedWorkspacePath(path string) string {
	if r.Configuration.Driver == Driver.Native {
		return r.EmbeddedWorkspace.GetPath() + string(os.PathSeparator) + path
	}

	// make it relative to docker mounted dir
	path = strings.Replace(path, r.EmbeddedWorkspace.GetPath(), "", 1)
	return "/tmp/engine/" + path
}
