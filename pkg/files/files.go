package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LeanerCloud/f2c/pkg/processor"
)

// FileProcessor extends the base processor for file operations
type FileProcessor struct {
	*processor.Processor
	excludeList []string
}

// New creates a new FileProcessor with the given exclude list
func New(excludeFlag string) *FileProcessor {
	var excludeList []string
	if excludeFlag != "" {
		excludeList = strings.Split(excludeFlag, ",")
		for i, s := range excludeList {
			excludeList[i] = strings.TrimSpace(s)
		}
	}

	return &FileProcessor{
		Processor:   processor.New(),
		excludeList: excludeList,
	}
}

// Process processes the given paths
func (fp *FileProcessor) Process(paths []string) error {
	for _, dirPath := range paths {
		if err := fp.processDirectory(dirPath); err != nil {
			return fmt.Errorf("error processing directory %s: %w", dirPath, err)
		}
	}
	return nil
}

func (fp *FileProcessor) processDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && processor.IsTextFile(path) && !fp.isExcluded(path) {
			content, err := fp.ReadFileContent(path)
			if err != nil {
				return fmt.Errorf("error processing file %s: %w", path, err)
			}
			fp.AddToOutput(path, content)
			fp.AddProcessedItem(path)
		}
		return nil
	})
}

func (fp *FileProcessor) isExcluded(path string) bool {
	for _, exclude := range fp.excludeList {
		if exclude != "" && strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}
