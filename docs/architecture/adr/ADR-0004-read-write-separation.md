# ADR-0004: Read/Write Separation (CQRS-Inspired)

- Date: 2026-03-10
- Status: Accepted

## Context

The current service mixes write-critical logic with read-heavy endpoints and realtime fan-out. This coupling increases latency variance and makes horizontal scaling expensive.

## Decision

Separate write and read concerns:
1. `trading-core` handles all state mutations and risk logic.
2. `query-api` serves read endpoints from projection models.
3. `ws-gateway` consumes events for realtime delivery.
4. Read models are rebuilt from event streams when needed.

## Consequences

Positive:
- Protects command path from read-query pressure.
- Enables independent scaling for read and write workloads.
- Improves support for analytics and replay consumers.

Costs:
- Eventual consistency for some read endpoints.
- Requires projection rebuild tooling and monitoring.

## Alternatives Considered

1. Keep single DB model for both read and write paths.
2. Add only indexes and cache layers to current monolith.

Rejected because these do not isolate write path reliability at target scale.
