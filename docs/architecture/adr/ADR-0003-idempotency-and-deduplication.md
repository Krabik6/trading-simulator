# ADR-0003: Idempotency And Deduplication Strategy

- Date: 2026-03-10
- Status: Accepted

## Context

At-least-once delivery and client retries are expected in both live and replay paths. Without idempotency, duplicate commands/events can create duplicate orders or inconsistent balances.

## Decision

Adopt a two-layer idempotency model:
1. Command side idempotency:
   - External command endpoints require `Idempotency-Key`.
   - Key scope: tenant/user + operation + request hash.
   - Repeated command with same key returns original result.
2. Consumer deduplication:
   - All consumers track processed event IDs (or deterministic hash keys).
   - Duplicate events are acknowledged and ignored safely.

## Consequences

Positive:
- Protects against retries, network failures, and replay overlap.
- Makes at-least-once delivery operationally safe.

Costs:
- Extra storage for idempotency and dedup records.
- Additional key lifecycle and TTL management.

## Alternatives Considered

1. Exactly-once end-to-end transport guarantees only.
2. Stateless consumers with no dedup memory.

Rejected because transport guarantees alone do not cover all business-layer duplicates.
