package common

import (
	"context"

	authDomain "brokle/internal/core/domain/auth"
	orgDomain "brokle/internal/core/domain/organization"
	userDomain "brokle/internal/core/domain/user"
)

// TransactionManager coordinates transactions across multiple repositories.
// This interface maintains dependency inversion - defined in core, implemented in infrastructure.
//
// Usage:
//
//	err := txManager.WithTransaction(ctx, func(ctx context.Context, factory RepositoryFactory) error {
//	    userRepo := factory.UserRepository()
//	    orgRepo := factory.OrganizationRepository()
//	    // All operations share the same database transaction
//	    return nil
//	})
type TransactionManager interface {
	// WithTransaction executes the given function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, the transaction is committed.
	WithTransaction(ctx context.Context, fn func(context.Context, RepositoryFactory) error) error
}

// RepositoryFactory provides access to transaction-scoped repositories.
// All methods return domain interfaces only - never concrete infrastructure types.
// This maintains clean architecture by preventing core layer from depending on infrastructure.
type RepositoryFactory interface {
	// User domain repositories
	UserRepository() userDomain.Repository

	// Organization domain repositories
	OrganizationRepository() orgDomain.OrganizationRepository
	MemberRepository() orgDomain.MemberRepository
	ProjectRepository() orgDomain.ProjectRepository
	InvitationRepository() orgDomain.InvitationRepository

	// Auth domain repositories
	RoleRepository() authDomain.RoleRepository
	OrganizationMemberRepository() authDomain.OrganizationMemberRepository
}
