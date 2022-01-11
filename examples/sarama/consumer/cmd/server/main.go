package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	kafkatransport "github.com/SoftSwiss/go-kit-kafka/kafka/transport"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/endpoint"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/service"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/transport"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"

	"github.com/SoftSwiss/go-kit-kafka/examples/sarama/consumer/adapter"
)

func fatal(logger log.Logger, err error) {
	_ = level.Error(logger).Log("err", err)
	os.Exit(1)
}

func main() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	{
		ctx = context.Background()
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
	}

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = level.NewInjector(logger, level.InfoValue())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	_ = logger.Log("msg", "initializing services")

	var svc consumer.Service
	{
		storageSvc, err := service.NewStorageService(
			log.With(logger, "component", "storage_service"),
		)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init storage: %w", err))
		}
		svc = storageSvc
	}

	var endpoints endpoint.Endpoints
	{
		endpoints = endpoint.Endpoints{
			CreateEventEndpoint: endpoint.MakeCreateEventEndpoint(svc),
			ListEventsEndpoint:  endpoint.MakeListEventsEndpoint(svc),
		}
	}

	_ = logger.Log("msg", "initializing kafka handlers")

	kafkaHandler := transport.NewKafkaHandler(endpoints)

	_ = logger.Log("msg", "initializing kafka consumer")

	var kafkaListener *adapter.Listener
	{
		cfg := sarama.NewConfig()
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
		cfg.Consumer.Offsets.AutoCommit.Enable = true

		brokerAddr := domain.BrokerAddr
		if v, ok := os.LookupEnv("BROKER_ADDR"); ok {
			brokerAddr = v
		}

		client, err := sarama.NewClient(
			[]string{brokerAddr},
			cfg,
		)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka client: %w", err))
		}

		consumerGroup, err := sarama.NewConsumerGroupFromClient(domain.GroupID, client)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka consumer group: %w", err))
		}

		defer func() {
			if err := consumerGroup.Close(); err != nil {
				fatal(logger, fmt.Errorf("failed to close kafka consumer group: %w", err))
			}
		}()

		// use a router in case if there are many topics
		router := make(kafkatransport.Router)
		router.AddHandler(domain.Topic, kafkaHandler)

		topics := make([]string, 0)
		for topic := range router {
			topics = append(topics, topic)
		}

		consumerGroupHandler, err := adapter.NewConsumerGroupHandler(router)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka consumer group handler: %w", err))
		}

		kafkaListener, err = adapter.NewListener(topics, consumerGroup, consumerGroupHandler)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka listener: %w", err))
		}
	}

	_ = logger.Log("msg", "initializing http handler")

	httpHandler := transport.NewHTTPHandler(endpoints)

	errc := make(chan error, 1)

	go func() {
		if err := kafkaListener.Listen(ctx); err != nil {
			errc <- err
		}
	}()

	go func() {
		if err := http.ListenAndServe(":8081", httpHandler); err != nil {
			errc <- err
		}
	}()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-sigc)
	}()

	_ = logger.Log("msg", "application started")
	_ = logger.Log("msg", "application stopped", "exit", <-errc)
}
