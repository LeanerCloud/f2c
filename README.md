Sure! Here's a simplified README written in markdown format:

---

# File-to-Clipboard Tool

This tool helps you copy the contents of multiple files to the clipboard, with each file's content prefixed by a comment indicating the file name. It's useful for pasting entire projects into large language models (LLMs) for analysis.

## Features

- Reads contents of multiple files given as command-line arguments
- Prefixes each file's content with a comment showing the file name
- Copies the combined content to the clipboard

## Requirements

- Go 1.15 or higher

## Installation

1. Ensure you have Go installed from [golang.org](https://golang.org/).

2. Install the software
```shell
go install 
```

## Usage

1. Build the Go program:

    ```sh
    go build -o file-to-clipboard main.go
    ```

2. Run the program with the files you want to copy as arguments:

    ```sh
    ./file-to-clipboard file1.txt file2.txt file3.txt
    ```

    Replace `file1.txt`, `file2.txt`, and `file3.txt` with your file names.

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

Running the tool:

```sh
./file-to-clipboard file1.txt file2.txt
```

Results in the clipboard containing:

```txt
// file1.txt
Hello, this is file1.

// file2.txt
Hello, this is file2.
```

## Code

Here's the complete code for the tool:

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/atotto/clipboard"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please provide file names as command-line arguments.")
        os.Exit(1)
    }

    var output strings.Builder
    for _, fileName := range os.Args[1:] {
        if err := addFileToOutput(fileName, &output); err != nil {
            fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", fileName, err)
            return
        }
    }

    if err := clipboard.WriteAll(output.String()); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Content copied to clipboard.")
}

func addFileToOutput(fileName string, output *strings.Builder) error {
    file, err := os.Open(fileName)
    if err != nil {
        return fmt.Errorf("error opening file: %w", err)
    }
    defer file.Close()

    output.WriteString(fmt.Sprintf("// %s\n", fileName))
    reader := bufio.NewReader(file)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            return fmt.Errorf("error reading file: %w", err)
        }
        output.WriteString(line)
    }
    output.WriteString("\n")
    return nil
}
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

This simplified README provides clear instructions on what the tool does, how to install and use it, an example, the complete code, and mentions the license.