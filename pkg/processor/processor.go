package processor

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Processor handles the processing of files and functions
type Processor struct {
	output         strings.Builder
	processedItems []string
}

// New creates a new Processor
func New() *Processor {
	return &Processor{
		processedItems: make([]string, 0),
	}
}

// GetOutput returns the processed output and list of processed items
func (p *Processor) GetOutput() (string, []string) {
	return p.output.String(), p.processedItems
}

// AddToOutput adds content to the output with a header
func (p *Processor) AddToOutput(header, content string) {
	p.output.WriteString(fmt.Sprintf("// %s\n", header))
	p.output.WriteString(content)
	p.output.WriteString("\n\n")
}

// AddProcessedItem adds an item to the list of processed items
func (p *Processor) AddProcessedItem(item string) {
	p.processedItems = append(p.processedItems, item)
}

// ReadFileContent reads the content of a file and returns it as a string
func (p *Processor) ReadFileContent(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var content strings.Builder
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					content.WriteString(line)
				}
				break
			}
			return "", fmt.Errorf("error reading file: %w", err)
		}
		content.WriteString(line)
	}
	return content.String(), nil
}

// IsTextFile checks if a file is a text file
func IsTextFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	contentType := http.DetectContentType(buffer[:n])
	return strings.HasPrefix(contentType, "text/")
}
