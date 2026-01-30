package evaluation

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"brokle/internal/core/domain/evaluation"
)

// RegexScorer implements pattern matching scoring
type RegexScorer struct {
	logger *slog.Logger
}

// NewRegexScorer creates a new regex scorer
func NewRegexScorer(logger *slog.Logger) *RegexScorer {
	return &RegexScorer{
		logger: logger,
	}
}

func (s *RegexScorer) Type() evaluation.ScorerType {
	return evaluation.ScorerTypeRegex
}

func (s *RegexScorer) Execute(ctx context.Context, job *EvaluationJob) (*ScorerResult, error) {
	config, err := s.parseConfig(job.ScorerConfig)
	if err != nil {
		return nil, fmt.Errorf("invalid regex scorer config: %w", err)
	}

	text := s.getTargetText(job)
	if text == "" {
		s.logger.Debug("No text to match against",
			"job_id", job.JobID,
		)
		return &ScorerResult{Scores: []ScoreOutput{}}, nil
	}

	// ReDoS protection: validate pattern before compilation
	const maxPatternLength = 200
	const maxWildcards = 10

	if len(config.Pattern) > maxPatternLength {
		return nil, fmt.Errorf("regex pattern too long (max %d characters)", maxPatternLength)
	}
	if strings.Count(config.Pattern, "*") > maxWildcards {
		return nil, fmt.Errorf("regex pattern too complex (max %d wildcards)", maxWildcards)
	}

	// Compile regex
	re, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var value float64
	var reason string
	var stringValue *string

	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		value = config.MatchScore
		reason = fmt.Sprintf("Pattern matched: %s", config.Pattern)

		// If capture group specified, extract value
		if config.CaptureGroup != nil && *config.CaptureGroup < len(matches) {
			captured := matches[*config.CaptureGroup]
			stringValue = &captured

			// Try to parse as number if possible
			if numVal, err := strconv.ParseFloat(captured, 64); err == nil {
				value = numVal
			}
		}
	} else {
		value = config.NoMatchScore
		reason = fmt.Sprintf("Pattern did not match: %s", config.Pattern)
	}

	result := &ScorerResult{
		Scores: []ScoreOutput{
			{
				Name:        config.ScoreName,
				Value:       &value,
				StringValue: stringValue,
				Type:        "NUMERIC",
				Reason:      &reason,
			},
		},
	}

	return result, nil
}

func (s *RegexScorer) parseConfig(config map[string]any) (*evaluation.RegexScorerConfig, error) {
	pattern, ok := config["pattern"].(string)
	if !ok || pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	scoreName, ok := config["score_name"].(string)
	if !ok || scoreName == "" {
		scoreName = "regex_match"
	}

	matchScore := 1.0
	if v, ok := config["match_score"].(float64); ok {
		matchScore = v
	}

	noMatchScore := 0.0
	if v, ok := config["no_match_score"].(float64); ok {
		noMatchScore = v
	}

	var captureGroup *int
	if v, ok := config["capture_group"].(float64); ok {
		cg := int(v)
		captureGroup = &cg
	}

	return &evaluation.RegexScorerConfig{
		Pattern:      pattern,
		ScoreName:    scoreName,
		MatchScore:   matchScore,
		NoMatchScore: noMatchScore,
		CaptureGroup: captureGroup,
	}, nil
}

func (s *RegexScorer) getTargetText(job *EvaluationJob) string {
	// Check variables first
	if output, ok := job.Variables["output"]; ok && output != "" {
		return output
	}
	if input, ok := job.Variables["input"]; ok && input != "" {
		return input
	}

	// Fall back to span data
	if output, ok := job.SpanData["output"].(string); ok {
		return output
	}
	if input, ok := job.SpanData["input"].(string); ok {
		return input
	}

	return ""
}
