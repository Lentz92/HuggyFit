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

### Option 2: Building from Source

1. Clone the repository:
```bash
git clone https://github.com/Lentz92/huggyfit.git
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