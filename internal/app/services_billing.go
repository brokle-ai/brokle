package app

import (
	"log/slog"

	commonDomain "brokle/internal/core/domain/common"
	billingService "brokle/internal/core/services/billing"
)

// ProvideBillingServices wires the billing-domain services. The
// pricing service is built first because the usage service depends
// on it for tier resolution; the contract service is wrapped in a
// transactor so org → contract → history writes commit atomically.
func ProvideBillingServices(
	transactor commonDomain.Transactor,
	billingRepos *BillingRepositories,
	orgRepos *OrganizationRepositories,
	logger *slog.Logger,
) *BillingServices {
	pricingService := billingService.NewPricingService(
		billingRepos.OrganizationBilling,
		billingRepos.Plan,
		billingRepos.Contract,
		billingRepos.VolumeTier,
		logger,
	)

	billableUsageService := billingService.NewBillableUsageService(
		billingRepos.BillableUsage,
		billingRepos.OrganizationBilling,
		pricingService,
		billingRepos.Plan,
		logger,
	)

	budgetService := billingService.NewBudgetService(
		billingRepos.UsageBudget,
		billingRepos.UsageAlert,
		orgRepos.Project,
		logger,
	)

	contractService := billingService.NewContractService(
		transactor,
		billingRepos.Contract,
		billingRepos.VolumeTier,
		billingRepos.ContractHistory,
		billingRepos.OrganizationBilling,
		logger,
	)

	return &BillingServices{
		BillableUsage: billableUsageService,
		Budget:        budgetService,
		Pricing:       pricingService,
		Contract:      contractService,
	}
}
