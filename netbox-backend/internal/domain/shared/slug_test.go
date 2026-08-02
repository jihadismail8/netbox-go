package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/shared"
)

func TestParseSlugMatchesDefaultDjangoSlugField(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"site-1", "Site_1", "ABC123"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			slug, err := shared.ParseSlug("  "+value+"  ", 100)
			require.NoError(t, err)
			assert.Equal(t, value, slug.String())
		})
	}
}

func TestParseSlugRejectsBlankUnicodeAndPunctuation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  string
		reason string
	}{
		{name: "blank", value: "  ", reason: "required"},
		{name: "unicode", value: "münchen", reason: "invalid"},
		{name: "punctuation", value: "site.one", reason: "invalid"},
		{name: "too long", value: "aaaaaaaaaaa", reason: "max_length"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := shared.ParseSlug(test.value, 10)
			require.Error(t, err)
			violations := shared.ViolationsOf(err)
			require.Len(t, violations, 1)
			assert.Equal(t, "slug", violations[0].Field)
			assert.Equal(t, test.reason, violations[0].Reason)
		})
	}
}
