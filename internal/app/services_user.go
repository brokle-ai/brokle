package app

import (
	"log/slog"

	userService "brokle/internal/core/services/user"
)

// ProvideUserServices wires the user-domain services. The first nil
// argument to NewUserService is a deprecated email-sender slot kept
// for binary-compat with the upstream signature; the org-member repo
// is the only cross-domain dep.
func ProvideUserServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	logger *slog.Logger,
) *UserServices {
	userSvc := userService.NewUserService(
		userRepos.User,
		nil,
		authRepos.OrganizationMember,
	)

	profileSvc := userService.NewProfileService(
		userRepos.User,
	)

	return &UserServices{
		User:    userSvc,
		Profile: profileSvc,
	}
}
