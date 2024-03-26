# Installing AST Metrics

## Downloading binaries

AST Metrics is built in Golang, and distributed as binary. 

You don't need anything, simply download the correct binary for your platform.

=== ":simple-linux: Linux"

    **Automatically:**

    Run the following command to download and install the latest version of AST Metrics:

    ```bash
    curl -L https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Linux_$(uname -m) -o ~/.local/bin/ast-metrics
    chmod +x ~/.local/bin/ast-metrics
    ```

    **Manually:**

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):

    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_x86_64) (most common)
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_arm64) (for Raspberry Pi)
    + [i386](https://github.com/Halleck45/ast-metrics/releases/download/v0.0.11-alpha/ast-metrics_Linux_i386) (for old 32-bit systems)

=== ":simple-apple: MacOS"

    **Automatically:**

    Run the following command to download and install the latest version of AST Metrics:
    
    ```bash
    curl -L https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_$(uname -m) -o ~/.local/bin/ast-metrics
    chmod +x ~/.local/bin/ast-metrics
    ```

    **Manually:**

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):
    
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_arm64) (for Apple Silicon)
    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Darwin_x86_64) (for Intel)


=== ":simple-windows10: Windows"

    **Manually:**

    Download the binary for your platform (run `uname -m` in your terminal to get your architecture):

    + [amd64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_x86_64.exe) (most common)
    + [arm64](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_arm64.exe) (for ARM)
    + [i386](https://github.com/Halleck45/ast-metrics/releases/download/--latest_version--/ast-metrics_Windows_i386.exe) (for old 32-bit systems)


## Verify Installation

Verify that the installation worked by opening a new terminal session and listing AST Metrics's available subcommands.

```bash
ast-metrics --help
```

You should see the help message with the available subcommands.

## Troubleshooting

If you get an error that the command `ast-metrics` is not found, you may need to add the directory where the binary is located to your PATH.