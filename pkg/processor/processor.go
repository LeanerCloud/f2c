// pkg/processor/processor.go
package processor

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// FileStats holds statistics about a file
type FileStats struct {
	SizeKB float64
	Lines  int
}

// ProcessedItem holds information about a processed item
type ProcessedItem struct {
	Name   string
	SizeKB float64
	Lines  int
}

// Processor handles the processing of files and functions
type Processor struct {
	output         strings.Builder
	processedItems []ProcessedItem
}

// New creates a new Processor
func New() *Processor {
	return &Processor{
		processedItems: make([]ProcessedItem, 0),
	}
}

// GetOutput returns the processed output and list of processed items
func (p *Processor) GetOutput() (string, []string) {
	return p.output.String(), p.formatProcessedItems()
}

// AddToOutput adds content to the output with a header
func (p *Processor) AddToOutput(header, content string) {
	p.output.WriteString(fmt.Sprintf("// %s\n", header))
	p.output.WriteString(content)
	p.output.WriteString("\n\n")
}

// AddProcessedItem adds an item to the list of processed items
func (p *Processor) AddProcessedItem(item string) {
	p.processedItems = append(p.processedItems, ProcessedItem{
		Name:   item,
		SizeKB: 0,
		Lines:  0,
	})
}

// AddProcessedItemWithStats adds an item with file statistics to the list of processed items
func (p *Processor) AddProcessedItemWithStats(item string, stats FileStats) {
	p.processedItems = append(p.processedItems, ProcessedItem{
		Name:   item,
		SizeKB: stats.SizeKB,
		Lines:  stats.Lines,
	})
}

// formatProcessedItems formats the processed items as a table
func (p *Processor) formatProcessedItems() []string {
	if len(p.processedItems) == 0 {
		return []string{}
	}

	// Find the maximum width for each column
	maxNameWidth := len("File/Item")
	maxSizeWidth := len("Size (KB)")
	maxLinesWidth := len("Lines")

	for _, item := range p.processedItems {
		if len(item.Name) > maxNameWidth {
			maxNameWidth = len(item.Name)
		}
		sizeStr := fmt.Sprintf("%.1f", item.SizeKB)
		if item.SizeKB == 0 {
			sizeStr = "-"
		}
		if len(sizeStr) > maxSizeWidth {
			maxSizeWidth = len(sizeStr)
		}
		linesStr := fmt.Sprintf("%d", item.Lines)
		if item.Lines == 0 {
			linesStr = "-"
		}
		if len(linesStr) > maxLinesWidth {
			maxLinesWidth = len(linesStr)
		}
	}

	// Create the formatted output
	var result []string

	// Header
	header := fmt.Sprintf("%-*s  %*s  %*s", maxNameWidth, "File/Item", maxSizeWidth, "Size (KB)", maxLinesWidth, "Lines")
	result = append(result, header)

	// Separator line
	separator := strings.Repeat("-", maxNameWidth) + "  " + strings.Repeat("-", maxSizeWidth) + "  " + strings.Repeat("-", maxLinesWidth)
	result = append(result, separator)

	// Data rows
	for _, item := range p.processedItems {
		sizeStr := "-"
		linesStr := "-"

		if item.SizeKB > 0 {
			sizeStr = fmt.Sprintf("%.1f", item.SizeKB)
		}
		if item.Lines > 0 {
			linesStr = fmt.Sprintf("%d", item.Lines)
		}

		row := fmt.Sprintf("%-*s  %*s  %*s", maxNameWidth, item.Name, maxSizeWidth, sizeStr, maxLinesWidth, linesStr)
		result = append(result, row)
	}

	return result
}

// GetFileStats returns file statistics (size in KB and line count)
func (p *Processor) GetFileStats(fileName string) (FileStats, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return FileStats{}, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return FileStats{}, fmt.Errorf("error getting file info: %w", err)
	}
	sizeKB := float64(info.Size()) / 1024.0

	// Count lines
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return FileStats{}, fmt.Errorf("error reading file: %w", err)
	}

	return FileStats{
		SizeKB: sizeKB,
		Lines:  lines,
	}, nil
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
