package units

// Billing unit sizes for usage-based pricing.
const (
	// SpansPer100K is the billing unit for spans (price per 100,000 spans).
	SpansPer100K int64 = 100_000

	// ScoresPer1K is the billing unit for scores (price per 1,000 scores).
	ScoresPer1K int64 = 1_000
)
