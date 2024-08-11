package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
)

var excludeFlag string

func main() {
	flag.StringVar(&excludeFlag, "exclude", "", "Comma-separated list of strings to exclude when appearing in file names (e.g., -exclude=.git,.tmp)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please provide file or directory names as command-line arguments.")
		os.Exit(1)
	}

	excludeList := strings.Split(excludeFlag, ",")
	for i, s := range excludeList {
		excludeList[i] = strings.TrimSpace(s)
	}

	var output strings.Builder
	var copiedFiles []string

	for _, dirPath := range flag.Args() {
		if err := processDirectory(dirPath, &output, &copiedFiles, excludeList); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing directory %s: %v\n", dirPath, err)
			return
		}
	}

	// Copy the accumulated output to the clipboard
	if err := clipboard.WriteAll(output.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Content copied to clipboard.")
	fmt.Println("Files processed:")
	for _, file := range copiedFiles {
		fmt.Println(file)
	}
}

func processDirectory(dirPath string, output *strings.Builder, copiedFiles *[]string, excludeList []string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isTextFile(path) && !isExcluded(path, excludeList) {
			if err := addFileToOutput(path, output); err != nil {
				return fmt.Errorf("error processing file %s: %w", path, err)
			}
			*copiedFiles = append(*copiedFiles, path)
		}
		return nil
	})
}

func isExcluded(path string, excludeList []string) bool {
	for _, exclude := range excludeList {
		if exclude != "" && strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}

func isTextFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read the first 512 bytes
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Use the http.DetectContentType function to detect the content type
	contentType := http.DetectContentType(buffer[:n])
	return strings.HasPrefix(contentType, "text/")
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
