package server

import (
	"slices"
	"testing"
)

func TestCORSHTTPOptionDefensiveCopy(t *testing.T) {
	origins := []string{
		"https://console.example.test",
		"http://localhost:3000",
	}
	option := WithHTTPCORSAllowedOrigins(origins)
	origins[0] = "https://mutated-before-apply.example.test"

	first := defaultHTTPOptions()
	first.apply(option)
	want := []string{
		"https://console.example.test",
		"http://localhost:3000",
	}
	if !slices.Equal(first.corsAllowedOrigins, want) {
		t.Fatalf("first CORS origins = %v, want %v", first.corsAllowedOrigins, want)
	}

	second := defaultHTTPOptions()
	second.apply(option)
	first.corsAllowedOrigins[0] = "https://mutated-after-apply.example.test"
	if !slices.Equal(second.corsAllowedOrigins, want) {
		t.Fatalf("second CORS origins = %v, want independent copy %v", second.corsAllowedOrigins, want)
	}
}
