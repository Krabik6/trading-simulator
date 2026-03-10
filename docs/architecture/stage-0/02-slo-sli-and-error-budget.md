# Stage 0: SLO, SLI, And Error Budget Policy

Date: 2026-03-10
Status: Baseline for implementation stages

## SLO Matrix

| Service Area | SLO | Measurement Window |
|---|---|---|
| Trading command API availability | >= 99.9% successful responses (non-5xx) | 30 days |
| Trading command latency (`place/close/update`) | p95 <= 150 ms, p99 <= 400 ms | rolling 1 hour |
| Read API latency | p95 <= 120 ms, p99 <= 300 ms | rolling 1 hour |
| Market ingest freshness | p99 tick age <= 2 seconds | rolling 15 minutes |
| Event publication delay (`DB commit -> Kafka`) | p99 <= 1 second | rolling 15 minutes |
| WS delivery latency (`event accepted -> sent`) | p99 <= 250 ms | rolling 15 minutes |
| Replay determinism | 100% deterministic result for same input/version | per run |

## Required SLIs

1. `http_requests_total`, `http_request_duration_seconds` by route, status, method.
2. `kafka_consumer_lag` per topic/partition.
3. `outbox_pending_events` and `outbox_oldest_event_age_seconds`.
4. `db_pool_in_use`, `db_pool_wait_count`, `db_pool_wait_duration`.
5. `ws_connections_active`, `ws_messages_dropped_total`.
6. `simulation_runs_total`, `simulation_runs_failed_total`, `replay_events_processed_total`.

## Error Budget Policy

For 99.9% monthly availability:
- Total error budget: 43 minutes 12 seconds per 30-day month.

Policy:
1. Burn rate > 2x over 1 hour: freeze non-critical releases.
2. Burn rate > 5x over 15 minutes: incident mode, assign incident commander.
3. Budget consumed > 50% before day 15: reliability sprint takes priority.
4. Budget consumed > 80% before month end: allow only risk-reduction changes.

## Release Gates

A stage cannot be promoted to production unless:
1. All mandatory SLI signals are emitted.
2. Alerting exists for burn-rate and hard-SLO breaches.
3. Dashboards cover command path, event path, and data path.
