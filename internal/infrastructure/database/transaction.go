package database

import (
	"context"

	"gorm.io/gorm"

	"brokle/internal/core/domain/common"
)

// transactionManager implements common.TransactionManager (private struct)
type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager instance.
// Returns the interface type to maintain dependency inversion.
func NewTransactionManager(db *gorm.DB) common.TransactionManager {
	return &transactionManager{db: db}
}

// WithTransaction implements common.TransactionManager interface.
// Executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// The original context is passed through to preserve request-scoped values
// (request ID, user context, etc.) while GORM manages the transaction state.
func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(context.Context, common.RepositoryFactory) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		factory := NewRepositoryFactory(tx)
		// Pass original ctx (preserves request values), not GORM's derived context
		return fn(ctx, factory)
	})
}
