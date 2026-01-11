package shared

import (
	"context"

	"gorm.io/gorm"
)

// txKey is an unexported type for context keys to prevent collisions.
// Using a struct type ensures type safety and prevents accidental key conflicts.
// This key is shared between transactor (injection) and repositories (extraction).
type txKey struct{}

// GetDB returns the transaction-aware GORM DB from context.
// If a transaction exists in the context, it returns the transactional DB.
// Otherwise, it returns the default (non-transactional) DB.
//
// This helper enables repositories to transparently support both:
//   - Transactional calls: When called within WithinTransaction
//   - Non-transactional calls: When called directly
//
// Repository usage pattern:
//
//	func (r *repository) getDB(ctx context.Context) *gorm.DB {
//	    return shared.GetDB(ctx, r.db)
//	}
//
//	func (r *repository) Create(ctx context.Context, entity *Entity) error {
//	    return r.getDB(ctx).WithContext(ctx).Create(entity).Error
//	}
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}

// InjectTx injects a transaction into the context.
// This is used by the Transactor implementation to make the transaction
// available to repositories.
//
// Transactor usage:
//
//	txCtx := shared.InjectTx(ctx, tx)
//	return fn(txCtx)
func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
