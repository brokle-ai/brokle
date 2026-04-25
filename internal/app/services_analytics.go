package app

import (
	analyticsService "brokle/internal/core/services/analytics"
)

// ProvideAnalyticsServices builds the provider-pricing service used
// by the prompt execution path and the observability cost annotator.
//
// The Overview service (consumed by the dashboard /api/v1/overview
// route) needs services not yet built at this point — projectService
// + credentials repo — so the orchestrator (services.go) constructs
// it later and stamps it onto AnalyticsServices.Overview.
func ProvideAnalyticsServices(
	analyticsRepos *AnalyticsRepositories,
) *AnalyticsServices {
	providerPricingServiceImpl := analyticsService.NewProviderPricingService(analyticsRepos.ProviderModel)

	return &AnalyticsServices{
		ProviderPricing: providerPricingServiceImpl,
	}
}
