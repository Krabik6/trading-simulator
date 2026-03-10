# ADR-0001: Event-Driven Core For Trading State Propagation

- Date: 2026-03-10
- Status: Accepted

## Context

The platform must support two parallel workloads: live trading simulation and bot training/testing. The current architecture couples request handling, in-memory distribution, and downstream consumers too tightly for safe scale.

## Decision

Adopt event-driven propagation as the default integration model:
1. Trading state transitions are committed to write DB.
2. Corresponding immutable domain events are published to Kafka.
3. Downstream read services (`query-api`, `ws-gateway`, `analytics`, `bot-simulation-service`) consume events instead of direct write-model reads for propagation.

## Consequences

Positive:
- Decouples producers from read/fan-out workloads.
- Enables replay and deterministic simulation pipelines.
- Simplifies horizontal scaling boundaries.

Costs:
- Requires schema governance and compatibility checks.
- Increases operational complexity (Kafka health, lag handling).

## Alternatives Considered

1. Expand synchronous RPC mesh between services.
2. Keep single-service architecture with shared DB reads.

Both alternatives were rejected due to tighter coupling and poor replay support.
