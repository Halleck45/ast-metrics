# Installing AST Metrics

## Downloading binaries

AST Metrics is built in Golang, and distributed as binary. 

You don't need anything, simply download the correct binary for your platform.


=== ":magic_wand: Automatically"


    Run the following command to download the latest version of AST Metrics:

    ```bash
    curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|bash
    ```

    > Be careful when running scripts from the internet. Always check the content of the script before running it.


=== ":simple-linux: Linux"


    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):

    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_x86_64) (most common)
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_arm64) (for Raspberry Pi)
    + [i386](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_i386) (for old 32-bit systems)

=== ":simple-apple: MacOS"

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):
    
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_arm64) (for Apple Silicon)
    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_x86_64) (for Intel)


=== ":simple-windows10: Windows"

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):

    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_x86_64.exe) (most common)
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_arm64.exe) (for ARM)
    + [i386](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_i386.exe) (for old 32-bit systems)


=== ":simple-docker: inside containers (Docker)"

    If you don't know what is your image architecture, the simplest way consists in running the following command in your container:

    ```bash
    # run this command in your container. 
    # For example, execute `docker exec -it my-container bash` to open a shell in your container

    echo "OS: $(uname -s), Arch: $(uname -m)"
    ```

    It will show you the OS and architecture of the container. 

    For example, if the output is `OS: Linux, Arch: x86_64`, you should download the `amd64` binary for Linux.


=== ":simple-go: With Go"

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
