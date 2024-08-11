# File-to-Clipboard Tool (f2c)

This tool helps you copy the contents of multiple files to the clipboard, with each file's content prefixed by a comment indicating the file name.

It's useful for pasting entire projects into large language models (LLMs) for analysis.

## Features

- Reads contents of multiple files given as command-line arguments.
- Prefixes each file's content with a comment showing the file name.
- Copies the combined content to the clipboard.
- walks directories recursively and gathers all their files.
- Ignores non-text files.
- Can exclude files matching a list of paths, given as comma-separated list.

## Requirements

- Go 1.15 or higher

## Installation

1. Ensure you have Go installed from [golang.org](https://golang.org/).

2. Install the software

```shell
go install github.com:LeanerCloud/f2c@latest
```

## Usage

Assuming the GOPATH/bin is in your PATH, you can run the program with the files you want to copy as arguments:

```shell
f2c *.txt
```

or

```shell
f2c -exclude .git .
```

3. The combined content will be copied to your clipboard, ready to paste.

## Example

Given these files:

**file1.txt**:

```txt
Hello, this is file1.
```

**file2.txt**:

```txt
Hello, this is file2.
```

Results in the clipboard containing:

```txt
// file1.txt
Hello, this is file1.

// file2.txt
Hello, this is file2.
```

## License

This project is licensed under the MIT License.
