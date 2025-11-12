package database

import (
	"gorm.io/gorm"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/common"
	orgDomain "brokle/internal/core/domain/organization"
	userDomain "brokle/internal/core/domain/user"

	authRepo "brokle/internal/infrastructure/repository/auth"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	userRepo "brokle/internal/infrastructure/repository/user"
)

// repositoryFactory implements common.RepositoryFactory (private struct).
// Provides transaction-scoped repository instances with lazy initialization and caching.
type repositoryFactory struct {
	db *gorm.DB

	// Cached repositories (lazy initialization for performance)
	userRepo       userDomain.Repository
	orgRepo        orgDomain.OrganizationRepository
	memberRepo     orgDomain.MemberRepository
	projectRepo    orgDomain.ProjectRepository
	invitationRepo orgDomain.InvitationRepository
	roleRepo       authDomain.RoleRepository
	orgMemberRepo  authDomain.OrganizationMemberRepository
}

// NewRepositoryFactory creates a new repository factory instance.
// Returns the interface type to maintain abstraction.
func NewRepositoryFactory(db *gorm.DB) common.RepositoryFactory {
	return &repositoryFactory{db: db}
}

// UserRepository returns a transaction-scoped user repository (cached)
func (f *repositoryFactory) UserRepository() userDomain.Repository {
	if f.userRepo == nil {
		f.userRepo = userRepo.NewUserRepository(f.db)
	}
	return f.userRepo
}

// OrganizationRepository returns a transaction-scoped organization repository (cached)
func (f *repositoryFactory) OrganizationRepository() orgDomain.OrganizationRepository {
	if f.orgRepo == nil {
		f.orgRepo = orgRepo.NewOrganizationRepository(f.db)
	}
	return f.orgRepo
}

// MemberRepository returns a transaction-scoped member repository (cached)
func (f *repositoryFactory) MemberRepository() orgDomain.MemberRepository {
	if f.memberRepo == nil {
		f.memberRepo = orgRepo.NewMemberRepository(f.db)
	}
	return f.memberRepo
}

// ProjectRepository returns a transaction-scoped project repository (cached)
func (f *repositoryFactory) ProjectRepository() orgDomain.ProjectRepository {
	if f.projectRepo == nil {
		f.projectRepo = orgRepo.NewProjectRepository(f.db)
	}
	return f.projectRepo
}

// InvitationRepository returns a transaction-scoped invitation repository (cached)
func (f *repositoryFactory) InvitationRepository() orgDomain.InvitationRepository {
	if f.invitationRepo == nil {
		f.invitationRepo = orgRepo.NewInvitationRepository(f.db)
	}
	return f.invitationRepo
}

// RoleRepository returns a transaction-scoped role repository (cached)
func (f *repositoryFactory) RoleRepository() authDomain.RoleRepository {
	if f.roleRepo == nil {
		f.roleRepo = authRepo.NewRoleRepository(f.db)
	}
	return f.roleRepo
}

// OrganizationMemberRepository returns a transaction-scoped organization member repository (cached)
func (f *repositoryFactory) OrganizationMemberRepository() authDomain.OrganizationMemberRepository {
	if f.orgMemberRepo == nil {
		f.orgMemberRepo = authRepo.NewOrganizationMemberRepository(f.db)
	}
	return f.orgMemberRepo
}
