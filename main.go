package main

import (
    "fmt"
    "log"
    "os"
    "github.com/urfave/cli/v2"
    "go-php/src/hal/ast-metrics/components/Php"
    "github.com/apoorvam/goterminal"
    ct "github.com/daviddengcn/go-colortext"
)

func main() {

    // get an instance of writer (terminal)
    writer := goterminal.New(os.Stdout)

    app := &cli.App{
        Commands: []*cli.Command{
            {
                Name:    "analyze",
                Aliases: []string{"a"},
                Usage:   "Start analyzing the project",
                Action: func(cCtx *cli.Context) error {

                    // Ensure PHP is installed
                    fmt.Fprintln(writer, "Checking PHP version... ")
                    writer.Print()

                    phpVersion, err := Php.Ensure()
                    writer.Clear()
                    writer.Reset()

                    if err != nil {
                        ct.Foreground(ct.Red, false)
                        fmt.Fprintln(writer, err.Error())
                        writer.Print()
                        ct.ResetColor()
                    }

                    log.Println("PHP version: ", phpVersion)

                    path := cCtx.Args().First()
                    Php.DumpAST(writer, path)

                    return nil
                },
            },
            {
                Name:    "init",
                Aliases: []string{"i"},
                Usage:   "Creates a config file",
                Action: func(cCtx *cli.Context) error {
                    fmt.Println("completed task: ", cCtx.Args().First())
                    return nil
                },
            },
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}