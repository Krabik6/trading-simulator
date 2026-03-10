# Stage 2 Implementation: Observability And Operations

Date: 2026-03-10
Status: Implemented

## Objectives Covered

1. Request-level latency and status instrumentation for backend services.
2. Runtime metrics for Kafka lag/queue, DB pool pressure, and WebSocket delivery health.
3. Prometheus alert rules for API SLI breaches, ingest failures, and platform scrape/config issues.
4. Grafana dashboard coverage for trading command path and data path.
5. Operational runbooks mapped to alert annotations.

## Delivered Runtime Instrumentation

- `trading` service:
- HTTP request counters and latency histogram wired through chi middleware.
- Kafka consumer lag and queue depth collection from reader stats.
- DB connection pool usage/wait metrics collection.
- WebSocket active connection gauge and dropped-message counters.

- `market-data` service:
- HTTP request counters and latency histogram on health/readiness endpoints.
- Kafka producer in-flight operation gauge.

## Delivered Observability Assets

- Prometheus config updated to load rule files from `/etc/prometheus/rules`.
- Alert groups added:
- `trading-alerts.yml`
- `market-data-alerts.yml`
- `platform-alerts.yml`

- Compose profiles (`dev`, `stage`, `prod`, default) updated to mount rules directory.
- Grafana dashboard added: `trading-overview.json`.

## Delivered Operational Docs

- `docs/operations/runbooks/trading-api-slo-breach.md`
- `docs/operations/runbooks/market-data-ingest-stall.md`
- `docs/operations/runbooks/kafka-consumer-lag.md`
- `docs/operations/runbooks/platform-target-down.md`

## Notes

- This stage focuses on Prometheus/Grafana based observability. OTel traces/log correlation remains a follow-up enhancement.
- Alert thresholds are baseline defaults and should be tuned with production traffic data.
