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

	// Copy the accumulated output to the clipboard
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
