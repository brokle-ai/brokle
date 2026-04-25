package app

import (
	"context"
	"fmt"
	"log/slog"

	"brokle/internal/config"
	"brokle/pkg/email"
)

// HealthCheck reports the health status of every dependency that
// the active deployment mode opens. The map keys ("postgres",
// "redis", "clickhouse", "telemetry_stream_consumer", "evaluator_worker",
// "evaluation_worker", "mode") are the canonical wire shape consumed
// by /readyz and the ops dashboard.
//
// Worker entries are evaluated for "healthy" using a sampling rule:
// no activity = healthy (newly started); >0 successful processing
// with <10% error rate = healthy; otherwise degraded/unhealthy.
func (pc *ProviderContainer) HealthCheck() map[string]string {
	health := make(map[string]string)

	if pc.Core != nil && pc.Core.Databases != nil {
		if pc.Core.Databases.Pool != nil {
			if err := pc.Core.Databases.Pool.Ping(context.Background()); err != nil {
				health["postgres"] = "unhealthy: " + err.Error()
			} else {
				health["postgres"] = "healthy"
			}
		}

		if pc.Core.Databases.Redis != nil {
			if err := pc.Core.Databases.Redis.Health(); err != nil {
				health["redis"] = "unhealthy: " + err.Error()
			} else {
				health["redis"] = "healthy"
			}
		}

		if pc.Core.Databases.ClickHouse != nil {
			if err := pc.Core.Databases.ClickHouse.Health(); err != nil {
				health["clickhouse"] = "unhealthy: " + err.Error()
			} else {
				health["clickhouse"] = "healthy"
			}
		}
	}

	if pc.Workers != nil && pc.Workers.TelemetryConsumer != nil {
		stats := pc.Workers.TelemetryConsumer.GetStats()
		batchesProcessed := stats["batches_processed"]
		errorsCount := stats["errors_count"]

		switch {
		case batchesProcessed == 0 && errorsCount == 0:
			health["telemetry_stream_consumer"] = "healthy (no activity yet)"
		case batchesProcessed > 0:
			errorRate := float64(errorsCount) / float64(batchesProcessed)
			if errorRate < 0.10 {
				health["telemetry_stream_consumer"] = fmt.Sprintf("healthy (processed: %d, errors: %d, streams: %d)",
					batchesProcessed, errorsCount, stats["active_streams"])
			} else {
				health["telemetry_stream_consumer"] = fmt.Sprintf("degraded (high error rate: %.1f%%)", errorRate*100)
			}
		default:
			health["telemetry_stream_consumer"] = fmt.Sprintf("unhealthy (errors: %d, no successful processing)", errorsCount)
		}
	}

	if pc.Workers != nil && pc.Workers.EvaluatorWorker != nil {
		stats := pc.Workers.EvaluatorWorker.GetStats()
		spansProcessed := stats["spans_processed"]
		errorsCount := stats["errors_count"]

		switch {
		case spansProcessed == 0 && errorsCount == 0:
			health["evaluator_worker"] = "healthy (no activity yet)"
		case spansProcessed > 0:
			health["evaluator_worker"] = fmt.Sprintf("healthy (spans_processed: %d, jobs_emitted: %d, errors: %d)",
				spansProcessed, stats["jobs_emitted"], errorsCount)
		default:
			health["evaluator_worker"] = fmt.Sprintf("unhealthy (errors: %d)", errorsCount)
		}
	}

	if pc.Workers != nil && pc.Workers.EvaluationWorker != nil {
		stats := pc.Workers.EvaluationWorker.GetStats()
		jobsProcessed := stats["jobs_processed"]
		errorsCount := stats["errors_count"]

		switch {
		case jobsProcessed == 0 && errorsCount == 0:
			health["evaluation_worker"] = "healthy (no activity yet)"
		case jobsProcessed > 0:
			health["evaluation_worker"] = fmt.Sprintf("healthy (processed: %d, scores: %d, llm: %d, builtin: %d, regex: %d)",
				jobsProcessed, stats["scores_created"], stats["llm_calls"], stats["builtin_calls"], stats["regex_calls"])
		default:
			health["evaluation_worker"] = fmt.Sprintf("unhealthy (errors: %d)", errorsCount)
		}
	}

	health["mode"] = string(pc.Mode)

	return health
}

// Shutdown stops every long-lived resource the provider container
// owns. Errors during shutdown are logged and the last one returned;
// we do NOT bail early — every resource gets its chance to close so
// stuck pools don't keep the process alive.
func (pc *ProviderContainer) Shutdown() error {
	var lastErr error
	logger := pc.Core.Logger

	if pc.Workers != nil {
		if pc.Workers.TelemetryConsumer != nil {
			logger.Info("Stopping telemetry stream consumer...")
			pc.Workers.TelemetryConsumer.Stop()
			logger.Info("Telemetry stream consumer stopped")
		}

		if pc.Workers.EvaluatorWorker != nil {
			logger.Info("Stopping evaluator worker...")
			pc.Workers.EvaluatorWorker.Stop()
			logger.Info("Evaluator worker stopped")
		}

		if pc.Workers.EvaluationWorker != nil {
			logger.Info("Stopping evaluation worker...")
			pc.Workers.EvaluationWorker.Stop()
			logger.Info("Evaluation worker stopped")
		}
	}

	if pc.Core != nil && pc.Core.Databases != nil {
		if pc.Core.Databases.Pool != nil {
			logger.Info("Closing pgx pool...")
			pc.Core.Databases.Pool.Close()
		}

		if pc.Core.Databases.Redis != nil {
			if err := pc.Core.Databases.Redis.Close(); err != nil {
				logger.Error("Failed to close Redis connection", "error", err)
				lastErr = err
			}
		}

		if pc.Core.Databases.ClickHouse != nil {
			if err := pc.Core.Databases.ClickHouse.Close(); err != nil {
				logger.Error("Failed to close ClickHouse connection", "error", err)
				lastErr = err
			}
		}
	}

	return lastErr
}

// createEmailSender resolves the configured email transport.
// Returns NoOpEmailSender when no provider is configured so callers
// don't need to nil-check the sender at every send site.
func createEmailSender(cfg *config.EmailConfig, logger *slog.Logger) (email.EmailSender, error) {
	if cfg.Provider == "" {
		logger.Warn("Email sender not configured, invitations will not be sent via email")
		return &email.NoOpEmailSender{}, nil
	}

	logger.Info("Initializing email sender", "provider", cfg.Provider)

	switch cfg.Provider {
	case "resend":
		return email.NewResendClient(email.ResendConfig{
			APIKey:    cfg.ResendAPIKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		}), nil

	case "smtp":
		return email.NewSMTPClient(email.SMTPConfig{
			Host:      cfg.SMTPHost,
			Port:      cfg.SMTPPort,
			Username:  cfg.SMTPUsername,
			Password:  cfg.SMTPPassword,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
			UseTLS:    cfg.SMTPUseTLS,
		}), nil

	case "ses":
		client, err := email.NewSESClient(email.SESConfig{
			Region:    cfg.SESRegion,
			AccessKey: cfg.SESAccessKey,
			SecretKey: cfg.SESSecretKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create SES client: %w", err)
		}
		return client, nil

	case "sendgrid":
		return email.NewSendGridClient(email.SendGridConfig{
			APIKey:    cfg.SendGridAPIKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		}), nil

	default:
		return nil, fmt.Errorf("unknown email provider: %s", cfg.Provider)
	}
}
