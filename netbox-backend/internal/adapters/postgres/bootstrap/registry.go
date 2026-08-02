// Package bootstrap creates only database tables that do not already exist.
package bootstrap

import (
	"fmt"
	"reflect"
	"strings"
)

// Entry describes one model in database bootstrap order. Dependencies must
// refer to entries that appear earlier in the registry. This makes the
// declared order explicit and prevents accidentally introducing an invalid
// migration order as the registry grows.
type Entry struct {
	Name         string
	Model        any
	Dependencies []string
}

// Registry is an immutable, validated, topologically ordered set of models.
type Registry struct {
	entries []Entry
}

// NewRegistry validates entries and preserves their declared order.
func NewRegistry(entries ...Entry) (Registry, error) {
	seen := make(map[string]struct{}, len(entries))
	validated := make([]Entry, 0, len(entries))

	for index, entry := range entries {
		entry.Name = strings.TrimSpace(entry.Name)
		if entry.Name == "" {
			return Registry{}, fmt.Errorf("bootstrap registry entry %d has an empty name", index)
		}
		if isNil(entry.Model) {
			return Registry{}, fmt.Errorf("bootstrap registry entry %q has a nil model", entry.Name)
		}
		if _, exists := seen[entry.Name]; exists {
			return Registry{}, fmt.Errorf("bootstrap registry contains duplicate entry %q", entry.Name)
		}

		dependencies := make([]string, 0, len(entry.Dependencies))
		entryDependencies := make(map[string]struct{}, len(entry.Dependencies))
		for _, dependency := range entry.Dependencies {
			dependency = strings.TrimSpace(dependency)
			if dependency == "" {
				return Registry{}, fmt.Errorf("bootstrap registry entry %q has an empty dependency", entry.Name)
			}
			if _, duplicate := entryDependencies[dependency]; duplicate {
				return Registry{}, fmt.Errorf(
					"bootstrap registry entry %q repeats dependency %q",
					entry.Name,
					dependency,
				)
			}
			if _, ordered := seen[dependency]; !ordered {
				return Registry{}, fmt.Errorf(
					"bootstrap registry entry %q depends on %q, which must appear earlier",
					entry.Name,
					dependency,
				)
			}
			entryDependencies[dependency] = struct{}{}
			dependencies = append(dependencies, dependency)
		}

		entry.Dependencies = dependencies
		seen[entry.Name] = struct{}{}
		validated = append(validated, entry)
	}

	return Registry{entries: validated}, nil
}

// Entries returns a defensive copy of the ordered entries.
func (registry Registry) Entries() []Entry {
	entries := make([]Entry, len(registry.entries))
	for index, entry := range registry.entries {
		entry.Dependencies = append([]string(nil), entry.Dependencies...)
		entries[index] = entry
	}
	return entries
}

func isNil(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
