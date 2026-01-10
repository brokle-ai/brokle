package database

import (
	"context"

	"gorm.io/gorm"

	"brokle/internal/infrastructure/shared"
)

// gormTransactor implements the Transactor interface using GORM.
// This is the idiomatic Go pattern for transaction management in clean architecture.
type gormTransactor struct {
	db *gorm.DB
}

// NewTransactor creates a new GORM-based transactor.
// This is the constructor used in the DI layer (providers.go).
func NewTransactor(db *gorm.DB) *gormTransactor {
	return &gormTransactor{db: db}
}

// WithinTransaction executes fn within a database transaction.
// The transaction is injected into the context and can be extracted by repositories
// using the GetDB helper function.
//
// Transaction semantics:
//   - Commits automatically when fn returns nil
//   - Rolls back automatically when fn returns an error
//   - Rolls back automatically on panic (GORM handles this)
//
// Example usage in services:
//
//	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
//	    if err := s.repo.Create(ctx, entity); err != nil {
//	        return err // Triggers rollback
//	    }
//	    return s.repo.Update(ctx, other) // Commits on success
//	})
func (t *gormTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Inject the transaction into context using shared helper
		txCtx := shared.InjectTx(ctx, tx)
		return fn(txCtx)
	})
}
