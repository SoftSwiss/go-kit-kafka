package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	kafkatransport "github.com/SoftSwiss/go-kit-kafka/kafka/transport"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/endpoint"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/service"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/transport"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"

	"github.com/SoftSwiss/go-kit-kafka/examples/confluent/consumer/adapter"
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
		brokerAddr := domain.BrokerAddr
		if v, ok := os.LookupEnv("BROKER_ADDR"); ok {
			brokerAddr = v
		}

		c, err := ckafka.NewConsumer(&ckafka.ConfigMap{
			"bootstrap.servers":  brokerAddr,
			"group.id":           domain.GroupID,
			"enable.auto.commit": true,
		})
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka consumer: %w", err))
		}

		defer func() {
			if err := c.Close(); err != nil {
				fatal(logger, fmt.Errorf("failed to close kafka consumer: %w", err))
			}
		}()

		// use a router in case if there are many topics
		router := make(kafkatransport.Router)
		router.AddHandler(domain.Topic, kafkaHandler)

		topics := make([]string, 0)
		for topic := range router {
			topics = append(topics, topic)
		}

		if err := c.SubscribeTopics(topics, nil); err != nil {
			fatal(logger, fmt.Errorf("failed to subscribe to topics: %w", err))
		}

		kafkaListener, err = adapter.NewListener(
			c,
			router,
			adapter.ListenerErrorLogger(
				log.With(logger, "component", "listener"),
			),
		)
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
