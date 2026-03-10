# Stage 3 Implementation: Core Correctness

Date: 2026-03-10
Status: Implemented

## Objectives Covered

1. Transaction boundaries for critical write flows.
2. Concurrency control with row-level locking in command path.
3. Command idempotency for order/position mutation endpoints.
4. Database-level invariant constraints for financial entities.
5. Runtime behavior hardening for invalid partial-close quantities.

## Delivered Changes

### Transaction Boundaries

- Added transaction manager contract (`domain.TxManager`) and implementation in PostgreSQL DB wrapper (`WithinTx`).
- Wrapped critical write flows in transactions:
- auth registration (`user` + `account` creation)
- order placement / cancel / update
- position close / liquidation / TP-SL triggers / TP-SL update

### Concurrency Control

- Added `FOR UPDATE` repository methods:
- `accounts`: `GetByUserIDForUpdate`
- `orders`: `GetByIDForUpdate`
- `positions`: `GetByIDForUpdate`, `GetOpenByUserIDAndSymbolForUpdate`
- Standardized lock ordering for money-sensitive command flows (lock account first, then position/order).
- Replaced balance rollback hack with atomic guarded update (`balance + delta >= 0`).

### Idempotency

- Added idempotency storage model in PostgreSQL (`idempotency_keys`) and store implementation.
- Added HTTP idempotency middleware for command routes (`/orders*`, `/positions*`, methods `POST/PATCH/DELETE`):
- requires `Idempotency-Key`
- hashes request payload
- replays stored response for duplicate requests
- rejects same key with different payload (`409`)
- rejects in-flight duplicate (`409`)

### DB Invariants

- Added migration `003_core_correctness` with check constraints:
- non-negative account balance
- positive quantity/price checks on orders, positions, trades
- close-percent range checks (1..100)
- Added idempotency table, indexes, and update trigger.

### Integration Test Coverage

- Added idempotency integration tests:
- replay with same key returns same response and no duplicate side effects
- same key + different payload returns conflict
- missing `Idempotency-Key` returns bad request
- Added position close test for invalid quantity greater than current position size.
- Updated integration request helper to auto-attach idempotency key for order/position mutations.

## Validation

Passed:
- `./scripts/ci/check-go-format.sh`
- `./scripts/ci/check-go-tests.sh`
- `./scripts/ci/check-migration-pairs.sh`
- `./scripts/ci/check-compose-config.sh`

Note:
- `go test ./internal/integration_test` build remains blocked by existing toolchain/dependency mismatch in testcontainers/docker modules in current environment.
