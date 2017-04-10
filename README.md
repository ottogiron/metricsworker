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


For testing purposes you can run the docker compose development file [docker-compose.dev.yml](docker-compose.dev.yml)

```bash
docker-compose -f docker-compose.dev.yml up -d
```

##  Example

```bash

mworker --concurrency=1 \
    --wait-timeout=600 \
    --rabbit-uri=amqp://guest:guest@localhost:5672 \
    --rabbit-queue_name=hello \
    --rabbit-consumer_auto_ack=true \
    --rabbit-exchange="test-exchange" \
    --rabbit-routing_key="test-key"
```


## Usage 

```
mworker [flags]

Flags :
  -concurrency int
        Number of concurrent set of workers running (default 1)
  -rabbit-binding_wait
        Binding wait
  -rabbit-consumer_auto_ack
        Consumer Auto ACK
  -rabbit-consumer_no_local
        Consumer no local
  -rabbit-consumer_no_wait
        Consumer no wait
  -rabbit-consumer_tag string
        Consumer tag
  -rabbit-exchange string
        Exchange name. If exchange name is empty all other exchange flags are ignored
  -rabbit-exchange_delete_when_complete
        Exchange delete when complete
  -rabbit-exchange_durable
        Exchange durable (default true)
  -rabbit-exchange_internal
        Exchange internal
  -rabbit-exchange_no_wait
        Exchange no wait
  -rabbit-exchange_type string
        Exchange type - direct|fanout|topic|x-custom (default "direct")
  -rabbit-queue_delete_when_used
        Queue delete queue when used
  -rabbit-queue_durable
        Queue durable
  -rabbit-queue_exclusive
        Queue exclusive
  -rabbit-queue_name string
        Rabbit queue name (default "hello-queue")
  -rabbit-queue_no_wait
        Queue no wait
  -rabbit-routing_key string
        Routing Key
  -rabbit-uri string
        Rabbit instance uri e.g. amqp://guest:guest@localhost:5672/ (default "amqp://guest:guest@localhost:5672/")
  -wait-timeout int
        Time to wait in miliseconds until new jobs are available in rabbit  (default 500)
```