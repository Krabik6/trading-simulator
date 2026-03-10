# ADR-0002: Transactional Outbox For Reliable Event Publication

- Date: 2026-03-10
- Status: Accepted

## Context

Direct publish to Kafka from business code creates inconsistency risk when DB write succeeds but event publish fails (or the reverse). Trading flows require atomic state transition plus guaranteed event emission.

## Decision

Use transactional outbox in all write paths that emit domain events:
1. Domain mutation and outbox insert happen in one DB transaction.
2. Outbox relay publishes pending events to Kafka asynchronously.
3. Relay marks events as published with retry/backoff and dead-letter policy.

## Consequences

Positive:
- Eliminates dual-write inconsistency.
- Supports retry without re-executing business mutation.
- Improves auditability of pending/failed publications.

Costs:
- Additional table/storage and relay operational component.
- Requires monitoring (`outbox_pending`, lag, retry counters).

## Alternatives Considered

1. Direct in-transaction Kafka publish.
2. Best-effort publish with compensating actions.

Rejected because correctness requirements are stricter than best-effort semantics.
