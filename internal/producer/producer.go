package producer

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/list-service/config"
	"github.com/weeb-vip/list-service/internal/logger"
)

type Producer[T any] interface {
	Send(ctx context.Context, data []byte) error
}

type ProducerImpl[T any] struct {
	client pulsar.Client
	config config.PulsarConfig
}

func NewProducer[T any](ctx context.Context, cfg config.PulsarConfig) Producer[T] {
	log := logger.FromCtx(ctx)
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.URL,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Error creating pulsar client")
		return nil
	}

	return &ProducerImpl[T]{
		config: cfg,
		client: client,
	}
}

func (p *ProducerImpl[T]) Send(ctx context.Context, data []byte) error {
	log := logger.FromCtx(ctx)
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic: p.config.ProducerTopic,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Error creating pulsar producer")
		return err
	}

	defer producer.Close()

	msg := pulsar.ProducerMessage{
		Payload: data,
	}

	_, err = producer.Send(ctx, &msg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error sending message")
		return err
	}

	return nil
}
