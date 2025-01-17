# HuggyFit

HuggyFit is a command-line tool that helps you determine the GPU memory requirements for running HuggingFace models. It provides estimation of memory usage considering different quantization methods and system overhead.

## Features

- 🧮 GPU memory requirement calculations
- 📊 Support for different quantization types (FP16, INT8, INT4)
- 🔍 Detailed model information from HuggingFace
- 💻 Cross-platform support (Linux, macOS, Windows)
- 🎯 System overhead consideration for real-world compatibility

## Requirements

- Go 1.21 or higher
- Internet connection (for fetching model information)

## Installation

### Option 1: Using Go Install

If you have Go installed on your system, you can install HuggyFit directly using:

```bash
go install github.com/Lentz92/huggyfit/cmd/huggyfit@latest
```

After installation, you'll need to ensure the Go binary directory is in your system's PATH. The steps depend on your shell:

#### Fish Shell (Recommended)
1. Create the Fish config directory and file (if they don't exist):
```bash
mkdir -p ~/.config/fish
```

2. Open the config file in your preferred editor:
```bash
nano ~/.config/fish/config.fish
```

3. Add this line to the file:
```fish
set -gx PATH $PATH ~/go/bin
```

4. Save the file:
   - If using nano: Press Ctrl+X, then Y, then Enter
   - If using vim: Press Esc, type `:wq`, then Enter

5. Apply the changes to your current session:
```bash
source ~/.config/fish/config.fish
```

Now `huggyfit` will be available in your PATH permanently, even after system restarts.

#### Other Shells

For Bash (add to `~/.bashrc` or `~/.bash_profile`):
```bash
export PATH=$PATH:~/go/bin
```

For Zsh (add to `~/.zshrc`):
```zsh
export PATH=$PATH:~/go/bin
```

After adding the path, reload your shell configuration:
- Bash: `source ~/.bashrc` or `source ~/.bash_profile`
- Zsh: `source ~/.zshrc`

#### Windows

1. The default Go binary location is `%USERPROFILE%\go\bin`
2. Add to PATH:
   - Open System Properties > Advanced > Environment Variables
   - Under "User variables", edit PATH
   - Add `%USERPROFILE%\go\bin`
   - Click OK and restart your terminal

### Option 2: Building from Source

1. Clone the repository:
```bash
git clone https://github.com/lentz92/huggyfit.git
cd huggyfit
```

2. Build the binary:
```bash
# Linux/macOS
go build -o huggyfit ./cmd/huggyfit

# Windows (PowerShell)
go build -o huggyfit.exe .\cmd\huggyfit
```

3. (Optional) Move the binary to your system's PATH:

Linux/macOS:
```bash
sudo mv huggyfit /usr/local/bin/
```

Windows:
1. Create a directory for binaries (if it doesn't exist):
```powershell
mkdir C:\Users\<YourUsername>\AppData\Local\Programs\huggyfit
```
2. Move the binary:
```powershell
move huggyfit.exe C:\Users\<YourUsername>\AppData\Local\Programs\huggyfit
```
3. Add to PATH:
   - Open System Properties > Advanced > Environment Variables
   - Under "User variables", edit PATH
   - Add `C:\Users\<YourUsername>\AppData\Local\Programs\huggyfit`

## Usage

### Basic Usage

```bash
huggyfit -model Qwen/Qwen2.5-0.5B
```

### Specify Quantization Type

```bash
huggyfit -model Qwen/Qwen2.5-0.5B -dtype q4
```

### Show Detailed Information

```bash
huggyfit -model Qwen/Qwen2.5-0.5B -verbose
```

### Supported Data Types

- float16 (or f16): 16-bit floating point
- int8 (or q8): 8-bit integer quantization
- int4 (or q4): 4-bit integer quantization

## Examples

1. Check basic model requirements:
```bash
huggyfit -model facebook/opt-350m
```

2. Use 4-bit quantization:
```bash
huggyfit -model facebook/opt-350m -dtype q4
```

3. Get detailed model information:
```bash
huggyfit -model facebook/opt-350m -verbose
```

## Help

For a full list of options:
```bash
huggyfit -help
```

## Credit
This tool is inspired by [Philipp Schmid](https://github.com/philschmid) from Hugging Face who made a similar tool in python.