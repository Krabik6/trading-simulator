# Stage 0: Service Boundaries And Ownership

Date: 2026-03-10
Status: Target architecture boundary contract

## Service Boundary Table

| Service | Primary Responsibility | Owns Write Data | Reads From | Writes To |
|---|---|---|---|---|
| `market-data` | Ingest and normalize live market prices | No | External exchanges | Kafka `market_ticks` |
| `api-gateway` | Edge auth, routing, rate limiting | No | Trading/query services | HTTP/gRPC to internal services |
| `trading-core` | Command handling, risk, execution, account state transitions | Yes (`users`, `accounts`, `orders`, `positions`, `trades`) | Kafka `market_ticks`, command API | DB + Kafka `trading_events` (through outbox) |
| `query-api` | Read-optimized endpoints and projections | Yes (read model only) | Kafka `trading_events` | HTTP read responses |
| `ws-gateway` | Realtime fan-out for UI and bots | No | Kafka `market_ticks`, `trading_events` | WebSocket streams |
| `analytics` | Aggregations, reporting, performance stats | Yes (analytics model) | Kafka topics | Metrics, reports |
| `replay-service` | Deterministic playback of historical ticks/events | Yes (replay metadata) | Historical store | Kafka `replay_ticks` |
| `bot-simulation-service` | Run training/backtest/paper sessions | Yes (run metadata/results) | Replay streams + query APIs | Kafka `order_commands`, results store |

## Interaction Rules

1. Write model mutations happen only in `trading-core`.
2. Read-heavy queries must use `query-api` projections.
3. Service-to-service communication defaults to asynchronous events for state propagation.
4. Synchronous calls are allowed only for command acknowledgement or strict request/response needs.
5. WebSocket fan-out does not read from `trading-core` memory directly in target architecture.

## Data Ownership Rules

1. No direct cross-service writes into another service schema.
2. Data sharing uses event contracts with schema versioning.
3. Every consumer must tolerate at-least-once delivery via idempotent handling.
4. Replay inputs are immutable once finalized.

## Security Boundary Rules

1. External traffic enters only through edge/gateway surfaces.
2. Internal service authentication is mandatory (service identity).
3. Tokens in query string are disallowed in target production profile.
