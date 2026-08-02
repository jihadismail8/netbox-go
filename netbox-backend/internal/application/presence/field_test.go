package presence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"netbox-go/internal/application/presence"
)

func TestFieldPreservesOmittedNullAndPresentZero(t *testing.T) {
	t.Parallel()

	var omitted presence.Field[string]
	assert.Equal(t, presence.Omitted, omitted.State())
	_, ok := omitted.Get()
	assert.False(t, ok)

	null := presence.NullField[string]()
	assert.Equal(t, presence.Null, null.State())
	_, ok = null.Get()
	assert.False(t, ok)

	present := presence.Value("")
	assert.Equal(t, presence.Present, present.State())
	value, ok := present.Get()
	assert.True(t, ok)
	assert.Empty(t, value)
}
