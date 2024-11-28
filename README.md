# Installation Guide for "get" CLI

The "get" CLI allows you to quickly open GitHub repositories with your preferred editor. Here's how to install it.

## Installation

### Method 1: Using `curl` (Linux/macOS)

Run the following command in your terminal:

```bash
curl -fsSL https://get.aio-web.xyz/install.sh | bash
```

### Method 2: Using `iwr` (Windows)

On Windows, run this command in PowerShell:

```powershell
iwr -useb https://get.aio-web.xyz/install.ps1 | iex
```

### Method 3: Manual Installation

1. Download the latest release from [here](https://github.com/AIO-Develope/get/releases).
2. Move the binary to a directory in your PATH (e.g., `/usr/local/bin` for Linux/macOS or `C:\Program Files\get` on Windows).

## Usage

Once installed, you can use the CLI to open GitHub repositories.

### Open a Repository Interactively

Run the following command:

```bash
get
```

### Open a specific Repository

Run the following command:

```bash
get get [name]
```

### Uninstall the CLI

To uninstall, run:

```bash
get uninstall
```

This will remove the CLI from your system.

## License

This project is licensed under the MIT License.