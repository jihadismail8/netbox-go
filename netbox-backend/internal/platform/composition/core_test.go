package composition_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
	"netbox-go/internal/platform/composition"
)

func TestNewCoreUsesFailClosedPermissionAuthorizer(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(
		sqlite.Open("file:composition_default_authorizer?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)

	core := composition.NewCore(db)
	_, err = core.Sites.CreateSite(
		t.Context(),
		identity.Principal{ID: 7, Username: "ungranted"},
		applicationdcim.CreateSiteCommand{},
	)
	require.Error(t, err)
	assert.Equal(t, shared.ErrorReasonForbidden, shared.ReasonOf(err))
}
