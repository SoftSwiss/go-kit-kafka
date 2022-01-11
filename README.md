# go-kit-kafka

> Apache Kafka integration module for go-kit

[![build](https://img.shields.io/github/workflow/status/SoftSwiss/go-kit-kafka/CI)](https://github.com/SoftSwiss/go-kit-kafka/actions?query=workflow%3ACI)
[![version](https://img.shields.io/github/go-mod/go-version/SoftSwiss/go-kit-kafka)](https://golang.org/)
[![report](https://goreportcard.com/badge/github.com/SoftSwiss/go-kit-kafka)](https://goreportcard.com/report/github.com/SoftSwiss/go-kit-kafka)
[![coverage](https://img.shields.io/codecov/c/github/SoftSwiss/go-kit-kafka)](https://codecov.io/github/SoftSwiss/go-kit-kafka)
[![tag](https://img.shields.io/github/tag/SoftSwiss/go-kit-kafka.svg)](https://github.com/SoftSwiss/go-kit-kafka/tags)
[![reference](https://pkg.go.dev/badge/github.com/SoftSwiss/go-kit-kafka.svg)](https://pkg.go.dev/github.com/SoftSwiss/go-kit-kafka)

## Getting started

Go modules are supported.

Manual install:

```bash
go get -u github.com/SoftSwiss/go-kit-kafka
```

Golang import:

```go
import "github.com/SoftSwiss/go-kit-kafka/kafka"
```

## Usage

To use consumer/producer transport abstractions converters to the following types from the chosen Apache Kafka
client library should be implemented:

```go
type Message struct {
    Topic     string
    Partition int32
    Offset    int64
    Key       []byte
    Value     []byte
    Headers   []Header
    Timestamp time.Time
}

type Header struct {
    Key   []byte
    Value []byte
}
```

## Examples

Go to [Examples](examples).
