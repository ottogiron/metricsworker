# Metrics Workers

A set of workers processing metrics from RabbitMQ.

[![Build Status](https://travis-ci.org/ottogiron/metricsworker.svg?branch=master)](https://travis-ci.org/ottogiron/metricsworker)
[![GoDoc](https://godoc.org/github.com/ottogiron/metricsworker?status.svg)](https://godoc.org/github.com/ottogiron/metricsworker)
[![Go Report Card](https://goreportcard.com/badge/github.com/ottogiron/metricsworker)](https://goreportcard.com/report/github.com/ottogiron/metricsworker)



## Install 

Download and install from the [releases](http://github.com/ottogiron/metricsworker/releases) page


## Dependencies

* RabbitMQ
* Redis
* MongoDB
* PostgreSQL

## Running 

```bash

mworker --concurrency=1 \
    --wait-timeout=600 \
    --rabbit-uri=amqp://guest:guest@localhost:5672 \
    --rabbit-queue_name=hello \
    --rabbit-consumer_auto_ack=true \
    --rabbit-exchange="test-exchange" \
    --rabbit-routing_key="text-key"
```