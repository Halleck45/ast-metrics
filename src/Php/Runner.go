package Php

import (
    "sync"
    "errors"
    "embed"
    "io/fs"
    "os"
    "path/filepath"
    "log"
    "os/exec"
    "strings"
    "strconv"
    "github.com/yargevad/filepathx"
    "crypto/md5"
    "encoding/hex"
    "io"
    "io/ioutil"
    "bytes"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
)

func Ensure(progressbar *pterm.SpinnerPrinter, phpSources embed.FS ) (string, error) {

    // clean up
    cleanup(phpSources)

    // Install sources locally (vendors)
    tempDir := ".temp"
    if err := os.Mkdir(tempDir, 0755); err != nil {
        log.Fatal(err)
    }

    // Extract PHP sources for directories "runner/php/vendor", "runner/php/generated" and file "runner/php/dump.php"
    if err := fs.WalkDir(phpSources, "runner/php", func(path string, d fs.DirEntry, err error) error {
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

	// Get PHP binary path. IF env PHP_BINARY_PATH is not set, use default value
    phpBinaryPath := getPHPBinaryPath()

    // Get PHP version
    phpVersion := getPHPVersion(phpBinaryPath)

    // if version is empty, throw error
    if phpVersion == "" {
        return "", errors.New("Cannot get PHP version using the PHP binary path: " + phpBinaryPath + ". Please check if PHP is installed, or set the PHP_BINARY_PATH environment variable to the correct path.")
    }

    progressbar.Info("PHP is ready (v" + phpVersion + ")")
    progressbar.Stop()

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
        maxParallelCommands = "10"
    }
    // to int
    maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
    if err != nil {
        progressbar.Fail("Error while parsing MAX_PARALLEL_COMMANDS env variable")
    }

    workDir := Storage.Path()

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
                executePHPCommandForFile(workDir, file)

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
    progressbar.UpdateText("")
    progressbar.Info("PHP analysis finished")
    progressbar.Stop()
}

func Finish(progressbar *pterm.SpinnerPrinter, phpSources embed.FS ) (string, error) {
    cleanup(phpSources)
    progressbar.Info("Cleaned up")
    progressbar.Stop()
    return "", nil
}

func cleanup(phpSources embed.FS ) (string, error) {
    // Remove temp directory
    tempDir := ".temp"

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

func executePHPCommandForFile(tmpDir string, file string) {

    hash, err := getFileHash(file)
    if err != nil {
        log.Printf("Cannot get hash for file %s : %v\n", file, err)
        return
    }
    outputFilePath := filepath.Join(tmpDir, hash + ".bin")

    // if file already exists, skip
    if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
        return
    }

    phpBinaryPath := getPHPBinaryPath()
    cmd := exec.Command(phpBinaryPath, ".temp/runner/php/dump.php", file)
    cmd.Env = os.Environ()
    cmd.Env = append(cmd.Env, "OUTPUT_FORMAT=binary")
    var out bytes.Buffer
    cmd.Stdout = io.MultiWriter(ioutil.Discard, &out)

    if err := cmd.Run(); err != nil {
        log.Printf("Cannot execute command %s : %v\n", cmd.String(), err)

        // output
        log.Printf("Output : %s\n", out.String())
        return
    }

    jsonBytes := out.Bytes()

    if err := ioutil.WriteFile(outputFilePath, jsonBytes, 0644); err != nil {
        log.Printf("Cannot write file %s : %v\n", outputFilePath, err)
    }

    // Redirige la sortie de la commande vers /dev/null
    cmd.Stdout = ioutil.Discard
    cmd.Stderr = os.Stderr
}

func getPHPBinaryPath() string {
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