package Php

import (
    "errors"
    "os"
    "path/filepath"
    "log"
    "os/exec"
    "strings"
    "strconv"
    "github.com/apoorvam/goterminal"
    "github.com/yargevad/filepathx"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "bytes"
    "github.com/gosuri/uiprogress"
)

func Ensure() (string, error) {
	// Get PHP binary path. IF env PHP_BINARY_PATH is not set, use default value
    phpBinaryPath := getPHPBinaryPath()

    // Get PHP version
    phpVersion := getPHPVersion(phpBinaryPath)

    // if version is empty, throw error
    if phpVersion == "" {
        return "", errors.New("Cannot get PHP version using the PHP binary path: " + phpBinaryPath + ". Please check if PHP is installed, or set the PHP_BINARY_PATH environment variable to the correct path.")
    }

    return phpVersion, nil
}

func DumpAST(writer *goterminal.Writer, path string) (string, error) {
    //phpBinaryPath := os.Getenv("PHP_BINARY_PATH")


    // list all .php file in path, recursively
    path = strings.TrimRight(path, "/")
    fmt.Fprintln(writer, "Parsing PHP files in " + path + "... ")
    writer.Print()

    matches, err := filepathx.Glob(path + "/**/*.php")
    if err != nil {
        return "", err
    }


    maxParallelCommands := os.Getenv("MAX_PARALLEL_COMMANDS")
    // if maxParallelCommands is empty, set default value
    if maxParallelCommands == "" {
        maxParallelCommands = "10"
    }
    // to int
    maxParallelCommandsInt, err := strconv.Atoi(maxParallelCommands)
    if err != nil {
        return "", err
    }

    // workdir: folder ".ast-metrics" in the current directory
    workDir, err := os.Getwd()
    if err != nil {
        return "", err
    }
    workDir = filepath.Join(workDir, ".ast-metrics")
    // create workdir if not exists
    if _, err := os.Stat(workDir); os.IsNotExist(err) {
        os.Mkdir(workDir, 0755)
    }

    log.Printf("Dossier temporaire : %s\n", workDir)


    uiprogress.Start()            // start rendering
	bar := uiprogress.AddBar(len(matches) - 1) // Add a new bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
	    return fmt.Sprintf("Parsing PHP Code")
    })

	bar.AppendCompleted()
	bar.PrependElapsed()



    sem := make(chan struct{}, maxParallelCommandsInt)
    for _, file := range matches {
        if !strings.Contains(file, "/vendor/") {
            sem <- struct{}{}
            go func(file string) {
                executePHPCommandForFile(workDir, file)
                bar.Incr()
                <-sem
            }(file)
        }
    }

    // Attendez que les commandes se terminent (vous pouvez ajouter une synchronisation ici)
    fmt.Println("Analyzing...")
    for i := 0; i < maxParallelCommandsInt; i++ {
        sem <- struct{}{}
    }

    uiprogress.Stop()

    // cleanup current cli line
    fmt.Print("\033[2K\r")
    

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
        log.Printf("Erreur lors du calcul du hachage du fichier %s : %v\n", file, err)
        return
    }
    outputFilePath := filepath.Join(tmpDir, hash + ".json")

    // if file already exists, skip
    if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
        return
    }

    cmd := exec.Command("php", "runner/php/vendor/nikic/php-parser/bin/php-parse", "--json-dump", file)

    var out bytes.Buffer
    cmd.Stdout = io.MultiWriter(ioutil.Discard, &out)

    if err := cmd.Run(); err != nil {
        log.Printf("Erreur lors de l'exécution de la commande pour %s : %v\n", file, err)
        return
    }

    jsonBytes := out.Bytes()

    if err := ioutil.WriteFile(outputFilePath, jsonBytes, 0644); err != nil {
        log.Printf("Erreur lors de la sauvegarde du fichier %s : %v\n", outputFilePath, err)
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