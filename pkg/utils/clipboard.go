package utils

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the given content to the system clipboard
func CopyToClipboard(content string) error {
	if err := clipboard.WriteAll(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}
