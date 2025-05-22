// cmd/f2c/main.go
package main

import (
	"fmt"
	"os"

	"github.com/LeanerCloud/f2c/pkg/code"
	"github.com/LeanerCloud/f2c/pkg/files"
	"github.com/LeanerCloud/f2c/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	excludeFlag  string
	functionFlag string
)

var rootCmd = &cobra.Command{
	Use:   "f2c [flags] [file/directory...]",
	Short: "File-to-Clipboard Tool",
	Long: `f2c is a tool that copies contents to the clipboard.
It can process either files or Go functions:
- Without -function flag: copies contents of multiple files
- With -function flag: analyzes and copies a Go function and its dependencies`,
	Run: run,
}

func init() {
	rootCmd.Flags().StringVarP(&excludeFlag, "exclude", "e", "", "Comma-separated list of strings to exclude when appearing in file names")
	rootCmd.Flags().StringVarP(&functionFlag, "function", "f", "", "Name of the Go function to analyze")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	if functionFlag == "" {
		// File processing mode
		if len(args) < 1 {
			fmt.Println("Please provide file or directory names as arguments.")
			os.Exit(1)
		}

		processor := files.New(excludeFlag)
		if err := processor.Process(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output, processedItemsTable := processor.GetOutput()
		if err := utils.CopyToClipboard(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Content copied to clipboard.")
		fmt.Println("\nFiles processed:")
		for _, line := range processedItemsTable {
			fmt.Println(line)
		}
	} else {
		// Function processing mode
		analyzer := code.New(functionFlag)
		if err := analyzer.Process(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output, processedItemsTable := analyzer.GetOutput()
		if err := utils.CopyToClipboard(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully copied to clipboard")
		fmt.Println("\nItems processed:")
		for _, line := range processedItemsTable {
			fmt.Println(line)
		}
	}
}
