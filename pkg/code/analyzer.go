// pkg/code/analyzer.go
package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/LeanerCloud/f2c/pkg/processor"
)

type Dependency struct {
	name  string
	file  string
	start int
	end   int
}

// CodeAnalyzer extends the base processor for code analysis
type CodeAnalyzer struct {
	*processor.Processor
	functionName string
	dependencies map[string]Dependency
	targetFile   string
	targetStart  int
	targetEnd    int
}

// New creates a new CodeAnalyzer for the given function
func New(functionName string) *CodeAnalyzer {
	return &CodeAnalyzer{
		Processor:    processor.New(),
		functionName: functionName,
		dependencies: make(map[string]Dependency),
	}
}

// Process finds and processes the target function. The paths parameter is ignored.
func (ca *CodeAnalyzer) Process(paths ...string) error {
	if err := ca.findTargetFunction(); err != nil {
		return err
	}

	if ca.targetFile == "" {
		return fmt.Errorf("function %s not found in any .go files", ca.functionName)
	}

	if err := ca.collectDependencies(); err != nil {
		return err
	}

	return nil
}

func (ca *CodeAnalyzer) findTargetFunction() error {
	return filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			fSet := token.NewFileSet()
			node, err := parser.ParseFile(fSet, path, nil, parser.ParseComments)
			if err != nil {
				return nil
			}

			for _, decl := range node.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if ok && funcDecl.Name.String() == ca.functionName {
					ca.targetFile = path
					ca.targetStart = fSet.Position(funcDecl.Pos()).Line
					ca.targetEnd = fSet.Position(funcDecl.End()).Line
					// Add target function to dependencies to be processed
					ca.dependencies[ca.functionName] = Dependency{
						name:  ca.functionName,
						file:  path,
						start: ca.targetStart,
						end:   ca.targetEnd,
					}
					ca.analyzeDependencies(funcDecl, path)
				}
			}
		}
		return nil
	})
}

func (ca *CodeAnalyzer) analyzeDependencies(node ast.Node, currentFile string) {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch x := n.(type) {
		case *ast.Ident:
			if !x.IsExported() && x.Name != "_" {
				if x.Name != "" {
					ca.dependencies[x.Name] = Dependency{name: x.Name, file: currentFile}
				}
			}

		case *ast.SelectorExpr:
			if x.Sel != nil {
				ca.dependencies[x.Sel.Name] = Dependency{name: x.Sel.Name, file: currentFile}
			}

		case *ast.StructType:
			if x.Fields != nil {
				for _, field := range x.Fields.List {
					if field.Type != nil {
						if ident, ok := field.Type.(*ast.Ident); ok {
							ca.dependencies[ident.Name] = Dependency{name: ident.Name, file: currentFile}
						}
					}
				}
			}

		case *ast.InterfaceType:
			if x.Methods != nil {
				for _, method := range x.Methods.List {
					if method.Type != nil {
						if ident, ok := method.Type.(*ast.Ident); ok {
							ca.dependencies[ident.Name] = Dependency{name: ident.Name, file: currentFile}
						}
					}
				}
			}
		}
		return true
	})
}

func (ca *CodeAnalyzer) collectDependencies() error {
	// Track what we've already found to avoid duplicates
	found := make(map[string]bool)

	return filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			fSet := token.NewFileSet()
			node, err := parser.ParseFile(fSet, path, nil, parser.ParseComments)
			if err != nil {
				return nil
			}

			for _, decl := range node.Decls {
				if err := ca.processDeclaration(decl, path, fSet, found); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (ca *CodeAnalyzer) processDeclaration(decl ast.Decl, path string, fSet *token.FileSet, found map[string]bool) error {
	// Extract declaration details
	var name string
	var pos, end token.Pos

	switch d := decl.(type) {
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			name = typeSpec.Name.String()
			pos = typeSpec.Pos()
			end = typeSpec.End()
		}
	case *ast.FuncDecl:
		name = d.Name.String()
		pos = d.Pos()
		end = d.End()
	default:
		return nil
	}

	// Check if this is a dependency we're looking for and haven't processed yet
	_, exists := ca.dependencies[name]
	if !exists || found[name] {
		return nil
	}

	// Mark as found
	found[name] = true

	// Read and process the content
	content, err := ca.ReadFileContent(path)
	if err != nil {
		return err
	}

	startLine := fSet.Position(pos).Line
	endLine := fSet.Position(end).Line

	header := fmt.Sprintf("%s:%d %s", path, startLine, name)
	lines := strings.Split(content, "\n")
	snippet := strings.Join(lines[startLine-1:endLine], "\n")

	ca.AddToOutput(header, snippet)

	// Get file stats and add with enhanced info
	stats, err := ca.GetFileStats(path)
	if err != nil {
		ca.AddProcessedItem(name)
	} else {
		ca.AddProcessedItemWithStats(name, stats)
	}

	return nil
}
