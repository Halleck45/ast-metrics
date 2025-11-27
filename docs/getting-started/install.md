# Installing AST Metrics

AST Metrics is built in Golang and distributed as a single binary. It has no dependencies.

## 🚀 Quick Install

Choose your preferred method below.

??? info ":magic_wand: Automatic Install (Linux/MacOS/Windows)"

    Run the following command to download the latest version:

    ```bash
    curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|sh
    ```

    Then move the `./ast-metrics` binary to a directory in your `PATH` (e.g. `/usr/local/bin` for Linux/MacOS).

    > Be careful when running scripts from the internet. Always check the content of the script before running it.

??? info ":simple-linux: Linux (Manual)"

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):

    - [amd64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_x86_64) (most common)
    - [arm64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_arm64) (for Raspberry Pi)
    - [i386](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_i386) (for old 32-bit systems)

??? info ":simple-apple: MacOS (Manual)"

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):
    
    - [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_arm64) (for Apple Silicon / M1 / M2)
    - [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_x86_64) (for Intel Macs)

??? info ":fontawesome-brands-windows: Windows (Manual)"

    Download the executable for your platform:

    - [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_x86_64.exe) (most common)
    - [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_arm64.exe) (for ARM)
    - [i386](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_i386.exe) (for old 32-bit systems)



??? info ":elephant: PHP Project (Composer)"

    If you are working on a PHP project, you can install AST Metrics as a dev dependency via Composer.
    This is the recommended way for PHP developers as it manages the binary version for you.

    ```bash
    composer require --dev halleck45/ast-metrics
    ```

    Then you can run it using:

    ```bash
    php vendor/bin/ast-metrics analyze .
    ```

??? info ":simple-go: Go Install"

    If you have Go installed:

    ```bash
    go install github.com/halleck45/ast-metrics@latest
    ```

## Verify Installation

Verify that the installation worked by opening a new terminal session and listing AST Metrics's available subcommands.

```bash
ast-metrics --help
```

You should see the help message with the available subcommands.

## Troubleshooting

If you get an error that the command `ast-metrics` is not found, you may need to add the directory where the binary is located to your PATH.

## Updating

Update is really easy. Just run:

```bash
ast-metrics self-update
```