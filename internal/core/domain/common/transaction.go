package common

import (
	"context"
)

// Transactor provides transaction management without exposing database details.
// This is the idiomatic Go approach recommended by experts (Three Dots Labs, community consensus).
//
// Usage:
//
//	err := transactor.WithinTransaction(ctx, func(ctx context.Context) error {
//	    // All repository calls within this function automatically use the transaction
//	    if err := repo.Create(ctx, entity); err != nil {
//	        return err // Automatic rollback
//	    }
//	    return repo.Update(ctx, other) // Automatic commit on success
//	})
//
// How it works:
//   - Transaction is injected into the context
//   - Repositories extract the transaction using a helper function
//   - Commits on nil return, rolls back on error or panic
//   - No factory pattern needed - services use their existing repository fields
type Transactor interface {
	// WithinTransaction executes fn within a database transaction.
	// The transaction is injected into the context and automatically extracted by repositories.
	// Commits on nil return, rolls back on error or panic.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
