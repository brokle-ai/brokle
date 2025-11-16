package analytics

import (
	"math"
	"time"
)

// CalculatorConfig represents calculator configuration
type CalculatorConfig struct {
	DefaultPeriod       time.Duration `json:"default_period"`
	DecimalPrecision    int           `json:"decimal_precision"`
	EnableNormalization bool          `json:"enable_normalization"`
}

// Calculator provides advanced analytics calculations
type Calculator struct {
	config *CalculatorConfig
}

// NewCalculator creates a new analytics calculator
func NewCalculator(config *CalculatorConfig) *Calculator {
	if config == nil {
		config = &CalculatorConfig{
			DefaultPeriod:       24 * time.Hour,
			DecimalPrecision:    2,
			EnableNormalization: true,
		}
	}

	return &Calculator{
		config: config,
	}
}

// AI Platform Specific Calculations

// CalculateAICostEfficiency calculates cost efficiency metrics for AI requests
func (c *Calculator) CalculateAICostEfficiency(requests []AIRequest) *CostEfficiencyMetrics {
	if len(requests) == 0 {
		return &CostEfficiencyMetrics{}
	}

	var totalCost, totalTokens float64
	var totalRequests int64
	providerCosts := make(map[string]float64)
	modelCosts := make(map[string]float64)

	for _, req := range requests {
		totalCost += req.Cost
		totalTokens += float64(req.TokensUsed)
		totalRequests++

		providerCosts[req.Provider] += req.Cost
		modelCosts[req.Model] += req.Cost
	}

	return &CostEfficiencyMetrics{
		TotalCost:         c.roundToDecimal(totalCost),
		TotalTokens:       int64(totalTokens),
		TotalRequests:     totalRequests,
		AvgCostPerRequest: c.roundToDecimal(totalCost / float64(totalRequests)),
		AvgCostPerToken:   c.roundToDecimal(totalCost / totalTokens),
		TokensPerRequest:  c.roundToDecimal(totalTokens / float64(totalRequests)),
		CostByProvider:    c.roundMapValues(providerCosts),
		CostByModel:       c.roundMapValues(modelCosts),
		EfficiencyScore:   c.calculateEfficiencyScore(requests),
	}
}

// CalculateProviderPerformance calculates performance metrics for AI providers
func (c *Calculator) CalculateProviderPerformance(requests []AIRequest) map[string]*ProviderMetrics {
	providerData := make(map[string][]AIRequest)

	// Group by provider
	for _, req := range requests {
		providerData[req.Provider] = append(providerData[req.Provider], req)
	}

	result := make(map[string]*ProviderMetrics)

	for provider, providerRequests := range providerData {
		metrics := &ProviderMetrics{
			Provider:      provider,
			TotalRequests: int64(len(providerRequests)),
		}

		var totalLatency, totalCost, totalQuality float64
		var successCount, errorCount int64
		latencies := make([]float64, 0, len(providerRequests))

		for _, req := range providerRequests {
			totalLatency += req.Latency
			totalCost += req.Cost
			totalQuality += req.Quality
			latencies = append(latencies, req.Latency)

			if req.Success {
				successCount++
			} else {
				errorCount++
			}
		}

		metrics.AvgLatency = c.roundToDecimal(totalLatency / float64(len(providerRequests)))
		metrics.AvgCost = c.roundToDecimal(totalCost / float64(len(providerRequests)))
		metrics.AvgQuality = c.roundToDecimal(totalQuality / float64(len(providerRequests)))
		metrics.SuccessRate = c.roundToDecimal(float64(successCount) / float64(len(providerRequests)) * 100)
		metrics.ErrorRate = c.roundToDecimal(float64(errorCount) / float64(len(providerRequests)) * 100)
		metrics.P95Latency = c.percentile(latencies, 0.95)
		metrics.TotalCost = c.roundToDecimal(totalCost)

		result[provider] = metrics
	}

	return result
}

// CalculateUsageMetrics calculates usage metrics over time
func (c *Calculator) CalculateUsageMetrics(requests []AIRequest, timeWindow TimeWindow) *UsageMetrics {
	if len(requests) == 0 {
		return &UsageMetrics{
			TimeWindow: timeWindow,
		}
	}

	// Filter requests within time window
	filteredRequests := make([]AIRequest, 0)
	for _, req := range requests {
		if req.Timestamp.After(timeWindow.Start) && req.Timestamp.Before(timeWindow.End) {
			filteredRequests = append(filteredRequests, req)
		}
	}

	if len(filteredRequests) == 0 {
		return &UsageMetrics{
			TimeWindow: timeWindow,
		}
	}

	metrics := &UsageMetrics{
		TimeWindow:    timeWindow,
		TotalRequests: int64(len(filteredRequests)),
	}

	var totalTokens int64
	var totalCost float64
	hourlyUsage := make(map[int]int64)
	dailyUsage := make(map[string]int64)

	for _, req := range filteredRequests {
		totalTokens += int64(req.TokensUsed)
		totalCost += req.Cost

		// Hourly usage
		hour := req.Timestamp.Hour()
		hourlyUsage[hour]++

		// Daily usage
		day := req.Timestamp.Format("2006-01-02")
		dailyUsage[day]++
	}

	metrics.TotalTokens = totalTokens
	metrics.TotalCost = c.roundToDecimal(totalCost)
	metrics.AvgTokensPerRequest = c.roundToDecimal(float64(totalTokens) / float64(len(filteredRequests)))
	metrics.AvgCostPerRequest = c.roundToDecimal(totalCost / float64(len(filteredRequests)))

	// Calculate usage rate (requests per hour)
	duration := timeWindow.End.Sub(timeWindow.Start)
	hoursInWindow := duration.Hours()
	if hoursInWindow > 0 {
		metrics.RequestsPerHour = c.roundToDecimal(float64(len(filteredRequests)) / hoursInWindow)
	}

	metrics.PeakHour = c.findPeakUsageHour(hourlyUsage)
	metrics.PeakDay = c.findPeakUsageDay(dailyUsage)

	return metrics
}

// CalculateCacheEffectiveness calculates cache performance metrics
func (c *Calculator) CalculateCacheEffectiveness(cacheEvents []CacheEvent) *CacheMetrics {
	if len(cacheEvents) == 0 {
		return &CacheMetrics{}
	}

	var hits, misses int64
	var totalSavedCost, totalSavedTime float64

	for _, event := range cacheEvents {
		if event.Hit {
			hits++
			totalSavedCost += event.SavedCost
			totalSavedTime += event.SavedTime
		} else {
			misses++
		}
	}

	total := hits + misses
	hitRate := float64(hits) / float64(total) * 100

	return &CacheMetrics{
		TotalRequests:  total,
		CacheHits:      hits,
		CacheMisses:    misses,
		HitRate:        c.roundToDecimal(hitRate),
		MissRate:       c.roundToDecimal(100 - hitRate),
		TotalCostSaved: c.roundToDecimal(totalSavedCost),
		TotalTimeSaved: c.roundToDecimal(totalSavedTime),
		AvgCostSaved:   c.roundToDecimal(totalSavedCost / float64(hits)),
		AvgTimeSaved:   c.roundToDecimal(totalSavedTime / float64(hits)),
	}
}

// CalculateQualityScores calculates AI response quality metrics
func (c *Calculator) CalculateQualityScores(requests []AIRequest) *QualityMetrics {
	if len(requests) == 0 {
		return &QualityMetrics{}
	}

	qualities := make([]float64, 0, len(requests))
	providerQualities := make(map[string][]float64)
	modelQualities := make(map[string][]float64)

	for _, req := range requests {
		qualities = append(qualities, req.Quality)
		providerQualities[req.Provider] = append(providerQualities[req.Provider], req.Quality)
		modelQualities[req.Model] = append(modelQualities[req.Model], req.Quality)
	}

	return &QualityMetrics{
		OverallAverage:    c.roundToDecimal(c.average(qualities)),
		OverallMedian:     c.roundToDecimal(c.percentile(qualities, 0.5)),
		StandardDeviation: c.roundToDecimal(c.standardDeviation(qualities)),
		MinScore:          c.roundToDecimal(c.min(qualities)),
		MaxScore:          c.roundToDecimal(c.max(qualities)),
		P25Score:          c.roundToDecimal(c.percentile(qualities, 0.25)),
		P75Score:          c.roundToDecimal(c.percentile(qualities, 0.75)),
		P90Score:          c.roundToDecimal(c.percentile(qualities, 0.90)),
		P95Score:          c.roundToDecimal(c.percentile(qualities, 0.95)),
		QualityByProvider: c.calculateProviderQuality(providerQualities),
		QualityByModel:    c.calculateModelQuality(modelQualities),
	}
}

// CalculateRoutingEfficiency calculates routing decision effectiveness
func (c *Calculator) CalculateRoutingEfficiency(decisions []RoutingDecision) *RoutingMetrics {
	if len(decisions) == 0 {
		return &RoutingMetrics{}
	}

	var totalLatency, totalCost, totalAccuracy float64
	routingReasons := make(map[string]int64)
	providerSelections := make(map[string]int64)

	for _, decision := range decisions {
		totalLatency += decision.DecisionLatency
		totalCost += decision.RoutedRequestCost
		totalAccuracy += decision.AccuracyScore

		routingReasons[decision.Reason]++
		providerSelections[decision.SelectedProvider]++
	}

	total := float64(len(decisions))

	return &RoutingMetrics{
		TotalDecisions:     int64(len(decisions)),
		AvgDecisionLatency: c.roundToDecimal(totalLatency / total),
		AvgRoutedCost:      c.roundToDecimal(totalCost / total),
		AvgAccuracy:        c.roundToDecimal(totalAccuracy / total),
		RoutingReasons:     routingReasons,
		ProviderSelections: providerSelections,
		EfficiencyScore:    c.calculateRoutingEfficiency(decisions),
	}
}

// Business Intelligence Calculations

// CalculateGrowthMetrics calculates growth metrics over periods
func (c *Calculator) CalculateGrowthMetrics(currentPeriod, previousPeriod []AIRequest) *GrowthMetrics {
	currentMetrics := c.calculatePeriodMetrics(currentPeriod)
	previousMetrics := c.calculatePeriodMetrics(previousPeriod)

	return &GrowthMetrics{
		RequestsGrowth:     c.calculateGrowthRate(float64(previousMetrics.TotalRequests), float64(currentMetrics.TotalRequests)),
		TokensGrowth:       c.calculateGrowthRate(float64(previousMetrics.TotalTokens), float64(currentMetrics.TotalTokens)),
		CostGrowth:         c.calculateGrowthRate(previousMetrics.TotalCost, currentMetrics.TotalCost),
		QualityImprovement: c.calculateGrowthRate(previousMetrics.AvgQuality, currentMetrics.AvgQuality),
		LatencyImprovement: c.calculateGrowthRate(currentMetrics.AvgLatency, previousMetrics.AvgLatency), // Inverted for improvement
		CurrentPeriod:      currentMetrics,
		PreviousPeriod:     previousMetrics,
	}
}

// CalculateAnomalies detects anomalies in metrics
func (c *Calculator) CalculateAnomalies(values []float64, threshold float64) []Anomaly {
	if len(values) < 3 {
		return nil
	}

	mean := c.average(values)
	stddev := c.standardDeviation(values)
	upperBound := mean + threshold*stddev
	lowerBound := mean - threshold*stddev

	var anomalies []Anomaly
	for i, value := range values {
		if value > upperBound || value < lowerBound {
			severity := "medium"
			if math.Abs(value-mean) > 3*stddev {
				severity = "high"
			} else if math.Abs(value-mean) > 2*stddev {
				severity = "medium"
			} else {
				severity = "low"
			}

			anomalies = append(anomalies, Anomaly{
				Index:     i,
				Value:     value,
				Expected:  mean,
				Deviation: math.Abs(value - mean),
				Severity:  severity,
				ZScore:    (value - mean) / stddev,
			})
		}
	}

	return anomalies
}

// CalculateTrends calculates trends in time series data
func (c *Calculator) CalculateTrends(timeSeries TimeSeries) *TrendAnalysis {
	if len(timeSeries.DataPoints) < 2 {
		return &TrendAnalysis{
			Direction: "insufficient_data",
		}
	}

	values := make([]float64, len(timeSeries.DataPoints))
	for i, point := range timeSeries.DataPoints {
		values[i] = point.Value
	}

	slope, intercept := c.linearRegression(values)
	direction := "stable"

	if slope > 0.1 {
		direction = "increasing"
	} else if slope < -0.1 {
		direction = "decreasing"
	}

	correlation := c.calculateCorrelation(values)

	return &TrendAnalysis{
		Direction:   direction,
		Slope:       c.roundToDecimal(slope),
		Intercept:   c.roundToDecimal(intercept),
		Correlation: c.roundToDecimal(correlation),
		Strength:    c.getTrendStrength(correlation),
		DataPoints:  len(values),
	}
}

// Helper calculation methods

func (c *Calculator) calculateEfficiencyScore(requests []AIRequest) float64 {
	if len(requests) == 0 {
		return 0
	}

	// Calculate efficiency based on cost, latency, and quality
	var totalScore float64
	for _, req := range requests {
		// Normalize factors (lower cost and latency is better, higher quality is better)
		costScore := math.Max(0, 100-req.Cost*10)       // Assumes cost is in reasonable range
		latencyScore := math.Max(0, 100-req.Latency/10) // Assumes latency is in ms
		qualityScore := req.Quality

		// Weighted average (quality weighted more)
		score := (costScore*0.3 + latencyScore*0.2 + qualityScore*0.5)
		totalScore += score
	}

	return c.roundToDecimal(totalScore / float64(len(requests)))
}

func (c *Calculator) calculateRoutingEfficiency(decisions []RoutingDecision) float64 {
	if len(decisions) == 0 {
		return 0
	}

	var totalScore float64
	for _, decision := range decisions {
		// Consider decision speed, accuracy, and cost optimization
		speedScore := math.Max(0, 100-decision.DecisionLatency) // Lower latency is better
		accuracyScore := decision.AccuracyScore
		costScore := math.Max(0, 100-decision.RoutedRequestCost*10) // Lower cost is better

		score := (speedScore*0.2 + accuracyScore*0.5 + costScore*0.3)
		totalScore += score
	}

	return c.roundToDecimal(totalScore / float64(len(decisions)))
}

func (c *Calculator) calculatePeriodMetrics(requests []AIRequest) PeriodMetrics {
	if len(requests) == 0 {
		return PeriodMetrics{}
	}

	var totalTokens int64
	var totalCost, totalLatency, totalQuality float64

	for _, req := range requests {
		totalTokens += int64(req.TokensUsed)
		totalCost += req.Cost
		totalLatency += req.Latency
		totalQuality += req.Quality
	}

	count := float64(len(requests))

	return PeriodMetrics{
		TotalRequests: int64(len(requests)),
		TotalTokens:   totalTokens,
		TotalCost:     c.roundToDecimal(totalCost),
		AvgLatency:    c.roundToDecimal(totalLatency / count),
		AvgQuality:    c.roundToDecimal(totalQuality / count),
	}
}

func (c *Calculator) calculateGrowthRate(previous, current float64) float64 {
	if previous == 0 {
		return 0
	}
	return c.roundToDecimal(((current - previous) / previous) * 100)
}

func (c *Calculator) calculateProviderQuality(providerQualities map[string][]float64) map[string]float64 {
	result := make(map[string]float64)
	for provider, qualities := range providerQualities {
		result[provider] = c.roundToDecimal(c.average(qualities))
	}
	return result
}

func (c *Calculator) calculateModelQuality(modelQualities map[string][]float64) map[string]float64 {
	result := make(map[string]float64)
	for model, qualities := range modelQualities {
		result[model] = c.roundToDecimal(c.average(qualities))
	}
	return result
}

func (c *Calculator) findPeakUsageHour(hourlyUsage map[int]int64) int {
	var maxUsage int64
	var peakHour int

	for hour, usage := range hourlyUsage {
		if usage > maxUsage {
			maxUsage = usage
			peakHour = hour
		}
	}

	return peakHour
}

func (c *Calculator) findPeakUsageDay(dailyUsage map[string]int64) string {
	var maxUsage int64
	var peakDay string

	for day, usage := range dailyUsage {
		if usage > maxUsage {
			maxUsage = usage
			peakDay = day
		}
	}

	return peakDay
}

// Mathematical helper functions

func (c *Calculator) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (c *Calculator) min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func (c *Calculator) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (c *Calculator) percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Create a copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple sort implementation
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (c *Calculator) standardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	mean := c.average(values)
	sumSquaredDiff := 0.0

	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values)-1)
	return math.Sqrt(variance)
}

func (c *Calculator) linearRegression(values []float64) (slope, intercept float64) {
	n := float64(len(values))
	if n < 2 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0, sumY / n
	}

	slope = (n*sumXY - sumX*sumY) / denominator
	intercept = (sumY - slope*sumX) / n

	return slope, intercept
}

func (c *Calculator) calculateCorrelation(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	// Create x values (indices)
	x := make([]float64, len(values))
	for i := range x {
		x[i] = float64(i)
	}

	meanX := c.average(x)
	meanY := c.average(values)

	var numerator, denomX, denomY float64

	for i := range values {
		diffX := x[i] - meanX
		diffY := values[i] - meanY

		numerator += diffX * diffY
		denomX += diffX * diffX
		denomY += diffY * diffY
	}

	if denomX == 0 || denomY == 0 {
		return 0
	}

	return numerator / math.Sqrt(denomX*denomY)
}

func (c *Calculator) getTrendStrength(correlation float64) string {
	absCorr := math.Abs(correlation)
	if absCorr >= 0.8 {
		return "strong"
	} else if absCorr >= 0.5 {
		return "moderate"
	} else if absCorr >= 0.3 {
		return "weak"
	}
	return "negligible"
}

func (c *Calculator) roundToDecimal(value float64) float64 {
	multiplier := math.Pow(10, float64(c.config.DecimalPrecision))
	return math.Round(value*multiplier) / multiplier
}

func (c *Calculator) roundMapValues(m map[string]float64) map[string]float64 {
	result := make(map[string]float64)
	for k, v := range m {
		result[k] = c.roundToDecimal(v)
	}
	return result
}
