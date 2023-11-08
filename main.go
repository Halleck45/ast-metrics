package main

import (
    "path/filepath"
    "bufio"
    "os"
    "github.com/urfave/cli/v2"
    "github.com/halleck45/ast-metrics/src/Php"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Command"
    "github.com/halleck45/ast-metrics/src/Engine"
    "github.com/halleck45/ast-metrics/src/Driver"
    log "github.com/sirupsen/logrus"
)

func main() {

    log.SetLevel(log.TraceLevel)
    var driverSelected string

    // Prepare accepted languages
    runnerPhp := Php.PhpRunner{}
    runners := []Engine.Engine{&runnerPhp}

    app := &cli.App{
        Commands: []*cli.Command{
            {
                Name:    "analyze",
                Aliases: []string{"a"},
                Usage:   "Start analyzing the project",
                Flags: []cli.Flag{
                    &cli.BoolFlag{
                        Name:  "verbose",
                        Aliases:  []string{"v"},
                        Usage: "Enable verbose mode",
                    },
                    &cli.StringFlag{
                        Name:        "driver",
                        Value:       "docker",
                        Usage:       "Driver to use (docker or native)",
                        Destination: &driverSelected,
                    },
                },
                Action: func(cCtx *cli.Context) error {

                    // get option --verbose
                    if cCtx.Bool("verbose") {
                        log.SetLevel(log.DebugLevel)
                    }

                    // Stdout
                    outWriter := bufio.NewWriter(os.Stdout)
                    pterm.DefaultBasicText.Println(pterm.LightMagenta(" AST Metrics ") + "is a language-agnostic static code analyzer.")

                    // Valide args
                    if cCtx.Args().Len() == 0 {
                        pterm.Error.Println("Please provide a path to analyze")
                        return nil
                    }

                    // Path
                    path := cCtx.Args().First()
                    // make path absolute
                    if !filepath.IsAbs(path) {
                        var err error
                        path, err = filepath.Abs(path)
                        if err != nil {
                            pterm.Error.Println(err.Error())
                            return err
                        }
                    }

                    // Driver
                    var driver Driver.Driver
                    driver = Driver.Native
                    if driverSelected == "docker" {
                        driver = Driver.Docker
                    }

                    // Run command
                    command := Command.NewAnalyzeCommand(path, driver, outWriter, runners)
                    err := command.Execute()
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }

                    return nil
                },
            },
            {
                Name:    "clean",
                Aliases: []string{"c"},
                Usage:   "Clean workdir",
                Action: func(cCtx *cli.Context) error {
                    // Run command
                    command := Command.NewCleanCommand()
                    err := command.Execute()
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }
                    return nil
                },
            },
        },
    }
    app.Suggest = true

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}