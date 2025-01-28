# File-to-Clipboard Tool (f2c)

A versatile tool that helps you copy file contents or Go code snippets to your clipboard. It's particularly useful when working with large language models (LLMs) for code analysis or sharing multiple files.

## Features

### File Processing Mode

- Reads contents of multiple files given as command-line arguments
- Walks directories recursively to gather all files
- Prefixes each file's content with a comment showing the file name
- Ignores non-text files
- Can exclude files matching specified patterns

### Go Code Analysis Mode

- Finds and extracts specified Go functions from your codebase
- Analyzes and includes function dependencies
- Captures related type definitions and helper functions
- Preserves file and line number information
- Formats output for easy copying into LLMs

## Requirements

- Go 1.15 or higher

## Installation

1. Ensure you have Go installed from [golang.org](https://golang.org/)
2. Install the tool:

```shell
go install github.com/LeanerCloud/f2c/cmd/f2c@latest
```

## Usage

### File Processing Mode Usage

Copy contents of specific files or directories:

```shell
# Copy all text files in current directory
f2c .

# Copy specific files
f2c file1.txt file2.go

# Exclude certain paths
f2c --exclude .git,vendor,node_modules .

# Short form for exclude
f2c -e .git,vendor .
```

Example output:

```text
// file1.txt
Hello, this is file1.

// file2.txt
Hello, this is file2.
```

### Go Code Analysis Mode Usage

Extract a function and its dependencies from the current directory:

```shell
# Extract a specific function
f2c --function ProcessDirectory

# Short form
f2c -f ProcessDirectory
```

Example output:

```go
// main.go:40 ProcessDirectory
func ProcessDirectory(path string) error {
    // ... function code ...
}

// utils.go:15 helperFunction
func helperFunction() {
    // ... dependency code ...
}

// types.go:25 Config
type Config struct {
    // ... related type definition ...
}
```

## Command Line Options

```shell
f2c [flags] [file/directory...]

Flags:
  -e, --exclude string    Comma-separated list of strings to exclude when appearing in file names
  -f, --function string   Name of the Go function to analyze
  -h, --help             Help for f2c
```

## How It Works

### File Processing

1. Recursively scans provided paths
2. Identifies text files using content type detection
3. Skips excluded paths
4. Reads and formats file contents
5. Copies combined output to clipboard

### Go Code Analysis

1. Locates target function in Go source files
2. Analyzes function dependencies (types, helper functions)
3. Collects all related code snippets
4. Preserves source location information
5. Copies formatted output to clipboard

## License

This project is licensed under the MIT License.
