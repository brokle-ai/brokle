package app

import (
	"log/slog"
	"os"

	"brokle/internal/config"
	orgService "brokle/internal/core/services/organization"
)

// ProvideOrganizationServices builds the organization domain's
// member, project, invitation, and settings services. Returns the
// raw services rather than a container so the orchestrator (in
// services.go) can place them as flat fields on ServiceContainer —
// the existing dashboard handlers already consume them flat.
//
// The email-sender lifecycle is shared with the website service via
// createEmailSender (see lifecycle.go).
func ProvideOrganizationServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	orgRepos *OrganizationRepositories,
	billingRepos *BillingRepositories,
	authServices *AuthServices,
	databases *DatabaseContainer,
	cfg *config.Config,
	logger *slog.Logger,
) (
	*orgService.OrganizationService,
	*orgService.MemberService,
	*orgService.ProjectService,
	*orgService.InvitationService,
	*orgService.OrganizationSettingsService,
) {
	memberSvc := orgService.NewMemberService(
		orgRepos.Member,
		authRepos.ProjectMember,
		orgRepos.Organization,
		userRepos.User,
		authServices.Role,
		databases.TxManager,
	)

	projectSvc := orgService.NewProjectService(
		orgRepos.Project,
		orgRepos.Organization,
		orgRepos.Member,
		authServices.Role,
		authServices.ProjectMembers,
		databases.TxManager,
	)

	emailSender, err := createEmailSender(&cfg.External.Email, logger)
	if err != nil {
		logger.Error("Failed to create email sender", "error", err)
		os.Exit(1)
	}

	invitationSvc := orgService.NewInvitationService(
		orgRepos.Invitation,
		orgRepos.Organization,
		orgRepos.Member,
		userRepos.User,
		authServices.Role,
		emailSender,
		orgService.InvitationServiceConfig{
			AppURL: cfg.Server.AppURL,
		},
		logger.With("service", "invitation"),
	)

	orgSvc := orgService.NewOrganizationService(
		orgRepos.Organization,
		userRepos.User,
		memberSvc,
		projectSvc,
		authServices.Role,
		billingRepos.OrganizationBilling,
		billingRepos.Plan,
		logger,
	)

	settingsSvc := orgService.NewOrganizationSettingsService(
		orgRepos.Settings,
		orgRepos.Member,
	)

	return orgSvc, memberSvc, projectSvc, invitationSvc, settingsSvc
}
