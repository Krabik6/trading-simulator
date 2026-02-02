# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Trading Simulator - монорепозиторий с микросервисами для симуляции криптотрейдинга.

## Repository Structure

```
trading-simulator/
├── services/
│   ├── market-data/   # Go: стриминг цен криптовалют в Kafka
│   ├── trading/       # сервис исполнения ордеров
│   ├── analytics/     # аналитика и агрегация
│   └── api/           # API gateway
├── infrastructure/    # Prometheus, Grafana configs
├── deploy/docker/     # docker-compose.yml
├── frontend/          # UI
└── configs/           # shared configs
```

## Commands (from repo root)

```bash
make up                 # start all services
make down               # stop all
make ps                 # container status
make logs               # all logs
make clean              # stop + remove volumes
make kafka-topics       # list Kafka topics
make kafka-consume      # read 10 messages from crypto-prices
make kafka-consume-live # stream messages (Ctrl+C to exit)
make urls               # show service URLs
```

## Infrastructure

| Service | URL | Notes |
|---------|-----|-------|
| Kafka | localhost:9092 | message broker |
| Prometheus | http://localhost:9090 | metrics |
| Grafana | http://localhost:3000 | dashboards (admin/admin) |

## Tech Stack

- **Backend**: Go 1.22+
- **Messaging**: Apache Kafka
- **Observability**: Prometheus + Grafana
- **Containers**: Docker Compose

## Go Service Conventions

Каждый Go-сервис следует структуре:
```
service-name/
├── cmd/main.go           # entry point
├── config/               # env-based config
├── internal/
│   ├── app/              # orchestration, HTTP handlers
│   ├── domain/           # entities
│   └── ...               # feature packages
├── Dockerfile
├── Makefile              # service-specific commands
└── CLAUDE.md             # service-specific docs
```

Стандартные команды в каждом сервисе:
```bash
make run      # run locally
make build    # build binary
make tidy     # go mod tidy
make logs     # service logs (docker)
make restart  # restart container
make health   # check /health endpoint
```

## Kafka Topics

| Topic | Producer | Description |
|-------|----------|-------------|
| crypto-prices | market-data | real-time price updates |
