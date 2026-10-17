package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestBusinessLayersPointInward(testingTarget *testing.T) {
	walkError := filepath.WalkDir("../internal", func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		normalized := filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		isDomain := strings.Contains(normalized, "/domain/")
		isApplication := strings.Contains(normalized, "/application/")
		if !isDomain && !isApplication {
			return nil
		}
		parsed, parseError := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseError != nil {
			return parseError
		}
		for _, imported := range parsed.Imports {
			dependency, _ := strconv.Unquote(imported.Path.Value)
			forbidden := strings.Contains(dependency, ".") || strings.HasPrefix(dependency, "net/http") || dependency == "database/sql" || dependency == "os"
			if strings.HasPrefix(dependency, "qmanga/") {
				forbidden = !strings.Contains(dependency, "/domain") && !(isApplication && strings.Contains(dependency, "/application"))
			}
			if forbidden {
				testingTarget.Errorf("%s imports outward dependency %s", path, dependency)
			}
		}
		return nil
	})
	if walkError != nil {
		testingTarget.Fatal(walkError)
	}
}
