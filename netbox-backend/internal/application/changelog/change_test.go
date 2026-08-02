package changelog_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/changelog"
	"netbox-go/internal/domain/shared"
)

func TestNewChangeCopiesSnapshots(t *testing.T) {
	t.Parallel()

	before := &testSnapshot{Name: "before"}
	after := &testSnapshot{Name: "after"}
	change, err := changelog.NewChange(
		3,
		"dcim.site",
		9,
		"after",
		changelog.ActionUpdate,
		before,
		after,
		shared.NewTimestamp(time.Now()),
	)
	require.NoError(t, err)
	before.Name = "mutated"
	after.Name = "mutated"
	assert.Equal(t, "before", change.Before.(testSnapshot).Name)
	assert.Equal(t, "after", change.After.(testSnapshot).Name)
}

func TestNewChangeRejectsSnapshotShapeForAction(t *testing.T) {
	t.Parallel()

	_, err := changelog.NewChange(
		3,
		"dcim.site",
		9,
		"site",
		changelog.ActionCreate,
		testSnapshot{Name: "unexpected"},
		testSnapshot{Name: "site"},
		shared.NewTimestamp(time.Now()),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

type testSnapshot struct {
	Name string
}

func (testSnapshot) ObjectType() string { return "dcim.site" }

func (snapshot testSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }
