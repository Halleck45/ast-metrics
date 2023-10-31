package Php

import (
    "sync"
    "errors"
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
    "github.com/docker/docker/api/types/mount"
)

func Ensure(progressbar *pterm.SpinnerPrinter, phpSources embed.FS, sourcesToAnalyzePath string) (string, error) {

    // clean up
    cleanup(phpSources)

    // Install sources locally (vendors)
    tempDir := Storage.Path() + "/.temp"
    if err := os.Mkdir(tempDir, 0755); err != nil {
        log.Fatal(err)
    }

    // Extract PHP sources for directories "engine/php/vendor", etc
    if err := fs.WalkDir(phpSources, "engine/php", func(path string, d fs.DirEntry, err error) error {
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
    }


    var useDocker bool = true // @todo add an option
    var phpVersion string
    if useDocker {

        imageName := "php:8.1-cli-alpine"
        // Pull
        var wg sync.WaitGroup
        wg.Add(1)
        progressbar.UpdateText("Pulling docker " + imageName + " image")
        go Docker.PullImage(&wg, progressbar, imageName)
        wg.Wait()

        // Ensure outdir exists
        if _, err := os.Stat(getLocalOutDirectory()); os.IsNotExist(err) {
            if err := os.Mkdir(getLocalOutDirectory(), 0755); err != nil {
                log.Fatal(err)
            }
        }

        // Run container
        // do not mount /tmp : permissions issues
        mounts := []mount.Mount{
           {
               Type:     mount.TypeBind,
               Source:   Storage.Path() + "/.temp/engine/php",
               Target:   "/tmp/engine",
               ReadOnly: true,
           },
           {
              Type:     mount.TypeBind,
                Source:   sourcesToAnalyzePath,
                Target:   "/tmp/sources",
                ReadOnly: true,
            },
           {
              Type:     mount.TypeBind,
                Source:   getLocalOutDirectory(),
                Target:   getContainerOutDirectory(),
                ReadOnly: false,

            },
        }
        // Create and start container. We want a deamonized, container, with an infinite loop. Loop stops when /tmp/engine is deleted
        loopString :=  []string{"sh", "-c", "until [ ! -f /tmp/engine/dump.php ]; do echo wait; sleep 1; done"}
        Docker.RunImage(imageName, "ast-php", mounts, loopString)

        // Execute command in container
        progressbar.UpdateText("Checking PHP version")
        command := []string{"sh", "-c", "php -r 'echo PHP_VERSION;' > " + getContainerOutDirectory() + "/php_version"}
        Docker.ExecuteInRunningContainer("ast-php", command)
        // get content of local file
        phpVersionBytes, err := os.ReadFile(getLocalOutDirectory() + "/php_version")
        if err != nil {
            log.Fatal(err)
            progressbar.Fail("Error while checking PHP version")
            return "", err
        }
        phpVersion = string(phpVersionBytes)

        progressbar.Info("PHP " + phpVersion+ " is ready")
        progressbar.Stop()

        return phpVersion, nil
    } else {
        // Get PHP binary path. IF env PHP_BINARY_PATH is not set, use default value
        phpBinaryPath := getPHPBinaryPath()

        // Get PHP version
        phpVersion := getPHPVersion(phpBinaryPath)

        // if version is empty, throw error
        if phpVersion == "" {
            return "", errors.New("Cannot get PHP version using the PHP binary path: " + phpBinaryPath + ". Please check if PHP is installed, or set the PHP_BINARY_PATH environment variable to the correct path.")
        }

        progressbar.UpdateText("PHP " + phpVersion)
        progressbar.Info("PHP " + phpVersion + " is ready")
        defer progressbar.Stop()
    }


    return phpVersion, nil
}

func DumpAST(progressbar *pterm.SpinnerPrinter, path string) {

    // list all .php file in path, recursively
    path = strings.TrimRight(path, "/")

    matches, err := filepathx.Glob(path + "/**/*.php")
    if err != nil {
        progressbar.Fail("Error while listing PHP files")
    }

    maxParallelCommands := os.Getenv("MAX_PARALLEL_COMMANDS")
    // if maxParallelCommands is empty, set default value
    if maxParallelCommands == "" {
        maxParallelCommands = "100"
    }
    // to int
    maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
    if err != nil {
        progressbar.Fail("Error while parsing MAX_PARALLEL_COMMANDS env variable")
    }

    workDir := getLocalOutDirectory()

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
                executePHPCommandForFile(workDir, file, path)

                // details is the number of files processed / total number of files
                details := strconv.Itoa(nbParsingFiles) + "/" + strconv.Itoa(len(matches))
                progressbar.UpdateText("Parsing PHP files (" + details + ")")
                <-sem
            }(file)
        }
    }

    // Wait for all goroutines to finish
    for i := 0; i < maxParallelCommandsInt; i++ {
        sem <- struct{}{}
    }

    wg.Wait()
    progressbar.Info("PHP analysis finished")
}

func Finish(progressbar *pterm.SpinnerPrinter, phpSources embed.FS ) (string, error) {
    cleanup(phpSources)
    //Docker.RemoveContainer("ast-php")
    progressbar.Info("AST dumped for PHP files")
    //progressbar.Stop()

    return "", nil
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

func executePHPCommandForFile(tmpDir string, file string, path string) {

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

    phpBinaryPath := getPHPBinaryPath()
    containerOutputFilePath := getContainerOutDirectory() + "/" + hash + ".bin"
    command := "(" + phpBinaryPath + " /tmp/engine/dump.php /tmp/sources/" + relativePath + " > " + containerOutputFilePath + ") || rm " + containerOutputFilePath
    Docker.ExecuteInRunningContainer("ast-php", []string{"sh", "-c", command})
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

func getPHPVersion(phpBinaryPath string) string {
    cmd := exec.Command(phpBinaryPath, "-v")
    out, err := cmd.CombinedOutput()
    if err != nil {
        return ""
    }

    outString  := string(out)
    outString  = outString[0:10]
    outString  = outString[4:10]

    // trim
    outString = strings.TrimSpace(outString)

    return outString
}

func getContainerOutDirectory() string {
    return "/root/output"
}
func getLocalOutDirectory() string {
    return Storage.Path() + "/output"
}