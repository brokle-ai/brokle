# Analytics & Billing Integration

This directory contains the complete analytics worker and billing service implementation for the Brokle AI Gateway.

## Overview

The analytics and billing system provides:
- Real-time usage tracking and metrics collection
- Comprehensive cost calculation and billing management
- Advanced discount and promotional pricing
- Professional invoice generation
- Usage quota management and enforcement
- Background processing for scalability

## Components

### 1. Gateway Analytics Worker (`gateway_analytics_worker.go`)

**Purpose**: Processes gateway usage data for analytics and billing

**Key Features**:
- Batched processing of request metrics for performance
- Real-time cost tracking and billing integration
- Background workers for data aggregation
- In-memory buffering with configurable flush intervals
- Health monitoring and metrics collection

**Metrics Tracked**:
- Request metrics (ID, organization, provider, model, tokens, costs)
- Usage metrics (aggregated by time periods)
- Cost metrics (detailed cost breakdown and billing data)

**Configuration**:
```go
config := &WorkerConfig{
    BatchSize:     1000,
    FlushInterval: 30 * time.Second,
    BufferSize:    10000,
    WorkerCount:   3,
}
```

### 2. Billing Service (`billing_service.go`)

**Purpose**: Core billing operations and usage management

**Key Features**:
- Usage recording and cost calculation
- Billing summary generation
- Payment processing integration
- Usage quota management
- Billing tier support (free, pro, business, enterprise)

**Data Models**:
- `UsageRecord`: Individual usage entries
- `UsageQuota`: Organization usage limits
- `BillingSummary`: Billing period summaries
- `BillingRecord`: Payment transaction records

### 3. Usage Tracker (`usage_tracker.go`)

**Purpose**: Real-time usage tracking and quota enforcement

**Key Features**:
- In-memory quota caching for performance
- Automatic monthly usage resets
- Background synchronization
- Quota violation detection
- Usage metrics and reporting

**Quota Management**:
- Monthly request limits
- Monthly token limits
- Monthly cost limits
- Real-time usage tracking
- Automatic reset on billing periods

### 4. Discount Calculator (`discount_calculator.go`)

**Purpose**: Advanced discount and promotional pricing

**Key Features**:
- Multiple discount types (percentage, fixed, tiered)
- Conditional discounts based on:
  - Billing tiers
  - Usage thresholds
  - Time-based rules
  - Provider/model specific
  - First-time customer discounts
- Volume-based discount tiers
- Priority-based discount application

**Discount Types**:
- **Percentage**: `10%` off total cost
- **Fixed**: `$50` fixed discount
- **Tiered**: Volume-based sliding discounts

### 5. Invoice Generator (`invoice_generator.go`)

**Purpose**: Professional invoice generation and management

**Key Features**:
- HTML invoice generation with professional styling
- Tax calculation based on billing address
- Line item breakdown by provider/model
- Invoice status management (draft, sent, paid, overdue)
- Payment reminder generation
- Late fee calculations

**Invoice Features**:
- Professional branded templates
- Detailed usage breakdowns
- Tax and discount calculations
- Payment terms and due dates
- Metadata tracking

## Integration Points

### With Gateway Service
```go
// Record usage after API request
err := analyticsWorker.RecordRequest(ctx, &gateway.RequestMetrics{
    RequestID:      requestID,
    OrganizationID: orgID,
    ProviderID:     providerID,
    ModelID:        modelID,
    RequestType:    "chat_completion",
    InputTokens:    request.TokenCount,
    OutputTokens:   response.TokenCount,
    ActualCost:     calculatedCost,
})
```

### With Billing System
```go
// Generate monthly bill
summary, err := billingService.CalculateBill(ctx, orgID, "monthly")
if err != nil {
    return err
}

// Create invoice if charges exist
if summary.NetCost > 0 {
    invoice, err := invoiceGenerator.GenerateInvoice(ctx, summary, orgName, address)
    if err != nil {
        return err
    }
    
    // Generate HTML for invoice
    html, err := invoiceGenerator.GenerateInvoiceHTML(ctx, invoice)
}
```

### With Quota Enforcement
```go
// Check quotas before processing request
status, err := billingService.CheckUsageQuotas(ctx, orgID)
if err != nil {
    return err
}

if !status.RequestsOK {
    return errors.New("monthly request limit exceeded")
}
```

## Database Schema Requirements

The system requires tables for:
- `usage_records`: Individual usage entries
- `usage_quotas`: Organization quotas and limits
- `billing_records`: Payment transactions
- `billing_summaries`: Billing period summaries
- `discount_rules`: Promotional discount rules
- `invoices`: Generated invoices
- `invoice_line_items`: Invoice details

## Performance Considerations

### Batching
- Request metrics are batched (default: 1000 records)
- Automatic flushing every 30 seconds
- Immediate flush when buffer is full

### Caching
- Usage quotas cached in memory for 5 minutes
- Background sync every minute
- Automatic cache invalidation

### Background Processing
- Analytics aggregation runs every minute
- Usage quota sync runs every minute
- Graceful shutdown with data persistence

## Monitoring & Health Checks

### Health Endpoints
```go
// Analytics worker health
health := analyticsWorker.GetHealth()
// Returns: running status, buffer sizes, configuration

// Billing service health
health := billingService.GetHealth()
// Returns: service status, component health

// Usage tracker metrics
metrics := usageTracker.GetUsageMetrics(ctx, orgID)
// Returns: detailed quota usage and limits
```

### Metrics Collection
- Request processing rates
- Buffer utilization
- Error rates and retry attempts
- Billing calculation latency
- Quota enforcement statistics

## Configuration

### Environment Variables
```bash
# Analytics Worker
ANALYTICS_BATCH_SIZE=1000
ANALYTICS_FLUSH_INTERVAL=30s
ANALYTICS_BUFFER_SIZE=10000

# Billing Service
BILLING_DEFAULT_CURRENCY=USD
BILLING_GRACE_PERIOD=7d
BILLING_AUTO_BILLING=true

# Usage Tracking
USAGE_CACHE_EXPIRY=5m
USAGE_SYNC_INTERVAL=1m
```

### Deployment Considerations

1. **Database Performance**: Ensure proper indexing on organization_id, timestamps
2. **Memory Usage**: Monitor buffer sizes and cache usage
3. **Background Workers**: Configure appropriate worker counts based on load
4. **Error Handling**: Implement retry logic and dead letter queues
5. **Data Retention**: Configure appropriate TTL for analytics data

## Security & Compliance

- All cost calculations use precise decimal arithmetic
- PII data is properly anonymized in analytics
- Payment data integration follows PCI DSS guidelines
- Audit trails for all billing operations
- Data encryption at rest and in transit

## Testing Strategy

- Unit tests for all calculation logic
- Integration tests with mock databases
- Load testing for high-volume scenarios
- End-to-end billing workflow testing
- Quota enforcement validation

This implementation provides a production-ready analytics and billing system that can scale with the Brokle AI Gateway platform while maintaining accuracy and compliance requirements.