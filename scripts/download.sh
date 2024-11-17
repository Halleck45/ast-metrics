#!/bin/sh

# Function to get the version of the latest release
get_latest_release() {
  # Get latest release from GitHub API
  curl --silent "https://api.github.com/repos/Halleck45/ast-metrics/releases/latest" | \
  # Get tag line
  grep '"tag_name":' | \
  # Pluck JSON value
  sed -E 's/.*"([^"]+)".*/\1/'
}

# Function to check OS and architecture
get_os_arch() {
  os=$(uname -s)
  arch=$(uname -m)

  case "$os" in
    Linux)
      case "$arch" in
        i686|i386) echo "Linux_i386" ;;
        x86_64) echo "Linux_x86_64" ;;
        aarch64) echo "Linux_arm64" ;;
        *) echo "Unsupported Linux architecture" ;;
      esac
      ;;
    Darwin)
      case "$arch" in
        x86_64) echo "Darwin_x86_64" ;;
        arm64) echo "Darwin_arm64" ;;
        *) echo "Unsupported macOS architecture" ;;
      esac
      ;;
    MINGW32*|MSYS*)
      # Assuming 32-bit executable for MSYS/MINGW32
      echo "Windows_i386.exe"
      ;;
    MINGW64*|CYGWIN*)
      if [ "$arch" = "x86_64" ]; then
        echo "Windows_x86_64.exe"
      else
        echo "Windows_i386.exe"
      fi
      ;;
    Windows_NT)
      if echo "$PROCESSOR_ARCHITECTURE" | grep -q "64"; then
        echo "Windows_x86_64.exe"
      else
        echo "Windows_i386.exe"
      fi
      ;;
    *)
      echo "Unsupported OS"
      ;;
  esac
}

os_arch=$(get_os_arch)

if echo "$os_arch" | grep -q "Unsupported"; then
  echo "No download available for your system: $os_arch"
  exit 1
fi

version=$(get_latest_release)

download_url=""
destination="ast-metrics"
if echo "$os_arch" | grep -q "Linux\|Darwin"; then
  download_url="https://github.com/Halleck45/ast-metrics/releases/download/$version/ast-metrics_$os_arch"
elif echo "$os_arch" | grep -q "Windows"; then
  download_url="https://github.com/Halleck45/ast-metrics/releases/download/$version/$os_arch"
  destination="ast-metrics.exe"
fi

if [ -n "$download_url" ]; then
  echo "📦 Downloading $download_url"
  curl -L -o "$destination" "$download_url"
else
  echo "Failed to construct the download URL."
  exit 1
fi

# permissions
if echo "$os_arch" | grep -q "Linux\|Darwin"; then
  chmod +x "$destination"
fi

echo
echo "📁 File downloaded: $destination"
echo
echo "You can move the executable to a directory in your PATH to make it easier to run."
echo
if echo "$os_arch" | grep -q "Windows"; then
  echo "    Example: "
  echo "    mv $destination C:\\Windows\\System32"
fi
if echo "$os_arch" | grep -q "Linux\|Darwin"; then
  echo "    Example: "
  echo
  echo "    mv $destination /usr/local/bin"
  echo "    chmod +x /usr/local/bin/$destination"
fi
echo
echo
echo "🌟 If you like AST Metrics, please consider starring the project on GitHub: https://github.com/Halleck45/ast-metrics/"
echo
