package bootstrap

import (
	"strings"
	"testing"
)

type registryModel struct {
	ID uint64
}

func TestNewRegistryAcceptsTopologicalOrder(t *testing.T) {
	registry, err := NewRegistry(
		Entry{Name: "parent", Model: &registryModel{}},
		Entry{Name: "child", Model: &struct{ ID uint64 }{}, Dependencies: []string{"parent"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	entries := registry.Entries()
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	entries[1].Dependencies[0] = "changed"
	if registry.Entries()[1].Dependencies[0] != "parent" {
		t.Fatal("Entries returned mutable registry state")
	}
}

func TestNewRegistryRejectsInvalidEntries(t *testing.T) {
	var typedNil *registryModel
	tests := []struct {
		name    string
		entries []Entry
		want    string
	}{
		{name: "empty name", entries: []Entry{{Model: &registryModel{}}}, want: "empty name"},
		{name: "nil model", entries: []Entry{{Name: "nil", Model: nil}}, want: "nil model"},
		{name: "typed nil model", entries: []Entry{{Name: "nil", Model: typedNil}}, want: "nil model"},
		{
			name: "duplicate name",
			entries: []Entry{
				{Name: "same", Model: &registryModel{}},
				{Name: "same", Model: &struct{ ID uint64 }{}},
			},
			want: "duplicate entry",
		},
		{
			name: "dependency declared later",
			entries: []Entry{
				{Name: "child", Model: &registryModel{}, Dependencies: []string{"parent"}},
				{Name: "parent", Model: &struct{ ID uint64 }{}},
			},
			want: "must appear earlier",
		},
		{
			name: "duplicate dependency",
			entries: []Entry{
				{Name: "parent", Model: &registryModel{}},
				{
					Name:         "child",
					Model:        &struct{ ID uint64 }{},
					Dependencies: []string{"parent", "parent"},
				},
			},
			want: "repeats dependency",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRegistry(test.entries...)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewRegistry error = %v, want error containing %q", err, test.want)
			}
		})
	}
}
