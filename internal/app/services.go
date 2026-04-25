package app

// Service-orchestration layer. Each Provide<Domain>Services lives in
// its own file (services_<domain>.go); ProvideServerServices and
// ProvideWorkerServices below choreograph them in dependency order.
//
// The dependency order is load-bearing — Analytics → Observability →
// Auth/User → Org → Prompt → Credentials → Playground → Evaluation →
// Annotation → Comment / Website. Reordering without checking
// constructor signatures will break the build.

import (
	analyticsService "brokle/internal/core/services/analytics"
	commentService "brokle/internal/core/services/comment"
	registrationService "brokle/internal/core/services/registration"
	websiteService "brokle/internal/core/services/website"
	commentRepo "brokle/internal/infrastructure/repository/comment"
	"brokle/pkg/email"
)

// ProvideServerServices builds every service the HTTP plane needs.
// The Overview analytics service is stamped onto AnalyticsServices
// after construction because it depends on services not yet wired
// at the time AnalyticsServices itself is built.
func ProvideServerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	billingServices := ProvideBillingServices(core.Transactor, repos.Billing, repos.Organization, logger)
	analyticsServices := ProvideAnalyticsServices(repos.Analytics)
	observabilityServices := ProvideObservabilityServices(repos.Observability, repos.Storage, analyticsServices, databases.Redis, cfg, logger)
	authServices := ProvideAuthServices(cfg, repos.User, repos.Auth, repos.Organization, databases, logger)
	userServices := ProvideUserServices(repos.User, repos.Auth, logger)
	orgSvc, memberService, projectService, invitationService, settingsService :=
		ProvideOrganizationServices(repos.User, repos.Auth, repos.Organization, repos.Billing, authServices, cfg, logger)

	// Registration orchestrates user, org, project creation atomically.
	registrationSvc := registrationService.NewRegistrationService(
		core.Transactor,
		repos.User.User,
		repos.Organization.Organization,
		repos.Organization.Member,
		repos.Organization.Project,
		repos.Organization.Invitation,
		authServices.Role,
		authServices.Auth,
		billingServices.BillableUsage,
	)

	promptServices := ProvidePromptServices(core.Transactor, repos.Prompt, analyticsServices.ProviderPricing, cfg, logger)

	// Config validation ensures AI_KEY_ENCRYPTION_KEY is valid, so
	// credentials service is guaranteed to initialize.
	credentialsServices := ProvideCredentialsServices(repos.Credentials, repos.Analytics, cfg, logger)

	playgroundServices := ProvidePlaygroundServices(
		repos.Playground,
		credentialsServices.ProviderCredential,
		promptServices.Compiler,
		promptServices.Execution,
		logger,
	)

	evaluationServices := ProvideEvaluationServices(core.Transactor, repos.Evaluation, repos.Observability, observabilityServices, repos.Prompt, databases.Redis, logger)

	dashboardServices := ProvideDashboardServices(repos.Dashboard, logger)

	annotationServices := ProvideAnnotationServices(core.Transactor, repos.Annotation, evaluationServices, observabilityServices, repos.Organization, logger)

	// Comment service for trace/span comments (with reactions support).
	commentSvc := commentService.NewCommentService(
		commentRepo.NewCommentRepository(databases.TxManager),
		commentRepo.NewReactionRepository(databases.TxManager),
		repos.Observability.Trace,
		logger,
	)

	// Website service (contact form). Reuses the same email sender
	// shape as invitations; degrades to a no-op sender if the
	// transport is unconfigured.
	websiteEmailSender, err := createEmailSender(&cfg.External.Email, logger)
	if err != nil {
		logger.Warn("Failed to create email sender for website service, notifications disabled", "error", err)
		websiteEmailSender = &email.NoOpEmailSender{}
	}
	websiteSvc := websiteService.NewWebsiteService(
		repos.Website.ContactSubmission,
		websiteEmailSender,
		cfg.Notifications.WebsiteNotificationEmail,
		logger,
	)

	// Overview is constructed last because it needs projectService
	// (built by ProvideOrganizationServices) and the credentials repo;
	// stamp it onto the already-built AnalyticsServices container.
	overviewSvc := analyticsService.NewOverviewService(
		repos.Analytics.Overview,
		projectService,
		repos.Credentials.ProviderCredential,
		logger,
	)
	analyticsServices.Overview = overviewSvc

	return &ServiceContainer{
		User:                userServices,
		Auth:                authServices,
		Registration:        registrationSvc,
		OrganizationService: orgSvc,
		MemberService:       memberService,
		ProjectService:      projectService,
		InvitationService:   invitationService,
		SettingsService:     settingsService,
		Observability:       observabilityServices,
		Billing:             billingServices,
		Analytics:           analyticsServices,
		Prompt:              promptServices,
		Credentials:         credentialsServices,
		Playground:          playgroundServices,
		Evaluation:          evaluationServices,
		Dashboard:           dashboardServices,
		Annotation:          annotationServices,
		Comment:             commentSvc,
		Website:             websiteSvc,
	}
}

// ProvideWorkerServices builds the subset of services the worker
// pool needs: observability + billing + analytics + prompt +
// evaluation, plus optional credentials (gated on encryption-key
// availability — disabled gracefully when AI_KEY_ENCRYPTION_KEY is
// not configured).
//
// Auth, user, organization, dashboard, playground, comment, and
// website services are deliberately nil — workers never touch the
// HTTP plane.
func ProvideWorkerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	billingServices := ProvideBillingServices(core.Transactor, repos.Billing, repos.Organization, logger)
	analyticsServices := ProvideAnalyticsServices(repos.Analytics)
	observabilityServices := ProvideObservabilityServices(repos.Observability, repos.Storage, analyticsServices, databases.Redis, cfg, logger)

	promptServices := ProvidePromptServices(core.Transactor, repos.Prompt, analyticsServices.ProviderPricing, cfg, logger)

	var credentialsServices *CredentialsServices
	if cfg.Encryption.AIKeyEncryptionKey != "" {
		credentialsServices = ProvideCredentialsServices(repos.Credentials, repos.Analytics, cfg, logger)
	} else {
		logger.Warn("AI_KEY_ENCRYPTION_KEY not configured, LLM scorer will be disabled")
	}

	evaluationServices := ProvideEvaluationServices(core.Transactor, repos.Evaluation, repos.Observability, observabilityServices, repos.Prompt, databases.Redis, logger)

	return &ServiceContainer{
		Prompt:        promptServices,
		Credentials:   credentialsServices,
		Observability: observabilityServices,
		Billing:       billingServices,
		Analytics:     analyticsServices,
		Evaluation:    evaluationServices,
	}
}
