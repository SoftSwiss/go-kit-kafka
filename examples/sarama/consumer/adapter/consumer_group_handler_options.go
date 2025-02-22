package adapter

import (
	"github.com/go-kit/kit/transport"
	"github.com/go-kit/log"
)

type ConsumerGroupHandlerOption func(*ConsumerGroupHandler)

func ConsumerGroupHandlerErrorLogger(logger log.Logger) ConsumerGroupHandlerOption {
	return func(l *ConsumerGroupHandler) {
		l.errorHandler = transport.NewLogErrorHandler(logger)
	}
}

func ConsumerGroupHandlerErrorHandler(errHandler transport.ErrorHandler) ConsumerGroupHandlerOption {
	return func(l *ConsumerGroupHandler) {
		l.errorHandler = errHandler
	}
}
