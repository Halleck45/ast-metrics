package Php

import (
    "sync"
    "embed"
    "io/fs"
    "os"
    "path/filepath"
    log "github.com/sirupsen/logrus"
    "os/exec"
    "strings"
    "strconv"
    "github.com/yargevad/filepathx"
    "crypto/md5"
    "encoding/hex"
    "io"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
    "github.com/halleck45/ast-metrics/src/Docker"
    "github.com/halleck45/ast-metrics/src/Driver"
    "github.com/docker/docker/api/types/mount"
)

// This allows to embed PHP sources in GO binary
//go:embed phpsources
var phpSources embed.FS

type PhpRunner struct {
    progressbar *pterm.SpinnerPrinter
    sourcesToAnalyzePath string
    driver Driver.Driver
}

func (r PhpRunner) IsRequired() (bool) {
    return true
}

func (r *PhpRunner) SetProgressbar(progressbar *pterm.SpinnerPrinter) {
    (*r).progressbar = progressbar
}

func (r *PhpRunner) SetSourcesToAnalyzePath(path string) {
    (*r).sourcesToAnalyzePath = path
}

func (r *PhpRunner) SetDriver(driver Driver.Driver) {
    (*r).driver = driver
}

func (r *PhpRunner) getContainerOutDirectory() string {
    return "/root/output"
}
func (r *PhpRunner) getLocalOutDirectory() string {
    return Storage.Path() + "/output"
}

func (r PhpRunner) Ensure() (error) {

    // clean up
    cleanup(phpSources)

    // Install sources locally (vendors)
    tempDir := Storage.Path() + "/.temp"
    if err := os.Mkdir(tempDir, 0755); err != nil {
        log.Fatal(err)
        return err
    }

    // Extract PHP sources for directories "vendor", etc
    if err := fs.WalkDir(phpSources, ".", func(path string, d fs.DirEntry, err error) error {

        if err != nil {
            return err
        }

        if d.Type().IsRegular() {
            content, err := phpSources.ReadFile(path)
            if err != nil {
                return err
            }
            outputPath := tempDir + "/" + path
            if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
                return err
            }
            if err := os.WriteFile(outputPath, content, 0644); err != nil {
                return err
            }
        }
        return nil
    }); err != nil {
        log.Fatal(err)
        return err
    }

    // Ensure outdir exists
    if _, err := os.Stat(r.getLocalOutDirectory()); os.IsNotExist(err) {
        if err := os.Mkdir(r.getLocalOutDirectory(), 0755); err != nil {
            log.Fatal(err)
            return err
        }
    }

    var phpVersion string
    if r.driver == Driver.Docker {
        // Pull
        imageName := "php:8.1-cli-alpine"
        var wg sync.WaitGroup
        wg.Add(1)
        r.progressbar.UpdateText("🐘 Pulling docker " + imageName + " image")
        go Docker.PullImage(&wg, r.progressbar, imageName)
        wg.Wait()

        // Run container
        // do not mount /tmp : permissions issues
        mounts := []mount.Mount{
           {
               Type:     mount.TypeBind,
               Source:   Storage.Path() + "/.temp/phpsources",
               Target:   "/tmp/engine",
               ReadOnly: true,
           },
           {
              Type:     mount.TypeBind,
                Source:   r.sourcesToAnalyzePath,
                Target:   "/tmp/sources",
                ReadOnly: true,
            },
           {
              Type:     mount.TypeBind,
                Source:   r.getLocalOutDirectory(),
                Target:   r.getContainerOutDirectory(),
                ReadOnly: false,

            },
        }
        // Create and start container. We want a deamonized, container, with an infinite loop. Loop stops when /tmp/engine is deleted
        loopString :=  []string{"sh", "-c", "until [ ! -f /tmp/engine/dump.php ]; do echo wait; sleep 1; done"}
        Docker.RunImage(imageName, "ast-php", mounts, loopString)
    }

    // Execute command
    r.progressbar.UpdateText("Checking PHP version")

    if r.driver == Driver.Docker {
        command := []string{"sh", "-c", "php -r 'echo PHP_VERSION;' > " + r.getContainerOutDirectory() + "/php_version"}
        Docker.ExecuteInRunningContainer("ast-php", command)
    } else {
        phpBinaryPath := getPHPBinaryPath()
        cmd := exec.Command("sh", "-c" , phpBinaryPath +  " -r 'echo PHP_VERSION;' > " +  r.getLocalOutDirectory() + "/php_version")
        if err := cmd.Run(); err != nil {
            log.Fatal(err)
            return err
        }
    }

    // get content of local file
    phpVersionBytes, err := os.ReadFile(r.getLocalOutDirectory() + "/php_version")
    if err != nil {
        log.Fatal(err)
        r.progressbar.Fail("Error while checking PHP version")
        return  err
    }
    phpVersion = string(phpVersionBytes)

    r.progressbar.Info("🐘 PHP " + phpVersion+ " is ready")
    r.progressbar.Stop()

    return nil
}

func (r PhpRunner) DumpAST() {

    // list all .php file in path, recursively
    path := strings.TrimRight(r.sourcesToAnalyzePath, "/")

    matches, err := filepathx.Glob(path + "/**/*.php")
    if err != nil {
        r.progressbar.Fail("Error while listing PHP files")
        return
    }

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

    nbParsingFiles := 0
    sem := make(chan struct{}, maxParallelCommandsInt)
    for _, file := range matches {
        if !strings.Contains(file, "/vendor/") {
            wg.Add(1)
            nbParsingFiles++
            sem <- struct{}{}
            go func(file string) {
                defer wg.Done()
                r.executePHPCommandForFile(workDir, file, path)

                // details is the number of files processed / total number of files
                details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(len(matches))
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

func (r PhpRunner) Finish() (error) {
    cleanup(phpSources)
    return nil
}

func cleanup(phpSources embed.FS ) (string, error) {
    // Remove temp directory
    tempDir := Storage.Path() + "/.temp"

    // check if tempDir exists
    if _, err := os.Stat(tempDir); os.IsNotExist(err) {
        return "", nil
    }
    if err := os.RemoveAll(tempDir); err != nil {
        log.Fatal(err)
        return "", err
    }

    return "", nil
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

func (r PhpRunner)  executePHPCommandForFile(tmpDir string, file string, path string) {

    hash, err := getFileHash(file)
    if err != nil {
        log.Printf("Cannot get hash for file %s : %v\n", file, err)
        return
    }
    outputFilePath := filepath.Join(tmpDir, hash + ".bin")

    relativePath := strings.Replace(file, path, "", 1)
    relativePath = strings.TrimLeft(relativePath, "/")

    // if file already exists, skip
    if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
        return
    }

    if r.driver == Driver.Docker {
        containerOutputFilePath := r.getContainerOutDirectory() + "/" + hash + ".bin"
        command := "(php /tmp/engine/dump.php /tmp/sources/" + relativePath + " > " + containerOutputFilePath + ") || rm " + containerOutputFilePath
        Docker.ExecuteInRunningContainer("ast-php", []string{"sh", "-c", command})
    } else {
        phpBinaryPath := getPHPBinaryPath()
        tempDir := Storage.Path() + "/.temp"
        cmd := exec.Command("sh", "-c" , phpBinaryPath + " " + tempDir + "/phpsources/dump.php " + file + " > " + outputFilePath)
        if err := cmd.Run(); err != nil {
            log.Printf("Cannot execute command %s : %v\n", cmd.String(), err)
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