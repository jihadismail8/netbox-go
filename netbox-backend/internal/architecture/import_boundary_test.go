package architecture_test

import (
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestDomainAndApplicationImportsPointInward makes the architecture rule
// executable: business code cannot quietly acquire transport, generated API,
// GORM, global database/configuration, adapter, or platform dependencies.
func TestDomainAndApplicationImportsPointInward(t *testing.T) {
	t.Parallel()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test location")
	}
	internalRoot := filepath.Dir(filepath.Dir(filename))

	for _, layer := range []string{"domain", "application"} {
		layer := layer
		t.Run(layer, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(internalRoot, layer)
			err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				checkImports(t, layer, path)
				return nil
			})
			if err != nil {
				t.Fatalf("scan %s imports: %v", layer, err)
			}
		})
	}
}

// TestGenericWorkflowPackagesAreRetired makes the completed typed cutover
// irreversible. Neither production code nor tests may recreate or import the
// retired map-shaped domain, application, or PostgreSQL workflow packages.
func TestGenericWorkflowPackagesAreRetired(t *testing.T) {
	t.Parallel()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test location")
	}
	internalRoot := filepath.Dir(filepath.Dir(filename))
	backendRoot := filepath.Dir(internalRoot)
	retired := map[string]struct{}{
		"netbox-go/internal/domain/workflow":            {},
		"netbox-go/internal/application/workflow":       {},
		"netbox-go/internal/adapters/postgres/workflow": {},
	}

	for importPath := range retired {
		relativePath := strings.TrimPrefix(importPath, "netbox-go/")
		if _, err := fs.Stat(os.DirFS(backendRoot), filepath.FromSlash(relativePath)); !errors.Is(err, fs.ErrNotExist) {
			if err == nil {
				t.Errorf("retired package directory still exists: %s", relativePath)
			} else {
				t.Errorf("inspect retired package directory %s: %v", relativePath, err)
			}
		}
	}

	for _, relativeRoot := range []string{"internal", "test"} {
		root := filepath.Join(backendRoot, relativeRoot)
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			parsed, parseErr := parser.ParseFile(
				token.NewFileSet(), path, nil, parser.ImportsOnly,
			)
			if parseErr != nil {
				t.Errorf("parse imports from %s: %v", path, parseErr)
				return nil
			}
			for _, imported := range parsed.Imports {
				importPath, unquoteErr := strconv.Unquote(imported.Path.Value)
				if unquoteErr != nil {
					t.Errorf("unquote import %s in %s: %v", imported.Path.Value, path, unquoteErr)
					continue
				}
				if _, forbidden := retired[importPath]; forbidden {
					t.Errorf("%s imports retired workflow dependency %q", path, importPath)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s imports: %v", relativeRoot, err)
		}
	}
}

func checkImports(t *testing.T, layer, path string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Errorf("parse imports from %s: %v", path, err)
		return
	}

	for _, imported := range parsed.Imports {
		importPath, unquoteErr := strconv.Unquote(imported.Path.Value)
		if unquoteErr != nil {
			t.Errorf("unquote import %s in %s: %v", imported.Path.Value, path, unquoteErr)
			continue
		}
		if forbiddenImport(importPath) {
			t.Errorf("%s imports forbidden outward dependency %q", path, importPath)
		}
		if layer == "domain" && strings.HasPrefix(importPath, "netbox-go/internal/application/") {
			t.Errorf("domain file %s imports application dependency %q", path, importPath)
		}
	}
}

func forbiddenImport(importPath string) bool {
	for _, prefix := range []string{
		"github.com/gin-gonic/gin",
		"google.golang.org/protobuf",
		"gorm.io/",
		"netbox-go/api/",
		"netbox-go/gen/go/",
		"netbox-go/internal/adapters/",
		"netbox-go/internal/cache",
		"netbox-go/internal/config",
		"netbox-go/internal/dao",
		"netbox-go/internal/database",
		"netbox-go/internal/handler",
		"netbox-go/internal/model",
		"netbox-go/internal/platform/",
		"netbox-go/internal/routers",
		"netbox-go/internal/server",
		"netbox-go/internal/service",
	} {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}
