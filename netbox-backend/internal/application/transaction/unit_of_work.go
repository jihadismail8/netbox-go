// Package transaction owns the transaction boundary consumed by application
// use cases. Infrastructure carries its transaction handle through the child
// context; application code never imports GORM.
package transaction

import "context"

type UnitOfWork interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
