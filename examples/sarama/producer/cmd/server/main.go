package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	kitendpoint "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/producer"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/producer/endpoint"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/producer/service"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/producer/transport"

	"github.com/SoftSwiss/go-kit-kafka/examples/sarama/producer/adapter"
)

func fatal(logger log.Logger, err error) {
	_ = logger.Log("err", err)
	os.Exit(1)
}

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = level.NewInjector(logger, level.InfoValue())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	_ = logger.Log("msg", "initializing services")

	var svc producer.Service
	{
		var err error
		svc, err = service.NewGeneratorService(logger)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to create generator: %w", err))
		}
	}

	_ = logger.Log("msg", "initializing kafka producer")

	var producerMiddleware kitendpoint.Middleware
	{
		cfg := sarama.NewConfig()
		cfg.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
		cfg.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
		cfg.Producer.Return.Successes = true

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

		p, err := sarama.NewSyncProducerFromClient(client)
		if err != nil {
			fatal(logger, fmt.Errorf("failed to init kafka producer: %w", err))
		}

		defer func() {
			if err := p.Close(); err != nil {
				fatal(logger, fmt.Errorf("failed to close kafka producer: %w", err))
			}
		}()

		e := transport.NewKafkaProducer(
			adapter.NewProducer(p),
			domain.Topic,
		).Endpoint()

		producerMiddleware = endpoint.ProducerMiddleware(e)
	}

	var endpoints endpoint.Endpoints
	{
		endpoints = endpoint.Endpoints{
			GenerateEvent: producerMiddleware(
				endpoint.MakeGenerateEventEndpoint(svc),
			),
		}
	}

	_ = logger.Log("msg", "initializing http handler")

	var httpHandler http.Handler
	{
		httpHandler = transport.NewHTTPHandler(endpoints)
	}

	errc := make(chan error, 1)

	go func() {
		if err := http.ListenAndServe(":8080", httpHandler); err != nil {
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
