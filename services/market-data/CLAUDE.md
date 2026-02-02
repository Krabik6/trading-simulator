# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### From repo root (`trading-simulator/`)
```bash
make up                 # start all services
make down               # stop all
make ps                 # container status
make logs               # all logs
make clean              # stop + remove volumes
make kafka-topics       # list Kafka topics
make kafka-consume      # read 10 messages
make kafka-consume-live # stream messages (Ctrl+C to exit)
make urls               # show service URLs
```

### From service dir (`services/market-data/`)
```bash
make run        # run locally
make build      # build binary to bin/
make tidy       # go mod tidy
make logs       # service logs
make restart    # restart container
make health     # check /health
make metrics    # show Prometheus metrics
```

## Architecture

Go microservice streaming crypto prices from mock client to Kafka.

```
cmd/main.go → config.Load() → logger.Init() → app.New()
                                                ├─ client.NewClient()
                                                ├─ kafka.NewKafkaProducer()
                                                └─ app.Run()
                                                     ├─ HTTP server (/health, /ready, /metrics)
                                                     ├─ client.StreamPrices() → channel
                                                     └─ processPrices() → kafka.Send()
```

## Package Structure

- `cmd/` - entry point
- `config/` - env-based configuration
- `internal/app/` - orchestration, HTTP handlers
- `internal/client/` - PriceClient interface + factory
- `internal/client/mock/` - mock price generator
- `internal/domain/` - Price entity
- `internal/kafka/` - Kafka producer
- `internal/logger/` - structured logging (slog)
- `internal/metrics/` - Prometheus metrics

## Configuration

Env vars loaded from `.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| CLIENT_TYPE | mock | client type |
| SYMBOLS | BTCUSDT,ETHUSDT,SOLUSDT | trading pairs |
| KAFKA_BROKERS | localhost:9092 | Kafka brokers |
| KAFKA_TOPIC | crypto-prices | Kafka topic |
| LOG_LEVEL | info | log level |
| HTTP_PORT | 8080 | HTTP port |

## Endpoints

- `GET /health` - liveness
- `GET /ready` - readiness (checks Kafka)
- `GET /metrics` - Prometheus metrics

## Observability

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
