// Package routers provides tests for the NetBox-compatible (DRF-style) API endpoints.
package routers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"netbox-go/internal/database"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestBuildSearchFilter verifies search filter construction.
// The operator is dialect-aware: ILIKE on PostgreSQL, LIKE on SQLite.
func TestBuildSearchFilter(t *testing.T) {
	// Determine the expected operator based on the active dialect
	op := "ILIKE"
	if db := database.GetDB(); db != nil && db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		op = "LIKE"
	}

	tests := []struct {
		name      string
		cols      []string
		q         string
		wantEmpty bool
		wantSub   string
	}{
		{
			name:      "empty columns",
			cols:      []string{},
			q:         "test",
			wantEmpty: true,
		},
		{
			name:      "empty query",
			cols:      []string{"name"},
			q:         "",
			wantEmpty: true,
		},
		{
			name:    "single column",
			cols:    []string{"name"},
			q:       "test",
			wantSub: fmt.Sprintf("name %s '%%test%%'", op),
		},
		{
			name:    "multiple columns joined by OR",
			cols:    []string{"name", "slug", "description"},
			q:       "test",
			wantSub: fmt.Sprintf("name %s '%%test%%' OR slug %s '%%test%%' OR description %s '%%test%%'", op, op, op),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSearchFilter(tt.cols, tt.q)
			if tt.wantEmpty && result != "" {
				t.Errorf("buildSearchFilter() = %q, want empty", result)
			}
			if !tt.wantEmpty && result != tt.wantSub {
				t.Errorf("buildSearchFilter() = %q, want %q", result, tt.wantSub)
			}
		})
	}
}

// TestEscapeSQL verifies SQL injection prevention in search queries.
func TestEscapeSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello", "hello"},
		{"single quote", "O'Brien", "O''Brien"},
		{"semicolon", "test; DROP", "test DROP"},
		{"sql comment", "test--evil", "testevil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeSQL(tt.input)
			if result != tt.want {
				t.Errorf("escapeSQL(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

// TestParsePagination verifies pagination parameter parsing.
func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		limitStr   string
		offsetStr  string
		wantLimit  int
		wantOffset int
	}{
		{"default", "50", "0", 50, 0},
		{"custom", "10", "20", 10, 20},
		{"invalid limit", "abc", "0", 50, 0},
		{"invalid offset", "10", "abc", 10, 0},
		{"zero limit", "0", "0", 50, 0},
		{"negative offset", "10", "-5", 10, 0},
		{"too large limit", "5000", "0", 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?limit="+tt.limitStr+"&offset="+tt.offsetStr, nil)

			limit, offset := parsePagination(c)
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}

// TestParseOrdering verifies sort/ordering parameter parsing.
func TestParseOrdering(t *testing.T) {
	cfg := ModelEndpointConfig{
		OrderFields: map[string]bool{
			"name": true, "id": true, "created": true,
		},
		DefaultSort: "name ASC",
	}

	tests := []struct {
		name     string
		ordering string
		want     string
	}{
		{"default", "", "name ASC"},
		{"ascending", "name", "name ASC"},
		{"descending", "-name", "name DESC"},
		{"non-whitelisted field", "evil_field", "name ASC"},
		{"descending non-whitelisted", "-evil_field", "name ASC"},
		{"id ascending", "id", "id ASC"},
		{"created descending", "-created", "created DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			url := "/"
			if tt.ordering != "" {
				url += "?ordering=" + tt.ordering
			}
			c.Request = httptest.NewRequest(http.MethodGet, url, nil)

			result := parseOrdering(c, cfg)
			if result != tt.want {
				t.Errorf("parseOrdering() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestParseIDList verifies ID list parsing.
func TestParseIDList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single id", "1", []string{"1"}},
		{"comma-separated", "1,2,3", []string{"1", "2", "3"}},
		{"with spaces", "1, 2, 3", []string{"1", " 2", " 3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIDList(tt.input)
			if len(result) != len(tt.want) {
				t.Errorf("parseIDList(%q) = %v, want %v", tt.input, result, tt.want)
				return
			}
			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("parseIDList(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.want[i])
				}
			}
		})
	}
}

// TestNormalizeRequestBody verifies nested FK object flattening.
func TestNormalizeRequestBody(t *testing.T) {
	cfg := ModelEndpointConfig{
		NestedFields: map[string]NestedRef{
			"region": {
				Table:      "dcim_region",
				IDCol:      "region_id",
				DisplayCol: "name",
				SlugCol:    "slug",
			},
		},
	}

	tests := []struct {
		name string
		body map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "nested object with id",
			body: map[string]interface{}{
				"name":   "Test Site",
				"slug":   "test-site",
				"region": map[string]interface{}{"id": float64(5)},
			},
			want: map[string]interface{}{
				"name":      "Test Site",
				"slug":      "test-site",
				"region_id": float64(5),
			},
		},
		{
			name: "nested object with numeric id",
			body: map[string]interface{}{
				"name":   "Test Site",
				"region": float64(10),
			},
			want: map[string]interface{}{
				"name":      "Test Site",
				"region_id": int64(10),
			},
		},
		{
			name: "null value skipped",
			body: map[string]interface{}{
				"name":   "Test Site",
				"region": nil,
			},
			want: map[string]interface{}{
				"name": "Test Site",
			},
		},
		{
			name: "internal fields skipped",
			body: map[string]interface{}{
				"name":    "Test Site",
				"display": "Test Site",
				"_meta":   map[string]interface{}{"foo": "bar"},
			},
			want: map[string]interface{}{
				"name": "Test Site",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRequestBody(tt.body, cfg)
			if len(result) != len(tt.want) {
				t.Errorf("normalizeRequestBody() = %v (len %d), want %v (len %d)",
					result, len(result), tt.want, len(tt.want))
				return
			}
			for k, wantV := range tt.want {
				if gotV, ok := result[k]; !ok || gotV != wantV {
					t.Errorf("normalizeRequestBody()[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

// TestBuildPaginationURLs verifies next/previous URL construction.
func TestBuildPaginationURLs(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/dcim/sites/?limit=10&offset=10", nil)
	c.Request.Host = "localhost:8080"

	t.Run("has next and previous", func(t *testing.T) {
		next, prev := buildPaginationURLs(c, "/api/dcim/sites", 10, 10, 100)
		if next == nil {
			t.Error("expected next URL to be non-nil")
		}
		if prev == nil {
			t.Error("expected prev URL to be non-nil")
		}
		if next != nil && *next == "" {
			t.Error("expected next URL to be non-empty")
		}
	})

	t.Run("no next page", func(t *testing.T) {
		next, _ := buildPaginationURLs(c, "/api/dcim/sites", 90, 10, 100)
		if next != nil {
			t.Errorf("expected next URL to be nil, got %q", *next)
		}
	})

	t.Run("no previous on first page", func(t *testing.T) {
		_, prev := buildPaginationURLs(c, "/api/dcim/sites", 0, 10, 100)
		if prev != nil {
			t.Errorf("expected prev URL to be nil, got %q", *prev)
		}
	})
}

// TestStrPtr verifies the helper creates a proper pointer.
func TestStrPtr(t *testing.T) {
	s := "hello"
	p := strPtr(s)
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != s {
		t.Errorf("strPtr(%q) = %q, want %q", s, *p, s)
	}
}

// TestNormalizeListResults verifies slice type conversion.
func TestNormalizeListResults(t *testing.T) {
	records := []map[string]interface{}{
		{"id": 1, "name": "Site A"},
		{"id": 2, "name": "Site B"},
	}

	result := normalizeListResults(records)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	first := result[0].(map[string]interface{})
	if first["name"] != "Site A" {
		t.Errorf("expected 'Site A', got %v", first["name"])
	}
}

// TestRegisterModelEndpoint verifies endpoint registration.
func TestRegisterModelEndpoint(t *testing.T) {
	cfg := ModelEndpointConfig{
		TableName: "test_table",
	}

	// Register a test endpoint
	RegisterModelEndpoint("/api/test/widgets", cfg)

	if _, exists := netboxModelRegistry["/api/test/widgets"]; !exists {
		t.Error("expected /api/test/widgets to be registered")
	}

	// Clean up
	delete(netboxModelRegistry, "/api/test/widgets")
}

// TestRegionsRegistry verifies a representative frozen endpoint remains
// available while promoted first-profile resources stay physically retired.
func TestRegionsRegistry(t *testing.T) {
	cfg, exists := netboxModelRegistry["/api/dcim/regions"]
	if !exists {
		t.Fatal("expected /api/dcim/regions to be registered in the frozen registry")
	}

	if cfg.TableName != "dcim_region" {
		t.Errorf("TableName = %q, want 'dcim_region'", cfg.TableName)
	}

	if cfg.DefaultSort != "name ASC" {
		t.Errorf("DefaultSort = %q, want 'name ASC'", cfg.DefaultSort)
	}

	// Verify essential search columns
	searchColFound := false
	for _, col := range cfg.SearchCols {
		if col == "name" {
			searchColFound = true
			break
		}
	}
	if !searchColFound {
		t.Error("expected 'name' in SearchCols")
	}

	// Verify filter columns.
	if _, ok := cfg.FilterCols["slug"]; !ok {
		t.Error("expected 'slug' in FilterCols")
	}

	// Region's parent relation is self-referential.
	if parent, ok := cfg.NestedFields["parent"]; !ok || parent.Table != "dcim_region" {
		t.Errorf("expected self-referential parent mapping, got %#v", parent)
	}

	// Verify order fields
	if !cfg.OrderFields["name"] {
		t.Error("expected 'name' to be sortable")
	}
}

// TestAllAutogenModelsRegisterWithoutPanic is a regression test for the
// duplicate route registration panic that occurred when both trailing-slash
// and non-trailing-slash variants were registered. It builds a full Gin engine
// with all ~189 autogen models and verifies no panic occurs.
func TestAllAutogenModelsRegisterWithoutPanic(t *testing.T) {
	t.Run("RegisterNetboxDRFRoutes", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("route registration panicked: %v", r)
			}
		}()
		r := gin.New()
		RegisterNetboxDRFRoutes(r)
	})

	t.Run("RegisterNetboxDRFRoutesWithGroup", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("group route registration panicked: %v", r)
			}
		}()
		r := gin.New()
		grp := r.Group("/api")
		RegisterNetboxDRFRoutesWithGroup(grp)
	})
}

// TestSchemeFromRequest verifies protocol detection.
func TestSchemeFromRequest(t *testing.T) {
	t.Run("plain http", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		if scheme := schemeFromRequest(c); scheme != "http" {
			t.Errorf("schemeFromRequest() = %q, want 'http'", scheme)
		}
	})

	t.Run("forwarded https", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")
		if scheme := schemeFromRequest(c); scheme != "https" {
			t.Errorf("schemeFromRequest() = %q, want 'https'", scheme)
		}
	})
}
