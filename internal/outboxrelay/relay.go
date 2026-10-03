// Package outboxrelay runs go-outbox-lib's relay for list-service.
package outboxrelay

import (
	"context"
	"time"

	epnats "github.com/ThatCatDev/ep/v2/drivers/nats"
	"github.com/weeb-vip/go-outbox-lib"

	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/internal/db"
	"github.com/weeb-vip/list-service/internal/logger"
	"github.com/weeb-vip/list-service/metrics"
)

// Run relays until ctx is cancelled. The NATS driver has no stream name, so
// it makes one stream per subject, which is how weeb-argocd's nats-streams
// values expect to find them.
func Run(ctx context.Context, cfg config.Config) error {
	log := logger.FromCtx(ctx)

	driver := epnats.NewNatsDriver(&epnats.Config{URL: cfg.NatsConfig.URL})
	defer func() {
		if err := driver.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing NATS driver")
		}
	}()

	database := db.NewDatabase(cfg.DBConfig)
	relay := outbox.NewRelay(
		database.DB,
		outbox.NewNatsPublisher(driver),
		outbox.Config{
			PollInterval:    time.Duration(cfg.OutboxConfig.PollIntervalMs) * time.Millisecond,
			BatchSize:       cfg.OutboxConfig.BatchSize,
			Retention:       time.Duration(cfg.OutboxConfig.RetentionHours) * time.Hour,
			CleanupInterval: time.Duration(cfg.OutboxConfig.CleanupMinutes) * time.Minute,
			BacklogInterval: time.Duration(cfg.OutboxConfig.BacklogSeconds) * time.Second,
		},
		outbox.WithLogger(log),
		outbox.WithMetrics(relayMetrics{}),
	)

	log.Info().Str("nats_url", cfg.NatsConfig.URL).Msg("Starting outbox relay")

	return relay.Run(ctx)
}

type relayMetrics struct{}

func (relayMetrics) Published(subject string) {
	_ = metrics.NewMetricsInstance().CountMetric("outbox_published_total", map[string]string{
		"service": metrics.GetServiceName(), "subject": subject, "result": "success", "env": metrics.GetCurrentEnv(),
	})
}

func (relayMetrics) PublishFailed(subject string) {
	_ = metrics.NewMetricsInstance().CountMetric("outbox_published_total", map[string]string{
		"service": metrics.GetServiceName(), "subject": subject, "result": "error", "env": metrics.GetCurrentEnv(),
	})
}

func (relayMetrics) Backlog(unpublished int64) {
	_ = metrics.NewMetricsInstance().GaugeMetric("outbox_unpublished", float64(unpublished), map[string]string{
		"service": metrics.GetServiceName(), "env": metrics.GetCurrentEnv(),
	})
}
