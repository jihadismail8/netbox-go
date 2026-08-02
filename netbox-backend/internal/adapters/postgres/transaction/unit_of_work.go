// Package transaction implements the application transaction boundary without
// exposing GORM to application or domain code.
package transaction

import (
	"context"
	"errors"

	"gorm.io/gorm"

	applicationtransaction "netbox-go/internal/application/transaction"
)

var errOperationRequired = errors.New("postgres unit of work requires an operation")

type databaseContextKey struct{}

// UnitOfWork runs application mutations in one database transaction and binds
// the transaction handle to the callback context. PostgreSQL adapters resolve
// that handle with FromContext, so repositories and the change recorder commit
// or roll back atomically without passing *gorm.DB through application ports.
type UnitOfWork struct {
	db *gorm.DB
}

var _ applicationtransaction.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	if db == nil {
		panic("postgres unit of work requires a database")
	}
	return &UnitOfWork{db: db}
}

func (unit *UnitOfWork) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	if operation == nil {
		return errOperationRequired
	}
	if _, alreadyBound := FromContext(ctx); alreadyBound {
		return operation(ctx)
	}

	return unit.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactionContext := context.WithValue(ctx, databaseContextKey{}, tx)
		return operation(transactionContext)
	})
}

// FromContext returns the transaction bound by UnitOfWork. It is intentionally
// adapter-only: application and domain packages continue to depend solely on
// transaction.UnitOfWork.
func FromContext(ctx context.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}
	tx, ok := ctx.Value(databaseContextKey{}).(*gorm.DB)
	return tx, ok && tx != nil
}
