package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

// SiteListCriteria is validated by SiteService before reaching persistence.
type SiteListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []SiteSort
	Names                 []string
	Slugs                 []string
	Statuses              []dcimdomain.SiteStatus
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	// DeferPagination asks persistence to return the complete ordered result
	// after SQL predicates. The service uses this fail-closed path when an
	// authorizer cannot provide a complete pre-query object scope.
	DeferPagination bool
}

type SitePage struct {
	Count   uint64
	Results []*dcimdomain.Site
}

// SiteRepository is implemented by the PostgreSQL adapter. Mutation methods
// and GetForUpdate consume the transaction-aware context supplied by UnitOfWork.
type SiteRepository interface {
	List(context.Context, SiteListCriteria) (SitePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.Site, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Site, error)
	Create(context.Context, *dcimdomain.Site) error
	Update(context.Context, *dcimdomain.Site) error
	Delete(context.Context, *dcimdomain.Site) error
}
