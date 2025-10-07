package billing

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/workers/analytics"
	"brokle/pkg/ulid"
)

// InvoiceGenerator handles invoice generation and management
type InvoiceGenerator struct {
	logger *logrus.Logger
	config *BillingConfig
}

// Invoice represents a generated invoice
type Invoice struct {
	ID                ulid.ULID            `json:"id"`
	InvoiceNumber     string               `json:"invoice_number"`
	OrganizationID    ulid.ULID            `json:"organization_id"`
	OrganizationName  string               `json:"organization_name"`
	BillingAddress    *BillingAddress      `json:"billing_address"`
	Period            string               `json:"period"`
	PeriodStart       time.Time            `json:"period_start"`
	PeriodEnd         time.Time            `json:"period_end"`
	IssueDate         time.Time            `json:"issue_date"`
	DueDate           time.Time            `json:"due_date"`
	LineItems         []InvoiceLineItem    `json:"line_items"`
	Subtotal          float64              `json:"subtotal"`
	TaxAmount         float64              `json:"tax_amount"`
	DiscountAmount    float64              `json:"discount_amount"`
	TotalAmount       float64              `json:"total_amount"`
	Currency          string               `json:"currency"`
	Status            InvoiceStatus        `json:"status"`
	PaymentTerms      string               `json:"payment_terms"`
	Notes             string               `json:"notes,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
	PaidAt            *time.Time           `json:"paid_at,omitempty"`
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
)

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	ID           ulid.ULID `json:"id"`
	Description  string    `json:"description"`
	Quantity     float64   `json:"quantity"`
	UnitPrice    float64   `json:"unit_price"`
	Amount       float64   `json:"amount"`
	ProviderID   *ulid.ULID `json:"provider_id,omitempty"`
	ProviderName string    `json:"provider_name,omitempty"`
	ModelID      *ulid.ULID `json:"model_id,omitempty"`
	ModelName    string    `json:"model_name,omitempty"`
	RequestType  string    `json:"request_type,omitempty"`
	Tokens       int64     `json:"tokens,omitempty"`
	Requests     int64     `json:"requests,omitempty"`
}

// BillingAddress represents an organization's billing address
type BillingAddress struct {
	Company    string `json:"company"`
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	TaxID      string `json:"tax_id,omitempty"`
}

// TaxConfiguration represents tax configuration for billing
type TaxConfiguration struct {
	TaxRate     float64 `json:"tax_rate"`     // e.g., 0.08 for 8%
	TaxName     string  `json:"tax_name"`     // e.g., "VAT", "GST", "Sales Tax"
	TaxID       string  `json:"tax_id"`       // Tax identification number
	IsInclusive bool    `json:"is_inclusive"` // Whether tax is included in prices
}

// NewInvoiceGenerator creates a new invoice generator instance
func NewInvoiceGenerator(logger *logrus.Logger, config *BillingConfig) *InvoiceGenerator {
	return &InvoiceGenerator{
		logger: logger,
		config: config,
	}
}

// GenerateInvoice creates an invoice from a billing summary
func (g *InvoiceGenerator) GenerateInvoice(
	ctx context.Context,
	summary *analytics.BillingSummary,
	organizationName string,
	billingAddress *BillingAddress,
) (*Invoice, error) {
	
	if summary.NetCost <= 0 {
		return nil, fmt.Errorf("cannot generate invoice for zero or negative amount: %f", summary.NetCost)
	}

	invoice := &Invoice{
		ID:               ulid.New(),
		InvoiceNumber:    g.generateInvoiceNumber(summary.OrganizationID, summary.PeriodStart),
		OrganizationID:   summary.OrganizationID,
		OrganizationName: organizationName,
		BillingAddress:   billingAddress,
		Period:           summary.Period,
		PeriodStart:      summary.PeriodStart,
		PeriodEnd:        summary.PeriodEnd,
		IssueDate:        time.Now(),
		DueDate:          time.Now().Add(g.config.PaymentGracePeriod),
		Currency:         summary.Currency,
		Status:           InvoiceStatusDraft,
		PaymentTerms:     fmt.Sprintf("Net %d days", int(g.config.PaymentGracePeriod.Hours()/24)),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Generate line items from usage breakdown
	lineItems := g.generateLineItems(summary)
	invoice.LineItems = lineItems

	// Calculate totals
	subtotal := 0.0
	for _, item := range lineItems {
		subtotal += item.Amount
	}
	
	invoice.Subtotal = subtotal
	invoice.DiscountAmount = summary.Discounts
	
	// Apply tax if configured
	taxConfig := g.getTaxConfiguration(billingAddress)
	if taxConfig != nil {
		taxableAmount := subtotal - invoice.DiscountAmount
		invoice.TaxAmount = taxableAmount * taxConfig.TaxRate
	}
	
	invoice.TotalAmount = invoice.Subtotal - invoice.DiscountAmount + invoice.TaxAmount

	// Add metadata
	invoice.Metadata = map[string]interface{}{
		"total_requests":     summary.TotalRequests,
		"total_tokens":       summary.TotalTokens,
		"provider_breakdown": summary.ProviderBreakdown,
		"model_breakdown":    summary.ModelBreakdown,
		"billing_period":     summary.Period,
	}

	g.logger.WithFields(logrus.Fields{
		"invoice_id":      invoice.ID,
		"invoice_number":  invoice.InvoiceNumber,
		"organization_id": invoice.OrganizationID,
		"total_amount":    invoice.TotalAmount,
		"currency":        invoice.Currency,
	}).Info("Generated invoice")

	return invoice, nil
}

// GenerateInvoiceHTML generates HTML representation of an invoice
func (g *InvoiceGenerator) GenerateInvoiceHTML(ctx context.Context, invoice *Invoice) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Invoice {{.InvoiceNumber}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; color: #333; }
        .header { border-bottom: 2px solid #007acc; padding-bottom: 20px; margin-bottom: 30px; }
        .company-name { font-size: 28px; font-weight: bold; color: #007acc; }
        .invoice-title { font-size: 24px; margin: 20px 0; }
        .invoice-details { margin: 20px 0; }
        .billing-info { display: flex; justify-content: space-between; margin: 30px 0; }
        .billing-address { max-width: 300px; }
        .invoice-table { width: 100%; border-collapse: collapse; margin: 30px 0; }
        .invoice-table th, .invoice-table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .invoice-table th { background-color: #f8f9fa; font-weight: bold; }
        .amount { text-align: right; }
        .totals { max-width: 400px; margin-left: auto; margin-top: 30px; }
        .totals table { width: 100%; }
        .totals .total-row { font-weight: bold; font-size: 18px; border-top: 2px solid #333; }
        .payment-info { margin-top: 40px; padding: 20px; background-color: #f8f9fa; border-radius: 5px; }
        .footer { margin-top: 40px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <div class="company-name">Brokle</div>
        <div>The Open-Source AI Control Plane</div>
    </div>

    <div class="invoice-title">Invoice {{.InvoiceNumber}}</div>

    <div class="invoice-details">
        <strong>Issue Date:</strong> {{.IssueDate.Format "January 2, 2006"}}<br>
        <strong>Due Date:</strong> {{.DueDate.Format "January 2, 2006"}}<br>
        <strong>Period:</strong> {{.PeriodStart.Format "January 2, 2006"}} - {{.PeriodEnd.Format "January 2, 2006"}}
    </div>

    <div class="billing-info">
        <div>
            <strong>Bill To:</strong><br>
            {{.OrganizationName}}<br>
            {{if .BillingAddress}}
                {{.BillingAddress.Company}}<br>
                {{.BillingAddress.Address1}}<br>
                {{if .BillingAddress.Address2}}{{.BillingAddress.Address2}}<br>{{end}}
                {{.BillingAddress.City}}, {{.BillingAddress.State}} {{.BillingAddress.PostalCode}}<br>
                {{.BillingAddress.Country}}<br>
                {{if .BillingAddress.TaxID}}<strong>Tax ID:</strong> {{.BillingAddress.TaxID}}{{end}}
            {{end}}
        </div>
        <div>
            <strong>From:</strong><br>
            Brokle Inc.<br>
            123 AI Boulevard<br>
            San Francisco, CA 94105<br>
            United States
        </div>
    </div>

    <table class="invoice-table">
        <thead>
            <tr>
                <th>Description</th>
                <th>Quantity</th>
                <th>Unit Price</th>
                <th class="amount">Amount</th>
            </tr>
        </thead>
        <tbody>
            {{range .LineItems}}
            <tr>
                <td>
                    {{.Description}}
                    {{if .ProviderName}}<br><small>Provider: {{.ProviderName}}</small>{{end}}
                    {{if .ModelName}}<br><small>Model: {{.ModelName}}</small>{{end}}
                </td>
                <td>{{printf "%.0f" .Quantity}}</td>
                <td>${{printf "%.4f" .UnitPrice}}</td>
                <td class="amount">${{printf "%.2f" .Amount}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="totals">
        <table>
            <tr>
                <td><strong>Subtotal:</strong></td>
                <td class="amount">${{printf "%.2f" .Subtotal}}</td>
            </tr>
            {{if gt .DiscountAmount 0}}
            <tr>
                <td><strong>Discount:</strong></td>
                <td class="amount">-${{printf "%.2f" .DiscountAmount}}</td>
            </tr>
            {{end}}
            {{if gt .TaxAmount 0}}
            <tr>
                <td><strong>Tax:</strong></td>
                <td class="amount">${{printf "%.2f" .TaxAmount}}</td>
            </tr>
            {{end}}
            <tr class="total-row">
                <td><strong>Total:</strong></td>
                <td class="amount"><strong>${{printf "%.2f" .TotalAmount}} {{.Currency}}</strong></td>
            </tr>
        </table>
    </div>

    <div class="payment-info">
        <strong>Payment Terms:</strong> {{.PaymentTerms}}<br>
        <strong>Status:</strong> {{.Status}}<br>
        {{if .Notes}}<strong>Notes:</strong> {{.Notes}}<br>{{end}}
    </div>

    <div class="footer">
        <p>Thank you for using Brokle! For questions about this invoice, please contact support@brokle.com</p>
        <p>This invoice was generated automatically on {{.CreatedAt.Format "January 2, 2006 at 3:04 PM MST"}}</p>
    </div>
</body>
</html>
`

	t, err := template.New("invoice").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse invoice template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, invoice); err != nil {
		return "", fmt.Errorf("failed to execute invoice template: %w", err)
	}

	return buf.String(), nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (g *InvoiceGenerator) MarkInvoiceAsPaid(ctx context.Context, invoice *Invoice, paidAt time.Time) error {
	if invoice.Status == InvoiceStatusPaid {
		return fmt.Errorf("invoice %s is already marked as paid", invoice.InvoiceNumber)
	}

	invoice.Status = InvoiceStatusPaid
	invoice.PaidAt = &paidAt
	invoice.UpdatedAt = time.Now()

	g.logger.WithFields(logrus.Fields{
		"invoice_id":      invoice.ID,
		"invoice_number":  invoice.InvoiceNumber,
		"organization_id": invoice.OrganizationID,
		"paid_at":         paidAt,
	}).Info("Invoice marked as paid")

	return nil
}

// MarkInvoiceAsOverdue marks an invoice as overdue
func (g *InvoiceGenerator) MarkInvoiceAsOverdue(ctx context.Context, invoice *Invoice) error {
	if invoice.Status == InvoiceStatusPaid {
		return fmt.Errorf("cannot mark paid invoice %s as overdue", invoice.InvoiceNumber)
	}

	invoice.Status = InvoiceStatusOverdue
	invoice.UpdatedAt = time.Now()

	g.logger.WithFields(logrus.Fields{
		"invoice_id":      invoice.ID,
		"invoice_number":  invoice.InvoiceNumber,
		"organization_id": invoice.OrganizationID,
		"due_date":        invoice.DueDate,
	}).Warn("Invoice marked as overdue")

	return nil
}

// CancelInvoice cancels an invoice
func (g *InvoiceGenerator) CancelInvoice(ctx context.Context, invoice *Invoice, reason string) error {
	if invoice.Status == InvoiceStatusPaid {
		return fmt.Errorf("cannot cancel paid invoice %s", invoice.InvoiceNumber)
	}

	invoice.Status = InvoiceStatusCancelled
	invoice.UpdatedAt = time.Now()
	
	if invoice.Metadata == nil {
		invoice.Metadata = make(map[string]interface{})
	}
	invoice.Metadata["cancellation_reason"] = reason
	invoice.Metadata["cancelled_at"] = time.Now()

	g.logger.WithFields(logrus.Fields{
		"invoice_id":      invoice.ID,
		"invoice_number":  invoice.InvoiceNumber,
		"organization_id": invoice.OrganizationID,
		"reason":          reason,
	}).Info("Invoice cancelled")

	return nil
}

// GetInvoiceSummary generates a summary of multiple invoices
func (g *InvoiceGenerator) GetInvoiceSummary(ctx context.Context, invoices []*Invoice) *InvoiceSummary {
	summary := &InvoiceSummary{
		TotalInvoices: len(invoices),
		StatusCounts:  make(map[InvoiceStatus]int),
		TotalAmount:   0,
		PaidAmount:    0,
		OutstandingAmount: 0,
		OverdueAmount: 0,
	}

	for _, invoice := range invoices {
		// Count by status
		summary.StatusCounts[invoice.Status]++

		// Calculate amounts
		summary.TotalAmount += invoice.TotalAmount

		switch invoice.Status {
		case InvoiceStatusPaid:
			summary.PaidAmount += invoice.TotalAmount
		case InvoiceStatusOverdue:
			summary.OverdueAmount += invoice.TotalAmount
			summary.OutstandingAmount += invoice.TotalAmount
		case InvoiceStatusSent, InvoiceStatusDraft:
			summary.OutstandingAmount += invoice.TotalAmount
		}

		// Track earliest and latest dates
		if summary.EarliestDate == nil || invoice.IssueDate.Before(*summary.EarliestDate) {
			summary.EarliestDate = &invoice.IssueDate
		}
		if summary.LatestDate == nil || invoice.IssueDate.After(*summary.LatestDate) {
			summary.LatestDate = &invoice.IssueDate
		}
	}

	return summary
}

// InvoiceSummary represents a summary of multiple invoices
type InvoiceSummary struct {
	TotalInvoices     int                        `json:"total_invoices"`
	StatusCounts      map[InvoiceStatus]int      `json:"status_counts"`
	TotalAmount       float64                    `json:"total_amount"`
	PaidAmount        float64                    `json:"paid_amount"`
	OutstandingAmount float64                    `json:"outstanding_amount"`
	OverdueAmount     float64                    `json:"overdue_amount"`
	EarliestDate      *time.Time                 `json:"earliest_date,omitempty"`
	LatestDate        *time.Time                 `json:"latest_date,omitempty"`
}

// Internal methods

func (g *InvoiceGenerator) generateInvoiceNumber(orgID ulid.ULID, periodStart time.Time) string {
	// Format: BRKL-YYYY-MM-{ORG_SHORT}-{SEQUENCE}
	orgShort := orgID.String()[:8] // First 8 characters of org ID
	yearMonth := periodStart.Format("2006-01")
	
	// In a real implementation, you'd want to get the next sequence number from the database
	sequence := "001"
	
	return fmt.Sprintf("BRKL-%s-%s-%s", yearMonth, orgShort, sequence)
}

func (g *InvoiceGenerator) generateLineItems(summary *analytics.BillingSummary) []InvoiceLineItem {
	var lineItems []InvoiceLineItem

	// Create line items based on provider breakdown
	for providerKey, amount := range summary.ProviderBreakdown {
		if amount > 0 {
			lineItem := InvoiceLineItem{
				ID:          ulid.New(),
				Description: fmt.Sprintf("AI API Usage - Provider %s", providerKey),
				Quantity:    1,
				UnitPrice:   amount,
				Amount:      amount,
			}
			
			// In a real implementation, you'd look up provider details
			lineItem.ProviderName = fmt.Sprintf("Provider %s", providerKey[:8])
			
			lineItems = append(lineItems, lineItem)
		}
	}

	// If no provider breakdown, create a single line item
	if len(lineItems) == 0 {
		lineItems = append(lineItems, InvoiceLineItem{
			ID:          ulid.New(),
			Description: fmt.Sprintf("AI API Usage - %s", summary.Period),
			Quantity:    float64(summary.TotalRequests),
			UnitPrice:   summary.TotalCost / float64(summary.TotalRequests),
			Amount:      summary.TotalCost,
		})
	}

	return lineItems
}

func (g *InvoiceGenerator) getTaxConfiguration(billingAddress *BillingAddress) *TaxConfiguration {
	if billingAddress == nil {
		return nil
	}

	// Simple tax configuration based on country
	// In a real implementation, this would be more sophisticated
	taxConfigs := map[string]*TaxConfiguration{
		"US": {
			TaxRate:     0.08, // 8% sales tax
			TaxName:     "Sales Tax",
			IsInclusive: false,
		},
		"UK": {
			TaxRate:     0.20, // 20% VAT
			TaxName:     "VAT",
			IsInclusive: false,
		},
		"CA": {
			TaxRate:     0.13, // 13% HST (varies by province)
			TaxName:     "HST",
			IsInclusive: false,
		},
	}

	return taxConfigs[billingAddress.Country]
}

// Health check
func (g *InvoiceGenerator) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"service": "invoice_generator",
		"status":  "healthy",
		"config": map[string]interface{}{
			"default_currency":       g.config.DefaultCurrency,
			"payment_grace_period":   g.config.PaymentGracePeriod.String(),
			"invoice_generation":     g.config.InvoiceGeneration,
		},
	}
}

// Additional utility methods for invoice management

// IsOverdue checks if an invoice is overdue
func (g *InvoiceGenerator) IsOverdue(invoice *Invoice) bool {
	return invoice.Status != InvoiceStatusPaid && 
		   invoice.Status != InvoiceStatusCancelled && 
		   time.Now().After(invoice.DueDate)
}

// CalculateLateFee calculates late fee for an overdue invoice
func (g *InvoiceGenerator) CalculateLateFee(invoice *Invoice, lateFeeRate float64) float64 {
	if !g.IsOverdue(invoice) {
		return 0
	}

	daysOverdue := int(time.Since(invoice.DueDate).Hours() / 24)
	if daysOverdue <= 0 {
		return 0
	}

	return invoice.TotalAmount * lateFeeRate * float64(daysOverdue) / 365.0
}

// GeneratePaymentReminder generates text for a payment reminder
func (g *InvoiceGenerator) GeneratePaymentReminder(invoice *Invoice) string {
	daysOverdue := int(time.Since(invoice.DueDate).Hours() / 24)
	
	var message string
	if daysOverdue <= 0 {
		daysToDue := int(time.Until(invoice.DueDate).Hours() / 24)
		message = fmt.Sprintf("Your invoice %s for $%.2f is due in %d days.", 
			invoice.InvoiceNumber, invoice.TotalAmount, daysToDue)
	} else {
		message = fmt.Sprintf("Your invoice %s for $%.2f is %d days overdue. Please remit payment immediately.", 
			invoice.InvoiceNumber, invoice.TotalAmount, daysOverdue)
	}

	return message
}