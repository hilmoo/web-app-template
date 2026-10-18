package archtest_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const featuresImportPrefix = "web-app-template-core/internal/features/"

// TestFeaturesDoNotImportEachOther fails when a package under
// internal/features/<a> imports internal/features/<b>. Features talk through
// consumer-defined interfaces wired in internal/app; see docs/architecture.md.
func TestFeaturesDoNotImportEachOther(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "features")

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read features dir: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		violations, walkErr := foreignFeatureImports(filepath.Join(root, entry.Name()), entry.Name())
		if walkErr != nil {
			t.Fatalf("scan feature %q: %v", entry.Name(), walkErr)
		}

		for _, v := range violations {
			t.Error(v)
		}
	}
}

func foreignFeatureImports(dir, feature string) ([]string, error) {
	var violations []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		imports, err := fileImports(path)
		if err != nil {
			return err
		}

		for _, imp := range imports {
			rest, ok := strings.CutPrefix(imp, featuresImportPrefix)
			if !ok {
				continue
			}

			if other, _, _ := strings.Cut(rest, "/"); other != feature {
				msg := fmt.Sprintf("%s: feature %q imports feature %q (%s)", path, feature, other, imp)
				violations = append(violations, msg)
			}
		}

		return nil
	})

	return violations, err
}

func fileImports(path string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	imports := make([]string, 0, len(f.Imports))

	for _, spec := range f.Imports {
		imp, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr != nil {
			return nil, fmt.Errorf("unquote import in %s: %w", path, unquoteErr)
		}

		imports = append(imports, imp)
	}

	return imports, nil
}
