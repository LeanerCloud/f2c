// pkg/processor/processor.go
package processor

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

// GitIgnorePattern represents a gitignore pattern
type GitIgnorePattern struct {
	pattern    string
	regex      *regexp.Regexp
	isNegation bool
	isDir      bool
}

// Processor handles the processing of files and functions
type Processor struct {
	output            strings.Builder
	processedItems    []ProcessedItem
	gitignorePatterns []GitIgnorePattern
}

// New creates a new Processor
func New() *Processor {
	p := &Processor{
		processedItems:    make([]ProcessedItem, 0),
		gitignorePatterns: make([]GitIgnorePattern, 0),
	}

	// Load built-in ignore patterns first
	p.loadBuiltinIgnorePatterns()

	// Load .gitignore patterns if available (these can override built-ins with negation)
	p.loadGitignorePatterns()

	return p
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

// formatProcessedItems formats the processed items as a table with totals
func (p *Processor) formatProcessedItems() []string {
	if len(p.processedItems) == 0 {
		return []string{}
	}

	// Calculate totals
	var totalSizeKB float64
	var totalLines int
	var validStatsCount int

	for _, item := range p.processedItems {
		if item.SizeKB > 0 || item.Lines > 0 {
			totalSizeKB += item.SizeKB
			totalLines += item.Lines
			validStatsCount++
		}
	}

	// Find the maximum width for each column
	maxNameWidth := len("File/Item")
	maxSizeWidth := len("Size (KB)")
	maxLinesWidth := len("Lines")

	// Check against "TOTAL" row as well
	if len("TOTAL") > maxNameWidth {
		maxNameWidth = len("TOTAL")
	}

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

	// Check total values for width
	totalSizeStr := fmt.Sprintf("%.1f", totalSizeKB)
	if len(totalSizeStr) > maxSizeWidth {
		maxSizeWidth = len(totalSizeStr)
	}
	totalLinesStr := fmt.Sprintf("%d", totalLines)
	if len(totalLinesStr) > maxLinesWidth {
		maxLinesWidth = len(totalLinesStr)
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

	// Add separator before totals
	result = append(result, separator)

	// Add totals row
	totalRow := fmt.Sprintf("%-*s  %*s  %*s", maxNameWidth, "TOTAL", maxSizeWidth, totalSizeStr, maxLinesWidth, totalLinesStr)
	result = append(result, totalRow)

	// Add summary information
	result = append(result, "")
	result = append(result, fmt.Sprintf("Summary: %d files processed, %.1f KB total, %d lines total",
		len(p.processedItems), totalSizeKB, totalLines))

	if validStatsCount < len(p.processedItems) {
		result = append(result, fmt.Sprintf("Note: %d items had no size/line statistics available",
			len(p.processedItems)-validStatsCount))
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

// IsGitIgnored checks if a file path matches any gitignore pattern
func (p *Processor) IsGitIgnored(filePath string) bool {
	// Convert to relative path for consistent matching
	relPath := strings.TrimPrefix(filePath, "./")

	ignored := false

	for _, pattern := range p.gitignorePatterns {
		// Check if pattern matches the full path or just the filename
		matches := pattern.regex.MatchString(relPath) || pattern.regex.MatchString(filepath.Base(relPath))

		// For directory patterns, also check if any parent directory matches
		if !matches && pattern.isDir {
			pathParts := strings.Split(relPath, "/")
			for i := range pathParts {
				partialPath := strings.Join(pathParts[:i+1], "/")
				if pattern.regex.MatchString(partialPath) {
					matches = true
					break
				}
			}
		}

		if matches {
			if pattern.isNegation {
				ignored = false // Negation patterns override previous ignores
			} else {
				ignored = true
			}
		}
	}

	return ignored
}

// loadBuiltinIgnorePatterns loads common patterns that should typically be ignored
func (p *Processor) loadBuiltinIgnorePatterns() {
	builtinPatterns := []string{
		// Git directory
		".git/",

		// Package managers and lock files
		"package.json",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"composer.lock",
		"Gemfile.lock",
		"Pipfile.lock",
		"poetry.lock",

		// Go specific
		"go.sum",
		"go.work.sum",
		"vendor/",

		// Node.js
		"node_modules/",

		// Python
		"__pycache__/",
		"*.pyc",
		"*.pyo",
		"*.pyd",
		".Python",
		"pip-log.txt",
		"pip-delete-this-directory.txt",
		".venv/",
		"venv/",
		"ENV/",
		"env/",

		// Java/Maven/Gradle
		"target/",
		".gradle/",
		"build/",
		"*.class",
		"*.jar",
		"*.war",

		// .NET
		"bin/",
		"obj/",
		"*.dll",
		"*.exe",
		"*.pdb",

		// Build and dist directories
		"dist/",
		"build/",
		"out/",
		".next/",
		".nuxt/",

		// IDE and editor files
		".vscode/",
		".idea/",
		"*.swp",
		"*.swo",
		"*~",
		".DS_Store",
		"Thumbs.db",

		// Logs
		"*.log",
		"logs/",
		"npm-debug.log*",
		"yarn-debug.log*",
		"yarn-error.log*",

		// Environment and config
		".env",
		".env.local",
		".env.*.local",

		// Coverage and test results
		"coverage/",
		".nyc_output/",
		".coverage",
		"htmlcov/",
		".pytest_cache/",
		"test-results/",

		// Temporary files
		"*.tmp",
		"*.temp",
		".cache/",
		".temp/",
		"tmp/",
	}

	for _, pattern := range builtinPatterns {
		gitPattern := p.parseGitignorePattern(pattern)
		if gitPattern.regex != nil {
			p.gitignorePatterns = append(p.gitignorePatterns, gitPattern)
		}
	}
}

// loadGitignorePatterns loads patterns from .gitignore file if it exists
func (p *Processor) loadGitignorePatterns() {
	gitignoreFile := ".gitignore"
	if _, err := os.Stat(gitignoreFile); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(gitignoreFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := p.parseGitignorePattern(line)
		if pattern.regex != nil {
			p.gitignorePatterns = append(p.gitignorePatterns, pattern)
		}
	}
}

// parseGitignorePattern converts a gitignore pattern to a compiled regex
func (p *Processor) parseGitignorePattern(pattern string) GitIgnorePattern {
	original := pattern
	isNegation := false
	isDir := false

	// Handle negation patterns (starting with !)
	if strings.HasPrefix(pattern, "!") {
		isNegation = true
		pattern = pattern[1:]
	}

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		isDir = true
		pattern = pattern[:len(pattern)-1]
	}

	// Convert gitignore pattern to regex
	regexPattern := p.gitignoreToRegex(pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return GitIgnorePattern{}
	}

	return GitIgnorePattern{
		pattern:    original,
		regex:      regex,
		isNegation: isNegation,
		isDir:      isDir,
	}
}

// gitignoreToRegex converts a gitignore pattern to a regex pattern
func (p *Processor) gitignoreToRegex(pattern string) string {
	// Escape special regex characters except for gitignore wildcards
	pattern = regexp.QuoteMeta(pattern)

	// Convert gitignore wildcards to regex equivalents
	pattern = strings.ReplaceAll(pattern, `\*\*`, `.*`)  // ** matches any number of directories
	pattern = strings.ReplaceAll(pattern, `\*`, `[^/]*`) // * matches anything except /
	pattern = strings.ReplaceAll(pattern, `\?`, `.`)     // ? matches any single character

	// Handle leading slash (absolute path from repo root)
	if strings.HasPrefix(pattern, "/") {
		pattern = "^" + pattern[1:] + "$"
	} else {
		// Pattern can match anywhere in the path
		pattern = "(^|/)" + pattern + "($|/)"
	}

	return pattern
}
